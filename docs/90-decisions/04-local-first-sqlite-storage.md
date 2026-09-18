# 04 — Model Penyimpanan Berbasis Kedaulatan Lokal (Local-First SQLite)

**Status:** ACCEPTED  
**Tanggal:** 2026-09-16  
**Pengambil Keputusan:** Core Team  

---

## 1. Konteks & Permasalahan

Niskava Agent membutuhkan media penyimpanan persisten untuk:
1. Riwayat sesi investigasi (`investigations`).
2. Titik anomali yang terdeteksi (`anomalies`).
3. Kartu temuan dan jejak bukti (*findings* & *evidence_items*).
4. Respons cache API Sectors (`sectors_cache`).

Pilihan arsitektur penyimpanan mencakup: database cloud terkelola (PostgreSQL / Supabase) vs penyimpanan lokal di mesin pengguna (*local-first file storage*).

---

## 2. Keputusan

**Mengadopsi pendekatan Local-First murni menggunakan SQLite database lokal yang disimpan pada direktori home pengguna: `~/.niskava/niskava.db`.**

Go Core Daemon berinteraksi dengan SQLite menggunakan driver murni Go (`modernc.org/sqlite`) yang dikonfigurasi dengan mode **Write-Ahead Logging (WAL)**:
```sql
PRAGMA journal_mode = WAL;
PRAGMA synchronous = NORMAL;
PRAGMA foreign_keys = ON;
```

---

## 3. Alternatif yang Dipertimbangkan

| Alternatif | Alasan Ditolak |
|---|---|
| **Cloud Managed DB (Supabase / Neon Postgres)** | Mengharuskan pengguna login/koneksi internet hanya untuk melihat riwayat investigasi lama, menambah biaya hosting, dan menimbulkan kekhawatiran privasi bagi investor institusi. |
| **JSON Filesystem Storage** | Sulit di-query dengan filter kompleks (misal: mencari anomali $V_z > 3.0$ dari seluruh sesi) dan rentan terhadap korupsi data saat konkurensi write. |
| **DuckDB** | Bagus untuk analitik OLAP masif, namun SQLite jauh lebih ringan, stabil, dan didukung native oleh driver Go tanpa CGO. |

---

## 4. Konsekuensi

### Positif
* **Privasi Mutlak**: Data portofolio, ticker yang diselidiki, dan catatan analisis pengguna tidak pernah meninggalkan laptop mereka.
* **100% Akses Offline**: Seluruh sesi masa lalu dapat diinspeksi kembali di CLI (`niskava sessions`) dan Web Dashboard tanpa internet.
* **Zero Infrastructure Cost**: Tidak ada tagihan server database bulanan.
* **Kemudahan Backup**: Pengguna cukup menyalin 1 file `~/.niskava/niskava.db`.

### Negatif / Kompromi yang Diterima
* Tidak ada sinkronisasi multi-device otomatis (misal: investigasi di laptop kantor tidak otomatis muncul di PC rumah kecuali file db disinkronkan manual).
