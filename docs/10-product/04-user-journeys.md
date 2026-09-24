# 04 — Alur Perjalanan Pengguna (User Journeys)

**Status:** ACCEPTED  
**Versi Dokumen:** 1.0.0  
**Terakhir Diperbarui:** 2026-09-16  

Dokumen ini memetakan alur interaksi pengguna ujung-ke-ujung (*end-to-end user journeys*) dalam berbagai skenario penggunaan nyata.

---

## 1. Journey 1: Investigasi Kilat via Terminal (The Fast Ticker Check)

* **Pengguna**: Arya (Retail Swing Trader).
* **Kondisi Awal**: Arya melihat ticker `ANTM` mendadak naik +8.2% dengan bar volume menjulang tinggi di aplikasi sekuritasnya.
* **Tujuan**: Memastikan apakah kenaikan ini didorong berita nyata sebelum memutuskan masuk posisi beli.

```text
[1] User mengeksekusi di terminal:
    $ niskava investigate ANTM --days 30

[2] CLI menampilkan animated spinner & progress bar:
    - Menarik data Sectors v2 (0.3s)
    - Menghitung anomali matematis deterministik (0.05s) -> Alert: Volume Z-Score 3.84x
    - Mengaktifkan News Harvester terarah (2.1s) -> 4 dokumen ditemukan
    - Menyusun korelasi waktu & validasi bukti (1.8s)

[3] Terminal mencetak ringkasan bukti terverifikasi:
    - [SUPPORTED] Katalis teridentifikasi: Pengumuman uji coba smelter feronikel.
    - [SUPPORTED] Aliran dana broker: Asing mencatatkan Net Buy IDR 84 Miliar.
    - [UNCERTAIN] Isu tambahan: Spekulasi dividen interim (belum ada konfirmasi bursa).

[4] Hasil:
    Arya memperoleh kepastian fakta dalam waktu < 5 detik tanpa perlu browsing manual.
```

---

## 2. Journey 2: Forensic Deep Dive via Local Web Workspace

* **Pengguna**: Clara (Junior Equity Research Associate).
* **Kondisi Awal**: Clara diminta membuat profil risiko dan kronologi anomali saham `ANTM` untuk dimasukkan ke laporan riset mingguan direksi.
* **Tujuan**: Menginspeksi visual grafik harga, timeline peristiwa, dan mengekspor kartu bukti.

```text
[1] User menjalankan server lokal:
    $ niskava serve --open
    Browser otomatis membuka: http://localhost:8080

[2] User memilih sesi 'INV-2026-0042 (ANTM)' dari panel riwayat sesi (Sidebar).

[3] Workspace menampilkan tampilan komprehensif:
    - Candlestick & Volume Chart dengan marker anomali tanggal 12 September.
    - Mengklik marker anomali langsung menyorot kartu bukti terkait di panel kanan.
    - Panel Timeline menyusun urutan kronologis peristiwa (pra-anomali hingga pasca-pengumuman).
    - Membaca kartu bukti dengan label warna transparan (SUPPORTED / UNCERTAIN).

[4] User mengklik tombol "Export Markdown Memo".

[5] Hasil:
    Clara mendapatkan dokumen rapi berisi ringkasan temuan dan tabel referensi yang siap dilampirkan ke laporan sekuritas.
```

---

## 3. Journey 3: Pengecekan Fakta Rumor Pasar (The Fact-Check Journey)

* **Pengguna**: Dimas (Jurnalis Finansial).
* **Kondisi Awal**: Beredar rumor di forum saham bahwa sebuah emiten terancam bangkrut karena gagal bayar kupon obligasi.
* **Tujuan**: Memverifikasi apakah klaim tersebut didukung oleh data laporan keuangan resmi atau merupakan disinformasi pasar.

```text
[1] User menjalankan investigasi berbasis profil kesehatan finansial:
    $ niskava investigate [TICKER] --skill financial-health

[2] Sistem menarik data laporan keuangan triwulanan dari Sectors API v2:
    - Cash and cash equivalents
    - Current Ratio & Quick Ratio
    - Debt-to-Equity Ratio (DER)
    - Tren laba bersih dan operating cash flow

[3] Niskava menghasilkan output verifikasi:
    - Status: [CONTRADICTED]
    - Klaim Pasar: "Emiten terancam gagal bayar kewajiban jangka pendek."
    - Bukti Sanggahan: Neraca Sectors v2 menunjukkan kas lancar IDR 4.2 Triliun dengan rasio lancar (Quick Ratio) 2.1x, membantah spekulasi kebangkrutan.

[4] Hasil:
    Dimas menulis artikel pelurusan fakta berbasis data empiris terpercaya.
```
