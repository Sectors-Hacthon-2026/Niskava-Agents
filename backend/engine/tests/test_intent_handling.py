"""Unit tests for Intent Routing and Ticker Disambiguation in ReAct Agent."""

import os
import unittest
from engine.agent.react_agent import NiskavaReActAgent
from engine.agent.tools import NiskavaToolRegistry


class TestIntentHandling(unittest.TestCase):
    def setUp(self):
        self.db_path = "/tmp/test_intent_agent.db"
        if os.path.exists(self.db_path):
            os.remove(self.db_path)
        self.registry = NiskavaToolRegistry(self.db_path, mock_mode=True)
        self.events = []
        self.agent = NiskavaReActAgent(
            tool_registry=self.registry,
            emitter=lambda ev: self.events.append(ev),
            mock_mode=True,
        )

    def tearDown(self):
        if os.path.exists(self.db_path):
            os.remove(self.db_path)

    def test_user_typo_domg_news_intent(self):
        """Prompt 'cek berita hari ini domg' must NOT be parsed as ticker DOMG."""
        prompt = "cek berita hari ini domg"
        result = self.agent.chat(prompt)

        # 1. Ticker DOMG must NOT appear as an anomaly target
        self.assertEqual(len(result["anomalies"]), 0)
        self.assertNotIn("DOMG", result["response"])

        # 2. Tool harvest_market_news should be called for general news
        tool_calls = [e.get("tool") for e in self.events if e.get("event") == "agent_tool_call"]
        self.assertIn("harvest_market_news", tool_calls)

        # 3. Response should contain market news overview
        self.assertIn("Rangkuman Berita Pasar", result["response"])

    def test_greeting_intent_natural_response(self):
        """Prompt 'halo' should return a warm, natural greeting without skill dump."""
        self.events.clear()
        result = self.agent.chat("halo selamat pagi")

        self.assertEqual(len(result["anomalies"]), 0)
        resp = result["response"]
        self.assertIn("Halo! Saya **Niskava Agent**", resp)
        # Should NOT dump raw skill identifiers like 'skill_market_anomaly_recon'
        self.assertNotIn("skill_market_anomaly_recon", resp)
        self.assertNotIn("Three-Tier Verification Taxonomy", resp)

    def test_genuine_stock_intent_triggers_quant_anomalies(self):
        """Prompt with genuine ticker like 'ANTM' must execute deterministic quant math."""
        self.events.clear()
        result = self.agent.chat("Apakah ada anomali volume pada saham ANTM?")

        self.assertTrue(len(result["anomalies"]) > 0)
        self.assertIn("ANTM", result["response"])
        self.assertIn("Z-Score", result["response"])
