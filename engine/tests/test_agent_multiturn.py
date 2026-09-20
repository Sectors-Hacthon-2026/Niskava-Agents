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


if __name__ == "__main__":
    unittest.main()
