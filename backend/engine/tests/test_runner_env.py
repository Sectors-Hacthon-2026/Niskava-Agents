"""Tests for load_dotenv_fallback ensuring empty keys do not shadow valid keys."""

import os
from unittest.mock import patch
from engine.runner import load_dotenv_fallback, load_config_yaml_fallback, resolve_mock_mode

_real_open = open


def test_load_dotenv_fallback_skips_empty_values(tmp_path):
    local_env = tmp_path / ".env"
    global_env = tmp_path / "global.env"

    local_env.write_text("SECTORS_API_KEY=\nOPENAI_MODEL=hermes\n", encoding="utf-8")
    global_env.write_text("SECTORS_API_KEY=valid_key_12345\n", encoding="utf-8")

    def mock_exists(p):
        return p == ".env" or p == str(global_env)

    def mock_open(p, *args, **kwargs):
        if p == ".env":
            return _real_open(local_env, *args, **kwargs)
        elif p == str(global_env):
            return _real_open(global_env, *args, **kwargs)
        return _real_open(p, *args, **kwargs)

    with patch.dict(os.environ, {}, clear=True):
        with patch("engine.runner.os.path.expanduser", return_value=str(global_env)):
            with patch("engine.runner.os.path.exists", side_effect=mock_exists):
                with patch("builtins.open", side_effect=mock_open):
                    load_dotenv_fallback()
                    assert os.environ.get("SECTORS_API_KEY") == "valid_key_12345"
                    assert os.environ.get("OPENAI_MODEL") == "hermes"


def test_load_config_yaml_fallback(tmp_path):
    yaml_config = tmp_path / "config.yaml"
    yaml_config.write_text("""
auth:
  sectors_api_key: "sectors_yaml_key_999"
  gemini_api_key: "gemini_yaml_key_888"
  ai_provider: "gemini"
preferences:
  language: "en"
  offline_mode: true
""", encoding="utf-8")

    with patch.dict(os.environ, {}, clear=True):
        with patch("engine.runner.os.path.expanduser", return_value=str(yaml_config)):
            load_config_yaml_fallback()
            assert os.environ.get("SECTORS_API_KEY") == "sectors_yaml_key_999"
            assert os.environ.get("GEMINI_API_KEY") == "gemini_yaml_key_888"
            assert os.environ.get("AI_PROVIDER") == "gemini"
            assert os.environ.get("NISKAVA_LANG") == "en"
            assert os.environ.get("NISKAVA_OFFLINE") == "1"


def test_load_config_yaml_fallback_without_pyyaml(tmp_path):
    yaml_config = tmp_path / "config.yaml"
    yaml_config.write_text("""
auth:
  sectors_api_key: "sectors_yaml_key_777"
  gemini_api_key: "gemini_yaml_key_666"
  ai_provider: "gemini"
preferences:
  language: "id"
  offline_mode: false
""", encoding="utf-8")

    with patch.dict(os.environ, {}, clear=True):
        with patch("engine.runner.os.path.expanduser", return_value=str(yaml_config)):
            with patch.dict("sys.modules", {"yaml": None}):
                load_config_yaml_fallback()
                assert os.environ.get("SECTORS_API_KEY") == "sectors_yaml_key_777"
                assert os.environ.get("GEMINI_API_KEY") == "gemini_yaml_key_666"
                assert os.environ.get("AI_PROVIDER") == "gemini"
                assert os.environ.get("NISKAVA_LANG") == "id"
                assert os.environ.get("NISKAVA_OFFLINE") == "0"


def test_resolve_mock_mode_auto_toggles_live_with_valid_key():
    """Valid SECTORS_API_KEY must auto-toggle mock_mode to False even if MOCK_SECTORS=1 was set."""
    env = {
        "SECTORS_API_KEY": "valid_live_token_12345",
        "MOCK_SECTORS": "1",
        "NISKAVA_OFFLINE": "0",
    }
    assert resolve_mock_mode(args_offline=False, env=env) is False


def test_resolve_mock_mode_explicit_offline_env_overrides_key():
    """Explicit NISKAVA_OFFLINE=1 forces mock_mode to True even if SECTORS_API_KEY is present."""
    env = {
        "SECTORS_API_KEY": "valid_live_token_12345",
        "MOCK_SECTORS": "0",
        "NISKAVA_OFFLINE": "1",
    }
    assert resolve_mock_mode(args_offline=False, env=env) is True


def test_resolve_mock_mode_explicit_offline_arg_overrides_key():
    """Explicit --offline CLI flag forces mock_mode to True even if SECTORS_API_KEY is present."""
    env = {
        "SECTORS_API_KEY": "valid_live_token_12345",
        "MOCK_SECTORS": "0",
        "NISKAVA_OFFLINE": "0",
    }
    assert resolve_mock_mode(args_offline=True, env=env) is True


def test_resolve_mock_mode_empty_sectors_key_defaults_to_mock():
    """Missing or empty SECTORS_API_KEY runs in mock_mode if MOCK_SECTORS=1."""
    env = {
        "SECTORS_API_KEY": "   ",
        "MOCK_SECTORS": "1",
        "NISKAVA_OFFLINE": "0",
    }
    assert resolve_mock_mode(args_offline=False, env=env) is True


def test_resolve_mock_mode_empty_sectors_key_no_flags():
    """Missing SECTORS_API_KEY with no flags defaults to False unless MOCK_SECTORS or NISKAVA_OFFLINE is set."""
    env = {
        "SECTORS_API_KEY": "",
        "MOCK_SECTORS": "0",
        "NISKAVA_OFFLINE": "0",
    }
    assert resolve_mock_mode(args_offline=False, env=env) is False


