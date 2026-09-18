# 03 — Spesifikasi Lengkap Integrasi Sectors Financial API v2

**Status:** ACCEPTED  
**Versi Dokumen:** `2.0.0` (Complete Reference)  
**Terakhir Diperbarui:** 2026-09-16  
**Target Versi API:** `v2`  
**Base URL:** `https://api.sectors.app/v2`  
**Keputusan Arsitektur Terkait:** [`03-sectors-v2-and-credit-conservation.md`](../90-decisions/03-sectors-v2-and-credit-conservation.md), [`07-resilient-dual-engine-osint-architecture.md`](../90-decisions/07-resilient-dual-engine-osint-architecture.md)

---

## 1. Bukti Ketergantungan Inti (*Core Source Proof*)

Sesuai aturan resmi Sectors Hackathon 2026 (Official Rules Section 06 & Track 1 Requirements):  
> *"Projects must use Sectors MCP or the Sectors REST API as a core data source, not as a single decorative call. The product should lose its core functionality if Sectors data is removed."*

**Niskava Agent secara fungsional lumpuh total tanpa data Sectors API:**
1. **Deteksi Anomali Kuantitatif Deterministik (Stage 3):** Mesin matematika NumPy membutuhkan time series harga dan volume 30–90 hari dari `/v2/daily/{symbol}/` dan deret aliran modal asing `/v2/foreign-flow/{symbol}/` untuk menghitung $Z$-score volume ($V_z$), abnormal return ($R_t$), dan abnormal foreign flow ($F_z$).
2. **Audit Keterbukaan & Aksi Korporasi (Stage 2 & 5):** Sistem memanfaatkan `/v2/suspensions/` untuk melacak riwayat suspensi bursa beserta tautan PDF resmi BEI, `/v2/filings/` untuk transaksi orang dalam (*insider trading*), dan `/v2/corporate-actions/{symbol}/` untuk konfirmasi dividen/RUPS.
3. **Pilar Verifikasi Kausalitas (Stage 6):** Status bukti `SUPPORTED` mensyaratkan konfirmasi ganda antara sinyal eksternal dengan metrik fundamental dan rasio keuangan resmi dari `/v2/company/report/{symbol}/`.

---

## 2. Katalog Lengkap Endpoint v2 Berdasarkan Domain Fungsional

Sistem mengintegrasikan 7 domain endpoint resmi API v2 Indonesia:

```
┌────────────────────────────────────────────────────────────────────────┐
│                   SECTORS FINANCIAL API v2 CATALOG                     │
├────────────────────────────────────────────────────────────────────────┤
│ 1. Transaksi & Harga      │ /v2/daily/, /v2/daily-close/, /v2/top-chg/ │
│ 2. Fundamental Emiten     │ /v2/company/report/, /v2/quarterly-fin/   │
│ 3. Regulasi & Disclosures │ /v2/suspensions/, /v2/filings/, /v2/corp/  │
│ 4. Flow & Bandarmology    │ /v2/foreign-flow/, /v2/broker-summary-top/ │
│ 5. Ekstensi Komoditas/Mine│ /v2/mining-companies/, /v2/commodity-price/│
│ 6. Mesin Screener         │ /v2/companies/ (SQL-where & NL-query)      │
│ 7. Taksonomi & Helpers    │ /v2/subsectors/, /v2/industries/, /v2/tags/│
└────────────────────────────────────────────────────────────────────────┘
```

### Domain 1: Transaksi & Aksi Harga Pasar

| Endpoint Path | Parameter Kunci | Peran dalam Pipeline | Kebijakan Caching SQLite |
|---|---|---|---|
| `GET /v2/daily/{symbol}/` | `symbol` (e.g. `ANTM`), `start`, `end` | Menarik deret OHLCV harian s.d 90 hari untuk baseline $V_z$ & $R_t$. | **Permanen** untuk $T < \text{hari ini}$. |
| `GET /v2/daily-close/` | `date` (format `YYYY-MM-DD`) | Menarik harga penutupan seluruh emiten IDX dalam satu panggilan. | **Permanen** (data lampau). |
| `GET /v2/most-traded/` | `start`, `end`, `n_stock` (default 5) | Menemukan saham paling aktif diperdagangkan untuk *market screener*. | **24 Jam** (TTL). |
| `GET /v2/top-changes/` | `classification` (`top_gainers` / `top_losers`), `period` (`1d`,`7d`,`14d`,`30d`,`365d`) | Mendeteksi kandidat anomali lonjakan atau kejatuhan harga ekstrim. | **15 Menit** (hari bursa aktif). |
| `GET /v2/idx-total/` | `start`, `end` (s.d 90 hari) | Menghitung kapitalisasi pasar agregat IHSG sebagai pembanding makro. | **Permanen** (data lampau). |
| `GET /v2/index-daily/{index}/` | `index` (e.g. `LQ45`, `IDX30`, `KOMPAS100`) | Menghitung return benchmark indeks untuk kalkulasi *Abnormal Return* ($R_t$). | **Permanen** (data lampau). |

### Domain 2: Profil Fundamental & Kesehatan Finansial

| Endpoint Path | Parameter Kunci | Peran dalam Pipeline | Kebijakan Caching SQLite |
|---|---|---|---|
| `GET /v2/company/report/{symbol}/` | `symbol`, `sections` (`valuation`,`financials`,`peers`,`overview`) | Evaluasi fundamental, rasio profitabilitas, DER, dan PBV band. | **7 Hari** (TTL). |
| `GET /v2/quarterly-financials/{symbol}/` | `symbol`, `report_date` | Laporan keuangan triwulanan (khusus bank mencakup NII & loan deposit). | **30 Hari** (TTL). |
| `GET /v2/company-segments/{symbol}/` | `symbol`, `year` | Breakdown pendapatan dan beban per segmen usaha (format Sankey graph). | **90 Hari** (TTL). |
| `GET /v2/shareholders/{symbol}/` | `symbol` | Komposisi kepemilikan saham (pengendali, institusi, publik/ritel). | **14 Hari** (TTL). |
| `GET /v2/subsector/{subsector}/` | `subsector` (kebab-case slug) | Metrik rata-rata industri untuk mengukur *Sector Divergence* ($D_t$). | **7 Hari** (TTL). |

### Domain 3: Regulasi, Suspensi & Keterbukaan Informasi (Tier 1 Evidence)

| Endpoint Path | Parameter Kunci | Peran dalam Pipeline | Bobot Bukti |
|---|---|---|:---:|
| `GET /v2/suspensions/` | `symbol`, `start`, `end` | Mengambil catatan suspensi resmi BEI, alasan suspensi, dan **tautan langsung ke surat pengumuman PDF resmi bursa**. | **Tier 1 (1.00)** |
| `GET /v2/filings/` | `symbol`, `start`, `end`, `transaction_type` | Melacak transaksi insider trading (direksi, komisaris, pemegang saham $\ge 5\%$) saat anomali volume terjadi. | **Tier 1 (1.00)** |
| `GET /v2/corporate-actions/{symbol}/` | `symbol`, `action_type` (`dividend`,`split`,`right`) | Memverifikasi apakah anomali volume bertepatan dengan cum-date dividen, pemecahan saham, atau rights issue. | **Tier 1 (1.00)** |
| `GET /v2/news/?ticker={symbol}` | `ticker`, `extension` (`idx` / `mining`) | Menarik arsip berita finansial bursa terkurasi untuk memvalidasi kesenjangan informasi (*Evidence Gap*). | **Tier 2 (0.85)** |

### Domain 4: Arus Modal & Bandarmology (Capital Flow)

| Endpoint Path | Parameter Kunci | Peran dalam Pipeline | Kebijakan Caching SQLite |
|---|---|---|---|
| `GET /v2/foreign-flow/{symbol}/` | `symbol`, `start`, `end` (s.d 90 hari) | Deret harian Net Foreign Inflow (IDR) untuk menghitung anomali akumulasi asing ($F_z$). | **Permanen** untuk $T < \text{hari ini}$. |
| `GET /v2/broker-summary/{symbol}/` | `symbol`, `start`, `end` (s.d 14 hari) | Rincian transaksi harian seluruh broker per saham (lots, frekuensi, VWAP). | **Permanen** untuk $T < \text{hari ini}$. |
| `GET /v2/broker-summary-top/{symbol}/` | `symbol`, `start`, `end` | Peringkat broker akumulasi teratas (*top buyers*) dan distribusi (*top sellers*). | **Permanen** untuk $T < \text{hari ini}$. |
| `GET /v2/broker-registry/` | Tanpa parameter | Direktori resmi kode broker BEI dengan asal (*foreign/domestic*) dan kohort (*retail/institutional*). | **30 Hari** (TTL). |
| `GET /v2/top-brokers/` | `date`, `sort_by` (`gross` / `net`) | Peringkat broker paling aktif harian di seluruh bursa. | **Permanen** (data lampau). |

### Domain 5: Ekstensi Industri Pertambangan & Komoditas (Mining Extension)

| Endpoint Path | Parameter Kunci | Peran dalam Pipeline | Kebijakan Caching SQLite |
|---|---|---|---|
| `GET /v2/mining-companies/` | `commodity_type` (`nickel`,`gold`,`coal`) | Menemukan daftar emiten produsen komoditas terkait. | **30 Hari** (TTL). |
| `GET /v2/mining-company-detail/{slug}/` | `slug` (e.g. `aneka-tambang`) | Data operasional tambang, izin konsesi (IUP), dan jumlah lokasi tambang. | **30 Hari** (TTL). |
| `GET /v2/commodity-price/{commodity}/` | `commodity` (`nickel`,`gold`,`coal`), `start_year`, `end_year` | Harga historis komoditas bulanan/dwi-mingguan untuk korelasi pemicu harga. | **7 Hari** (TTL). |
| `GET /v2/mining-sites/` | `location`, `commodity_type` | Pemetaan fasilitas smelter dan lokasi operasional pertambangan emiten. | **30 Hari** (TTL). |

### Domain 6: Mesin Penyaring Saham (Screener)

| Endpoint Path | Parameter Kunci | Peran dalam Pipeline |
|---|---|---|
| `GET /v2/companies/` | `where` (SQL expression), `order_by`, `q` (Natural Language), `page` | Screening terarah untuk inisiasi penemuan saham yang memenuhi parameter anomali. |
| `GET /v2/screener/free-float/` | `sector`, `sub_sector`, `min_free_float` | Mengidentifikasi persentase saham publik beredar untuk mengukur risiko likuiditas. |

### Domain 7: Taksonomi & Berkas Pembantu (Helper Lists)

| Endpoint Path | Deskripsi |
|---|---|
| `GET /v2/subsectors/` | Mengambil seluruh pasangan sektor & subsektor resmi bursa dalam format *kebab-case*. |
| `GET /v2/industries/` & `GET /v2/subindustries/` | Daftar hierarki industri dan sub-industri resmi BEI. |
| `GET /v2/tags/` | Daftar tag resmi untuk memfilter artikel berita dan arsip pelaporan (*filings*). |
| `GET /v2/latest-quarterly-dates/` | Tanggal laporan keuangan triwulanan terbaru seluruh semesta emiten bursa. |

---

## 3. Skema Klien Python Terpadu (Production-Ready Client)

Klien Python diimplementasikan menggunakan `httpx` dengan dukungan koneksi HTTP/2, *exponential backoff*, *rate limiting*, dan *local caching* ke database SQLite:

```python
"""
niskava/sectors/client.py — Production-Grade Sectors Financial API v2 Client
"""
import os
import time
import hashlib
import json
import sqlite3
from typing import Any, Dict, List, Optional
import httpx

class SectorsAPIClient:
    BASE_URL = "https://api.sectors.app/v2"

    def __init__(self, db_path: str, api_key: Optional[str] = None):
        self.api_key = api_key or os.environ.get("SECTORS_API_KEY", "")
        self.db_path = db_path
        self.mock_mode = os.environ.get("MOCK_SECTORS", "0") == "1"
        self.client = httpx.Client(
            http2=True,
            timeout=10.0,
            headers={
                "Authorization": self.api_key,
                "User-Agent": "Niskava-Agent/2.0.0 (Track1-AI-Agents)"
            }
        )

    def _get_cache(self, cache_key: str) -> Optional[Any]:
        with sqlite3.connect(self.db_path) as conn:
            cursor = conn.cursor()
            cursor.execute(
                "SELECT response_json FROM sectors_cache WHERE cache_key = ? AND (expires_at IS NULL OR expires_at > datetime('now'))",
                (cache_key,)
            )
            row = cursor.fetchone()
            if row:
                return json.loads(row[0])
        return None

    def _set_cache(self, cache_key: str, endpoint: str, params_str: str, data: Any, ttl_seconds: Optional[int] = None):
        expires_at = f"datetime('now', '+{ttl_seconds} seconds')" if ttl_seconds else "NULL"
        payload_json = json.dumps(data)
        with sqlite3.connect(self.db_path) as conn:
            cursor = conn.cursor()
            query = f"""
                INSERT OR REPLACE INTO sectors_cache (cache_key, endpoint, params_hash, response_json, expires_at, created_at)
                VALUES (?, ?, ?, ?, {expires_at}, datetime('now'))
            """
            cursor.execute(query, (cache_key, endpoint, params_str, payload_json))
            conn.commit()

    def request(self, endpoint: str, params: Optional[Dict[str, Any]] = None, ttl_seconds: Optional[int] = None) -> Any:
        if self.mock_mode:
            return self._load_fixture(endpoint, params)

        params = params or {}
        params_str = json.dumps(params, sort_keys=True)
        cache_key = hashlib.sha256(f"{endpoint}:{params_str}".encode()).hexdigest()

        cached_data = self._get_cache(cache_key)
        if cached_data is not None:
            return cached_data

        url = f"{self.BASE_URL}{endpoint}"
        retries = 3
        backoff = 1.0

        for attempt in range(retries):
            try:
                resp = self.client.get(url, params=params)
                if resp.status_code == 429:
                    time.sleep(backoff)
                    backoff *= 2.0
                    continue
                resp.raise_for_status()
                data = resp.json()
                self._set_cache(cache_key, endpoint, params_str, data, ttl_seconds)
                return data
            except httpx.HTTPStatusError as e:
                if attempt == retries - 1:
                    raise e
                time.sleep(backoff)
                backoff *= 2.0
        raise RuntimeError(f"Gagal mengambil data dari Sectors API setelah {retries} percobaan: {endpoint}")

    def _load_fixture(self, endpoint: str, params: Optional[Dict[str, Any]]) -> Any:
        sanitized = endpoint.strip("/").replace("/", "_")
        fixture_path = os.path.join("tests", "fixtures", "sectors", f"{sanitized}.json")
        if os.path.exists(fixture_path):
            with open(fixture_path, "r") as f:
                return json.load(f)
        return {"mock": True, "endpoint": endpoint, "params": params}
```

---

## 4. Cuplikan Struktur JSON Endpoint Strategis

### A. Catatan Suspensi Bursa: `GET /v2/suspensions/?symbol=ANTM`
```json
[
  {
    "symbol": "ANTM",
    "company_name": "PT Aneka Tambang Tbk",
    "suspension_date": "2024-01-18",
    "unsuspension_date": "2024-01-19",
    "session": "Sesi I",
    "market_type": "Pasar Reguler dan Pasar Tunai",
    "reason": "Sehubungan dengan terjadinya peningkatan harga kumulatif yang signifikan pada Saham PT Aneka Tambang Tbk (ANTM), dalam rangka cooling down PT Bursa Efek Indonesia memandang perlu untuk melakukan penghentian sementara perdagangan.",
    "pdf_url": "https://www.idx.co.id/StaticData/NewsAndAnnouncement/ANNOUNCEMENTSTOCK/From_EREP/202401/Peng-SPT-00012_BEI.WAS_01-2024.pdf"
  }
]
```

### B. Transaksi Orang Dalam (Insider Filings): `GET /v2/filings/?symbol=ANTM`
```json
[
  {
    "symbol": "ANTM",
    "holder_name": "Nico Kanter",
    "holder_type": "DIRECTOR",
    "transaction_type": "BUY",
    "transaction_date": "2026-09-11",
    "shares_transacted": 250000,
    "price_per_share": 1455.0,
    "shares_after_transaction": 1250000,
    "percentage_after_transaction": 0.0052,
    "purpose": "Investasi Pribadi",
    "announcement_date": "2026-09-12"
  }
]
```

### C. Aliran Modal Asing Harian: `GET /v2/foreign-flow/ANTM/?start=2026-09-10&end=2026-09-13`
```json
[
  {
    "date": "2026-09-12",
    "foreign_buy_value": 142500000000,
    "foreign_sell_value": 31200000000,
    "net_foreign_inflow": 111300000000,
    "close_price": 1580,
    "symbol": "ANTM"
  },
  {
    "date": "2026-09-11",
    "foreign_buy_value": 24000000000,
    "foreign_sell_value": 26500000000,
    "net_foreign_inflow": -2500000000,
    "close_price": 1460,
    "symbol": "ANTM"
  }
]
```

### D. Jadwal Aksi Korporasi: `GET /v2/corporate-actions/ANTM/`
```json
[
  {
    "symbol": "ANTM",
    "action_type": "DIVIDEND",
    "cum_date": "2026-06-05",
    "ex_date": "2026-06-06",
    "recording_date": "2026-06-09",
    "payment_date": "2026-06-25",
    "dividend_per_share": 128.5,
    "currency": "IDR",
    "financial_year": "2025"
  }
]
```

---

## 5. Rincian Konservasi Kuota (Anggaran 1.000 Kredit)

Sistem menetapkan alokasi anggaran kredit yang ketat untuk menjamin ketersediaan kredit selama evaluasi juri:

| Aktivitas Sistem | Endpoint yang Dipanggil | Biaya Kredit Aktual (Cache Miss) | Biaya Kredit Berulang (Cache Hit) |
|---|---|:---:|:---:|
| **Investigasi Saham Tunggal** (`ANTM`) | `/v2/daily/ANTM/`<br>`/v2/company/report/ANTM/?sections=valuation,financials`<br>`/v2/foreign-flow/ANTM/`<br>`/v2/suspensions/?symbol=ANTM`<br>`/v2/news/?ticker=ANTM` | **5 Kredit** | **0 Kredit** |
| **Pengecekan Sektor Industri** | `/v2/subsector/metals-mining/` | **1 Kredit** | **0 Kredit** (TTL 7 hari) |
| **Verifikasi Insider Trading** | `/v2/filings/?symbol=ANTM` | **1 Kredit** | **0 Kredit** (TTL 24 jam) |

### Alokasi Kuota 1.000 Kredit:
* **Fase Development & Testing (19 Agu – 25 Sep):** Maksimal 100 kredit (selebihnya wajib menggunakan `MOCK_SECTORS=1`).
* **Perekaman Video Demo & Teaser (26 – 28 Sep):** Maksimal 50 kredit.
* **Cadangan Evaluasi Langsung Dewan Juri (1 – 8 Okt):** **Minimal 850 kredit utuh**.
