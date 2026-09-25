"""Unit tests for adaptive inference timeout calculation in ReAct agent."""

import pytest
from engine.agent.react_agent import DEFAULT_LLM_TIMEOUT, _adaptive_call_timeout


def test_zero_observations_returns_base():
    """Verify 0 observations returns the base timeout exactly."""
    base = 20.0
    result = _adaptive_call_timeout(base, obs_count=0)
    assert result == 20.0


def test_one_observation_adds_10s():
    """Verify 1 observation adds 10 seconds to base timeout."""
    base = 20.0
    result = _adaptive_call_timeout(base, obs_count=1)
    assert result == 30.0


def test_five_observations_adds_50s():
    """Verify 5 observations adds 50 seconds to base timeout."""
    base = 20.0
    result = _adaptive_call_timeout(base, obs_count=5)
    assert result == 70.0


def test_soft_cap_is_base_times_four():
    """Verify timeout is capped at 4x base timeout when below hard cap."""
    base = 15.0
    # raw = 15 + 10 * 10 = 115, soft cap = 15 * 4 = 60.0
    result = _adaptive_call_timeout(base, obs_count=10)
    assert result == 60.0

    base_20 = 20.0
    # raw = 20 + 8 * 10 = 100, soft cap = 20 * 4 = 80.0
    result_20 = _adaptive_call_timeout(base_20, obs_count=8)
    assert result_20 == 80.0


def test_hard_cap_is_300s():
    """Verify timeout never exceeds absolute hard cap of 300.0 seconds."""
    base = 100.0
    # raw = 100 + 35 * 10 = 450, soft cap = 400, hard cap = 300
    result = _adaptive_call_timeout(base, obs_count=35)
    assert result == 300.0

    # Large base: soft cap = 800, raw = 250, hard cap = 300
    base_large = 250.0
    result_large = _adaptive_call_timeout(base_large, obs_count=10)
    assert result_large == 300.0


def test_negative_obs_count_treated_as_zero():
    """Verify negative observation count is safely clamped to 0."""
    base = 20.0
    result = _adaptive_call_timeout(base, obs_count=-5)
    assert result == 20.0


def test_returns_float():
    """Verify return value is always of type float."""
    result = _adaptive_call_timeout(20, 2)
    assert isinstance(result, float)
    assert result == 40.0


def test_default_constant_is_60():
    """Verify default LLM timeout constant is configured to 60.0 seconds."""
    assert DEFAULT_LLM_TIMEOUT == 60.0
