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


class SectorsAPIError(Exception):
    """Raised when Sectors Financial API request fails in online mode."""

    def __init__(self, message: str, status_code: Optional[int] = None):
        super().__init__(message)
        self.status_code = status_code


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
        
        _is_testing = os.environ.get("NISKAVA_TESTING", "0") in ("1", "true", "True")

        if mock_mode is not None:
            if mock_mode and not _is_testing:
                raise RuntimeError(
                    "SECTORS_API_KEY is required. Mock mode is only available in test environments "
                    "(NISKAVA_TESTING=1). Obtain a free key at https://sectors.app"
                )
            self.mock_mode = mock_mode
        else:
            if not _is_testing and not self.api_key:
                raise RuntimeError(
                    "SECTORS_API_KEY is not configured. Niskava requires a valid Sectors "
                    "Financial API key to fetch real IDX market data. "
                    "Run 'niskava setup' or set SECTORS_API_KEY in your environment. "
                    "Obtain a free key at https://sectors.app"
                )
            self.mock_mode = _is_testing and (
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
        force_refresh: bool = False,
    ) -> Any:
        cache_key = self._generate_cache_key(endpoint, params)
        if not force_refresh:
            cached = self._get_cache(cache_key)
            if cached is not None:
                return cached

        # Mode mock HANYA aktif jika eksplisit diset mock_mode=True atau MOCK_SECTORS=1/NISKAVA_OFFLINE=1
        if self.mock_mode:
            mock_data = self._generate_mock_data(endpoint, params)
            self._set_cache(cache_key, endpoint, mock_data, ttl_seconds)
            return mock_data

        if not self.api_key:
            raise SectorsAPIError(
                "SECTORS_API_KEY belum dikonfigurasi. Silakan periksa file .env atau buka menu Pengaturan.",
                status_code=401,
            )

        try:
            url = f"{self.base_url}{endpoint}"
            cfg = RetryConfig(
                max_retries=2,
                initial_delay=0.5,
                max_delay=4.0,
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
        except requests.exceptions.HTTPError as e:
            status = e.response.status_code if e.response is not None else None
            raise SectorsAPIError(f"Sectors API HTTP {status}: {e}", status_code=status) from e
        except requests.exceptions.RequestException as e:
            raise SectorsAPIError(f"Gagal terhubung ke Sectors API: {e}") from e
        except Exception as e:
            raise SectorsAPIError(f"Terjadi kesalahan saat memproses data Sectors API: {e}") from e

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
        self,
        symbol: str,
        start: Optional[str] = None,
        end: Optional[str] = None,
        force_refresh: bool = False,
    ) -> List[Dict[str, Any]]:
        """Retrieve daily OHLCV candlestick data."""
        endpoint = f"/daily/{symbol.upper()}/"
        params = {}
        if start:
            params["start"] = start
        if end:
            params["end"] = end
        # Historical candlestick data is cached; force_refresh bypasses cache
        raw = self._request(endpoint, params, ttl_seconds=None, force_refresh=force_refresh)
        return self._normalize_list_response(raw)

    def get_company_report(
        self, symbol: str, sections: str = "valuation,financials,peers", force_refresh: bool = False
    ) -> Dict[str, Any]:
        """Fetch company fundamental report (cached for 24 hours)."""
        endpoint = f"/company/report/{symbol.upper()}/"
        params = {"sections": sections}
        res = self._request(endpoint, params, ttl_seconds=86400, force_refresh=force_refresh)
        return res if isinstance(res, dict) else {}

    def get_foreign_flow(self, symbol: str, force_refresh: bool = False) -> List[Dict[str, Any]]:
        """Retrieve Foreign Flow Net Inflow data."""
        endpoint = f"/foreign-flow/{symbol.upper()}/"
        raw = self._request(endpoint, ttl_seconds=86400, force_refresh=force_refresh)
        return self._normalize_list_response(raw)

    def get_news(self, symbol: Optional[str] = None, force_refresh: bool = False) -> List[Dict[str, Any]]:
        """Fetch curated financial news."""
        endpoint = "/news/"
        params: Dict[str, Any] = {}
        if symbol:
            clean = symbol.upper()
            params = {"ticker": clean}
        raw = self._request(endpoint, params, ttl_seconds=3600, force_refresh=force_refresh)
        return self._normalize_list_response(raw)

    def get_suspensions(self, symbol: str, force_refresh: bool = False) -> List[Dict[str, Any]]:
        """Fetch exchange suspension and UMA notices with official PDF links."""
        endpoint = "/suspensions/"
        params = {"symbol": symbol.upper()}
        raw = self._request(endpoint, params, ttl_seconds=86400, force_refresh=force_refresh)
        return self._normalize_list_response(raw)

    def get_corporate_actions(self, symbol: str, force_refresh: bool = False) -> List[Dict[str, Any]]:
        """Fetch scheduled corporate actions (dividends, splits, rights issue)."""
        endpoint = f"/company/corporate-actions/{symbol.upper()}/"
        raw = self._request(endpoint, ttl_seconds=86400, force_refresh=force_refresh)
        return self._normalize_list_response(raw)

    def get_filings(self, symbol: str, force_refresh: bool = False) -> List[Dict[str, Any]]:
        """Fetch insider trading and substantial shareholder filings."""
        endpoint = "/filings/"
        params = {"symbol": symbol.upper()}
        raw = self._request(endpoint, params, ttl_seconds=86400, force_refresh=force_refresh)
        return self._normalize_list_response(raw)

    def get_broker_summary(self, symbol: str, force_refresh: bool = False) -> Dict[str, Any]:
        """Fetch top broker accumulation and distribution summary."""
        endpoint = f"/broker-summary/{symbol.upper()}/top/"
        return self._request(endpoint, ttl_seconds=86400, force_refresh=force_refresh)

    def get_subsector_peers(self, subsector: str, force_refresh: bool = False) -> Dict[str, Any]:
        """Fetch industrial subsector peers and valuation benchmarks."""
        endpoint = f"/subsector/report/{subsector.lower()}/"
        return self._request(endpoint, ttl_seconds=604800, force_refresh=force_refresh)

    def get_mining_detail(self, slug: str, force_refresh: bool = False) -> Dict[str, Any]:
        """Fetch operational mining concession and smelter details."""
        endpoint = f"/mining/companies/{slug.lower()}/"
        return self._request(endpoint, ttl_seconds=2592000, force_refresh=force_refresh)

    def get_commodity_price(
        self,
        commodity: str,
        start_year: Optional[int] = None,
        end_year: Optional[int] = None,
        force_refresh: bool = False,
    ) -> List[Dict[str, Any]]:
        """Fetch historical commodity spot benchmark prices (e.g. nickel, coal, gold)."""
        endpoint = f"/mining/commodities/{commodity.lower()}/price/"
        params = {}
        if start_year:
            params["start_year"] = start_year
        if end_year:
            params["end_year"] = end_year
        return self._request(endpoint, params, ttl_seconds=604800, force_refresh=force_refresh)

    def get_quarterly_financials(
        self, symbol: str, report_date: Optional[str] = None, force_refresh: bool = False
    ) -> List[Dict[str, Any]]:
        """Fetch quarterly financial reports and balance sheet line items."""
        endpoint = f"/financials/quarterly/{symbol.upper()}/"
        params = {"report_date": report_date} if report_date else {}
        return self._request(endpoint, params, ttl_seconds=2592000, force_refresh=force_refresh)

    def get_broker_registry(self, force_refresh: bool = False) -> List[Dict[str, Any]]:
        """Fetch IDX broker directory with domicile (foreign/domestic) and cohort (retail/institution)."""
        endpoint = "/broker-registry/"
        return self._request(endpoint, ttl_seconds=2592000, force_refresh=force_refresh)

    def get_subsectors(self, force_refresh: bool = False) -> List[Dict[str, Any]]:
        """Fetch complete list of official IDX sectors and subsectors."""
        endpoint = "/subsectors/"
        return self._request(endpoint, ttl_seconds=2592000, force_refresh=force_refresh)

    def get_top_changes(
        self,
        classification: str = "top_gainers",
        period: str = "1d",
        n_stock: int = 5,
        force_refresh: bool = False,
    ) -> List[Dict[str, Any]]:
        """Fetch extreme market changes (top gainers or top losers) over specified period.

        Sectors API v2 official endpoint: /v2/companies/top-changes/
        Query params: classifications (top_gainers|top_losers), periods (1d|7d|14d|30d|365d), n_stock (1-10)

        Args:
            classification: 'top_gainers' or 'top_losers'.
            period: Duration window ('1d', '7d', '14d', '30d', '365d').
            n_stock: Number of stocks (default 5, max 10).
            force_refresh: Bypass SQLite cache if True.
        """
        endpoint = "/companies/top-changes/"
        clean_cls = "top_losers" if "loser" in classification.lower() else "top_gainers"
        valid_periods = {"1d", "7d", "14d", "30d", "365d"}
        clean_period = period.lower() if period.lower() in valid_periods else "1d"
        clean_n = max(1, min(10, int(n_stock)))

        params = {
            "classifications": clean_cls,
            "periods": clean_period,
            "n_stock": clean_n,
        }
        # 15 minutes TTL for real-time market action
        raw = self._request(endpoint, params, ttl_seconds=900, force_refresh=force_refresh)

        # Unpack nested dict response: {"top_gainers": {"1d": [...]}}
        if isinstance(raw, dict):
            if clean_cls in raw and isinstance(raw[clean_cls], dict):
                items = raw[clean_cls].get(clean_period, [])
                if isinstance(items, list):
                    return self._normalize_list_response(items)
            # Fallback if raw is already a list or directly contains results
            for k in ("results", "data"):
                if k in raw and isinstance(raw[k], list):
                    return self._normalize_list_response(raw[k])

        return self._normalize_list_response(raw)

    def get_most_traded(
        self,
        start: Optional[str] = None,
        end: Optional[str] = None,
        n_stock: int = 5,
        force_refresh: bool = False,
    ) -> List[Dict[str, Any]]:
        """Fetch the most actively traded stocks by volume/turnover on the exchange.

        Sectors API v2 official endpoint: /v2/most-traded/
        Returns either a list of stock objects or a dict keyed by date (e.g. {"2026-09-30": [...]}).

        Args:
            start: Start date YYYY-MM-DD (optional).
            end: End date YYYY-MM-DD (optional).
            n_stock: Number of top stocks to return (default 5).
            force_refresh: Bypass SQLite cache if True.
        """
        endpoint = "/most-traded/"
        clean_n = max(1, int(n_stock))
        params: Dict[str, Any] = {"n_stock": clean_n}
        if start:
            params["start"] = start
        if end:
            params["end"] = end
        # 1 hour TTL
        raw = self._request(endpoint, params, ttl_seconds=3600, force_refresh=force_refresh)

        # If response is a dict keyed by date (e.g. {"2026-09-30": [...]}), extract the latest date list
        if isinstance(raw, dict):
            dates = sorted(raw.keys(), reverse=True)
            for d in dates:
                if isinstance(raw[d], list) and raw[d]:
                    return self._normalize_list_response(raw[d])

        return self._normalize_list_response(raw)

    def _generate_mock_data(self, endpoint: str, params: Optional[Dict[str, Any]] = None) -> Any:
        """Generate realistic mock data fixtures for offline development and CI tests."""
        raw_sym = params.get("symbol") if params else None
        if not raw_sym and params and "ticker" in params:
            raw_sym = params.get("ticker")
        clean_sym = str(raw_sym).upper().strip() if raw_sym else None
        if not clean_sym:
            for prefix in (
                "/daily/",
                "/company/report/",
                "/foreign-flow/",
                "/financials/quarterly/",
                "/quarterly-financials/",
                "/broker-summary/",
                "/company/corporate-actions/",
                "/corporate-actions/",
            ):
                if prefix in endpoint:
                    parts = endpoint.split(prefix)[-1].strip("/").split("/")
                    if parts and parts[0]:
                        clean_sym = parts[0].upper()
                        break
        symbol = clean_sym or "ANTM"
        if "/daily/" in endpoint:
            # 30 days of synthetic candles with realistic profile per ticker
            candles = []
            base_date = datetime.now() - timedelta(days=35)
            
            ticker_profiles = {
                "ANTM": {"price": 1500.0, "norm_vol": 20_000_000.0, "spike_vol": 125_000_000.0, "spike_day": 25},
                "BBRI": {"price": 4980.0, "norm_vol": 85_000_000.0, "spike_vol": 245_000_000.0, "spike_day": 26},
                "BBCA": {"price": 10150.0, "norm_vol": 60_000_000.0, "spike_vol": 180_000_000.0, "spike_day": 24},
                "BUMI": {"price": 142.0, "norm_vol": 1_800_000_000.0, "spike_vol": 6_200_000_000.0, "spike_day": 22},
                "GOTO": {"price": 62.0, "norm_vol": 750_000_000.0, "spike_vol": 2_400_000_000.0, "spike_day": 25},
                "TLKM": {"price": 2950.0, "norm_vol": 45_000_000.0, "spike_vol": 140_000_000.0, "spike_day": 23},
                "ASII": {"price": 5125.0, "norm_vol": 28_000_000.0, "spike_vol": 88_000_000.0, "spike_day": 25},
            }
            prof = ticker_profiles.get(symbol)
            if not prof:
                seed = sum(ord(c) for c in symbol)
                price = float(500 + (seed % 35) * 100)
                norm_vol = float(15_000_000 + (seed % 20) * 2_000_000)
                spike_vol = norm_vol * (3.5 + (seed % 5) * 0.5)
                spike_day = 20 + (seed % 6)
            else:
                price = prof["price"]
                norm_vol = prof["norm_vol"]
                spike_vol = prof["spike_vol"]
                spike_day = prof["spike_day"]

            for day_idx in range(30):
                curr_date = (base_date + timedelta(days=day_idx)).strftime("%Y-%m-%d")
                if day_idx == spike_day:
                    volume = spike_vol
                    close = price * 1.082
                else:
                    volume = norm_vol + (day_idx % 5) * (norm_vol * 0.08)
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
            subsector_map = {
                "ANTM": ("Basic Materials", "Metals & Minerals", "PT Aneka Tambang Tbk"),
                "GOTO": ("Technology", "Software & IT Services", "PT GoTo Gojek Tokopedia Tbk"),
                "BUKA": ("Technology", "Software & IT Services", "PT Bukalapak.com Tbk"),
                "BBCA": ("Financials", "Banks", "PT Bank Central Asia Tbk"),
                "BBRI": ("Financials", "Banks", "PT Bank Rakyat Indonesia Tbk"),
                "BMRI": ("Financials", "Banks", "PT Bank Mandiri Tbk"),
                "BBNI": ("Financials", "Banks", "PT Bank Negara Indonesia Tbk"),
                "BUMI": ("Energy", "Oil, Gas & Coal", "PT Bumi Resources Tbk"),
                "TLKM": ("Telecommunication", "Telecommunication", "PT Telkom Indonesia Tbk"),
                "ASII": ("Industrials", "Automobiles & Components", "PT Astra International Tbk"),
            }
            sec, sub, name = subsector_map.get(symbol, ("Basic Materials", "Metals & Minerals", f"PT {symbol} Tbk"))
            return {
                "symbol": symbol,
                "company_name": name,
                "sector": sec,
                "sub_sector": sub,
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

        if "/broker-summary" in endpoint:
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

        if "/subsector/report/" in endpoint or ("/subsector/" in endpoint and "/subsectors" not in endpoint):
            return {
                "subsector": "metals-and-minerals-mining",
                "peer_count": 14,
                "median_pe": 16.8,
                "median_pb": 1.95,
                "peers": ["ANTM", "TINS", "INCO", "MBMA"],
            }

        if "/mining/companies/" in endpoint or "/mining-company-detail/" in endpoint:
            return {
                "slug": "aneka-tambang",
                "commodity": "NICKEL",
                "smelter_count": 3,
                "concession_area_ha": 45000,
                "operational_status": "ACTIVE",
            }

        if "/mining/commodities/" in endpoint or "/commodity-price/" in endpoint:
            # 30 daily/monthly benchmark spot prices
            base_date = datetime.now() - timedelta(days=35)
            prices = []
            curr_val = 16500.0  # e.g. USD/ton for nickel
            for i in range(30):
                d_str = (base_date + timedelta(days=i)).strftime("%Y-%m-%d")
                curr_val *= 1.0 + ((i % 4) - 1.5) * 0.008
                prices.append({"date": d_str, "price": round(curr_val, 2)})
            return prices

        if "/financials/quarterly/" in endpoint or "/quarterly-financials/" in endpoint:
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

        if "/top-changes/" in endpoint or "/companies/top-changes/" in endpoint:
            cls_param = (params or {}).get("classifications") or (params or {}).get("classification") or "top_gainers"
            clean_cls = "top_losers" if "loser" in str(cls_param).lower() else "top_gainers"
            period_param = (params or {}).get("periods") or (params or {}).get("period") or "1d"
            clean_period = str(period_param).lower()

            losers_1d = [
                {"name": "Fortune Indonesia Tbk", "symbol": "FORU.JK", "price_change": -0.145, "last_close_price": 224, "latest_close_date": "2026-09-30"},
                {"name": "PT GoTo Gojek Tokopedia Tbk", "symbol": "GOTO.JK", "price_change": -0.135, "last_close_price": 32, "latest_close_date": "2026-09-30"},
                {"name": "PT Transcoal Pacific Tbk", "symbol": "TCPI.JK", "price_change": -0.101, "last_close_price": 1550, "latest_close_date": "2026-09-30"},
                {"name": "PT MNC Digital Entertainment Tbk", "symbol": "MSIN.JK", "price_change": -0.097, "last_close_price": 186, "latest_close_date": "2026-09-30"},
                {"name": "PT Bank KB Indonesia Tbk", "symbol": "BBKP.JK", "price_change": -0.071, "last_close_price": 39, "latest_close_date": "2026-09-30"},
            ]
            gainers_1d = [
                {"name": "Bank of India Indonesia Tbk", "symbol": "BSWD.JK", "price_change": 0.246, "last_close_price": 2880, "latest_close_date": "2026-09-30"},
                {"name": "PT Sinar Mas Agro Resources and Technology Tbk", "symbol": "SMAR.JK", "price_change": 0.12, "last_close_price": 7700, "latest_close_date": "2026-09-30"},
                {"name": "Metropolitan Land Tbk", "symbol": "MTLA.JK", "price_change": 0.104, "last_close_price": 740, "latest_close_date": "2026-09-30"},
                {"name": "United Tractors Tbk", "symbol": "UNTR.JK", "price_change": 0.075, "last_close_price": 27075, "latest_close_date": "2026-09-30"},
                {"name": "PT Jhonlin Agro Raya Tbk", "symbol": "JARR.JK", "price_change": 0.068, "last_close_price": 3580, "latest_close_date": "2026-09-30"},
            ]

            return {
                "top_gainers": {clean_period: gainers_1d},
                "top_losers": {clean_period: losers_1d},
            }

        if "/most-traded/" in endpoint:
            return [
                {"symbol": "BBRI", "company_name": "Bank Rakyat Indonesia Tbk", "volume": 1_450_000_000, "turnover": 725_000_000_000, "price": 5000},
                {"symbol": "BBCA", "company_name": "Bank Central Asia Tbk", "volume": 890_000_000, "turnover": 910_000_000_000, "price": 10225},
                {"symbol": "BMRI", "company_name": "Bank Mandiri Tbk", "volume": 680_000_000, "turnover": 490_000_000_000, "price": 7200},
                {"symbol": "ANTM", "company_name": "Aneka Tambang Tbk", "volume": 450_000_000, "turnover": 724_500_000_000, "price": 1610},
                {"symbol": "ASII", "company_name": "Astra International Tbk", "volume": 310_000_000, "turnover": 155_000_000_000, "price": 5000},
            ]

        return {"status": "ok", "mock": True}

