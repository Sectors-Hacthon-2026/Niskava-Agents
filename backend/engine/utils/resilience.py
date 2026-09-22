"""Network resilience, rate limiting backoff, and retry handling."""

import email.utils
import random
import time
from dataclasses import dataclass, field
from datetime import datetime, timezone
from typing import Any, Callable, Dict, Mapping, Optional, Set

import requests


@dataclass
class RetryConfig:
    """Configuration for HTTP retry and backoff behavior."""

    max_retries: int = 3
    initial_delay: float = 1.0
    max_delay: float = 10.0
    backoff_factor: float = 2.0
    jitter: bool = True
    retryable_statuses: Set[int] = field(
        default_factory=lambda: {429, 500, 502, 503, 504}
    )


def parse_retry_after(
    headers: Optional[Mapping[str, str]],
    default_delay: Optional[float] = None,
    max_delay: float = 10.0,
) -> Optional[float]:
    """Extract and parse Retry-After header (seconds or HTTP-date).

    Returns clamped delay in seconds or default_delay if absent/unparseable.
    """
    if not headers:
        return default_delay

    # Case-insensitive lookup
    retry_header_val = None
    for k, v in headers.items():
        if k.lower() == "retry-after":
            retry_header_val = v.strip()
            break

    if not retry_header_val:
        return default_delay

    # 1. Try parsing integer/float seconds
    try:
        seconds = float(retry_header_val)
        if seconds > 0:
            return min(max(0.1, seconds), max_delay)
    except (ValueError, TypeError):
        pass

    # 2. Try parsing RFC 1123 / HTTP-date
    try:
        dt = email.utils.parsedate_to_datetime(retry_header_val)
        if dt:
            now = datetime.now(timezone.utc)
            diff = (dt - now).total_seconds()
            if diff > 0:
                return min(max(0.1, diff), max_delay)
    except Exception:
        pass

    return default_delay


def calculate_backoff(
    attempt: int,
    config: RetryConfig,
    retry_after: Optional[float] = None,
) -> float:
    """Calculate exponential backoff delay with jitter."""
    if retry_after is not None and retry_after > 0:
        return min(max(0.1, retry_after), config.max_delay)

    # Base exponential backoff: initial_delay * factor^(attempt - 1)
    exp_delay = config.initial_delay * (config.backoff_factor ** max(0, attempt - 1))

    # Add full jitter to prevent thundering herd
    if config.jitter:
        jitter_amount = random.uniform(0.0, 0.5)
        exp_delay += jitter_amount

    return min(max(0.1, exp_delay), config.max_delay)


def interruptible_sleep(seconds: float, check_interval: float = 0.1) -> None:
    """Sleep in small intervals to allow prompt responsive interruption (SIGINT/abort)."""
    remaining = max(0.0, seconds)
    while remaining > 0:
        step = min(remaining, check_interval)
        time.sleep(step)
        remaining -= step


def execute_with_retry(
    request_fn: Callable[[], requests.Response],
    config: Optional[RetryConfig] = None,
    on_retry_callback: Optional[Callable[[int, float, int, str], None]] = None,
) -> requests.Response:
    """Execute an HTTP request with automatic retry and exponential backoff for transient errors.

    Args:
        request_fn: Zero-argument callable that performs the requests call.
        config: RetryConfig options (defaults to standard config).
        on_retry_callback: Optional callback invoked before each sleep:
            callback(attempt, delay_seconds, status_code_or_zero, error_summary)

    Returns:
        The successful Response or the final Response if retries are exhausted.

    Raises:
        requests.exceptions.RequestException if network failure persists past max_retries.
    """
    cfg = config or RetryConfig()
    total_attempts = cfg.max_retries + 1

    for attempt in range(1, total_attempts + 1):
        try:
            resp = request_fn()
            if resp.status_code in cfg.retryable_statuses:
                if attempt <= cfg.max_retries:
                    retry_after = parse_retry_after(
                        resp.headers, default_delay=None, max_delay=cfg.max_delay
                    )
                    delay = calculate_backoff(attempt, cfg, retry_after)
                    if on_retry_callback:
                        summary = resp.text[:150] if resp.text else resp.reason or ""
                        on_retry_callback(attempt, delay, resp.status_code, summary)
                    interruptible_sleep(delay)
                    continue
                # Retries exhausted with retryable status code: return final response
                return resp

            # Non-retryable status (e.g. 200, 201, 400, 401, 403, 404)
            return resp

        except (requests.exceptions.Timeout, requests.exceptions.ConnectionError) as exc:
            if attempt <= cfg.max_retries:
                delay = calculate_backoff(attempt, cfg)
                if on_retry_callback:
                    on_retry_callback(attempt, delay, 0, str(exc)[:150])
                interruptible_sleep(delay)
                continue
            # Retries exhausted with network exception
            raise exc
