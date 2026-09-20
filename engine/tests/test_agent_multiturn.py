"""Unit tests for Python Agent multi-turn context compaction, anti-amnesia, and auto-title."""

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

    def test_gemini_multiturn_payload_and_react_loop(self) -> None:
        from unittest.mock import MagicMock, patch

        emitted_events = []
        gemini_agent = NiskavaReActAgent(
            tool_registry=self.registry,
            emitter=lambda ev: emitted_events.append(ev),
            ai_provider="gemini",
            api_key="AIzaSyMockTestKey",
            mock_mode=False,
        )

        # Mock responses from Gemini: Call 1 triggers a tool, Call 2 returns final response
        mock_resp_tool = MagicMock()
        mock_resp_tool.status_code = 200
        mock_resp_tool.json.return_value = {
            "candidates": [{
                "content": {
                    "parts": [{
                        "text": '<thought>Perlu cek anomali ANTM</thought><tool_call>{"name": "compute_quant_anomalies", "arguments": {"ticker": "ANTM"}}</tool_call>'
                    }]
                }
            }]
        }

        mock_resp_final = MagicMock()
        mock_resp_final.status_code = 200
        mock_resp_final.json.return_value = {
            "candidates": [{
                "content": {
                    "parts": [{
                        "text": '<thought>Sintesis hasil</thought><response>Analisis selesai. Ditemukan anomali volume ANTM.</response>'
                    }]
                }
            }]
        }

        with patch("requests.post", side_effect=[mock_resp_tool, mock_resp_final]) as mock_post:
            history = [
                {"role": "user", "content": "Halo apa kabar?"},
                {"role": "assistant", "content": "Kabar baik, silakan bertanya."},
            ]
            res = gemini_agent.chat(
                user_prompt="Cek anomali ANTM",
                session_id="SESS-GEM-01",
                history=history,
            )

            # Check that requests.post was called twice (ReAct loop)
            self.assertEqual(mock_post.call_count, 2)

            # Validate first call payload includes multi-turn history converted for Gemini
            first_payload = mock_post.call_args_list[0][1]["json"]
            self.assertIn("contents", first_payload)
            contents = first_payload["contents"]
            self.assertEqual(contents[0]["role"], "user")
            self.assertEqual(contents[0]["parts"][0]["text"], "Halo apa kabar?")
            self.assertEqual(contents[1]["role"], "model")
            self.assertEqual(contents[1]["parts"][0]["text"], "Kabar baik, silakan bertanya.")
            self.assertEqual(contents[2]["role"], "user")

            # Validate tool execution occurred
            tool_events = [ev for ev in emitted_events if ev.get("event") == "agent_tool_call"]
            self.assertEqual(len(tool_events), 1)
            self.assertEqual(tool_events[0]["tool"], "compute_quant_anomalies")

            # Validate final response
            self.assertIn("Analisis selesai. Ditemukan anomali volume ANTM.", res["response"])


if __name__ == "__main__":
    unittest.main()

