# 03 — Cakupan Produk & Antarmuka Pengguna (Surfaces)

**Status:** ACCEPTED  
**Versi Dokumen:** 1.0.0  
**Terakhir Diperbarui:** 2026-09-16  

Niskava Agent mengusung filosofi arsitektur **"Dual Surfaces, Single Engine"**:
1. **Terminal CLI** untuk kecepatan eksekusi, otomasi skrip, dan pengguna teknis (*power users*).
2. **Local Web Workspace** untuk eksplorasi interaktif, visualisasi grafik candlestick, timeline kejadian, dan kartu bukti.

Kedua antarmuka ditenagai oleh satu binary biner Go yang sama dan membaca database SQLite lokal yang sama.

---

## 1. Surface A: Terminal CLI (`niskava`)

Dibangun menggunakan Go (`spf13/cobra` untuk router command dan `charmbracelet/bubbletea` untuk TUI interaktif):

### Perintah Utama (CLI Commands)
```bash
# Menjalankan investigasi otonom terhadap emiten tertentu
niskava investigate ANTM --days 30

# Melihat daftar seluruh riwayat sesi investigasi lokal
niskava sessions

# Menampilkan kembali ringkasan sesi investigasi sebelumnya
niskava resume INV-2026-0042

# Menjalankan server web dashboard lokal dan membukanya otomatis di browser
niskava serve --port 8080 --open

# Menjalankan investigasi dalam mode offline (menggunakan mock/cache)
niskava investigate ANTM --offline
```

### Mockup Pengalaman Visual Terminal:
```text
$ niskava investigate ANTM --days 30

[●] NISKAVA INVESTIGATOR v1.0.0 — Target: ANTM (PT Aneka Tambang Tbk)
 ├── [1/4] Baseline Data Sectors v2 ............. [OK] 30 hari candle ditarik (Cache Hit)
 ├── [2/4] Deteksi Anomali Kuantitatif ......... [ALERT] Volume surge (3.84σ) pada 12 Sep
 ├── [3/4] Penelusuran OSINT Bertarget ......... [OK] 4 keterbukaan informasi & berita relevan
 └── [4/4] Validasi Bukti & Kausalitas ......... [OK] 3 temuan tervalidasi

─────────────────────────────────────────────────────────────────────────────
RINGKASAN TEMUAN (AUDIT TRAIL):
[SUPPORTED]   Lonjakan volume abnormal pada 12 Sep (184.5M lembar vs rata-rata 48.2M).
              Sumber: Sectors Daily API | Confidence: 1.00
[SUPPORTED]   Keterbukaan Informasi: Peresmian ekspansi smelter nikel baru di Halmahera Timur.
              Sumber: IDXnet / Sectors News | Causality: LIKELY_CATALYST | Confidence: 0.94
[UNCERTAIN]   Spekulasi pasar forum ritel terkait isu divestasi saham oleh induk holding.
              Sumber: Forum Komunitas (Unverified) | Causality: UNVERIFIED | Confidence: 0.35

─────────────────────────────────────────────────────────────────────────────
Sesi investigasi tersimpan sebagai: INV-2026-0042 (~/.niskava/niskava.db)
Ketik 'niskava serve --open' untuk membuka visual workspace interaktif di browser.
```

---

## 2. Surface B: Local Web Workspace (`localhost:8080`)

Frontend Single Page Application (SPA) modern yang dibangun dengan **Vite + React 18 + Tailwind CSS + shadcn/ui**, dikompilasi ke dalam biner Go menggunakan directive `//go:embed web/dist`.

### Komponen Kunci Web Dashboard:
1. **Header & Status Banner**: Menampilkan status koneksi agent, waktu investigasi, ticker aktif, dan tombol ekspor laporan (Markdown / JSON).
2. **Metrics & Anomaly Strip**: Kartu ringkasan cepat:
   * *Volume $Z$-Score* (misal: `+3.84σ` — Abnormal High)
   * *Abnormal Return* (misal: `+8.25%` vs Sektor `+0.45%`)
   * *Broker Dominance* (misal: Asing Net Buy IDR 84 Miliar)
3. **Interactive Financial Chart**:
   * Menampilkan grafik candlestick harga dan bar volume.
   * Pin anomali visual (*anomaly flag marker*) pada tanggal $T_{anomaly}$ yang dapat diklik untuk menyorot bukti terkait.
4. **Chronological Event Timeline**:
   * Urutan vertikal peristiwa: dari pergerakan volume awal $\rightarrow$ rilis pengumuman bursa $\rightarrow$ pemberitaan media massa $\rightarrow$ penutupan harga.
5. **Evidence Cards Matrix**:
   * Kartu temuan dengan penanda warna taksonomi:
     * Hijau (`SUPPORTED`): Data resmi Sectors / Pengumuman BEI.
     * Kuning (`UNCERTAIN`): Berita media pihak ketiga atau rumor pasar belum terverifikasi.
     * Merah (`CONTRADICTED`): Narasi yang terbantahkan oleh fakta laporan keuangan.
   * Setiap kartu memiliki tombol *"Lihat Sumber Asli"* dan indikator confidence score.
6. **Live SSE Terminal Drawer**:
   * Laci terminal tersembunyi di bagian bawah yang menampilkan stream log penalaran (*thinking steps*) agent secara real-time saat investigasi sedang diproses.
