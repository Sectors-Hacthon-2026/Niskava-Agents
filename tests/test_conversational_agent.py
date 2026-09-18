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


if __name__ == "__main__":
    unittest.main()
