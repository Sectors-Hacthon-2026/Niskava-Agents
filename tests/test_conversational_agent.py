import os
import unittest
from engine.agent.react_agent import NiskavaReActAgent
from engine.agent.tools import NiskavaToolRegistry


class TestConversationalAgent(unittest.TestCase):
    def setUp(self):
        self.db_path = "/tmp/test_niskava_chat.db"
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

    def test_conversational_prompt_extracts_ticker_and_enforces_law1(self):
        prompt = "Tolong periksa apakah saham ANTM mengalami lonjakan volume mencurigakan?"
        result = self.agent.chat(prompt)

        self.assertIn("response", result)
        self.assertIn("anomalies", result)
        self.assertIn("findings", result)
        self.assertTrue(len(result["anomalies"]) > 0)

        # Check emitted events
        event_types = [e.get("event") for e in self.events]
        self.assertIn("session_start", event_types)
        self.assertIn("agent_thought", event_types)
        self.assertIn("agent_tool_call", event_types)
        self.assertIn("anomaly_detected", event_types)
        self.assertIn("agent_observation", event_types)
        self.assertIn("agent_message_chunk", event_types)
        self.assertIn("session_complete", event_types)

        # Check response content
        resp = result["response"]
        self.assertIn("ANTM", resp)
        self.assertIn("Z-Score", resp)
        self.assertIn("DISCLAIMER FINANSIAL", resp)


    def test_conversational_prompt_no_ticker_greeting_mock_mode(self):
        # When user sends casual greetings like 'halo' or short words like 'da'
        for greeting in ["halo", "da", "test", "selamat pagi"]:
            self.events.clear()
            result = self.agent.chat(greeting)

            self.assertIn("response", result)
            self.assertEqual(len(result["anomalies"]), 0)
            self.assertEqual(len(result["findings"]), 0)

            # Response should guide the user to input a ticker, NOT fake ANTM data
            resp = result["response"]
            self.assertIn("Mode Offline/Mock", resp)
            self.assertIn("kode saham", resp)
            self.assertNotIn("35.71", resp)

            event_types = [e.get("event") for e in self.events]
            self.assertIn("session_start", event_types)
            self.assertIn("session_complete", event_types)
            self.assertNotIn("anomaly_detected", event_types)

    def test_ai_provider_connection_failure_emits_error_and_no_dummy_data(self):
        # Configure agent with an unreachable OpenAI endpoint and mock_mode=False
        error_events = []
        failing_agent = NiskavaReActAgent(
            tool_registry=self.registry,
            emitter=lambda ev: error_events.append(ev),
            mock_mode=False,
        )
        failing_agent.ai_provider = "openai"
        failing_agent.openai_base_url = "http://127.0.0.1:59999/v1"
        failing_agent.openai_model = "test-model"

        result = failing_agent.chat("da")

        self.assertEqual(result.get("status"), "ERROR")
        self.assertEqual(len(result.get("anomalies", [])), 0)
        self.assertEqual(len(result.get("findings", [])), 0)

        # Must report explicit error diagnostic, not fake ANTM anomalies
        self.assertIn("Gagal Terhubung ke Provider AI", result["response"])
        self.assertIn("http://127.0.0.1:59999/v1", result["response"])
        self.assertNotIn("35.71", result["response"])

        event_types = [e.get("event") for e in error_events]
        self.assertIn("session_error", event_types)
        self.assertNotIn("anomaly_detected", event_types)

    def test_gemini_missing_api_key_emits_error(self):
        error_events = []
        gemini_agent = NiskavaReActAgent(
            tool_registry=self.registry,
            emitter=lambda ev: error_events.append(ev),
            mock_mode=False,
            api_key="",
        )
        gemini_agent.ai_provider = "gemini"

        result = gemini_agent.chat("halo")

        self.assertEqual(result.get("status"), "ERROR")
        self.assertIn("GEMINI_API_KEY", result["response"])

        event_types = [e.get("event") for e in error_events]
        self.assertIn("session_error", event_types)
        self.assertNotIn("anomaly_detected", event_types)


if __name__ == "__main__":
    unittest.main()
