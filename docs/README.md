# Niskava Agent — Documentation Hub & Single Source of Truth

**Doc set version:** `1.1.0`  
**Status:** ACTIVE BASELINE  
**Terakhir Diperbarui:** 2026-09-16  
**Target Kompetisi:** [Sectors Hackathon Indonesia 2026](https://hackathon.sectors.app/)  
**Lintasan:** **Track 1 · AI Agents & Assistants** (*Qualifying Test Compliant*)  

---

## 1. Ringkasan Eksekutif & Problem Statement

**Niskava Agent** adalah platform orkestrasi investigasi finansial dan intelijen pasar modal Indonesia (IDX) otonom berbasis *evidence-first*, menggabungkan data kuantitatif **Sectors Financial API v2** dengan intelijen kualitatif eksternal **OSINT (News, Public Filings, Corporate Actions)**.

### 📌 1-Sentence Problem Statement (Resmi Submisi)
> *"Pelaku pasar modal Indonesia kehilangan momentum dan modal karena lambat memverifikasi penyebab lonjakan saham tidak wajar akibat terpisahnya data kuantitatif bursa (Sectors API) dengan konteks keterbukaan informasi dan berita pasar modal (OSINT)."*

Slogan inti: **"Don't just answer questions. Investigate them."**

---

## 2. Pernyataan Kepatuhan Regulasi & Aturan Kompetisi

Sesuai aturan resmi Sectors Hackathon 2026 (Official Rules Section 06 & Section 12):
1. **Larangan Eksekusi Trading Otomatis (Prohibition of Automated Trade Execution):**  
   Niskava Agent dirancang secara ketat sebagai sistem intelijen dan analisis bukti. Sistem **TIDAK memiliki kemampuan, modul, atau integrasi ke broker untuk menempatkan atau mengeksekusi order beli/jual secara otomatis**.
2. **Batasan Non-Advisory Finansial (Non-Advisory Boundary - [`05-strict-financial-non-advisory-boundary.md`](90-decisions/05-strict-financial-non-advisory-boundary.md)):**  
   Niskava **TIDAK memberikan nasihat keuangan, target harga pasti, atau rekomendasi BUY/SELL**. Seluruh temuan dikelompokkan dalam taksonomi bukti faktual (`SUPPORTED`, `UNCERTAIN`, `CONTRADICTED`) dan selalu menyertakan disclaimer risiko investasi pada CLI dan Web UI.

---

## 3. Panduan Evaluasi Khusus Dewan Juri (3-Minute Route)

Untuk menilai keselarasan Niskava dengan bobot rubrik penjurian (**40% Usability, 30% Storytelling Video, 30% Technical Depth**):

1. **Memeriksa Kualifikasi Track 1 & Strategi Penjurian (40% Usability & 30% Tech):**
   * Buka [`10-product/05-hackathon-strategy.md`](10-product/05-hackathon-strategy.md) — Matriks kepatuhan 6 syarat Track 1, alokasi kredit, dan strategi nilai maksimal.
2. **Melihat Alur Orkestrasi Multi-Step Reasoning (Core Innovation):**
   * Buka [`30-agent/01-investigation-pipeline.md`](30-agent/01-investigation-pipeline.md) — Alur kerja 7 tahap investigasi otonom dari deteksi anomali hingga verifikasi bukti.
3. **Memverifikasi Arsitektur Nyata (Not Faked / High Technical Depth):**
   * Buka [`20-architecture/01-system-overview.md`](20-architecture/01-system-overview.md) — Arsitektur hybrid Single Binary Go + Python Deterministic Quant + React SPA.
   * Buka [`20-architecture/03-sectors-v2-api.md`](20-architecture/03-sectors-v2-api.md) — Bukti ketergantungan mutlak (*Core Source Proof*) pada Sectors API v2.
4. **Membuktikan Akurasi pada Kasus Nyata BEI:**
   * Buka [`30-agent/04-eval-and-benchmarks.md`](30-agent/04-eval-and-benchmarks.md) — Evaluasi benchmark kasus historis nyata (`ANTM`, `BUMI`, `GOTO`, `BBRI`).
5. **Menguji Manajemen Memori & State Lintas Sesi (Track 1 Memory Criterion):**
   * Buka [`30-agent/05-conversational-memory-engine.md`](30-agent/05-conversational-memory-engine.md) dan [`90-decisions/06-local-conversational-graph-memory.md`](90-decisions/06-local-conversational-graph-memory.md) — Arsitektur graf memori lokal (SQLite + NetworkX) untuk retensi asosiasi entitas dan preferensi pengguna lintas sesi.

---

## 4. Aturan Hierarki Keputusan (*Precedence on Conflict*)

Jika ditemukan perbedaan pernyataan teknis antar dokumen:
```
Keputusan Langsung Tim Inti (Owner Decision)
  > 90-decisions/*.md (Architecture Decision Records yang disetujui)
    > 00-foundations/ (Prinsip operasional & batasan non-advisory yang terkunci)
      > 20-architecture/ & 30-agent/ (Spesifikasi teknis & kontrak sistem)
        > 10-product/ (Spesifikasi produk & strategi hackathon)
```
Setiap perubahan pada prinsip yang telah `LOCKED` wajib melalui persetujuan tim dan dicatat sebagai ADR baru.

---

## 5. Peta Area Dokumentasi

Untuk katalog lengkap setiap file beserta ringkasannya, buka **[`index.md`](index.md)**.

| Area | Folder | Isi Utama |
|---|---|---|
| **Fondasi** | [`00-foundations/`](00-foundations/) | Visi & tesis, 8 prinsip operasional mutlak, glosarium IDX & AI. |
| **Produk** | [`10-product/`](10-product/) | Lanskap masalah, target persona, antarmuka pengguna, strategi hackathon. |
| **Arsitektur** | [`20-architecture/`](20-architecture/) | Desain sistem hybrid, skema SQLite, Sectors v2, IPC contract, NFR, keamanan. |
| **Logika Agen** | [`30-agent/`](30-agent/) | 7-stage pipeline, katalog skills, taksonomi bukti, evaluasi & benchmark, engine memori graf lokal. |
| **Keputusan** | [`90-decisions/`](90-decisions/) | Architecture Decision Records (`01` s/d `07`). |
| **Konstitusi Agen** | [`AGENTS.md`](AGENTS.md) / [Root `AGENTS.md`](../AGENTS.md) | Konstitusi operasional agen AI, 6 hukum arsitektur, dan standar kode (Full English). |
| **Dinamika** | [`open-questions.md`](open-questions.md) | Daftar isu teknis terbuka yang sedang dievaluasi. |

---

## 6. Label Status yang Digunakan

Setiap dokumen memiliki header status resmi:
* **`LOCKED`**: Keputusan final yang mengikat. Hanya dapat dibuka kembali melalui ADR baru yang menggantikannya (*superseding ADR*).
* **`ACCEPTED`**: Disepakati oleh tim sebagai panduan implementasi aktif.
* **`PROPOSED`**: Rekomendasi desain yang sedang diuji coba.
* **`OPEN`**: Masih dalam perdebatan atau eksplorasi teknis (tercatat di `open-questions.md`).
