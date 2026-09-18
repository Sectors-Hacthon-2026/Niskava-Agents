# 02 — Skema Database SQLite Lokal

**Status:** ACCEPTED  
**Versi Dokumen:** 1.1.0  
**Terakhir Diperbarui:** 2026-09-16  
**Keputusan Terkait:** [`04-local-first-sqlite-storage.md`](../90-decisions/04-local-first-sqlite-storage.md), [`06-local-conversational-graph-memory.md`](../90-decisions/06-local-conversational-graph-memory.md)  

Database SQLite lokal disimpan di path direktori home pengguna: `~/.niskava/niskava.db`.

---

## 1. Diagram Relasi Entitas (ERD)

```text
┌─────────────────────────┐
│      investigations     │
├─────────────────────────┤
│ id (PK)                 │◀──────┐
│ ticker                  │       │
│ status                  │       │ 1:N
│ started_at              │       │
│ completed_at            │       │
│ summary_text            │       │
└───────────┬─────────────┘       │
            │ 1:N                 │
            ▼                     │
┌─────────────────────────┐       │
│        anomalies        │       │
├─────────────────────────┤       │
│ id (PK)                 │       │
│ investigation_id (FK)   │       │
│ anomaly_date            │       │
│ metric_type             │       │
│ z_score                 │       │
│ description             │       │
└─────────────────────────┘       │
                                  │
┌─────────────────────────┐       │
│        findings         │       │
├─────────────────────────┤       │
│ id (PK)                 │       │
│ investigation_id (FK)   │───────┤
│ title                   │       │
│ claim_text              │       │
│ verification_status     │       │
│ confidence_score        │       │
│ causality_status        │       │
└───────────┬─────────────┘       │
            │ 1:N                 │
            ▼                     │
┌─────────────────────────┐       │
│      evidence_items     │       │
├─────────────────────────┤       │
│ id (PK)                 │       │
│ finding_id (FK)         │       │
│ source_type             │       │
│ source_name             │       │
│ source_url              │       │
│ publication_date        │       │
│ snippet_text            │       │
└─────────────────────────┘       │
                                  │
┌─────────────────────────┐       │
│     timeline_events     │       │
├─────────────────────────┤       │
│ id (PK)                 │       │
│ investigation_id (FK)   │───────┘
│ event_timestamp         │
│ event_type              │
│ headline                │
│ details                 │
└─────────────────────────┘

┌─────────────────────────┐       ┌─────────────────────────┐
│      sectors_cache      │       │      memory_nodes       │ (Graph Memory)
├─────────────────────────┤       ├─────────────────────────┤
│ cache_key (PK)          │       │ id (PK)                 │◀──────┐
│ endpoint                │       │ label                   │       │
│ payload_json            │       │ node_type               │       │
│ expires_at              │       │ metadata_json           │       │
└─────────────────────────┘       │ last_observed_at        │       │
                                  └─────────────────────────┘       │ 1:N (Source & Target)
                                  ┌─────────────────────────┐       │
                                  │      memory_edges       │       │
                                  ├─────────────────────────┤       │
                                  │ source_id (PK, FK)      │───────┤
                                  │ target_id (PK, FK)      │───────┘
                                  │ relation (PK)           │
                                  │ context_snippet         │
                                  │ session_id (FK)         │───▶ investigations(id)
                                  │ weight                  │
                                  │ last_observed_at        │
                                  └─────────────────────────┘
```

---

## 2. Definisi DDL SQL Inti Sesi Investigasi

```sql
-- Konfigurasi Performa SQLite Murni Go (Zero-CGO)
PRAGMA journal_mode = WAL;
PRAGMA foreign_keys = ON;
PRAGMA synchronous = NORMAL;

-- Sesi Investigasi
CREATE TABLE IF NOT EXISTS investigations (
    id TEXT PRIMARY KEY,                       -- Format: 'INV-2026-0001'
    ticker TEXT NOT NULL,                      -- Contoh: 'ANTM'
    market TEXT NOT NULL DEFAULT 'IDX',
    timeframe_days INTEGER NOT NULL DEFAULT 30,
    status TEXT NOT NULL,                      -- 'PENDING', 'RUNNING', 'COMPLETED', 'FAILED'
    started_at TEXT NOT NULL,
    completed_at TEXT,
    summary_text TEXT,
    created_at TEXT NOT NULL DEFAULT CURRENT_TIMESTAMP
);

-- Anomali Kuantitatif Terdeteksi
CREATE TABLE IF NOT EXISTS anomalies (
    id TEXT PRIMARY KEY,
    investigation_id TEXT NOT NULL REFERENCES investigations(id) ON DELETE CASCADE,
    anomaly_date TEXT NOT NULL,                -- 'YYYY-MM-DD'
    metric_type TEXT NOT NULL,                 -- 'VOLUME_SPIKE', 'PRICE_BREAKOUT', 'SECTOR_DIVERGENCE'
    metric_value REAL NOT NULL,
    baseline_value REAL NOT NULL,
    z_score REAL NOT NULL,
    description TEXT,
    created_at TEXT NOT NULL DEFAULT CURRENT_TIMESTAMP
);

-- Temuan Investigasi (Findings)
CREATE TABLE IF NOT EXISTS findings (
    id TEXT PRIMARY KEY,
    investigation_id TEXT NOT NULL REFERENCES investigations(id) ON DELETE CASCADE,
    title TEXT NOT NULL,
    claim_text TEXT NOT NULL,
    verification_status TEXT NOT NULL,         -- 'SUPPORTED', 'UNCERTAIN', 'CONTRADICTED'
    confidence_score REAL NOT NULL,            -- 0.00 - 1.00
    causality_status TEXT NOT NULL,            -- 'LIKELY_CATALYST', 'PRECEDED_ANNOUNCEMENT', 'UNEXPLAINED_BY_NEWS'
    created_at TEXT NOT NULL DEFAULT CURRENT_TIMESTAMP
);

-- Bukti Pendukung (Evidence Items)
CREATE TABLE IF NOT EXISTS evidence_items (
    id TEXT PRIMARY KEY,
    finding_id TEXT NOT NULL REFERENCES findings(id) ON DELETE CASCADE,
    source_type TEXT NOT NULL,                 -- 'SECTORS_DATA', 'OFFICIAL_DISCLOSURE', 'PUBLIC_NEWS'
    source_name TEXT NOT NULL,                 -- 'Sectors v2 API', 'Keterbukaan BEI', 'Kontan'
    source_url TEXT,
    publication_date TEXT,
    snippet_text TEXT NOT NULL,
    created_at TEXT NOT NULL DEFAULT CURRENT_TIMESTAMP
);

-- Kronologi Kejadian (Timeline)
CREATE TABLE IF NOT EXISTS timeline_events (
    id TEXT PRIMARY KEY,
    investigation_id TEXT NOT NULL REFERENCES investigations(id) ON DELETE CASCADE,
    event_timestamp TEXT NOT NULL,
    event_type TEXT NOT NULL,                  -- 'DATA_ANOMALY', 'NEWS_RELEASE', 'DISCLOSURE', 'SUSPENSION', 'INSIDER_TRADE'
    headline TEXT NOT NULL,
    details TEXT,
    created_at TEXT NOT NULL DEFAULT CURRENT_TIMESTAMP
);

-- Catatan Suspensi Saham BEI Resmi (dari /v2/suspensions/)
CREATE TABLE IF NOT EXISTS suspension_records (
    id TEXT PRIMARY KEY,
    symbol TEXT NOT NULL,
    suspension_date TEXT NOT NULL,
    unsuspension_date TEXT,
    session TEXT,
    market_type TEXT,
    reason TEXT NOT NULL,
    pdf_url TEXT,                              -- Tautan langsung ke pengumuman PDF resmi BEI
    created_at TEXT NOT NULL DEFAULT CURRENT_TIMESTAMP
);

-- Transaksi Kepemilikan Orang Dalam / Insiders (dari /v2/filings/)
CREATE TABLE IF NOT EXISTS insider_filings (
    id TEXT PRIMARY KEY,
    symbol TEXT NOT NULL,
    holder_name TEXT NOT NULL,
    holder_type TEXT NOT NULL,                 -- 'DIRECTOR', 'COMMISSIONER', 'MAJOR_SHAREHOLDER'
    transaction_type TEXT NOT NULL,            -- 'BUY', 'SELL'
    transaction_date TEXT NOT NULL,
    shares_transacted REAL NOT NULL,
    price_per_share REAL,
    percentage_after_transaction REAL,
    purpose TEXT,
    created_at TEXT NOT NULL DEFAULT CURRENT_TIMESTAMP
);

-- Cache Data Sectors v2 (Penghematan Kuota Kredit 1.000 Grant)
CREATE TABLE IF NOT EXISTS sectors_cache (
    cache_key TEXT PRIMARY KEY,                -- Hash sha256 dari (endpoint + query_params)
    endpoint TEXT NOT NULL,                    -- Misal: '/v2/daily/ANTM/'
    params_hash TEXT,
    payload_json TEXT NOT NULL,                -- Respon JSON mentah dari Sectors
    expires_at TEXT,                           -- NULL jika data permanen (hari bursa lampau)
    created_at TEXT NOT NULL DEFAULT CURRENT_TIMESTAMP
);

-- Cache Data OSINT & Web Intelligence (07-resilient-dual-engine-osint-architecture)
CREATE TABLE IF NOT EXISTS osint_cache (
    cache_key TEXT PRIMARY KEY,                -- Hash sha256 dari query / URL
    source_type TEXT NOT NULL,                 -- 'GOOGLE_NEWS_RSS', 'WEB_ARTICLE'
    query_or_url TEXT NOT NULL,
    content_text TEXT NOT NULL,
    metadata_json TEXT,
    expires_at TEXT,                           -- TTL 24 jam untuk berita baru, NULL untuk historis
    created_at TEXT NOT NULL DEFAULT CURRENT_TIMESTAMP
);

-- Indeks Performa
CREATE INDEX IF NOT EXISTS idx_investigations_ticker ON investigations(ticker);
CREATE INDEX IF NOT EXISTS idx_investigations_created ON investigations(created_at DESC);
CREATE INDEX IF NOT EXISTS idx_anomalies_investigation ON anomalies(investigation_id);
CREATE INDEX IF NOT EXISTS idx_findings_investigation ON findings(investigation_id);
CREATE INDEX IF NOT EXISTS idx_evidence_finding ON evidence_items(finding_id);
CREATE INDEX IF NOT EXISTS idx_timeline_investigation ON timeline_events(investigation_id, event_timestamp ASC);
CREATE INDEX IF NOT EXISTS idx_suspensions_symbol ON suspension_records(symbol, suspension_date DESC);
CREATE INDEX IF NOT EXISTS idx_insider_filings_symbol ON insider_filings(symbol, transaction_date DESC);
CREATE INDEX IF NOT EXISTS idx_sectors_cache_endpoint ON sectors_cache(endpoint);
CREATE INDEX IF NOT EXISTS idx_osint_cache_type ON osint_cache(source_type);
```

---

## 3. Skema Tabel Graf Memori Percakapan & Lintas Sesi (06-local-conversational-graph-memory)

Tabel ini mendukung **Local Conversational Graph Memory Engine** agar agen dapat mengingat emiten, harga, anomali, dan preferensi pengguna lintas sesi:

```sql
-- Simpul Graf Memori (Entitas Percakapan & Riset)
CREATE TABLE IF NOT EXISTS memory_nodes (
    id TEXT PRIMARY KEY,                       -- Contoh: 'ticker:ANTM', 'user:default', 'catalyst:ANTM_INV-01'
    label TEXT NOT NULL,                      -- Contoh: 'ANTM', 'Smelter Halmahera'
    node_type TEXT NOT NULL,                  -- 'USER', 'TICKER', 'SECTOR', 'PRICE_LEVEL', 'CATALYST_EVENT'
    metadata_json TEXT,                       -- Properti fleksibel (JSON)
    created_at TEXT NOT NULL DEFAULT CURRENT_TIMESTAMP,
    last_observed_at TEXT NOT NULL DEFAULT CURRENT_TIMESTAMP
);

-- Relasi Graf Memori Berarah (Directed Edges)
CREATE TABLE IF NOT EXISTS memory_edges (
    source_id TEXT NOT NULL REFERENCES memory_nodes(id) ON DELETE CASCADE,
    target_id TEXT NOT NULL REFERENCES memory_nodes(id) ON DELETE CASCADE,
    relation TEXT NOT NULL,                   -- 'INVESTIGATED', 'WATCHES', 'HOLDS_AT', 'CATALYZED_BY'
    context_snippet TEXT,                     -- Ringkasan konteks percakapan/temuan
    session_id TEXT REFERENCES investigations(id) ON DELETE SET NULL,
    weight REAL NOT NULL DEFAULT 1.0,         -- Bobot relasi (mendukung temporal recency decay)
    confidence_score REAL NOT NULL DEFAULT 1.0,
    last_observed_at TEXT NOT NULL DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY (source_id, target_id, relation)
);

-- Indeks Graf untuk Traversal Cepat Subgraf
CREATE INDEX IF NOT EXISTS idx_memory_edges_source ON memory_edges(source_id);
CREATE INDEX IF NOT EXISTS idx_memory_edges_target ON memory_edges(target_id);
CREATE INDEX IF NOT EXISTS idx_memory_edges_relation ON memory_edges(relation);
CREATE INDEX IF NOT EXISTS idx_memory_edges_last_observed ON memory_edges(last_observed_at DESC);
CREATE INDEX IF NOT EXISTS idx_memory_nodes_type ON memory_nodes(node_type);
```
