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

Ketika pelaku pasar mencoba memanfaatkan AI generatif / Chatbot umum (seperti ChatGPT atau Gemini biasa), mereka menemui kegagalan fundamental:
1. **Pasif & Bergantung pada Prompt**: Hanya merangkum apa yang ditanyakan user tanpa inisiatif memverifikasi validitas data deret waktu harga.
2. **Kerapuhan Halusinasi (Hallucination Vulnerability)**: Sering mengaitkan berita usang dari tahun lalu dengan lonjakan harga hari ini (korelasi semu / *spurious correlation*).
3. **Ketiadaan Jejak Bukti (*No Evidence Audit Trail*)**: Menghasilkan teks opini yang terdengar meyakinkan tanpa referensi sumber resmi, stempel waktu (*timestamp*), atau pembedaan antara fakta regulasi vs rumor pasar.

---

## 2. Tesis & Pendekatan Niskava: Evidence-First Orchestration

Niskava Agent membalik paradigma dari *"Generative-First"* menjadi **"Evidence-First Autonomous Investigation"**:

```
Pendekatan Chatbot Konvensional:
User Query ──▶ Pencarian Web Bebas ──▶ Sintesis LLM Bebas ──▶ Risiko Halusinasi Tinggi

Pendekatan Niskava Agent:
User Query ──▶ Sectors v2 Ground Truth ──▶ Deteksi Anomali Deterministik
           ──▶ Evidence Gap Formulasi  ──▶ Targeted OSINT Harvester 
           ──▶ Cross-Verification Matrix ──▶ Audit Trail [SUPPORTED | UNCERTAIN | CONTRADICTED]
```

### Pilar Utama Pendekatan Niskava:
1. **Quantitative Ground Truth (Sectors API v2)**: Data harga harian, rasio valuasi, metrik fundamental, dan broker summary dari Sectors diperlakukan sebagai fakta dasar yang tak terbantahkan.
2. **Deterministic Before Generative**: Perhitungan anomali statistik ($Z$-score volume, abnormal return, divergensi sektoral) dihitung secara pasti melalui matematika murni sebelum model bahasa (LLM) diaktifkan.
3. **Evidence Gap Detection**: Agen secara otomatis merumuskan pertanyaan investigatif ketika terdapat gap antara pergerakan angka kuantitatif dengan ketersediaan informasi pasar.
4. **Targeted OSINT Harvesting**: Mesin pencarian tidak mencari informasi secara acak, melainkan mengunci jendela waktu anomali ($T_{anomaly} \pm 2\text{ hari}$) dengan kata kunci terstruktur.
5. **Causality vs Correlation Awareness**: Agen secara ketat membedakan apakah suatu berita memicu pergerakan volume (*Likely Catalyst*) atau volume melonjak mendahului pengumuman resmi (*Preceded Announcement / Potential Leak*).

---

## 3. Matriks Perbandingan: Chatbot Biasa vs Niskava Agent

| Dimensi Evaluasi | Chatbot Finansial Umum | Niskava Agent |
|---|---|---|
| **Cara Kerja Inti** | Meringkas teks jawaban (*Summarization*) | Menyelidiki anomali pasar (*Active Investigation*) |
| **Penanganan Angka Pasar** | Membaca deret angka via prompt LLM (rawan salah hitung) | Dihitung deterministik di Python ($Z$-Score, MA20) sebelum ke LLM |
| **Sumber Data Pasar** | Scraping web bebas / pencarian Google umum | Data resmi terstruktur via **Sectors API v2** |
| **Audit Trail Bukti** | Teks paragraf tanpa status verifikasi | Taksonomi bukti terstruktur (`SUPPORTED`, `UNCERTAIN`, `CONTRADICTED`) |
| **Kepatuhan Regulasi** | Rentan terpeleset memberikan saran *Buy/Sell* ilegal | *Strict Non-Advisory Guardrail* (hanya intelijen faktual & edukasi) |
| **Antarmuka Pengguna** | Kotak chat teks tunggal | Terminal CLI interaktif & Local Web Workspace dengan chart interaktif |
| **Kepemilikan Data** | Cloud vendor chat history | Database SQLite lokal (`~/.niskava/niskava.db`), *Local-First* |
