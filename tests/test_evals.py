"""Test suite for Niskava Ground-Truth Evaluation Benchmark."""

import unittest
from engine.evals.benchmark import BenchmarkEvaluator


class TestNiskavaEvalsBenchmark(unittest.TestCase):
    """Test suite executing benchmark evaluator in mock mode."""

    def test_benchmark_evaluator_metrics(self):
        evaluator = BenchmarkEvaluator(mock_mode=True)
        results = evaluator.run_benchmark()

        self.assertGreaterEqual(results["total_cases"], 1)
        self.assertEqual(results["guardrail_compliance_pct"], 100.0)
        self.assertEqual(results["hallucination_free_pct"], 100.0)
        self.assertIn("taxonomy_accuracy_pct", results)
        self.assertIn("causality_precision_pct", results)


if __name__ == "__main__":
    unittest.main()
