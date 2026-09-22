"""Targeted tests for NiskavaToolRegistry argument type defense."""

import pytest
from engine.agent.tools import NiskavaToolRegistry


def test_execute_tool_with_plain_string_ticker(tmp_path):
    """Verify execute_tool() handles string argument 'BBCA' gracefully without AttributeError."""
    registry = NiskavaToolRegistry(db_path=str(tmp_path / "test.db"), mock_mode=True)
    res = registry.execute_tool("get_daily_candles", "BBCA")
    assert isinstance(res, list)


def test_execute_tool_with_json_string(tmp_path):
    """Verify execute_tool() parses JSON string arguments."""
    registry = NiskavaToolRegistry(db_path=str(tmp_path / "test.db"), mock_mode=True)
    res = registry.execute_tool("get_daily_candles", '{"ticker": "ANTM", "days": 15}')
    assert isinstance(res, list)


def test_execute_tool_with_none_arguments(tmp_path):
    """Verify execute_tool() handles None arguments safely."""
    registry = NiskavaToolRegistry(db_path=str(tmp_path / "test.db"), mock_mode=True)
    res = registry.execute_tool("harvest_market_news", None)
    assert isinstance(res, list)


def test_execute_tool_with_list_or_int_arguments(tmp_path):
    """Verify execute_tool() handles non-dict, non-string types safely."""
    registry = NiskavaToolRegistry(db_path=str(tmp_path / "test.db"), mock_mode=True)
    res = registry.execute_tool("harvest_market_news", 12345)
    assert isinstance(res, list)
