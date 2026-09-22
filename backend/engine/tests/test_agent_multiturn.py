"""Unit tests for Python Agent multi-turn context compaction, anti-amnesia, and auto-title."""

import json
import os
import sqlite3
import tempfile
import unittest

from engine.agent.react_agent import NiskavaReActAgent
from engine.agent.tools import NiskavaToolRegistry


class TestAgentMultiTurnAndSessions(unittest.TestCase):
    def setUp(self) -> None:
        self.tmp_db = tempfile.mktemp(suffix=".db")
        conn = sqlite3.connect(self.tmp_db)
        conn.execute(
            """
            CREATE TABLE IF NOT EXISTS chat_sessions (
                id TEXT PRIMARY KEY,
                title TEXT NOT NULL,
                model TEXT NOT NULL DEFAULT 'hermes',
                status TEXT NOT NULL DEFAULT 'IDLE',
                message_count INTEGER NOT NULL DEFAULT 0,
                last_message_preview TEXT,
                is_pinned INTEGER NOT NULL DEFAULT 0,
                parent_session_id TEXT,
                created_at TEXT NOT NULL DEFAULT CURRENT_TIMESTAMP,
                updated_at TEXT NOT NULL DEFAULT CURRENT_TIMESTAMP
            );
            """
        )
        conn.execute(
            """
            CREATE TABLE IF NOT EXISTS chat_messages (
                id TEXT PRIMARY KEY,
                session_id TEXT NOT NULL,
                role TEXT NOT NULL,
                content TEXT NOT NULL,
                thought TEXT,
                tool_calls_json TEXT,
                status TEXT NOT NULL DEFAULT 'COMPLETED',
                created_at TEXT NOT NULL DEFAULT CURRENT_TIMESTAMP
            );
            """
        )
        conn.commit()
        conn.close()

        self.registry = NiskavaToolRegistry(db_path=self.tmp_db, mock_mode=True)
        self.agent = NiskavaReActAgent(tool_registry=self.registry, mock_mode=True)

    def tearDown(self) -> None:
        if os.path.exists(self.tmp_db):
            os.remove(self.tmp_db)

    def test_auto_title_and_anti_amnesia(self) -> None:
        session_id = "TEST-SESSION-001"

        # Turn 1: Explicit ticker
        res1 = self.agent.chat("Analisis lonjakan volume ANTM", session_id=session_id)
        self.assertIn("response", res1)

        # Simulate message recording in DB
        with sqlite3.connect(self.tmp_db) as conn:
            conn.execute(
                "INSERT INTO chat_messages (id, session_id, role, content) VALUES (?, ?, ?, ?)",
                ("M1", session_id, "user", "Analisis lonjakan volume ANTM"),
            )
            conn.execute(
                "INSERT INTO chat_messages (id, session_id, role, content) VALUES (?, ?, ?, ?)",
                ("M2", session_id, "assistant", res1["response"]),
            )
            row = conn.execute("SELECT title FROM chat_sessions WHERE id = ?", (session_id,)).fetchone()
            self.assertIsNotNone(row)
            self.assertEqual(row[0], "Anomali & Volume ANTM")

        # Turn 2: Follow-up question without explicit ticker (tests anti-amnesia)
        res2 = self.agent.chat("Bagaimana valuasi rasio keuangannya?", session_id=session_id)
        self.assertIn("response", res2)
        # Verify that ANTM was recalled from context
        self.assertTrue(
            "ANTM" in res2["response"] or "ANTM" in str(res2),
            "Agent should remember ANTM as target ticker from previous turns",
        )

    def test_context_compaction_on_long_chat(self) -> None:
        session_id = "TEST-SESSION-LONG"
        with sqlite3.connect(self.tmp_db) as conn:
            for i in range(12):
                role = "user" if i % 2 == 0 else "assistant"
                content = f"Turn {i+1} message content about market and tickers"
                conn.execute(
                    "INSERT INTO chat_messages (id, session_id, role, content) VALUES (?, ?, ?, ?)",
                    (f"MSG-{i+1}", session_id, role, content),
                )

        compacted = self.agent._get_compacted_history(session_id, max_turns=8)
        self.assertLessEqual(len(compacted), 8)
        self.assertTrue(any("Konteks Sebelumnya" in m.get("content", "") for m in compacted))

    def test_universal_model_agnostic_react_loop(self) -> None:
        from unittest.mock import MagicMock, patch

        emitted_events = []
        universal_agent = NiskavaReActAgent(
            tool_registry=self.registry,
            emitter=lambda ev: emitted_events.append(ev),
            model="deepseek-ai/deepseek-r1",
            base_url="http://localhost:11434/v1",
            api_key="mock-token-xyz",
            mock_mode=False,
        )

        # Mock responses from OpenAI-compatible endpoint: Call 1 triggers tool, Call 2 returns final response
        mock_resp_tool = MagicMock()
        mock_resp_tool.status_code = 200
        mock_resp_tool.text = json.dumps({
            "choices": [{
                "message": {
                    "content": '<thought>Perlu cek anomali ANTM</thought><tool_call>{"name": "compute_quant_anomalies", "arguments": {"ticker": "ANTM"}}</tool_call>'
                }
            }]
        })

        mock_resp_final = MagicMock()
        mock_resp_final.status_code = 200
        mock_resp_final.text = json.dumps({
            "choices": [{
                "message": {
                    "content": '<thought>Sintesis hasil</thought><response>Analisis selesai. Ditemukan anomali volume ANTM.</response>'
                }
            }]
        })

        with patch("requests.post", side_effect=[mock_resp_tool, mock_resp_final]) as mock_post:
            history = [
                {"role": "user", "content": "Halo apa kabar?"},
                {"role": "assistant", "content": "Kabar baik, silakan bertanya."},
            ]
            res = universal_agent.chat(
                user_prompt="Cek anomali ANTM",
                session_id="SESS-UNI-01",
                history=history,
            )

            # Check that requests.post was called twice (ReAct loop)
            self.assertEqual(mock_post.call_count, 2)

            # Validate target URL and headers
            target_url = mock_post.call_args_list[0][0][0]
            self.assertEqual(target_url, "http://localhost:11434/v1/chat/completions")
            headers = mock_post.call_args_list[0][1]["headers"]
            self.assertEqual(headers.get("Authorization"), "Bearer mock-token-xyz")

            # Validate model and messages in payload
            first_payload = mock_post.call_args_list[0][1]["json"]
            self.assertEqual(first_payload["model"], "deepseek-ai/deepseek-r1")
            self.assertIn("messages", first_payload)
            messages = first_payload["messages"]
            self.assertEqual(messages[0]["role"], "system")
            self.assertEqual(messages[1]["role"], "user")
            self.assertEqual(messages[1]["content"], "Halo apa kabar?")
            self.assertEqual(messages[2]["role"], "assistant")
            self.assertEqual(messages[2]["content"], "Kabar baik, silakan bertanya.")
            self.assertEqual(messages[3]["role"], "user")

            # Validate tool execution occurred
            tool_events = [ev for ev in emitted_events if ev.get("event") == "agent_tool_call"]
            self.assertEqual(len(tool_events), 1)
            self.assertEqual(tool_events[0]["tool"], "compute_quant_anomalies")

            # Validate final response
            self.assertIn("Analisis selesai. Ditemukan anomali volume ANTM.", res["response"])

    def test_history_deduplication_of_current_prompt(self) -> None:
        session_id = "TEST-DEDUP-001"
        prompt = "Berapa rasio PE dan PBV saham BBRI?"
        # Simulate server saving user prompt before agent execution
        with sqlite3.connect(self.tmp_db) as conn:
            conn.execute(
                "INSERT INTO chat_messages (id, session_id, role, content) VALUES (?, ?, ?, ?)",
                ("MSG-1", session_id, "user", prompt),
            )

        history = self.agent._get_compacted_history(session_id, max_turns=8)
        self.assertEqual(len(history), 1)
        self.assertEqual(history[0]["content"], prompt)

        # Inspect prepared messages inside agent
        prepared = self.agent._prepare_chat_messages(
            dynamic_system_prompt="SYSTEM PROMPT",
            user_prompt=prompt,
            history=history,
        )
        # Ensure user prompt only appears once at the end
        user_msgs = [m for m in prepared if m["role"] == "user"]
        self.assertEqual(len(user_msgs), 1, "User prompt must not be duplicated in LLM payload")
        self.assertEqual(user_msgs[0]["content"], prompt)


if __name__ == "__main__":
    unittest.main()

