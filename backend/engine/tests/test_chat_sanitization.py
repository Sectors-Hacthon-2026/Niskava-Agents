"""Targeted tests for chat message preparation and error card filtering."""

import pytest
from engine.agent.react_agent import NiskavaReActAgent
from engine.agent.tools import NiskavaToolRegistry


def test_prepare_chat_messages_strips_error_cards(tmp_path):
    """Verify _prepare_chat_messages drops assistant error markdown cards from history."""
    registry = NiskavaToolRegistry(db_path=str(tmp_path / "test.db"), mock_mode=True)
    agent = NiskavaReActAgent(tool_registry=registry, mock_mode=True)

    history = [
        {"role": "user", "content": "Halo apa kabar?"},
        {
            "role": "assistant",
            "content": "### ⚠️ Gagal Terhubung ke Provider AI\n\n- **Detail Error**: Connection refused",
        },
        {"role": "user", "content": "Cek harga ANTM"},
    ]

    messages = agent._prepare_chat_messages(
        dynamic_system_prompt="System Prompt",
        user_prompt="Cek harga ANTM",
        history=history,
    )

    # Error card should be filtered out
    error_contents = [m["content"] for m in messages if "Gagal Terhubung" in m["content"]]
    assert len(error_contents) == 0

    # User prompt should not be duplicated at the tail
    user_prompts = [m for m in messages if m["role"] == "user" and m["content"] == "Cek harga ANTM"]
    assert len(user_prompts) == 1


def test_prepare_chat_messages_cleans_dangling_tool_calls(tmp_path):
    """Verify _prepare_chat_messages sanitizes raw tool calls inside previous assistant messages."""
    registry = NiskavaToolRegistry(db_path=str(tmp_path / "test.db"), mock_mode=True)
    agent = NiskavaReActAgent(tool_registry=registry, mock_mode=True)

    history = [
        {"role": "user", "content": "Analisis IHSG"},
        {
            "role": "assistant",
            "content": 'Mencoba membaca data pasar.\n<tool_call>{"name": "harvest_market_news", "arguments": {"ticker": "IHSG"}',
        },
    ]

    messages = agent._prepare_chat_messages(
        dynamic_system_prompt="System Prompt",
        user_prompt="Lanjutkan",
        history=history,
    )

    assistant_msg = next((m for m in messages if m["role"] == "assistant"), None)
    assert assistant_msg is not None
    assert "<tool_call>" not in assistant_msg["content"]
    assert "harvest_market_news" not in assistant_msg["content"]
    assert "Mencoba membaca data pasar." in assistant_msg["content"]
