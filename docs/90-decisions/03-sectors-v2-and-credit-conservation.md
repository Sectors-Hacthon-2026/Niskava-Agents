# 03 — Adopsi Penuh Sectors API v2 & Strategi Konservasi Kredit

**Status:** ACCEPTED  
**Tanggal:** 2026-09-16  
**Pengambil Keputusan:** Core Team  
**Dokumen Terkait:** `docs/20-architecture/03-sectors-v2-api.md`, `docs/10-product/05-hackathon-strategy.md`

---

## 1. Konteks & Permasalahan

Dalam ajang Sectors Hackathon Indonesia 2026:
1. Penyelenggara mewajibkan penggunaan **Sectors Financial API v2** (`https://api.sectors.app/v2/`) sebagai pilar data inti (Rule 06).
2. Setiap tim peserta diberikan kuota terbatas sejumlah **1.000 API Credits**.
3. Jika kuota ini habis sebelum proses perekaman video demo submisi dan evaluasi langsung (*asynchronous judging*) oleh dewan juri selesai, proyek akan gagal dinilai.

Sistem membutuhkan spesifikasi integrasi yang memanfaatkan seluruh kapabilitas mutakhir API v2 (termasuk fitur baru yang krusial untuk intelijen pasar modal Indonesia) sekaligus menjamin kuota kredit terlindungi secara disiplin.

---

## 2. Keputusan

1. **Adopsi Penuh Fitur Kunci API v2 Indonesia**:
   * **Transaksi & Harga**: `GET /v2/daily/{symbol}/` (OHLCV s.d 90 hari) dan `GET /v2/daily-close/`.
   * **Fundamental Terarah**: `GET /v2/company/report/{symbol}/?sections=valuation,financials` untuk meminimalkan ukuran payload dan latensi.
   * **Intelijen Regulasi & Suspensi**: `GET /v2/suspensions/?symbol={symbol}` untuk melacak riwayat suspensi bursa beserta tautan dokumen PDF resmi BEI.
   * **Transaksi Orang Dalam (Insider Trading)**: `GET /v2/filings/?symbol={symbol}` untuk mendeteksi transaksi direksi/komisaris saat anomali volume terjadi.
   * **Arus Modal Asing (Foreign Flow)**: `GET /v2/foreign-flow/{symbol}/` untuk deret Net Foreign Inflow harian guna mendeteksi anomali akumulasi asing ($F_z$).
   * **Aksi Korporasi**: `GET /v2/corporate-actions/{symbol}/` untuk konfirmasi cum-date dividen, rights issue, dan stock split.
   * **Berita Pasar Modal**: `GET /v2/news/?ticker={symbol}` sebagai sumber berita primer terkurasi.
   * **Ekstensi Komoditas**: `GET /v2/commodity-price/{commodity}/` dan `/v2/mining-company-detail/{slug}/`.
2. **Implementasi Local SQLite Caching Layer**:
   * Setiap respons dari Sectors di-cache di tabel `sectors_cache` (`cache_key = SHA256(endpoint + params)`).
   * Data historis harga dan transaksi masa lampau ($T < \text{hari ini}$) bersifat permanen (`expires_at = NULL`), sehingga pemanggilan berulang berbiaya **0 kredit**.
3. **Penyediaan Mode Mock Runner Deterministik**:
   * Sistem menyediakan mode pengujian offline (`MOCK_SECTORS=1`) dengan payload JSON statis di `tests/fixtures/sectors/` untuk pengujian CI/CD dan unit test tanpa memotong 1 kredit pun.

---

## 3. Alternatif yang Dipertimbangkan

| Alternatif | Alasan Ditolak |
|---|---|
| **Direct Pass-through (Tanpa Cache)** | Menghabiskan 1.000 kredit dalam 1–2 hari pengujian berulang dan membuat sistem rentan gagal saat demo jika API lambat. |
| **In-Memory Cache Saja (Redis / Python Dict)** | Cache hilang setiap kali proses CLI selesai dieksekusi (*ephemeral*), tidak persisten antar pemanggilan perintah terminal. |
| **Bulk Pre-download Seluruh Semesta IDX** | Menghabiskan ratusan kredit di awal untuk saham-saham yang tidak pernah diinvestigasi oleh pengguna. |
| **Hanya Mengambil OHLCV Saja** | Melewatkan data berharga tinggi seperti suspensi bursa resmi, transaksi insider, dan foreign flow yang membedakan Niskava dari chatbot biasa. |

---

## 4. Konsekuensi

### Positif
* **Disiplin Anggaran 1.000 Kredit**: Satu sesi investigasi mendalam ANTM hanya menghabiskan 5 kredit pada *cache miss* pertama, dan 0 kredit pada pengulangan. Lebih dari 850 kredit aman tersisa untuk evaluasi juri.
* **Kedalaman Bukti Tingkat Tinggi**: Adanya tautan dokumen PDF resmi BEI dari `/v2/suspensions/` dan data kepemilikan orang dalam dari `/v2/filings/` memperkuat skor *Technical Depth (30%)*.
* **Resiliensi Jaringan**: Aplikasi tetap dapat mendemokan investigasi saham yang sudah di-cache meskipun koneksi internet terputus di tengah jalan.
* **Kecepatan Sub-Milidetik**: Akses lokal dari SQLite mengembalikan data dalam tempo $< 5$ milidetik.

### Negatif / Kompromi yang Diterima
* Memerlukan validasi TTL berbeda: 15 menit untuk data bursa berjalan (*real-time quote*), 24 jam untuk filings, dan 7 hari untuk laporan keuangan triwulanan.
