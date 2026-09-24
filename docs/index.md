# Niskava Agent Documentation Catalogue (Index)

**Doc set version:** `1.2.0`  
**Status:** ACTIVE BASELINE  
**Terakhir Diperbarui:** 2026-09-22  

Katalog lengkap seluruh dokumen arsitektur, produk, logika agen, dan keputusan teknis Niskava Agent.

---

## 00 — Foundations (Fondasi Sistem)

| Path | Ringkasan Dokumen | Status |
|---|---|---|
| [`00-foundations/01-vision-and-thesis.md`](00-foundations/01-vision-and-thesis.md) | Visi inti *"Investigate, Don't Just Answer"*, masalah fragmentasi data IDX, dan matriks komparasi terhadap chatbot umum. | `LOCKED` |
| [`00-foundations/02-principles-and-rules.md`](00-foundations/02-principles-and-rules.md) | 8 Prinsip operasional mutlak, disiplin pembuktian fakta, dan batasan kepatuhan regulasi finansial non-advisory. | `LOCKED` |
| [`00-foundations/03-glossary.md`](00-foundations/03-glossary.md) | Kamus istilah resmi pasar modal Indonesia (IDX) dan rekayasa agen AI/OSINT. | `ACCEPTED` |

---

## 10 — Product (Spesifikasi Produk & Pasar)

| Path | Ringkasan Dokumen | Status |
|---|---|---|
| [`10-product/01-problem-and-market.md`](10-product/01-problem-and-market.md) | Analisis lanskap pasar modal Indonesia, kebutuhan investor, dan keunggulan data moat Sectors.app. | `ACCEPTED` |
| [`10-product/02-personas-and-jobs.md`](10-product/02-personas-and-jobs.md) | Profil target persona: Retail Swing Trader, Equity Research Associate, dan Financial Fact-Checker. | `ACCEPTED` |
| [`10-product/03-product-scope-and-surfaces.md`](10-product/03-product-scope-and-surfaces.md) | Spesifikasi antarmuka ganda: Terminal CLI interaktif (`bubbletea` + i18n) dan Local Web Dashboard (`server.go` embedded & `shadcn/ui`). | `ACCEPTED` |
| [`10-product/04-user-journeys.md`](10-product/04-user-journeys.md) | Alur interaksi pengguna ujung-ke-ujung (dari inisiasi CLI hingga inspeksi timeline bukti di web). | `ACCEPTED` |
| [`10-product/05-hackathon-strategy.md`](10-product/05-hackathon-strategy.md) | Penyelarasan kriteria Track 1 Sectors Hackathon 2026, strategi 1.000 API credit, dan skenario demo ANTM. | `LOCKED` |
| [`10-product/06-demo-video-script.md`](10-product/06-demo-video-script.md) | Storyboard dan naskah resmi video demo penjurian: Teaser 1-Menit & Judging Walkthrough 3-Menit. | `ACCEPTED` |

---

## 20 — Architecture (Arsitektur & Rekayasa Sistem)

| Path | Ringkasan Dokumen | Status |
|---|---|---|
| [`20-architecture/01-system-overview.md`](20-architecture/01-system-overview.md) | Desain arsitektur hybrid Go Core + Python Engine + React Web SPA, configurable monorepo, resolusi path dinamis, dan pemisahan repositori. | `ACCEPTED` |
| [`20-architecture/02-database-schema.md`](20-architecture/02-database-schema.md) | Skema database SQLite lokal (`~/.niskava/niskava.db`), tabel relasi, indeks performa, dan caching layer. | `ACCEPTED` |
| [`20-architecture/03-sectors-v2-api.md`](20-architecture/03-sectors-v2-api.md) | Spesifikasi lengkap Sectors v2 (7 domain endpoint: Transaksi, Fundamental, Suspensi, Insider Filings, Foreign Flow, Mining, Screener). | `ACCEPTED` |
| [`20-architecture/04-ipc-and-api-contract.md`](20-architecture/04-ipc-and-api-contract.md) | Spesifikasi protokol Subprocess IPC (JSON Lines) dan REST/SSE server lokal. | `ACCEPTED` |
| [`20-architecture/05-anomaly-detection-math.md`](20-architecture/05-anomaly-detection-math.md) | Formula matematis deterministik ($Z$-score volume, abnormal return, dan divergensi sektor). | `LOCKED` |
| [`20-architecture/06-osint-engine.md`](20-architecture/06-osint-engine.md) | Arsitektur Dual-Engine OSINT (Sectors v2 + Google News RSS), Secondary Disclosure Dorking, dan isolasi konteks anti-injeksi. | `ACCEPTED` |
| [`20-architecture/07-non-functional-requirements.md`](20-architecture/07-non-functional-requirements.md) | Standar NFR: Latency budget, ukuran biner, penggunaan memori, dan kapabilitas offline. | `ACCEPTED` |
| [`20-architecture/08-security-and-compliance.md`](20-architecture/08-security-and-compliance.md) | Manajemen kunci API lokal, proteksi prompt injection dari web, dan kepatuhan terhadap UU Pasar Modal. | `ACCEPTED` |
| [`20-architecture/09-git-workflow-and-branching-strategy.md`](20-architecture/09-git-workflow-and-branching-strategy.md) | Standarisasi Git branching flow, proteksi cabang `main` & `dev`, SOP penggabungan, dan Quality Gates. | `ACCEPTED` |
| [`20-architecture/10-npm-distribution-and-packaging-spec.md`](20-architecture/10-npm-distribution-and-packaging-spec.md) | Spesifikasi distribusi & packaging NPM (`npx niskava`), Node launcher wrapper, dan otomasi runtime hybrid. | `PROPOSED` |

---

## 30 — Agent (Logika Agen & Investigasi)

| Path | Ringkasan Dokumen | Status |
|---|---|---|
| [`30-agent/01-investigation-pipeline.md`](30-agent/01-investigation-pipeline.md) | Alur kerja 7 tahap investigasi otonom dari inisiasi hingga sintesis temuan dan streaming. | `ACCEPTED` |
| [`30-agent/02-skills-catalog.md`](30-agent/02-skills-catalog.md) | Katalog Standard Operating Procedure (SOP) analisa agen: `market-anomaly`, `company-research`, dll. | `ACCEPTED` |
| [`30-agent/03-evidence-and-causality.md`](30-agent/03-evidence-and-causality.md) | Taksonomi validasi bukti (`SUPPORTED`, `UNCERTAIN`, `CONTRADICTED`) dan evaluasi urutan kausalitas waktu. | `LOCKED` |
| [`30-agent/04-eval-and-benchmarks.md`](30-agent/04-eval-and-benchmarks.md) | Kerangka pengujian akurasi agen berbasis *ground truth* kasus historis nyata di IDX (ANTM, BUMI, GOTO). | `ACCEPTED` |
| [`30-agent/05-conversational-memory-engine.md`](30-agent/05-conversational-memory-engine.md) | Spesifikasi engine memori percakapan berbasis graf lokal (SQLite + NetworkX), siklus 4-tahap, dan retensi konteks lintas sesi. | `ACCEPTED` |

---

## 90 — Decisions (Architecture Decision Records)

| Path | Ringkasan Keputusan | Status |
|---|---|---|
| [`90-decisions/01-hybrid-stack-go-python-react.md`](90-decisions/01-hybrid-stack-go-python-react.md) | Memilih arsitektur hybrid Go (CLI/Daemon) + Python (Quant/AI) + React (Embedded UI) dengan pola decoupled repos & dynamic paths. | `ACCEPTED` |
| [`90-decisions/02-deterministic-quant-pre-llm.md`](90-decisions/02-deterministic-quant-pre-llm.md) | Menghitung statistik dan anomali deret waktu secara deterministik sebelum memanggil LLM. | `ACCEPTED` |
| [`90-decisions/03-sectors-v2-and-credit-conservation.md`](90-decisions/03-sectors-v2-and-credit-conservation.md) | Adopsi API v2 dan arsitektur SQLite caching lokal untuk mengamankan kuota 1.000 kredit API. | `ACCEPTED` |
| [`90-decisions/04-local-first-sqlite-storage.md`](90-decisions/04-local-first-sqlite-storage.md) | Menyimpan seluruh sesi investigasi secara lokal tanpa server cloud tersentralisasi. | `ACCEPTED` |
| [`90-decisions/05-strict-financial-non-advisory-boundary.md`](90-decisions/05-strict-financial-non-advisory-boundary.md) | Membatasi agen pada intelijen faktual dan melarang keras sinyal trading spekulatif *Buy/Sell*. | `ACCEPTED` |
| [`90-decisions/06-local-conversational-graph-memory.md`](90-decisions/06-local-conversational-graph-memory.md) | Mengadopsi graf memori lokal (SQLite + NetworkX) untuk retensi asosiasi entitas dan preferensi pengguna lintas sesi. | `ACCEPTED` |
| [`90-decisions/07-resilient-dual-engine-osint-architecture.md`](90-decisions/07-resilient-dual-engine-osint-architecture.md) | Mengadopsi arsitektur Dual-Engine OSINT (Sectors v2 + Google News RSS) untuk mengatasi sensor ISP lokal dan proteksi WAF bursa. | `ACCEPTED` |
| [`90-decisions/08-modular-skills-and-mcp-architecture.md`](90-decisions/08-modular-skills-and-mcp-architecture.md) | Mengadopsi arsitektur 4-layer pemisahan MCP primitives, compute gate, domain skills, dan cognitive ReAct loop. | `ACCEPTED` |
| [`90-decisions/09-internationalization-and-dual-language-ux.md`](90-decisions/09-internationalization-and-dual-language-ux.md) | Mengadopsi standardisasi i18n dwibahasa (Bahasa Indonesia & English) pada TUI dan prompt synthesis. | `ACCEPTED` |
| [`90-decisions/10-dual-surface-web-delivery-strategy.md`](90-decisions/10-dual-surface-web-delivery-strategy.md) | Mengadopsi strategi web delivery ganda: Embedded Go HTML Single-Binary Fallback + Vite React SPA. | `ACCEPTED` |
| [`90-decisions/11-progressive-skill-disclosure-and-lean-agent-architecture.md`](90-decisions/11-progressive-skill-disclosure-and-lean-agent-architecture.md) | Mengadopsi pola Progressive Skill Disclosure & Gateway Primitives (Antigravity/OpenCode) untuk memangkas prompt bloat. | `ACCEPTED` |

---

## Dokumen Navigasi Tambahan
* [`README.md`](README.md) — Halaman panduan utama, panduan bacaan (*reading order*), dan aturan tata kelola dokumentasi.
* [Root `AGENTS.md`](../AGENTS.md) — Konstitusi dan aturan operasional wajib bagi AI coding agent dan pengembang (Full English).
* [`open-questions.md`](open-questions.md) — Daftar pertanyaan terbuka dan keputusan arsitektur yang masih dalam proses evaluasi.
* [`TEAM_CLIENTS_INTEGRATION_GUIDE.md`](TEAM_CLIENTS_INTEGRATION_GUIDE.md) — Panduan integrasi teknis frontend dan CLI untuk pengembang klien.
* [`niskava_Project_Summary.md`](niskava_Project_Summary.md) — Ringkasan komprehensif produk, arsitektur, dan nilai tambah Niskava Agent.
