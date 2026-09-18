# 04 — Evaluasi Akurasi & Benchmark Kasus Nyata (Evaluation & Benchmarks)

**Status:** ACCEPTED  
**Versi Dokumen:** 1.0.0  
**Terakhir Diperbarui:** 2026-09-16  

Dokumen ini mendefinisikan metodologi pengujian empiris, metrik evaluasi keandalan, dan kumpulan kasus uji acuan (*ground truth benchmark*) berbasis kejadian nyata di Bursa Efek Indonesia (IDX).

---

## 1. Kumpulan Kasus Uji Acuan (Ground Truth Benchmark Cases)

Untuk menguji apakah Niskava Agent bekerja secara objektif tanpa halusinasi, sistem diuji pada 4 skenario historis nyata:

| ID Kasus | Ticker Emiten | Tanggal Kejadian | Peristiwa Pasar Aktual | Ekspektasi Output Agen |
|---|---|---|---|---|
| **TC-01** | `ANTM` | September 2026 | Uji coba operasional fasilitas smelter feronikel Halmahera Timur disertai lonjakan volume 3.8x. | Menemukan anomali $V_z \ge 3.5$, melabeli temuan sebagai `SUPPORTED`, dan menandai kausalitas `LIKELY_CATALYST`. |
| **TC-02** | `BUMI` | Oktober 2022 | Lonjakan volume ekstrem sebelum keterbukaan informasi kuasi-reorganisasi dan penambahan modal tanpa HMETD. | Menandai $V_z \ge 4.0$, mengidentifikasi bahwa volume mendahului pengumuman resmi (`PRECEDED_ANNOUNCEMENT`). |
| **TC-03** | `GOTO` | Desember 2022 | Tekanan jual masif pasca-berakhirnya periode penguncian saham (*lock-up expiry*) pemegang saham awal. | Menemukan *abnormal negative return*, mendeteksi korelasi jadwal lock-up bursa, dan mengonfirmasi distribusi broker besar. |
| **TC-04** | `BBRI` | Februari 2024 | Pengumuman pembagian dividen tunai rekor di tengah rally sektor perbankan. | Mengklasifikasikan pergerakan sebagai `SECTOR_BETA_RALLY` dengan divergensi rendah, dan melabeli berita sebagai `SUPPORTED`. |

---

## 2. Metrik Evaluasi Keandalan Sistem

Untuk menjamin kualitas sebelum submisi hackathon, Niskava diukur menggunakan 4 metrik kuantitatif:

### A. Zero Numerical Hallucination (Target: 100%)
* **Definisi**: Seluruh angka harga penutupan, volume, moving average, rasio keuangan, dan persentase return dalam laporan akhir wajib cocok 100% dengan payload mentah Sectors API v2.
* **Mekanisme**: Komputasi deterministik di Python murni menjamin rasio kesalahan hitung aritmatika adalah **0.0%**.

### B. Akurasi Presedensi Waktu (Target: $\ge 90\%$)
* **Definisi**: Persentase ketepatan agen dalam mengurutkan apakah berita terbit sebelum transaksi (*catalyst*) atau sesudah transaksi (*potential leak*).

### C. Validitas Jejak Bukti / Citation Precision (Target: 100%)
* **Definisi**: Seluruh temuan yang dipublikasikan wajib memiliki cuplikan teks rujukan asli dan tautan dokumen sumber yang valid. Tidak boleh ada temuan yang menggantung tanpa bukti.

---

## 3. Harness Pengujian Otomatis (Automated Benchmark Runner)

Sistem pengujian otomatis dijalankan melalui modul evaluasi:
```bash
python -m tests.eval_benchmarks --cases TC-01,TC-02,TC-03,TC-04 --strict
```
Skrip ini memutar ulang rekaman data pasar lokal (`tests/fixtures/`) tanpa memotong kuota kredit API Sectors langsung, membandingkan output agen dengan label ekspektasi pakar (*expert ground-truth labels*).
