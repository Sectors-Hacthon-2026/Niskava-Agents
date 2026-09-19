# 08 — Arsitektur Modular Domain Skills & MCP Primitives vs Monolithic AI Wrapper

**Status:** ACCEPTED  
**Tanggal:** 2026-09-19  
**Pengambil Keputusan:** Core Team  
**Dokumen Terkait:** [`../00-foundations/01-vision-and-thesis.md`](../00-foundations/01-vision-and-thesis.md), [`../30-agent/02-skills-catalog.md`](../30-agent/02-skills-catalog.md), [`02-deterministic-quant-pre-llm.md`](02-deterministic-quant-pre-llm.md)

---

## 1. Konteks & Permasalahan

Dalam membangun asisten kecerdasan buatan finansial, terdapat godaan besar untuk mengadopsi pendekatan **"Monolithic AI Wrapper"**:
* Menulis satu system prompt raksasa berisi semua aturan analisis.
* Meng-hardcode beberapa fungsi API bursa langsung ke dalam prompt function calling.
* Menyerahkan seluruh alur logika investigasi kepada LLM tanpa segmentasi metodologi terstruktur.

Kelemahan pendekatan wrapper monolitik ini:
1. **Kegagalan Investigasi Bertahap**: Ketika menemui skenario rumit (misal emiten tambang dengan anomali foreign flow dan rumor kebangkrutan sekaligus), agen monolitik kehilangan fokus (*instruction drift*), melewatkan data fundamental, atau mengambil kesimpulan terburu-buru.
2. **Ketiadaan Standarisasi Antarmuka**: Tidak ada isolasi antara cara menarik data bursa (*data fetching primitives*) dengan metodologi evaluasi analis (*analytical procedures*).
3. **Kerapuhan Pengujian (Untestable Monolith)**: Sangat sulit menguji apakah suatu logika analisis (misalnya deteksi insider bandarmology) bekerja secara akurat tanpa harus mengeksekusi seluruh rantai prompt.

---

## 2. Keputusan Arsitektur

Niskava Agent secara tegas mengadopsi **Arsitektur 4-Layer Terpisah**:

```
Layer 4: Cognitive ReAct Agent (Hermes/OpenCode loop, planning & synthesis)
   │
   ▼ (Memilih SOP Analisis)
Layer 3: Modular Domain Skills (Standardized SOP: market-anomaly, causality, insider, health)
   │
   ▼ (Firewall Numerik Wajib)
Layer 2: Deterministic Compute Gate (NumPy Z-Score, Moving Average, Correlation)
   │
   ▼ (Protokol Alat Standar)
Layer 1: MCP & OSINT Primitives (Sectors MCP Server + Dual-Engine OSINT Tools)
```

### Karakteristik Desain:
1. **Model Context Protocol (MCP) sebagai Standar Primitives (Layer 1):**  
   Seluruh interaksi data bursa diekspos sebagai primitive tools MCP yang stateless dan reusable.
2. **Deterministic Compute Gate (Layer 2):**  
   Firewall numerik NumPy yang memproses time series sebelum LLM melihat data, menjamin nol halusinasi aritmatika.
3. **Domain Skills sebagai SOP Analis Terisolasi (Layer 3):**  
   Setiap kapabilitas analisis dibungkus menjadi spesifikasi Skill mandiri dengan trigger condition, required MCP tools, deterministic gates, dan schema JSON output yang terikat kontrak.
4. **Cognitive ReAct Loop (Layer 4):**  
   LLM berperan murni sebagai orkestrator cerdas yang memilih Skill, mengevaluasi bukti, dan menyusun sintesis akhir dengan Three-Tier Verification Taxonomy.

---

## 3. Alternatif yang Dipertimbangkan

| Alternatif | Alasan Ditolak |
|---|---|
| **Monolithic System Prompt (Wrapper Tradisional)** | Cepat dibuat tetapi tidak dapat diskalakan, boros token, dan menghasilkan penalaran dangkal tanpa SOP analis terverifikasi. |
| **Pure LangChain/AutoGPT Agent Architecture** | Menambah dependensi pihak ketiga yang gemuk, overhead memori besar, dan sulit diintegrasikan ke dalam single executable Go. |
| **Hardcoded Sequential Pipeline Only (Tanpa ReAct Prompt Mode)** | Mengorbankan fleksibilitas eksplorasi interaktif REPL pengguna. |

---

## 4. Konsekuensi

### Positif
* **Investigasi Berkualitas Analis Ekuitas**: Agen mengikuti metodologi formal yang diakui di industri pasar modal, bukan tebak-tebakan teks.
* **Komposabilitas & Ekstensibilitas Tinggi**: Skill baru (misal *IPO forensic*, *options flow*, *ESG compliance*) dapat ditambahkan tanpa merusak skill yang sudah ada.
* **Kepatuhan Regulasi & Hackathon**: Memenuhi standar Track 1 (multi-step reasoning, custom tool-use pipelines, purpose-built interfaces).
* **Hemat Kuota & Bebas Halusinasi**: Data mentah disaring di Layer 2 sebelum masuk ke Layer 4.

### Negatif / Kompromi yang Diterima
* Membutuhkan arsitektur kode yang lebih terstruktur dan pendefinisian kontrak skema data JSON yang disiplin.
