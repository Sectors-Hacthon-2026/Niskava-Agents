# 01 — Visi & Tesis: Autonomous Financial OSINT Agent

**Status:** LOCKED  
**Versi Dokumen:** 1.0.0  
**Terakhir Diperbarui:** 2026-09-16  

> **"Don't just answer questions. Investigate them."**  
> Niskava Agent adalah platform orkestrasi intelijen pasar modal dan investigasi finansial otonom berbasis *evidence-first*, menggabungkan data kuantitatif pasar **Sectors API v2** dengan sinyal eksternal **OSINT (News, Public Filings, Corporate Actions)**.

---

## 1. Latar Belakang Masalah

Di pasar modal Indonesia (IDX), data kuantitatif bursa dan konteks kualitatif berada di tempat yang terpisah secara ekstrem:
* **Trader & Investor**: Melihat lonjakan harga atau volume abnormal di chart trading, tetapi terpaksa melakukan pencarian manual di Google, media sosial, atau forum saham untuk mencari tahu katalis penyebabnya.
* **Analis & Riset Ekuitas**: Membaca berita atau rumor aksi korporasi, namun kesulitan memverifikasi apakah sentimen tersebut didukung oleh aliran dana riil (*broker flow*) atau laporan fundamental.

Ketika pelaku pasar mencoba memanfaatkan AI generatif / Chatbot umum (seperti ChatGPT atau Gemini biasa tanpa arsitektur agen spesialis), mereka menemui kegagalan fundamental:
1. **Pasif & Bergantung pada Prompt**: Hanya merangkum apa yang ditanyakan user tanpa inisiatif memverifikasi validitas data deret waktu harga.
2. **Kerapuhan Halusinasi (Hallucination Vulnerability)**: Sering mengaitkan berita usang dari tahun lalu dengan lonjakan harga hari ini (korelasi semu / *spurious correlation*).
3. **Ketiadaan Jejak Bukti (*No Evidence Audit Trail*)**: Menghasilkan teks opini yang terdengar meyakinkan tanpa referensi sumber resmi, stempel waktu (*timestamp*), atau pembedaan antara fakta regulasi vs rumor pasar.

---

## 2. The Anti-Wrapper Manifesto: 5 Dosa Besar AI Wrapper Finansial

Sebagian besar produk "AI Finansial" di pasar saat ini hanyalah **wrapper tipis (*thin wrapper*)** di atas model bahasa besar (LLM). Wrapper jenis ini berbahaya jika diterapkan pada pasar modal karena 5 anti-pattern fundamental:

1. **Dosa 1: Raw JSON Dump & Prompt Stuffing**  
   Wrapper menarik ratusan baris candlestick mentah lalu menjejalkannya utuh ke prompt LLM: *"Apa pendapatmu tentang saham ANTM?"*.  
   *Dampak:* Model kehabisan context window, terdistraksi kebisingan (*noise*), dan menguras ribuan token secara sia-sia.
2. **Dosa 2: Mental Math & Numerical Hallucination**  
   Wrapper membiarkan LLM menghitung rata-rata bergerak, persentase return, volatilitas, atau Z-score secara mental.  
   *Dampak:* LLM bukan kalkulator numerik floating-point; angka yang dihasilkan sering fiktif tetapi disajikan dengan nada meyakinkan (*confident hallucination*).
3. **Dosa 3: Naive Non-Causal RAG (Atemporal Search)**  
   Wrapper melakukan pencarian semantik (vector database) tanpa kesadaran waktu (*temporal blindness*).  
   *Dampak:* Berita yang terbit 2 minggu *setelah* harga saham melonjak dianggap sebagai "katalis pemicu" lonjakan harga (*causality inversion*).
4. **Dosa 4: Monolithic Hardcoded Prompt (Ketiadaan Skills & MCP)**  
   Wrapper menjejalkan seluruh instruksi analisis ke dalam satu system prompt monolitik tanpa membedakan *Low-Level Primitive Tools* (I/O data bursa) dengan *Domain Skills* (SOP metodologi analis ekuitas).  
   *Dampak:* Sistem kaku, tidak dapat diaudit per tahap, dan gagal melakukan *backtracking* saat investigasi menemui jalan buntu.
5. **Dosa 5: Rekomendasi Spekulatif Ilegal (*Unregulated Financial Advice*)**  
   Wrapper sering tergoda memprediksi harga masa depan atau memberi saran *BUY/SELL* tanpa audit trail.  
   *Dampak:* Melanggar regulasi bursa (POJK) dan membahayakan modal pelaku pasar.

---

## 3. Tesis & Pendekatan Niskava: Evidence-First Autonomous Orchestration

Niskava Agent membalik paradigma dari *"Generative-First"* menjadi **"Evidence-First Autonomous Investigation"** bergaya Hermes / OpenCode yang dispesialisasi untuk pasar modal:

```
Pendekatan AI Wrapper Biasa:
User Query ──▶ Dump Raw JSON ke Prompt ──▶ Mental Math LLM ──▶ Halusinasi & Opini Spekulatif

Pendekatan Niskava Agent (4-Layer Hierarchy):
User Query ──▶ Layer 4: ReAct Cognitive Orchestrator
                 │ Memilih SOP Analisis yang Relevan
                 ▼
               Layer 3: Modular Skills Catalog (Domain SOPs)
                 │ Memanggil Tools & Kalkulasi Deterministik
                 ▼
               Layer 2: Deterministic Compute Gate (NumPy Firewall: Vz, Rt, Fz)
                 │ Protokol I/O Data Bursa & Web
                 ▼
               Layer 1: MCP & OSINT Primitives (Sectors MCP + Dual OSINT)
                 │
                 ▼
Evidence Layer: Cross-Verification ──▶ Audit Trail [SUPPORTED | UNCERTAIN | CONTRADICTED]
```

### Pilar Utama Pendekatan Niskava:
1. **Quantitative Ground Truth (Sectors API v2 & MCP)**: Data harga harian, rasio valuasi, metrik fundamental, dan broker summary dari Sectors diperlakukan sebagai fakta dasar yang tak terbantahkan.
2. **Deterministic Before Generative (Law 1)**: Perhitungan anomali statistik ($Z$-score volume, abnormal return, divergensi sektoral) dihitung secara pasti melalui matematika NumPy sebelum model bahasa (LLM) diaktifkan.
3. **Modular Domain Skills**: Investigasi tidak dijalankan serampangan, melainkan mengikuti Standar Operasional Prosedur (SOP) analis ekuitas terisolasi (`market-anomaly-recon`, `event-causality-audit`, `insider-bandarmology-forensic`).
4. **Targeted Temporal OSINT Harvesting**: Mesin pencarian mengunci jendela waktu anomali ($T_{anomaly} \pm 2\text{ hari}$) dengan kata kunci terstruktur untuk mencegah pembalikan kausalitas.
5. **Causality vs Correlation Awareness**: Agen secara ketat membedakan apakah suatu berita memicu pergerakan volume (*Likely Catalyst*) atau volume melonjak mendahului pengumuman resmi (*Preceded Announcement / Potential Leak*).

---

## 4. Matriks Komparasi: AI Wrapper Tipis vs Niskava Agent

| Dimensi Evaluasi | AI Wrapper Finansial Biasa | Niskava Agent (Hermes/OpenCode Paradigm) |
|---|---|---|
| **Paradigma Kerja** | Chatbot responsif pasif (*Prompt-in, Text-out*) | Agen investigasi otonom (*Autonomous ReAct Loop*) |
| **Arsitektur Tooling** | Fungsi Python monolitik di-hardcode ke prompt | **Layering 4 Tingkat:** MCP Primitives $\to$ Compute Gate $\to$ Skills $\to$ ReAct |
| **Kalkulasi Numerik** | Dihitung di dalam pikiran LLM (*Mental Math*) | **Deterministic Gate (NumPy):** LLM dilarang berhitung (Law 1) |
| **Konsumsi Token** | Memasukkan seluruh riwayat data mentah (boros token) | Hanya mengonsumsi metrik ringkas terverifikasi hasil gate |
| **Pencarian Informasi** | Semantic vector search tanpa filter tanggal | **Temporal-Aware OSINT:** Jendela $T \pm 2$ hari + stempel waktu bursa |
| **Status Temuan** | Opini teks bebas tanpa klasifikasi pembuktian | **3-Tier Taxonomy:** `SUPPORTED`, `UNCERTAIN`, `CONTRADICTED` |
| **Kepatuhan Regulasi** | Rentan terpeleset memberikan saran *Buy/Sell* ilegal | **Strict Non-Advisory (Law 2):** Platform intelijen & audit trail bukti |
| **Antarmuka Pengguna** | Kotak chat sederhana | **Dual Surfaces:** Terminal REPL (TUI Glamour) + Local Web Workspace |
| **Kedaulatan Data** | Riwayat disimpan di cloud vendor pihak ketiga | **Local-First (Law 4):** Database SQLite lokal (`~/.niskava/niskava.db`) |
