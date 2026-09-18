# 03 — Glosarium Istilah (Glossary)

**Status:** ACCEPTED  
**Versi Dokumen:** 1.0.0  
**Terakhir Diperbarui:** 2026-09-16  

Daftar istilah standar domain pasar modal Indonesia (IDX), rekayasa kecerdasan buatan (*AI Agents*), dan intelijen sumber terbuka (*OSINT*) yang digunakan dalam arsitektur dan dokumentasi Niskava Agent.

---

## 1. Domain Pasar Modal & Keuangan Indonesia (IDX Domain)

| Istilah | Definisi & Konteks Penggunaan |
|---|---|
| **Ticker (Kode Saham)** | Simbol 4 huruf pengenal perusahaan terbuka di Bursa Efek Indonesia (misal: `ANTM`, `BBRI`, `BUMI`). |
| **OHLCV** | Deret data perdagangan harian yang terdiri dari *Open*, *High*, *Low*, *Close*, dan *Volume*. |
| **Volume $Z$-Score ($V_z$)** | Ukuran statistik deviasi standar volume perdagangan harian terhadap rata-rata 20 hari bursa sebelumnya. Nilai $V_z \ge 2.5$ menandakan lonjakan volume luar biasa. |
| **Abnormal Return** | Perubahan harga saham harian yang menyimpang secara signifikan ($|R_t| \ge 5\%$) dibandingkan tren normal. |
| **Idiosyncratic Movement** | Pergerakan saham yang didorong oleh faktor internal spesifik emiten, dibuktikan dengan divergensi tajam terhadap indeks rata-rata subsektor industrinya. |
| **Bandarmology / Broker Summary** | Analisis struktur transaksi pihak pembeli dan penjual terbesar (broker) untuk mendeteksi akumulasi atau distribusi saham oleh institusi besar / asing. |
| **Foreign Flow (NBSA)** | *Net Buy/Sell Asing*, akumulasi selisih nilai transaksi beli dan jual oleh investor non-domestik. |
| **Keterbukaan Informasi (IDXnet)** | Pengumuman resmi wajib yang disampaikan oleh emiten kepada Bursa Efek Indonesia mengenai aksi korporasi, litigasi hukum, atau peristiwa material. |
| **UMA (Unusual Market Activity)** | Pengumuman pengawasan bursa atas pergerakan harga/volume saham yang bergerak di luar kebiasaan tanpa penjelasan material yang memadai. |
| **Mining Extension (Sectors)** | Modul data operasional spesifik komoditas di Sectors v2 yang mencatat volume produksi nikel/emas/batubara dan estimasi cadangan tambang. |
| **OJK (Otoritas Jasa Keuangan)** | Lembaga independen pengawas dan pengatur industri jasa keuangan dan pasar modal di Indonesia. |

---

## 2. Domain AI Agent & Open Source Intelligence (OSINT)

| Istilah | Definisi & Konteks Penggunaan |
|---|---|
| **Autonomous Agent** | Program perangkat lunak cerdas yang mampu merencanakan langkah (*planning*), memilih instrumen (*tool use*), dan mengeksekusi investigasi hingga tuntas tanpa intervensi manual setiap langkah. |
| **Evidence-First** | Paradigma arsitektur di mana setiap klaim harus diverifikasi terhadap fakta data terstruktur sebelum model bahasa (LLM) diizinkan menyimpulkan. |
| **Evidence Gap** | Kesenjangan antara temuan anomali kuantitatif (misal: volume meledak 3.8x) dengan ketiadaan penjelasan resmi pada data fundamental saat ini. |
| **Targeted OSINT** | Teknik pencarian data publik (berita, pengumuman bursa, laporan analis) yang dibatasi secara ketat pada jendela waktu anomali ($T_{anomaly} \pm 2\text{ hari}$) menggunakan query spesifik. |
| **Ground Truth** | Titik acuan fakta mutlak dalam sistem (dalam hal ini data numerik resmi dari Sectors API v2). |
| **Hallucination Mitigation** | Metode perlindungan sistem dari kesalahan fabrikasi fakta oleh LLM, dicapai dengan pemrosesan deterministik sebelum LLM dan taksonomi verifikasi bukti. |
| **Confidence Score** | Skor probabilitas validitas temuan (skala 0.00 hingga 1.00) yang dihitung berdasarkan kualitas dan konsistensi sumber bukti. |
| **Causality Label** | Label relasi waktu antara berita dan pergerakan pasar: `LIKELY_CATALYST` (berita memicu volume), `PRECEDED_ANNOUNCEMENT` (volume mendahului pengumuman), atau `UNEXPLAINED_BY_NEWS`. |
| **Verification Status** | Klasifikasi validitas temuan: `SUPPORTED` (didukung data resmi), `UNCERTAIN` (rumor/indikasi awal), atau `CONTRADICTED` (fakta membantah narasi pasar). |
| **IPC (Inter-Process Communication)** | Jalur komunikasi data streaming antara proses Go Core daemon dan Python Agent Engine melalui JSON Lines via STDIN/STDOUT. |
| **SSE (Server-Sent Events)** | Protokol HTTP streaming satu arah untuk mengirimkan pembaruan langkah pemikiran agen (*reasoning steps*) secara real-time dari Go server ke Web UI. |
