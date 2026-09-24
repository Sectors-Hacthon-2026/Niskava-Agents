"""Tests for load_dotenv_fallback ensuring empty keys do not shadow valid keys."""

import os
from unittest.mock import patch
from engine.runner import load_dotenv_fallback

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
