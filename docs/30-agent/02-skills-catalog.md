# 02 — Katalog Metodologi Analisa Agen (Standard Agent Skills Catalog)

**Status:** ACCEPTED  
**Versi Dokumen:** `2.0.0` (Standardized Agent Skills Architecture)  
**Terakhir Diperbarui:** 2026-09-19  
**Dokumen Terkait:** [`01-investigation-pipeline.md`](01-investigation-pipeline.md), [`03-evidence-and-causality.md`](03-evidence-and-causality.md), [`../90-decisions/08-modular-skills-and-mcp-architecture.md`](../90-decisions/08-modular-skills-and-mcp-architecture.md)

Dalam Niskava Agent, **Skill** adalah modul Standar Operasional Prosedur (SOP) analisis terstruktur yang dijalankan oleh agen secara deterministik dan orkestrasi ReAct. Berbeda dari AI wrapper biasa yang mencampurkan prompt, data bursa, dan kalkulasi ke dalam satu fungsi kaku, Niskava memperlakukan Skill sebagai **unit kapabilitas analisis yang terisolasi, dapat diuji secara mandiri (*composable & testable*), dan terikat kontrak antarmuka yang ketat**.

---

## 1. Spesifikasi Standar Antarmuka Skill (Agent Skill Specification)

Setiap Skill di Niskava Agent wajib mematuhi anatomi standar berikut:

```yaml
skill_id: string                 # Pengenal unik (kebab-case, e.g. market-anomaly-recon)
name: string                     # Nama resmi skill
version: string                  # Semantic versioning (e.g. 1.2.0)
description: string              # Deskripsi kapabilitas untuk routing oleh ReAct Agent
trigger_conditions: list[string] # Kondisi pemicu aktivasi skill
prerequisites:                   # Ketergantungan primitive tools
  mcp_tools: list[string]        # Tool MCP Sectors / OSINT yang wajib tersedia
  data_requirements: list[string]# Minimum rentang data yang dibutuhkan
deterministic_gate:              # Firewall matematika NumPy (Law 1)
  metrics: list[string]          # Formula metrik yang dihitung deterministik
  thresholds: dict               # Ambang batas anomali numerik
execution_protocol: list[step]   # Urutan langkah investigasi terstruktur (SOP)
output_schema: dict              # Struktur payload temuan bukti (JSON Schema)
verification_mapping: dict       # Pemetaan ke Three-Tier Verification Taxonomy
```

---

## 2. Katalog 6 Domain Skills Resmi Niskava

```
┌─────────────────────────────────────────────────────────────────────────┐
│                       NISKAVA SKILLS REGISTRY                           │
├────────────────────────────────────┬────────────────────────────────────┤
│ 1. market-anomaly-recon            │ 4. financial-health-stress-test    │
│    (Deteksi Anomali Kuantitatif)   │    (Audit Neraca & Solvabilitas)   │
├────────────────────────────────────┼────────────────────────────────────┤
│ 2. event-causality-audit           │ 5. mining-commodity-divergence     │
│    (Korelasi Berita & Disclosures) │    (Korelasi Harga Komoditas Global│
├────────────────────────────────────┼────────────────────────────────────┤
│ 3. insider-bandarmology-forensic   │ 6. peer-valuation-benchmark        │
│    (Broker Summary & Orang Dalam)  │    (Valuasi Relatif Subsektor)     │
└────────────────────────────────────┴────────────────────────────────────┘
```

---

### Skill 1: `market-anomaly-recon` (Deteksi Anomali Pasar & Divergensi)

* **Skill ID**: `market-anomaly-recon`
* **Trigger**: Inisiasi investigasi ticker, pertanyaan pengguna mengenai lonjakan harga/volume tidak wajar, atau screening harian.
* **Prerequisites (MCP Tools)**:
  * `sectors_get_daily_candles` (`/v2/daily/{symbol}/`)
  * `sectors_get_foreign_flow` (`/v2/foreign-flow/{symbol}/`)
  * `sectors_get_subsector_peers` (`/v2/subsector/{subsector}/`)
* **Deterministic Compute Gate (NumPy Firewall)**:
  $$\mu_{20} = \frac{1}{20}\sum_{i=1}^{20} V_{t-i}, \quad \sigma_{20} = \sqrt{\frac{1}{20}\sum_{i=1}^{20}(V_{t-i} - \mu_{20})^2}$$
  $$V_z = \frac{V_t - \mu_{20}}{\sigma_{20}}, \quad R_t = \frac{P_t - P_{t-1}}{P_{t-1}}, \quad F_z = \frac{F_t - \mu_{F,20}}{\sigma_{F,20}}$$
  * **Threshold Anomali**: $(V_z \ge 2.5) \lor (|R_t| \ge 5.0\%) \lor (|F_z| \ge 2.5)$.
* **Execution Protocol (SOP)**:
  1. Tarik time series 30–90 hari OHLCV dan net foreign inflow via Sectors MCP (cek cache SQLite lokal).
  2. Eksekusi NumPy Anomaly Gate untuk menghitung $V_z$, $R_t$, dan $F_z$ deterministik.
  3. Hitung divergensi terhadap rata-rata subsektor ($D_t = R_{t,\text{stock}} - R_{t,\text{subsector}}$).
  4. Jika anomali terkonfirmasi, isolasi tanggal kejadian $T_{\text{anomaly}}$ dan kirimkan ke `event-causality-audit`.
* **Output Schema**:
  ```json
  {
    "anomaly_detected": true,
    "anomaly_date": "2026-09-12",
    "metrics": {
      "volume": 184500000,
      "volume_ma20": 48200000,
      "volume_z_score": 3.84,
      "price_change_pct": 8.25,
      "foreign_flow_net_idr": 111300000000,
      "foreign_flow_z_score": 2.91,
      "sector_divergence_pct": 7.80
    },
    "verification_status": "SUPPORTED",
    "confidence_score": 1.00
  }
  ```

---

### Skill 2: `event-causality-audit` (Korelasi Kausalitas Berita & Keterbukaan)

* **Skill ID**: `event-causality-audit`
* **Trigger**: Adanya $T_{\text{anomaly}}$ dari `market-anomaly-recon` atau pertanyaan mengenai katalis penggerak harga.
* **Prerequisites (MCP Tools)**:
  * `sectors_get_suspensions` (`/v2/suspensions/`)
  * `sectors_get_corporate_actions` (`/v2/corporate-actions/{symbol}/`)
  * `osint_harvest_dual_engine` (Sectors `/v2/news/` + Google News RSS)
* **Deterministic Gate**:
  * Jendela waktu pencarian wajib dibatasi ketat: $[T_{\text{anomaly}} - 2\text{ hari}, T_{\text{anomaly}} + 1\text{ hari}]$.
  * Pencarian di luar jendela waktu ditolak untuk mencegah *temporal causality inversion*.
* **Execution Protocol (SOP)**:
  1. Periksa catatan suspensi resmi dan surat pengumuman BEI via `/v2/suspensions/`. Ambil tautan dokumen PDF resmi.
  2. Periksa jadwal aksi korporasi (RUPS, cum-date dividen, rights issue).
  3. Jalankan pencarian bertarget via Dual-Engine OSINT, ekstraksi teks berita via `trafilatura`, dan isolasi kutipan di tag `<evidence_context>`.
  4. Lakukan evaluasi urutan stempel waktu (*temporal precedence*):
     * Waktu rilis berita mendahului lonjakan volume $\to$ `LIKELY_CATALYST`.
     * Lonjakan volume mendahului rilis berita $\to$ `PRECEDED_ANNOUNCEMENT` (dugaan kebocoran informasi).
     * Tidak ada berita atau dokumen bursa dalam jendela waktu $\to$ `UNEXPLAINED_BY_NEWS`.
* **Output Schema**:
  ```json
  {
    "catalyst_identified": true,
    "causality_label": "LIKELY_CATALYST",
    "events": [
      {
        "headline": "ANTM Resmikan Unit Pemurnian Feronikel Halmahera Timur",
        "source": "Keterbukaan Informasi BEI (IDXnet)",
        "source_url": "https://www.idx.co.id/filings/ANTM-20260912.pdf",
        "published_at": "2026-09-12T08:15:00+07:00",
        "verification_status": "SUPPORTED",
        "confidence_score": 0.95
      }
    ]
  }
  ```

---

### Skill 3: `insider-bandarmology-forensic` (Forensik Broker & Insider Trading)

* **Skill ID**: `insider-bandarmology-forensic`
* **Trigger**: Lonjakan volume mendahului berita (`PRECEDED_ANNOUNCEMENT`), anomali $F_z$ asing ekstrim, atau pertanyaan mengenai siapa pelaku akumulasi.
* **Prerequisites (MCP Tools)**:
  * `sectors_get_filings` (`/v2/filings/?symbol={symbol}`)
  * `sectors_get_broker_summary_top` (`/v2/broker-summary-top/{symbol}/`)
  * `sectors_get_broker_registry` (`/v2/broker-registry/`)
* **Deterministic Gate**:
  * Menghitung rasio konsentrasi pembeli teratas (*Top 3 Buyer Concentration Ratio*):
    $$C_3 = \frac{\sum_{j=1}^3 \text{Volume Buyer}_j}{\text{Total Volume Market}}$$
  * Ambang batas akumulasi institusional: $C_3 \ge 65.0\%$.
* **Execution Protocol (SOP)**:
  1. Tarik riwayat pelaporan kepemilikan orang dalam (direksi, komisaris, PSP) dari `/v2/filings/`.
  2. Ambil 3 broker pembeli bersih (*top buyers*) dan 3 broker penjual bersih (*top sellers*).
  3. Cocokkan kode broker dengan direktori `/v2/broker-registry/` untuk memetakan asal domisili (asing vs domestik) dan tipe kohort (institusi vs ritel).
  4. Jika transaksi insider atau konsentrasi $C_3 \ge 65\%$ terjadi sebelum publikasi berita publik, labeli sebagai `PRECEDED_ANNOUNCEMENT` dengan tingkat keyakinan 0.95.
* **Output Schema**:
  ```json
  {
    "concentration_ratio_top3": 0.724,
    "accumulation_regime": "INSTITUTIONAL_ACCUMULATION",
    "top_buyers": [
      {"broker_code": "CS", "broker_name": "Credit Suisse Sekuritas", "type": "FOREIGN_INSTITUTION", "net_buy_shares": 52000000}
    ],
    "insider_transactions": [
      {"insider_name": "Direktur Operasional", "action": "BUY", "shares": 1500000, "filing_date": "2026-09-11"}
    ],
    "verification_status": "SUPPORTED",
    "confidence_score": 0.95
  }
  ```

---

### Skill 4: `financial-health-stress-test` (Stress Test Neraca & Sanggahan Rumor)

* **Skill ID**: `financial-health-stress-test`
* **Trigger**: Pertanyaan tentang solvabilitas emiten, penurunan tajam harga saham ($R_t \le -5\%$), atau klarifikasi isu gagal bayar / kepailitan.
* **Prerequisites (MCP Tools)**:
  * `sectors_get_company_report` (`/v2/company/report/{symbol}/?sections=valuation,financials`)
  * `sectors_get_quarterly_financials` (`/v2/quarterly-financials/{symbol}/`)
* **Deterministic Gate**:
  * Perhitungan rasio likuiditas: $\text{Current Ratio} = \frac{\text{Aset Lancar}}{\text{Liabilitas Jangka Pendek}}$, $\text{Quick Ratio} = \frac{\text{Kas} + \text{Setara Kas}}{\text{Liabilitas Jangka Pendek}}$.
  * Perhitungan solvabilitas: $\text{DER} = \frac{\text{Total Utang}}{\text{Ekuitas}}$, $\text{Interest Coverage} = \frac{\text{EBIT}}{\text{Beban Bunga}}$.
* **Execution Protocol (SOP)**:
  1. Ambil pos neraca dan arus kas resmi triwulanan dari Sectors MCP.
  2. Hitung rasio kesehatan deterministik di Python.
  3. Bandingkan dengan klaim rumor di pasar: jika beredar rumor gagal bayar surat utang tetapi Quick Ratio $> 1.8x$ dan kas melimpah, tetapkan status `CONTRADICTED`.
* **Output Schema**:
  ```json
  {
    "liquidity": {"quick_ratio": 2.15, "current_ratio": 2.80, "health_grade": "PRISTINE"},
    "solvency": {"der": 0.42, "interest_coverage": 8.4},
    "rumor_refutation": {
      "claim": "Isu kesulitan pembayaran kupon obligasi",
      "verification_status": "CONTRADICTED",
      "evidence": "Saldo kas dan setara kas perseroan tercatat Rp 9,4 Triliun per laporan keuangan resmi Q2 2026.",
      "confidence_score": 1.00
    }
  }
  ```

---

### Skill 5: `mining-commodity-divergence` (Korelasi Komoditas Pertambangan IDX)

* **Skill ID**: `mining-commodity-divergence`
* **Trigger**: Emiten sektor energi/pertambangan (ANTM, PTBA, ADRO, MEDC, TINS, MBMA) mengalami pergerakan harga signifikan.
* **Prerequisites (MCP Tools)**:
  * `sectors_get_mining_company_detail` (`/v2/mining-company-detail/{slug}/`)
  * `sectors_get_commodity_price` (`/v2/commodity-price/{commodity}/`)
* **Deterministic Gate**:
  * Menghitung koefisien korelasi Pearson ($r$) antara return saham harian dan delta harga komoditas acuan (London Metal Exchange / Newcastle Coal) selama 30 hari.
* **Execution Protocol (SOP)**:
  1. Identifikasi komoditas utama emiten dari Sectors Mining Extension.
  2. Tarik deret harga spot komoditas global.
  3. Uji apakah pergerakan saham didorong oleh kenaikan komoditas global (*Commodity-Driven*) atau pergerakan anomali internal (*Company-Specific Anomaly*).
* **Output Schema**:
  ```json
  {
    "commodity_tracked": "NICKEL",
    "commodity_30d_return_pct": 1.2,
    "stock_30d_return_pct": 14.8,
    "divergence_class": "IDIOSYNCRATIC_COMPANY_ALPHA",
    "conclusion": "Pergerakan ANTM bukan akibat pergerakan harga nikel acuan global, melainkan katalis spesifik korporasi.",
    "verification_status": "SUPPORTED",
    "confidence_score": 0.85
  }
  ```

---

### Skill 6: `peer-valuation-benchmark` (Valuasi Relatif & Posisi Subsektor)

* **Skill ID**: `peer-valuation-benchmark`
* **Trigger**: Pertanyaan mengenai kewajaran harga saham, analisis komparatif emiten sejenis, atau rotasi sektor.
* **Prerequisites (MCP Tools)**:
  * `sectors_get_subsector_peers` (`/v2/subsector/{subsector}/`)
  * `sectors_get_company_report` (`/v2/company/report/{symbol}/?sections=valuation,peers`)
* **Deterministic Gate**:
  * Menghitung persentil valuasi: Posisi PER dan PBV saham terhadap median subsektor ($z_{\text{val}} = \frac{\text{PER}_{\text{stock}} - \text{Median}_{\text{subsector}}}{\text{IQR}_{\text{subsector}}}$).
* **Execution Protocol (SOP)**:
  1. Tarik seluruh emiten dalam subsektor yang sama dari Sectors API v2.
  2. Hitung median dan dispersi valuasi (P/E, P/B, EV/EBITDA, ROE).
  3. Petakan apakah saham diperdagangkan pada valuasi premium atau diskon terhadap rekan seindustri tanpa memberikan target harga spekulatif.
* **Output Schema**:
  ```json
  {
    "subsector": "metals-and-minerals-mining",
    "peer_count": 14,
    "metrics": {
      "target_pe": 12.4,
      "subsector_median_pe": 16.8,
      "valuation_posture": "TRADING_AT_DISCOUNT",
      "roe_rank": "TOP_25_PERCENTILE"
    },
    "verification_status": "SUPPORTED",
    "confidence_score": 1.00
  }
  ```

---

## 3. Dynamic Skill Loading & ReAct Orchestration

Agen ReAct [`engine/agent/react_agent.py`](../../engine/agent/react_agent.py) tidak memanggil raw endpoint bursa secara acak. Agen bekerja dengan alur:

```
[User Prompt]
      │
      ▼
1. REASONING: Agen menganalisis intensi pertanyaan dan memeriksa Skills Registry.
      │
2. SKILL ROUTING: Agen memilih satu atau beberapa Skill yang relevan:
   - "Kenapa ANTM melonjak?" ──▶ [market-anomaly-recon] ──▶ [event-causality-audit]
   - "Apakah rumor gagal bayar benar?" ──▶ [financial-health-stress-test]
      │
3. EXECUTION: Agen mengeksekusi SOP Skill:
   - Memanggil MCP Tools (Sectors & OSINT Primitives)
   - Melewatkan data melalui Deterministic Compute Gate
      │
4. EVIDENCE HARVEST: Output schema terstruktur disimpan ke SQLite (`findings` & `anomalies`).
      │
5. SYNTHESIS: Agen merangkum jawaban dengan taksonomi bukti [SUPPORTED | UNCERTAIN | CONTRADICTED].
```

---

## 4. Progressive Skill Disclosure & Proactive Recommendations (ADR-11)

Sesuai arsitektur **ADR-11**, Niskava Agent tidak melakukan *eager dump* terhadap seluruh definisi detail skill ke dalam system prompt (menghindari *prompt bloat* ~1.450 token). Sebagai gantinya:

### A. Gateway Primitive (`execute_skill`)
LLM hanya melihat satu fungsi pintu gerbang:
```json
{
  "name": "execute_skill",
  "arguments": {
    "skill_id": "market_anomaly_recon",
    "arguments": {"ticker": "ANTM", "days": 30}
  }
}
```
Deskripsi fungsi `execute_skill` memuat **Compact Skills Manifest** 1-baris untuk masing-masing dari 6 SOP di atas, memangkas ukuran prompt sistem hingga ~716 token.

### B. Proactive Follow-Up Graph (`_SKILL_FOLLOWUP_GRAPH`)
Setelah skill selesai dieksekusi dan hasil sintesis siap dikirim, agen secara proaktif menavigasikan langkah investigasi logis berikutnya via directed graph:

```
market_anomaly_recon
  ├──▶ event_causality_audit
  └──▶ insider_bandarmology_forensic

event_causality_audit
  ├──▶ insider_bandarmology_forensic
  └──▶ financial_health_stress_test

insider_bandarmology_forensic
  ├──▶ peer_valuation_benchmark
  └──▶ financial_health_stress_test

mining_commodity_divergence
  ├──▶ market_anomaly_recon
  └──▶ peer_valuation_benchmark
```

Di akhir respon markdown, agen menyematkan rekomendasi langkah berikutnya (tanpa mengulang skill yang sudah pernah dijalankan pada sesi tersebut) dalam format dwibahasa (*ID/EN*).


