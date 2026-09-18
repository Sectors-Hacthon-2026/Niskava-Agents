"""Production-Grade Sectors Financial API v2 Client with SQLite Caching.

Complies with Law 5 (Credit Budget Discipline & Local Caching):
Never execute duplicate HTTP requests. Historical candlestick data
(T < today) is cached permanently (expires_at = NULL).
"""

import hashlib
import json
import os
import sqlite3
import time
from datetime import datetime, timedelta
from typing import Any, Dict, List, Optional
import requests


class SectorsAPIClient:
    """Client for Sectors Financial API v2 with transparent local SQLite caching."""

    BASE_URL = "https://api.sectors.app/v2"

    def __init__(
        self,
        db_path: str,
        api_key: Optional[str] = None,
        mock_mode: Optional[bool] = None,
        base_url: Optional[str] = None,
    ):
        self.api_key = api_key or os.environ.get("SECTORS_API_KEY", "")
        self.base_url = base_url or os.environ.get("SECTORS_BASE_URL", self.BASE_URL)
        self.db_path = os.path.expanduser(db_path)
        
        if mock_mode is not None:
            self.mock_mode = mock_mode
        else:
            self.mock_mode = (
                os.environ.get("MOCK_SECTORS", "0") in ("1", "true", "True")
                or os.environ.get("NISKAVA_OFFLINE", "0") in ("1", "true", "True")
                or not self.api_key
            )

        self.session = requests.Session()
        self.session.headers.update({
            "Authorization": self.api_key,
            "User-Agent": "Niskava-Agent/1.0.0 (Track1-AI-Agents)",
            "Accept": "application/json",
        })

    def _generate_cache_key(self, endpoint: str, params: Optional[Dict[str, Any]] = None) -> str:
        param_str = json.dumps(params or {}, sort_keys=True)
        raw = f"{endpoint}:{param_str}"
        return hashlib.sha256(raw.encode("utf-8")).hexdigest()

    def _get_cache(self, cache_key: str) -> Optional[Any]:
        if not os.path.exists(self.db_path):
            return None
        try:
            with sqlite3.connect(self.db_path) as conn:
                cursor = conn.cursor()
                cursor.execute(
                    """
                    SELECT payload_json FROM sectors_cache
                    WHERE cache_key = ? AND (expires_at IS NULL OR expires_at > datetime('now'))
                    """,
                    (cache_key,),
                )
                row = cursor.fetchone()
                if row:
                    return json.loads(row[0])
        except sqlite3.Error:
            return None
        return None

    def _set_cache(
        self,
        cache_key: str,
        endpoint: str,
        data: Any,
        ttl_seconds: Optional[int] = None,
    ) -> None:
        if not os.path.exists(self.db_path):
            os.makedirs(os.path.dirname(self.db_path), exist_ok=True)
        try:
            payload_json = json.dumps(data)
            with sqlite3.connect(self.db_path) as conn:
                cursor = conn.cursor()
                if ttl_seconds:
                    cursor.execute(
                        """
                        INSERT OR REPLACE INTO sectors_cache (cache_key, endpoint, payload_json, expires_at)
                        VALUES (?, ?, ?, datetime('now', ?))
                        """,
                        (cache_key, endpoint, payload_json, f"+{ttl_seconds} seconds"),
                    )
                else:
                    cursor.execute(
                        """
                        INSERT OR REPLACE INTO sectors_cache (cache_key, endpoint, payload_json, expires_at)
                        VALUES (?, ?, ?, NULL)
                        """,
                        (cache_key, endpoint, payload_json),
                    )
                conn.commit()
        except sqlite3.Error:
            pass

    def _request(
        self,
        endpoint: str,
        params: Optional[Dict[str, Any]] = None,
        ttl_seconds: Optional[int] = None,
    ) -> Any:
        cache_key = self._generate_cache_key(endpoint, params)
        cached = self._get_cache(cache_key)
        if cached is not None:
            return cached

        if self.mock_mode:
            mock_data = self._generate_mock_data(endpoint, params)
            self._set_cache(cache_key, endpoint, mock_data, ttl_seconds)
            return mock_data

        url = f"{self.base_url}{endpoint}"
        resp = self.session.get(url, params=params, timeout=12.0)
        resp.raise_for_status()
        data = resp.json()

        self._set_cache(cache_key, endpoint, data, ttl_seconds)
        return data

    def get_daily_candles(
        self, symbol: str, start: Optional[str] = None, end: Optional[str] = None
    ) -> List[Dict[str, Any]]:
        """Retrieve daily OHLCV candlestick data."""
        endpoint = f"/daily/{symbol.upper()}/"
        params = {}
        if start:
            params["start"] = start
        if end:
            params["end"] = end
        # Historical candlestick data is permanently cached (ttl=None)
        return self._request(endpoint, params, ttl_seconds=None)

    def get_company_report(
        self, symbol: str, sections: str = "valuation,financials,peers"
    ) -> Dict[str, Any]:
        """Fetch company fundamental report (cached for 24 hours)."""
        endpoint = f"/company/report/{symbol.upper()}/"
        params = {"sections": sections}
        return self._request(endpoint, params, ttl_seconds=86400)

    def get_foreign_flow(self, symbol: str) -> List[Dict[str, Any]]:
        """Retrieve Foreign Flow Net Inflow data."""
        endpoint = f"/foreign-flow/{symbol.upper()}/"
        return self._request(endpoint, ttl_seconds=86400)

    def get_news(self, symbol: Optional[str] = None) -> List[Dict[str, Any]]:
        """Fetch curated financial news."""
        endpoint = "/news/"
        params = {"symbol": symbol.upper()} if symbol else {}
        return self._request(endpoint, params, ttl_seconds=3600)

    def _generate_mock_data(self, endpoint: str, params: Optional[Dict[str, Any]] = None) -> Any:
        """Generate realistic mock data fixtures for offline development and CI tests."""
        symbol = "ANTM"
        if "/daily/" in endpoint:
            # 30 days of synthetic candles with a volume surge on day 25
            candles = []
            base_date = datetime.now() - timedelta(days=35)
            price = 1500.0
            for day_idx in range(30):
                curr_date = (base_date + timedelta(days=day_idx)).strftime("%Y-%m-%d")
                if day_idx == 25:
                    volume = 125_000_000.0  # Massive spike
                    close = price * 1.082  # +8.2%
                else:
                    volume = 20_000_000.0 + (day_idx % 5) * 2_000_000.0
                    close = price * (1.0 + ((day_idx % 3) - 1) * 0.01)
                candles.append({
                    "date": curr_date,
                    "open": round(price, 2),
                    "high": round(close * 1.02, 2),
                    "low": round(price * 0.98, 2),
                    "close": round(close, 2),
                    "volume": round(volume, 2),
                })
                price = close
            return candles

        if "/company/report/" in endpoint:
            return {
                "symbol": symbol,
                "company_name": "PT Aneka Tambang Tbk",
                "sector": "Basic Materials",
                "sub_sector": "Metals & Minerals",
                "market_cap": 38_500_000_000_000,
                "pe_ratio": 12.4,
                "pb_ratio": 1.65,
            }

        if "/foreign-flow/" in endpoint:
            return [
                {"date": "2026-09-12", "net_foreign_buy": 84_500_000_000.0},
                {"date": "2026-09-11", "net_foreign_buy": -12_000_000_000.0},
            ]

        if "/news/" in endpoint:
            return [
                {
                    "title": "ANTM Resmikan Uji Coba Smelter Feronikel Baru di Halmahera",
                    "source": "IDX Channel",
                    "url": "https://idxchannel.com/market/antm-smelter-halmahera",
                    "publish_date": "2026-09-12T07:30:00Z",
                    "snippet": "PT Aneka Tambang Tbk (ANTM) mengumumkan penyelesaian proyek hilirisasi nikel...",
                }
            ]

        return {"status": "ok", "mock": True}
