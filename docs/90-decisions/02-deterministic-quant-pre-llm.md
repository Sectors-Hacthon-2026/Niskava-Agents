# 02 — Komputasi Deterministik Kuantitatif Sebelum LLM

**Status:** ACCEPTED  
**Tanggal:** 2026-09-16  
**Pengambil Keputusan:** Core Team  
**Menggantikan:** `docs/decisions/deterministic-quant.md`  

---

## 1. Konteks & Permasalahan

Pendekatan umum pada aplikasi AI finansial generasi awal adalah menyuntikkan seluruh tabel data harga dan volume deret waktu (30–90 hari baris angka) langsung ke dalam prompt LLM dan meminta model:
> *"Tolong baca tabel harga di atas dan temukan tanggal berapa saja yang mengalami lonjakan volume tidak wajar."*

Pendekatan ini memiliki kelemahan fatal:
1. **Halusinasi Aritmatika**: Model bahasa (LLM) pada dasarnya adalah *next-token predictor*, bukan kalkulator statistik. LLM sering salah menghitung deviasi standar dan rata-rata bergerak.
2. **Pemborosan Biaya & Kuota Token**: Memasukkan ratusan baris data angka menghabiskan 3.000–6.000 token konteks per prompt.
3. **Latensi Lambat**: Membaca konteks teks yang panjang menambah latensi respons hingga 5–10 detik.

---

## 2. Keputusan

**Seluruh perhitungan statistik time series (moving average volume 20 hari, standar deviasi, volume $Z$-score, return harian, dan divergensi sektor) wajib dieksekusi secara deterministik menggunakan script matematika Python (`numpy`) sebelum LLM dipanggil.**

Model bahasa (LLM) hanya diaktifkan **setelah** titik anomali terbukti secara matematis. LLM hanya menerima payload ringkas berupa tanggal anomali yang valid, nilai metrik deviasi, dan teks berita untuk dianalisis konteks kualitatifnya.

---

## 3. Alternatif yang Dipertimbangkan

| Alternatif | Alasan Ditolak |
|---|---|
| **LLM Direct Reasoning (Zero Python Math)** | Akurasi numerik sangat rendah (<65%), rawan salah mengidentifikasi tanggal lonjakan volume. |
| **LLM with Code Interpreter (Sandboxed Python by LLM)** | Terlalu lambat (menambah round-trip eksekusi kode 5-15 detik) dan memerlukan infrastruktur sandbox container yang rumit. |
| **Database-Side Calculation (SQL Window Functions)** | Keterbatasan fungsi statistik standar di SQLite bawaan tanpa ekstensi matematika CGO. |

---

## 4. Konsekuensi

### Positif
* **100% Akurasi Numerik**: Menjamin tidak ada kesalahan hitung angka atau tanggal anomali palsu (*zero mathematical hallucination*).
* **Hemat Token Hingga 70%**: Ukuran prompt berkurang drastis karena hanya mengirimkan parameter anomali ringkas.
* **Kecepatan Eksekusi Seketika**: Komputasi numerik Python selesai dalam tempo $< 15$ milidetik.

### Negatif / Kompromi yang Diterima
* Formula anomali harus didefinisikan secara eksplisit di awal (misal ambang batas $V_z \ge 2.5$) dan tidak adaptif secara dinamis kecuali diatur ulang parameternya.
