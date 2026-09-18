"""Unit tests for Sectors API Client and SQLite Caching layer."""

import os
import sqlite3
import pytest
from engine.sectors.client import SectorsAPIClient


def test_sectors_client_mock_and_cache(tmp_path):
    db_path = str(tmp_path / "test_sectors.db")
    
    # Initialize SQLite database with sectors_cache table
    with sqlite3.connect(db_path) as conn:
        conn.execute("""
            CREATE TABLE sectors_cache (
                cache_key TEXT PRIMARY KEY,
                endpoint TEXT NOT NULL,
                payload_json TEXT NOT NULL,
                created_at TEXT NOT NULL DEFAULT CURRENT_TIMESTAMP,
                expires_at TEXT
            )
        """)

    client = SectorsAPIClient(db_path=db_path, mock_mode=True)
    candles = client.get_daily_candles("ANTM")

    assert len(candles) == 30
    assert candles[25]["volume"] == 125_000_000.0

    # Verify that cache is populated in SQLite
    with sqlite3.connect(db_path) as conn:
        cursor = conn.cursor()
        cursor.execute("SELECT COUNT(*) FROM sectors_cache WHERE endpoint LIKE '%/daily/%'")
        count = cursor.fetchone()[0]
        assert count == 1

    # Second call should read directly from SQLite cache
    candles_cached = client.get_daily_candles("ANTM")
    assert len(candles_cached) == 30
    assert candles_cached == candles
