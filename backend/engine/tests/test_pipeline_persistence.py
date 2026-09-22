import sqlite3
import pytest
from engine.agent.pipeline import InvestigationPipeline

def test_investigation_pipeline_persists_full_lifecycle(tmp_path):
    db_path = str(tmp_path / "test_niskava.db")
    
    with sqlite3.connect(db_path) as conn:
        conn.execute("""
            CREATE TABLE IF NOT EXISTS investigations (
                id TEXT PRIMARY KEY,
                ticker TEXT NOT NULL,
                market TEXT NOT NULL DEFAULT 'IDX',
                timeframe_days INTEGER NOT NULL DEFAULT 30,
                status TEXT NOT NULL,
                started_at TEXT NOT NULL,
                completed_at TEXT,
                summary_text TEXT,
                created_at TEXT NOT NULL DEFAULT CURRENT_TIMESTAMP
            )
        """)
        conn.execute("""
            CREATE TABLE IF NOT EXISTS anomalies (
                id TEXT PRIMARY KEY,
                investigation_id TEXT NOT NULL REFERENCES investigations(id) ON DELETE CASCADE,
                anomaly_date TEXT NOT NULL,
                metric_type TEXT NOT NULL,
                metric_value REAL NOT NULL,
                baseline_value REAL NOT NULL,
                z_score REAL NOT NULL,
                description TEXT,
                created_at TEXT NOT NULL DEFAULT CURRENT_TIMESTAMP
            )
        """)
        conn.execute("""
            CREATE TABLE IF NOT EXISTS findings (
                id TEXT PRIMARY KEY,
                investigation_id TEXT NOT NULL REFERENCES investigations(id) ON DELETE CASCADE,
                title TEXT NOT NULL,
                claim_text TEXT NOT NULL,
                verification_status TEXT NOT NULL,
                confidence_score REAL NOT NULL,
                causality_status TEXT NOT NULL,
                created_at TEXT NOT NULL DEFAULT CURRENT_TIMESTAMP
            )
        """)
        conn.execute("""
            CREATE TABLE IF NOT EXISTS evidence_items (
                id TEXT PRIMARY KEY,
                finding_id TEXT NOT NULL REFERENCES findings(id) ON DELETE CASCADE,
                source_type TEXT NOT NULL,
                source_name TEXT NOT NULL,
                source_url TEXT NOT NULL,
                publication_date TEXT NOT NULL,
                snippet_text TEXT NOT NULL,
                created_at TEXT NOT NULL DEFAULT CURRENT_TIMESTAMP
            )
        """)
        conn.execute("""
            CREATE TABLE IF NOT EXISTS sectors_cache (
                cache_key TEXT PRIMARY KEY,
                endpoint TEXT NOT NULL,
                payload_json TEXT NOT NULL,
                created_at TEXT NOT NULL DEFAULT CURRENT_TIMESTAMP,
                expires_at TEXT
            )
        """)

    session_id = "INV-TEST-001"
    pipeline = InvestigationPipeline(db_path=db_path, mock_mode=True)
    res = pipeline.run(ticker="ANTM", timeframe_days=30, session_id=session_id)

    with sqlite3.connect(db_path) as conn:
        row = conn.execute("SELECT status, summary_text FROM investigations WHERE id = ?", (session_id,)).fetchone()
        assert row is not None, "Investigation record must exist"
        assert row[0] == "COMPLETED", f"Expected COMPLETED status, got {row[0]}"
        assert row[1] is not None and len(row[1]) > 0, "Summary text must not be empty"

        anom_count = conn.execute("SELECT count(*) FROM anomalies WHERE investigation_id = ?", (session_id,)).fetchone()[0]
        assert anom_count > 0, "Anomalies must be persisted to SQLite"

        find_count = conn.execute("SELECT count(*) FROM findings WHERE investigation_id = ?", (session_id,)).fetchone()[0]
        assert find_count > 0, "Findings must be persisted to SQLite"
