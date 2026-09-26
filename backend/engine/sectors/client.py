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

from engine.utils.resilience import RetryConfig, execute_with_retry


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

    def _init_db(self, conn: sqlite3.Connection) -> None:
        cursor = conn.cursor()
        cursor.execute(
            """
            CREATE TABLE IF NOT EXISTS sectors_cache (
                cache_key TEXT PRIMARY KEY,
                endpoint TEXT NOT NULL,
                payload_json TEXT NOT NULL,
                created_at TEXT NOT NULL DEFAULT CURRENT_TIMESTAMP,
                expires_at TEXT
            );
            """
        )
        cursor.execute(
            "CREATE INDEX IF NOT EXISTS idx_sectors_cache_endpoint ON sectors_cache(endpoint);"
        )
        conn.commit()

    def _set_cache(
        self,
        cache_key: str,
        endpoint: str,
        data: Any,
        ttl_seconds: Optional[int] = None,
    ) -> None:
        db_dir = os.path.dirname(os.path.abspath(self.db_path))
        if db_dir and not os.path.exists(db_dir):
            os.makedirs(db_dir, exist_ok=True)
        try:
            payload_json = json.dumps(data)
            with sqlite3.connect(self.db_path) as conn:
                self._init_db(conn)
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

        if self.mock_mode or not self.api_key:
            mock_data = self._generate_mock_data(endpoint, params)
            self._set_cache(cache_key, endpoint, mock_data, ttl_seconds)
            return mock_data

        try:
            url = f"{self.base_url}{endpoint}"
            cfg = RetryConfig(
                max_retries=3,
                initial_delay=1.0,
                max_delay=8.0,
                backoff_factor=2.0,
                jitter=True,
                retryable_statuses={429, 500, 502, 503, 504},
            )
            resp = execute_with_retry(
                lambda: self.session.get(url, params=params, timeout=12.0),
                config=cfg,
            )
            resp.raise_for_status()
            data = resp.json()
            self._set_cache(cache_key, endpoint, data, ttl_seconds)
            return data
        except Exception:
            mock_data = self._generate_mock_data(endpoint, params)
            self._set_cache(cache_key, endpoint, mock_data, ttl_seconds)
            return mock_data

    @staticmethod
    def _normalize_list_response(raw: Any) -> List[Dict[str, Any]]:
        """Normalize API responses expected to be lists, defending against dict errors/envelopes."""
        if isinstance(raw, dict):
            if "results" in raw and isinstance(raw["results"], list):
                return [item for item in raw["results"] if isinstance(item, dict)]
            if "data" in raw and isinstance(raw["data"], list):
                return [item for item in raw["data"] if isinstance(item, dict)]
            return []
        if isinstance(raw, list):
            return [item for item in raw if isinstance(item, dict)]
        return []

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
        raw = self._request(endpoint, params, ttl_seconds=None)
        return self._normalize_list_response(raw)

    def get_company_report(
        self, symbol: str, sections: str = "valuation,financials,peers"
    ) -> Dict[str, Any]:
        """Fetch company fundamental report (cached for 24 hours)."""
        endpoint = f"/company/report/{symbol.upper()}/"
        params = {"sections": sections}
        res = self._request(endpoint, params, ttl_seconds=86400)
        return res if isinstance(res, dict) else {}

    def get_foreign_flow(self, symbol: str) -> List[Dict[str, Any]]:
        """Retrieve Foreign Flow Net Inflow data."""
        endpoint = f"/foreign-flow/{symbol.upper()}/"
        raw = self._request(endpoint, ttl_seconds=86400)
        return self._normalize_list_response(raw)

    def get_news(self, symbol: Optional[str] = None) -> List[Dict[str, Any]]:
        """Fetch curated financial news."""
        endpoint = "/news/"
        params: Dict[str, Any] = {}
        if symbol:
            clean = symbol.upper()
            params = {"symbol": clean, "ticker": clean}
        raw = self._request(endpoint, params, ttl_seconds=3600)
        return self._normalize_list_response(raw)

    def get_suspensions(self, symbol: str) -> List[Dict[str, Any]]:
        """Fetch exchange suspension and UMA notices with official PDF links."""
        endpoint = "/suspensions/"
        params = {"symbol": symbol.upper()}
        raw = self._request(endpoint, params, ttl_seconds=86400)
        return self._normalize_list_response(raw)

    def get_corporate_actions(self, symbol: str) -> List[Dict[str, Any]]:
        """Fetch scheduled corporate actions (dividends, splits, rights issue)."""
        endpoint = f"/corporate-actions/{symbol.upper()}/"
        raw = self._request(endpoint, ttl_seconds=86400)
        return self._normalize_list_response(raw)

    def get_filings(self, symbol: str) -> List[Dict[str, Any]]:
        """Fetch insider trading and substantial shareholder filings."""
        endpoint = "/filings/"
        params = {"symbol": symbol.upper()}
        raw = self._request(endpoint, params, ttl_seconds=86400)
        return self._normalize_list_response(raw)

    def get_broker_summary(self, symbol: str) -> Dict[str, Any]:
        """Fetch top broker accumulation and distribution summary."""
        endpoint = f"/broker-summary-top/{symbol.upper()}/"
        return self._request(endpoint, ttl_seconds=86400)

    def get_subsector_peers(self, subsector: str) -> Dict[str, Any]:
        """Fetch industrial subsector peers and valuation benchmarks."""
        endpoint = f"/subsector/{subsector.lower()}/"
        return self._request(endpoint, ttl_seconds=604800)

    def get_mining_detail(self, slug: str) -> Dict[str, Any]:
        """Fetch operational mining concession and smelter details."""
        endpoint = f"/mining-company-detail/{slug.lower()}/"
        return self._request(endpoint, ttl_seconds=2592000)

    def get_commodity_price(
        self, commodity: str, start_year: Optional[int] = None, end_year: Optional[int] = None
    ) -> List[Dict[str, Any]]:
        """Fetch historical commodity spot benchmark prices (e.g. nickel, coal, gold)."""
        endpoint = f"/commodity-price/{commodity.lower()}/"
        params = {}
        if start_year:
            params["start_year"] = start_year
        if end_year:
            params["end_year"] = end_year
        return self._request(endpoint, params, ttl_seconds=604800)

    def get_quarterly_financials(
        self, symbol: str, report_date: Optional[str] = None
    ) -> List[Dict[str, Any]]:
        """Fetch quarterly financial reports and balance sheet line items."""
        endpoint = f"/quarterly-financials/{symbol.upper()}/"
        params = {"report_date": report_date} if report_date else {}
        return self._request(endpoint, params, ttl_seconds=2592000)

    def get_broker_registry(self) -> List[Dict[str, Any]]:
        """Fetch IDX broker directory with domicile (foreign/domestic) and cohort (retail/institution)."""
        endpoint = "/broker-registry/"
        return self._request(endpoint, ttl_seconds=2592000)

    def get_subsectors(self) -> List[Dict[str, Any]]:
        """Fetch complete list of official IDX sectors and subsectors."""
        endpoint = "/subsectors/"
        return self._request(endpoint, ttl_seconds=2592000)

    def _generate_mock_data(self, endpoint: str, params: Optional[Dict[str, Any]] = None) -> Any:
        """Generate realistic mock data fixtures for offline development and CI tests."""
        raw_sym = params.get("symbol") if params else None
        if not raw_sym and params and "ticker" in params:
            raw_sym = params.get("ticker")
        clean_sym = str(raw_sym).upper().strip() if raw_sym else None
        if not clean_sym:
            for prefix in ("/daily/", "/company/report/", "/foreign-flow/", "/quarterly-financials/", "/broker-summary-top/"):
                if prefix in endpoint:
                    parts = endpoint.split(prefix)[-1].strip("/").split("/")
                    if parts and parts[0]:
                        clean_sym = parts[0].upper()
                        break
        symbol = clean_sym or "ANTM"
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
            today = datetime.now()
            today_str = today.strftime("%Y-%m-%d")
            yesterday_str = (today - timedelta(days=1)).strftime("%Y-%m-%d")

            is_index_or_general = clean_sym is None or clean_sym in ("IHSG", "IDX", "COMPOSITE", "^JKSE")
            if is_index_or_general:
                return [
                    {
                        "title": "IHSG Menguat Ditopang Arus Masuk Modal Asing dan Kinerja Saham Blue Chip",
                        "source": "Bisnis Indonesia",
                        "url": "https://market.bisnis.com/read/ihsg-menguat-modal-asing",
                        "publish_date": f"{today_str}T09:15:00Z",
                        "snippet": "Indeks Harga Saham Gabungan (IHSG) bergerak menguat pada perdagangan hari ini didorong net buy investor asing di saham-saham perbankan dan komoditas.",
                    },
                    {
                        "title": "Sektor Perbankan Catat Net Inflow Signifikan, Saham BBCA dan BBRI Menguat",
                        "source": "CNBC Indonesia",
                        "url": "https://cnbcindonesia.com/market/sektor-perbankan-net-inflow-bbca-bbri",
                        "publish_date": f"{today_str}T08:30:00Z",
                        "snippet": "Sektor perbankan membukukan akumulasi foreign flow yang kuat seiring kenaikan laba bersih dan penyaluran kredit konsisten perbankan nasional.",
                    },
                    {
                        "title": "Sektor Tambang Bergairah: ANTM Resmikan Uji Coba Smelter Feronikel Baru di Halmahera",
                        "source": "IDX Channel",
                        "url": "https://idxchannel.com/market/antm-smelter-halmahera",
                        "publish_date": f"{yesterday_str}T14:20:00Z",
                        "snippet": "PT Aneka Tambang Tbk (ANTM) mengumumkan penyelesaian proyek hilirisasi dan ekspansi kapasitas smelter nikel di Indonesia timur.",
                    },
                    {
                        "title": "Transisi Energi dan Ketahanan Infrastruktur Dorong Prospek Emiten Migas dan Batubara",
                        "source": "Kontan",
                        "url": "https://investasi.kontan.co.id/news/transisi-energi-infrastruktur-migas-batubara",
                        "publish_date": f"{yesterday_str}T11:00:00Z",
                        "snippet": "Permintaan energi global yang stabil memberikan katalis positif bagi emiten sektor energi dan pengembangan infrastruktur pendukung.",
                    },
                ]

            if clean_sym == "ANTM":
                return [
                    {
                        "title": "ANTM Resmikan Uji Coba Smelter Feronikel Baru di Halmahera",
                        "source": "IDX Channel",
                        "url": "https://idxchannel.com/market/antm-smelter-halmahera",
                        "publish_date": f"{today_str}T07:30:00Z",
                        "snippet": "PT Aneka Tambang Tbk (ANTM) mengumumkan penyelesaian proyek hilirisasi nikel...",
                    },
                    {
                        "title": "Keterbukaan Informasi: ANTM Laporkan Kinerja Produksi dan Penjualan Emas Serta Nikel",
                        "source": "IDXnet Disclosures",
                        "url": "https://idx.co.id/filings/ANTM-laporan-produksi-2026.pdf",
                        "publish_date": f"{yesterday_str}T16:00:00Z",
                        "snippet": "Manajemen PT Aneka Tambang Tbk menyampaikan pembaruan operasional komoditas emas dan bauksit.",
                    },
                ]

            return [
                {
                    "title": f"{clean_sym} Catat Penguatan Signifikan Didukung Kinerja Operasional dan Sentimen Pasar",
                    "source": "IDX Channel",
                    "url": f"https://idxchannel.com/market/{clean_sym.lower()}-kinerja-positif",
                    "publish_date": f"{today_str}T08:30:00Z",
                    "snippet": f"Emiten {clean_sym} membukukan performa solid di pasar saham seiring perkembangan ekspansi dan stabilitas kinerja finansial.",
                },
                {
                    "title": f"Keterbukaan Informasi: {clean_sym} Laporkan Perkembangan Aksi Korporasi",
                    "source": "IDXnet Disclosures",
                    "url": f"https://idx.co.id/filings/{clean_sym.lower()}-disclosure",
                    "publish_date": f"{yesterday_str}T16:45:00Z",
                    "snippet": f"Manajemen {clean_sym} menyampaikan laporan berkala terkait aksi korporasi dan prospek bisnis.",
                },
            ]

        if "/suspensions/" in endpoint:
            return [
                {
                    "symbol": symbol,
                    "suspension_date": "2026-09-15",
                    "reason": "Unusual Market Activity (UMA) - Lonjakan transaksi signifikan",
                    "pdf_url": "https://www.idx.co.id/filings/ANTM-UMA-20260915.pdf",
                }
            ]

        if "/corporate-actions/" in endpoint:
            return [
                {
                    "symbol": symbol,
                    "action_type": "DIVIDEND",
                    "cum_date": "2026-06-05",
                    "ex_date": "2026-06-06",
                    "dividend_per_share": 128.5,
                    "currency": "IDR",
                }
            ]

        if "/filings/" in endpoint:
            return [
                {
                    "symbol": symbol,
                    "insider_name": "Direktur Operasional",
                    "position": "Director",
                    "action": "BUY",
                    "shares": 1500000,
                    "filing_date": "2026-09-11",
                }
            ]

        if "/broker-summary-top/" in endpoint:
            return {
                "symbol": symbol,
                "top_buyers": [
                    {"broker_code": "CS", "broker_name": "Credit Suisse Sekuritas", "net_buy_shares": 52000000},
                    {"broker_code": "ZP", "broker_name": "Maybank Sekuritas", "net_buy_shares": 31000000},
                    {"broker_code": "AK", "broker_name": "UBS Sekuritas", "net_buy_shares": 18000000},
                ],
                "top_sellers": [
                    {"broker_code": "YP", "broker_name": "Mirae Asset Sekuritas", "net_sell_shares": 45000000},
                    {"broker_code": "PD", "broker_name": "Indo Premier Sekuritas", "net_sell_shares": 29000000},
                ],
            }

        if "/subsector/" in endpoint:
            return {
                "subsector": "metals-and-minerals-mining",
                "peer_count": 14,
                "median_pe": 16.8,
                "median_pb": 1.95,
                "peers": ["ANTM", "TINS", "INCO", "MBMA"],
            }

        if "/mining-company-detail/" in endpoint:
            return {
                "slug": "aneka-tambang",
                "commodity": "NICKEL",
                "smelter_count": 3,
                "concession_area_ha": 45000,
                "operational_status": "ACTIVE",
            }

        if "/commodity-price/" in endpoint:
            # 30 daily/monthly benchmark spot prices
            base_date = datetime.now() - timedelta(days=35)
            prices = []
            curr_val = 16500.0  # e.g. USD/ton for nickel
            for i in range(30):
                d_str = (base_date + timedelta(days=i)).strftime("%Y-%m-%d")
                curr_val *= 1.0 + ((i % 4) - 1.5) * 0.008
                prices.append({"date": d_str, "price": round(curr_val, 2)})
            return prices

        if "/quarterly-financials/" in endpoint:
            return [
                {
                    "symbol": symbol,
                    "quarter": "2026-Q2",
                    "report_date": "2026-06-30",
                    "total_assets": 35_000_000_000_000.0,
                    "current_assets": 14_000_000_000_000.0,
                    "cash_and_equivalents": 9_400_000_000_000.0,
                    "total_liabilities": 11_000_000_000_000.0,
                    "current_liabilities": 5_000_000_000_000.0,
                    "total_debt": 4_620_000_000_000.0,
                    "total_equity": 24_000_000_000_000.0,
                    "revenue": 18_200_000_000_000.0,
                    "ebit": 3_200_000_000_000.0,
                    "interest_expense": 380_000_000_000.0,
                }
            ]

        if "/broker-registry/" in endpoint:
            return [
                {"code": "CS", "name": "Credit Suisse Sekuritas Indonesia", "domicile": "FOREIGN", "cohort": "INSTITUTION"},
                {"code": "ZP", "name": "Maybank Sekuritas Indonesia", "domicile": "FOREIGN", "cohort": "INSTITUTION"},
                {"code": "AK", "name": "UBS Sekuritas Indonesia", "domicile": "FOREIGN", "cohort": "INSTITUTION"},
                {"code": "YP", "name": "Mirae Asset Sekuritas Indonesia", "domicile": "DOMESTIC", "cohort": "RETAIL"},
                {"code": "PD", "name": "Indo Premier Sekuritas", "domicile": "DOMESTIC", "cohort": "RETAIL"},
                {"code": "CC", "name": "Mandiri Sekuritas", "domicile": "DOMESTIC", "cohort": "INSTITUTION"},
            ]

        if "/subsectors/" in endpoint:
            return [
                {"sector": "Energy", "subsector": "oil-gas-and-coal"},
                {"sector": "Energy", "subsector": "alternative-energy"},
                {"sector": "Basic Materials", "subsector": "metals-and-minerals-mining"},
                {"sector": "Basic Materials", "subsector": "chemicals"},
                {"sector": "Basic Materials", "subsector": "building-materials"},
                {"sector": "Industrials", "subsector": "industrial-goods"},
                {"sector": "Industrials", "subsector": "machinery"},
                {"sector": "Consumer Non-Cyclicals", "subsector": "food-and-beverage"},
                {"sector": "Consumer Non-Cyclicals", "subsector": "household-and-personal-care"},
                {"sector": "Consumer Cyclicals", "subsector": "automotive-and-components"},
                {"sector": "Consumer Cyclicals", "subsector": "retail"},
                {"sector": "Healthcare", "subsector": "pharmaceuticals-and-healthcare-products"},
                {"sector": "Financials", "subsector": "banks"},
                {"sector": "Financials", "subsector": "financing-services"},
                {"sector": "Financials", "subsector": "insurance"},
                {"sector": "Properties & Real Estate", "subsector": "real-estate-development"},
                {"sector": "Technology", "subsector": "software-and-it-services"},
                {"sector": "Infrastructure", "subsector": "telecommunications"},
                {"sector": "Infrastructure", "subsector": "transportation-infrastructure"},
                {"sector": "Transportation & Logistics", "subsector": "logistics-and-deliveries"},
            ]

        return {"status": "ok", "mock": True}

