"""Unit tests for network resilience and retry backoff module."""

import email.utils
from datetime import datetime, timedelta, timezone
from unittest.mock import MagicMock, call

import pytest
import requests

from engine.utils.resilience import (
    RetryConfig,
    calculate_backoff,
    execute_with_retry,
    interruptible_sleep,
    parse_retry_after,
)


def _create_mock_response(status_code: int, text: str = "", headers: dict = None) -> requests.Response:
    resp = requests.Response()
    resp.status_code = status_code
    resp._content = text.encode("utf-8")
    resp.headers.update(headers or {})
    return resp


def test_parse_retry_after_integer_seconds():
    headers = {"Retry-After": "3"}
    assert parse_retry_after(headers, default_delay=1.0) == 3.0

    headers_lower = {"retry-after": "5"}
    assert parse_retry_after(headers_lower, default_delay=1.0) == 5.0

    # Test max delay clamping
    headers_large = {"Retry-After": "120"}
    assert parse_retry_after(headers_large, default_delay=1.0, max_delay=10.0) == 10.0


def test_parse_retry_after_http_date():
    future_time = datetime.now(timezone.utc) + timedelta(seconds=4)
    date_str = email.utils.format_datetime(future_time)
    headers = {"Retry-After": date_str}
    parsed = parse_retry_after(headers, default_delay=1.0)
    assert parsed is not None
    assert 2.0 <= parsed <= 6.0


def test_parse_retry_after_invalid_or_missing():
    assert parse_retry_after(None, default_delay=1.5) == 1.5
    assert parse_retry_after({}, default_delay=1.5) == 1.5
    assert parse_retry_after({"Retry-After": "invalid-date"}, default_delay=1.5) == 1.5


def test_calculate_backoff():
    cfg = RetryConfig(initial_delay=1.0, backoff_factor=2.0, max_delay=10.0, jitter=False)
    # Attempt 1: 1.0 * (2^0) = 1.0
    assert calculate_backoff(1, cfg) == 1.0
    # Attempt 2: 1.0 * (2^1) = 2.0
    assert calculate_backoff(2, cfg) == 2.0
    # Attempt 3: 1.0 * (2^2) = 4.0
    assert calculate_backoff(3, cfg) == 4.0

    # With jitter
    cfg_jitter = RetryConfig(initial_delay=1.0, backoff_factor=2.0, max_delay=10.0, jitter=True)
    d1 = calculate_backoff(1, cfg_jitter)
    assert 1.0 <= d1 <= 1.6

    # With explicit retry_after
    assert calculate_backoff(1, cfg, retry_after=7.5) == 7.5


def test_retry_on_429_success_after_backoff():
    responses = [
        _create_mock_response(429, "Rate limited", {"Retry-After": "0.1"}),
        _create_mock_response(429, "Rate limited", {"Retry-After": "0.1"}),
        _create_mock_response(200, '{"result": "success"}'),
    ]
    mock_fn = MagicMock(side_effect=responses)
    mock_callback = MagicMock()

    cfg = RetryConfig(max_retries=3, initial_delay=0.05, jitter=False)
    resp = execute_with_retry(mock_fn, config=cfg, on_retry_callback=mock_callback)

    assert resp.status_code == 200
    assert mock_fn.call_count == 3
    assert mock_callback.call_count == 2
    # Verify callback args: attempt 1 and 2, status 429
    first_call_args = mock_callback.call_args_list[0][0]
    assert first_call_args[0] == 1  # attempt
    assert first_call_args[2] == 429  # status


def test_non_retryable_status_fails_fast():
    responses = [_create_mock_response(401, "Unauthorized")]
    mock_fn = MagicMock(side_effect=responses)
    mock_callback = MagicMock()

    cfg = RetryConfig(max_retries=3, initial_delay=0.05)
    resp = execute_with_retry(mock_fn, config=cfg, on_retry_callback=mock_callback)

    assert resp.status_code == 401
    assert mock_fn.call_count == 1
    assert mock_callback.call_count == 0


def test_max_retries_exhaustion():
    responses = [
        _create_mock_response(429, "Rate limit 1"),
        _create_mock_response(429, "Rate limit 2"),
        _create_mock_response(429, "Rate limit 3"),
        _create_mock_response(429, "Rate limit 4"),
    ]
    mock_fn = MagicMock(side_effect=responses)
    mock_callback = MagicMock()

    cfg = RetryConfig(max_retries=3, initial_delay=0.01, jitter=False)
    resp = execute_with_retry(mock_fn, config=cfg, on_retry_callback=mock_callback)

    assert resp.status_code == 429
    assert mock_fn.call_count == 4  # 1 initial + 3 retries
    assert mock_callback.call_count == 3


def test_network_timeout_retry_and_exhaustion():
    mock_fn = MagicMock(side_effect=requests.exceptions.Timeout("Connection timed out"))
    mock_callback = MagicMock()

    cfg = RetryConfig(max_retries=2, initial_delay=0.01, jitter=False)
    with pytest.raises(requests.exceptions.Timeout):
        execute_with_retry(mock_fn, config=cfg, on_retry_callback=mock_callback)

    assert mock_fn.call_count == 3  # 1 initial + 2 retries
    assert mock_callback.call_count == 2


def test_interruptible_sleep():
    # Verify interruptible sleep runs and completes for small interval
    interruptible_sleep(0.05, check_interval=0.02)


def test_react_agent_emits_rate_limit_thought(monkeypatch):
    from unittest.mock import patch
    from engine.agent.react_agent import NiskavaReActAgent
    from engine.agent.tools import NiskavaToolRegistry

    db_path = "/tmp/test_resilience_agent.db"
    events = []
    registry = NiskavaToolRegistry(db_path, mock_mode=True)
    agent = NiskavaReActAgent(
        tool_registry=registry,
        emitter=lambda ev: events.append(ev),
        mock_mode=False,
        ai_provider="openai",
        api_key="sk-dummy",
        base_url="http://mock-llm.local",
    )

    success_content = "<thought>Memproses sapaan pengguna</thought><response>Halo! Saya Niskava Agent, asisten riset pasar Anda.</response>"
    success_json = '{"choices": [{"message": {"content": "' + success_content + '"}}]}'

    responses = [
        _create_mock_response(429, "Rate limited", {"Retry-After": "0.05"}),
        _create_mock_response(200, success_json),
    ]

    with patch("requests.post", side_effect=responses):
        res = agent.chat("Halo Niskava!")

    # Verify response
    assert "Halo! Saya Niskava Agent" in res.get("response", "")

    # Verify agent emitted rate limit thought
    thought_events = [e for e in events if e.get("event") == "agent_thought"]
    rate_limit_thoughts = [
        e["thought"] for e in thought_events if "HTTP 429" in e.get("thought", "")
    ]
    assert len(rate_limit_thoughts) == 1
    assert "Attempt 1/3" in rate_limit_thoughts[0]


def test_sectors_client_retry_on_429(monkeypatch):
    import os
    from unittest.mock import patch
    from engine.sectors.client import SectorsAPIClient

    db_path = "/tmp/test_resilience_sectors.db"
    if os.path.exists(db_path):
        os.remove(db_path)

    client = SectorsAPIClient(db_path, api_key="sec_test_key", mock_mode=False)

    sample_candles = [{"date": "2026-09-01", "close": 1800, "volume": 5000000}]
    responses = [
        _create_mock_response(429, "Too Many Requests", {"Retry-After": "0.05"}),
        _create_mock_response(200, str(sample_candles).replace("'", '"')),
    ]

    with patch.object(client.session, "get", side_effect=responses) as mock_get:
        candles = client.get_daily_candles("BBCA")

    assert mock_get.call_count == 2
    assert len(candles) == 1
    assert candles[0]["close"] == 1800

    if os.path.exists(db_path):
        os.remove(db_path)
