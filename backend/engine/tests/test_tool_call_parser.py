"""Targeted tests for tool call parsing, unclosed repair, and final response sanitization."""

import pytest
from engine.agent.react_agent import (
    DEFAULT_MAX_TOKENS,
    DEFAULT_LLM_TIMEOUT,
    parse_single_tool_call,
    extract_tool_calls,
    sanitize_final_response,
)


def test_token_and_timeout_constants():
    """Verify default max tokens is 30,000 and default timeout is 90s."""
    assert DEFAULT_MAX_TOKENS == 30000
    assert DEFAULT_LLM_TIMEOUT == 90.0


def test_parse_closed_json_tool_call():
    """Verify standard closed JSON tool call parsing."""
    raw = '{"name": "harvest_market_news", "arguments": {"ticker": "IHSG"}}'
    parsed = parse_single_tool_call(raw)
    assert parsed is not None
    name, args = parsed
    assert name == "harvest_market_news"
    assert args == {"ticker": "IHSG"}


def test_parse_tool_alias_json():
    """Verify parsing when model uses 'tool' and 'args' keys."""
    raw = '{"tool": "get_daily_candles", "args": {"ticker": "ANTM", "days": 10}}'
    parsed = parse_single_tool_call(raw)
    assert parsed is not None
    name, args = parsed
    assert name == "get_daily_candles"
    assert args.get("ticker") == "ANTM"
    assert args.get("days") == 10


def test_parse_unclosed_json_tool_call():
    """Verify unclosed JSON tool call repair."""
    raw = '{"name": "harvest_market_news", "arguments": {"ticker": "IHSG"'
    parsed = parse_single_tool_call(raw)
    assert parsed is not None
    name, args = parsed
    assert name == "harvest_market_news"
    assert args.get("ticker") == "IHSG"


def test_parse_xml_format_tool_call():
    """Verify XML-style tool call parsing."""
    raw = "<name>harvest_market_news</name><ticker>IHSG</ticker>"
    parsed = parse_single_tool_call(raw)
    assert parsed is not None
    name, args = parsed
    assert name == "harvest_market_news"
    assert args == {"ticker": "IHSG"}

    raw_multiline = "compute_quant_anomalies\n<ticker>BBCA</ticker>\n<days>30</days>"
    parsed2 = parse_single_tool_call(raw_multiline)
    assert parsed2 is not None
    name2, args2 = parsed2
    assert name2 == "compute_quant_anomalies"
    assert args2.get("ticker") == "BBCA"


def test_extract_tool_calls_unclosed_tag():
    """Verify extract_tool_calls recovers when model leaves <tool_call> unclosed."""
    content = (
        "<thought>Saya perlu mengecek berita pasar.</thought>\n"
        '<tool_call>{"name": "harvest_market_news", "arguments": {"ticker": "IHSG"}}'
    )
    calls = extract_tool_calls(content)
    assert len(calls) == 1
    assert calls[0][0] == "harvest_market_news"
    assert calls[0][1] == {"ticker": "IHSG"}


def test_sanitize_final_response_explicit_tags():
    """Verify sanitize_final_response extracts content from <response> tags."""
    content = (
        "<thought>Analisis selesai.</thought>\n"
        "<response>\n## Ringkasan Pasar\nIHSG menguat 0.5% didorong sektor perbankan.\n</response>"
    )
    res = sanitize_final_response(content)
    assert "## Ringkasan Pasar" in res
    assert "<thought>" not in res
    assert "<response>" not in res


def test_sanitize_final_response_removes_dangling_tool_calls():
    """Verify sanitize_final_response prevents unclosed <tool_call> leakage."""
    content = (
        "Berikut adalah analisis awal.\n"
        '<tool_call>{"name": "harvest_market_news", "arguments": {"ticker": "IHSG"}'
    )
    res = sanitize_final_response(content)
    assert "<tool_call>" not in res
    assert "harvest_market_news" not in res
    assert "Berikut adalah analisis awal." in res


def test_parse_xml_key_value_pairs():
    """Verify parsing when model emits <arg_key>ticker</arg_key><arg_value>IHSG</arg_value>."""
    raw = "<name>query_sectors</name><arg_key>ticker</arg_key><arg_value>IHSG</arg_value>"
    parsed = parse_single_tool_call(raw)
    assert parsed is not None
    name, args = parsed
    assert name == "query_sectors"
    assert args == {"ticker": "IHSG"}


def test_parse_xml_attribute_format():
    """Verify parsing when model emits <arg name="ticker">IHSG</arg>."""
    raw = '<name>query_sectors</name><arg name="ticker">IHSG</arg><arg name="domain">candles</arg>'
    parsed = parse_single_tool_call(raw)
    assert parsed is not None
    name, args = parsed
    assert name == "query_sectors"
    assert args.get("ticker") == "IHSG"
    assert args.get("domain") == "candles"

