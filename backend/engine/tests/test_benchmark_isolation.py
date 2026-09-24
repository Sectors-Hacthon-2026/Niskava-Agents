import os
from engine.evals.benchmark import BenchmarkEvaluator


def test_benchmark_evaluator_uses_isolated_db_by_default():
    evaluator = BenchmarkEvaluator()
    # Default DB must not be the production ~/.niskava/niskava.db
    prod_path = os.path.expanduser("~/.niskava/niskava.db")
    assert evaluator.pipeline.db_path != prod_path
    assert os.path.exists(os.path.dirname(evaluator.pipeline.db_path))
