# Audit Komprehensif Kesenjangan Fitur & Roadmap Implementasi Web Dashboard — Niskava Agent

> **Single Source of Truth (SSoT) Kesiapan Antarmuka Web Workspace Dashboard**  
> **Target:** Pengembang Web/Frontend (Nabil dkk.), Core Backend Engineer, & Evaluasi Hackathon Indonesia 2026.  
> **Status Basis Data & REST/SSE Server:** 100% Siap (Seluruh Endpoint `/api/...` Aktif & Teruji).

---

> [!CAUTION]
> ## 🚨 ARAHAN MUTLAK TIM: PEMBEKUAN BACKEND (BE CODE FREEZE)
> 1. **DILARANG KERAS MENYENTUH KODE BACKEND:**  
>    Seluruh direktori backend (`backend/core/`, `backend/engine/`, `backend/core/db/`, `backend/core/server/`, `backend/core/ipc/`) berstatus **STABIL, TERUJI, & LENGKAP**. Tim Frontend Web dilarang melakukan modifikasi, penambahan logic, atau refactoring pada berkas backend Go maupun Python Engine secara mandiri/sepihak.
> 2. **JALUR KOMUNIKASI & KOORDINASI WAJIB VIA IKHSAN:**  
>    Jika terdapat kendala kontrak data, kebutuhan penyesuaian event SSE, atau bug pada layer backend, **ANGGOTA TIM WAJIB MENGABARI DAN MENGHUBUNGI IKHSAN** terlebih dahulu sebelum melakukan tindakan apa pun. Penyesuaian backend hanya boleh dilakukan melalui koordinasi langsung bersama Ikhsan untuk mencegah regresi arsitektur, kerusakan 272 suite uji coba otomatis, atau pelanggaran terhadap 6 Hukum Niskava & Aturan Resmi Hackathon.

---

## 1. Executive Summary & Audit Scope

Dokumen ini memuat audit teknis mendalam terhadap antarmuka grafis berbasis web (**Web Workspace Dashboard**) pada repositori **Niskava Agent**. Audit mengevaluasi secara kritis kesenjangan (*gaps*) antara endpoint REST dan Server-Sent Events (SSE) yang telah dibangun oleh **Backend (BE)** pada Go Core Server (`backend/core/server/server.go`) dan basis data SQLite (`backend/core/db/`) dengan apa yang saat ini telah terhubung pada antarmuka web.

### Kondisi Arsitektur Web Saat Ini:
Terdapat **dua representasi permukaan web** di dalam repositori:
1. **`clients/web/` (Target Arsitektur React SPA):**  
   Dirancang menggunakan React 18, Vite, TypeScript, dan Tailwind CSS sesuai panduan integrasi (`docs/TEAM_CLIENTS_INTEGRATION_GUIDE.md`).  
   *Status:* Masih berupa kerangka kosong (*skeleton*). Berkas `clients/web/src/App.tsx` hanya berisi `return null;` dan `clients/web/package.json` belum memiliki dependensi terpasang (`"dependencies": {}`).
2. **`backend/core/server/workspace.html` & `clients/web/index.html` (Active Embedded Workspace):**  
   Aplikasi web mandiri (*single-file SPA*) berbasis HTML5, CSS kustom (Binance Institutional Dark/Light Palette), dan Vanilla JavaScript murni berukuran 4.249 baris (196 KB). Berkas ini tersemat langsung ke dalam biner kompilasi Go melalui `//go:embed` dan dilayani secara default pada `http://localhost:20128`.  
   *Status:* Menjadi antarmuka web aktif yang saat ini berjalan ketika pengguna mengeksekusi `niskava serve` atau memilih menu "Web Workspace" pada CLI launcher.

Audit ini berfokus pada **fitur-fitur backend yang sudah dibuat namun belum diimplementasikan atau belum dihubungkan pada Web Workspace aktif**, serta rekomendasi arsitektur penyelesaiannya.

---

## 2. Matrix Kesenjangan Fitur: Backend API vs Web Dashboard

Berikut adalah pemetaan komprehensif antara REST/SSE endpoints yang telah disediakan Backend dengan kondisi integrasi pada Web Workspace:

| Endpoint Backend (`server.go`) | Metode | Deskripsi Fitur Backend | Status di Web Workspace | Urgensi | Catatan Teknis / Gaps |
|---|---|---|---|---|---|
| **Panel Samping Kanan (`#insightPanel`)** | UI DOM | Visualisasi data pasar, sinyal anomali & timeline aktif | 🔴 **Data Hardcoded** | **P0 (Kritis)** | Berisi data mockup statik ANTM / Nikel / IHSG buatan yang tidak pernah diperbarui secara dinamis dari API. |
| `/api/investigations` | `GET` | Mengambil daftar riwayat audit pipeline 7-tahap | ❌ **Belum Ada** | **P0 (Kritis)** | Tidak ada Halaman/Tab "Investigations" di web; pengguna tidak bisa melihat histori audit formal. |
| `/api/investigations` | `POST` | Memicu eksekusi investigasi otonom 7-tahap pada ticker | ❌ **Belum Ada** | **P0 (Kritis)** | Tidak ada tombol/modal untuk memicu pipeline terstruktur (misal: ANTM 30 hari) dari antarmuka web. |
| `/api/investigations/{id}` | `GET` | Mengambil dossier lengkap (anomali, bukti, temuan, timeline) | ❌ **Belum Ada** | **P0 (Kritis)** | Data dossier investigasi tidak memiliki representasi visual di web. |
| `/api/investigations/{id}/anomalies` | `GET` | Mengambil metrik anomali kuantitatif terdeteksi | ❌ **Belum Ada** | **P1 (Tinggi)** | Metrik deviasi statistik hanya muncul di streaming text, tidak ada tabel ringkasan anomali per sesi. |
| `/api/investigations/{id}/findings` | `GET` | Mengambil temuan terverifikasi 3-tier taxonomy | ❌ **Belum Ada** | **P1 (Tinggi)** | Temuan tidak terorganisir ke dalam matriks verifikasi terstruktur (*Evidence Matrix*). |
| `event: anomaly_detected` (SSE) | SSE | Data kuantitatif deviasi volume Z-Score & harga | ⚠️ **Parsial** | **P0 (Kritis)** | **Ketiadaan Chart:** Hanya dirender sebagai 1 item teks di right-panel; tidak ada candlestick chart harga & bar volume. |
| `/api/chat/sessions/{id}/fork` | `POST` | Percabangan percakapan (*branching/fork*) ke sesi baru | ❌ **Belum Ada** | **P1 (Tinggi)** | Pola OpenCode branching percakapan belum memiliki tombol aksi pada bubble chat. |
| `/api/chat/search` | `GET` | Pencarian teks pesan global di seluruh riwayat SQLite | ❌ **Belum Ada** | **P1 (Tinggi)** | Input pencarian di sidebar hanya memfilter DOM lokal judul sesi yang termuat, tidak memanggil endpoint pencarian global. |
| `/api/chat/sessions/{id}` | `PATCH` | Mengubah judul sesi (*rename*) atau status pin (*is_pinned*) | ❌ **Belum Ada** | **P2 (Sedang)** | Pengguna tidak dapat me-rename judul sesi dari antarmuka web atau mem-pin sesi favorit ke atas sidebar. |
| `/api/chat/sessions/{id}/reset` | `POST` | Mengosongkan riwayat pesan dalam sesi aktif | ⚠️ **Parsial** | **P2 (Sedang)** | Tombol di UI hanya "New Research" (buat sesi baru), tidak ada tombol "Reset Current Session". |
| `/api/graph/data` | `GET` | Data nodes & edges memori asosiatif lokal | ⚠️ **Parsial** | **P1 (Tinggi)** | Ticker filter sudah terpasang, namun tombol pembersihan data uji coba (`PruneMockTestData`) belum ada. |
| `/api/graph/stats` | `GET` | Statistik agregasi graf dan top hub nodes | ⚠️ **Parsial** | **P2 (Sedang)** | Top hub nodes belum interaktif (mengklik hub node belum memfilter canvas graf ke emiten tersebut). |
| `db.PruneMockTestData` | Core | Pembersihan node/edge data mock evaluasi | ❌ **Belum Ada** | **P2 (Sedang)** | Tidak ada tombol aksi di UI Memory Graph untuk membersihkan entri pengujian sintetis. |
| `/api/settings/telegram` | `PATCH` | Konfigurasi bot Telegram termasuk `allowed_users` | ⚠️ **Parsial** | **P2 (Sedang)** | Form pengaturan Telegram di modal Settings belum memiliki input tag/chip untuk mengelola daftar whitelist user. |
| `/api/system/sectors-usage` | `GET` | Pelacakan disiplin kuota kredit 1.000 panggilan Sectors | ⚠️ **Parsial** | **P2 (Sedang)** | Belum menampilkan visual progress bar utilisasi kredit, rasio hit/miss cache, dan pemisahan OHLCV vs Fundamentals. |

---

## 3. Analisis Teknis Detail Kesenjangan (Root Cause & Code Audit)

### 3.1 Kesenjangan Kritis: Mockup Statik Hardcoded pada Panel Samping Kanan (`#insightPanel`)
* **Lokasi Berkas:** `backend/core/server/workspace.html` (Baris 2330–2450)
* **Analisis Kode:**  
  Panel samping kanan (`#insightPanel`) memuat tiga kartu utama yang saat ini di-hardcode secara statis dalam HTML murni:
  1. **Kartu "Metrik & Candlestick":** Berisi teks statik `Volume Anomaly Z-Score +3.42σ`, `Foreign Flow Net Buy +Rp 111,3 M`, `News Correlation Keterbukaan Info Hilirisasi BEI`, `RSI 64.2`, dan `Basic Materials +1.8%`.
  2. **Kartu "Timeline Kejadian":** Berisi teks statik `09:00 WIB Keterbukaan Informasi BEI Dirilis` dan `10:15 WIB Lonjakan Volume Anomali +3.42σ`.
  3. **Kartu "Data Pasar":** Berisi angka statik `IHSG 7.320,40 (+0,74%)`, `Nilai Transaksi Rp 12,4 T`, `Harga Nikel $16.420`, dan `IDR / USD 15.680`.
* **Dampak:** Ketika pengguna menganalisis saham perbankan (misalnya `BBCA` atau `BBRI`), panel kanan tetap menampilkan informasi tambang nikel `ANTM` dan komoditas nikel! Ini merusak kredibilitas profesional platform riset.
* **Solusi Arsitektural:**  
  1. Kosongkan isi awal panel tersebut (*Empty State* elegan: *"Menunggu investigasi emiten..."*).
  2. Buat fungsi reaktif JavaScript `updateInsightPanel(data)` yang mengisi panel tersebut secara dinamis ketika event `anomaly_detected` diterima atau saat sesi investigasi emiten dipilih.
  3. Jika pengguna sedang membahas emiten perbankan, tampilkan metrik perbankan, bukan komoditas nikel.

---

### 3.2 Kesenjangan 2: Ketiadaan Antarmuka Pipeline Investigasi 7-Tahap (Headless Surface)
* **Lokasi Backend:** `backend/core/server/server.go` (Baris 1386–1485)
* **Analisis Endpoint:**  
  Backend telah menyediakan rangkaian endpoint lengkap untuk mengelola investigasi formal:
  - `POST /api/investigations` (menerima payload `{"ticker": "ANTM", "timeframe_days": 30}`)
  - `GET /api/investigations` (mengembalikan daftar investigasi beserta status `RUNNING`, `COMPLETED`, `FAILED`)
  - `GET /api/investigations/{id}` (mengembalikan objek investigasi, anomalies, findings, evidence_items, timeline_events)
* **Kondisi Web Saat Ini:**  
  Di `workspace.html` hanya terdapat 3 tampilan utama (`sections`):
  1. `#heroView` (Layar beranda sambutan)
  2. `#chatView` (Ruang percakapan ReAct streaming)
  3. `#graphPageView` (Visualisasi kanvas Memory Graph)  
  **Sama sekali tidak ada tampilan atau tab untuk melihat atau memicu Investigasi Formal.**
* **Dampak:** Analis institusional yang ingin menjalankan audit terstruktur satu emiten dan mengunduh dossier lengkap terpaksa menggunakan CLI terminal (`niskava investigate ANTM --days 30`). Web workspace kehilangan salah satu dari dua pilar interaksi utama Niskava.
* **Solusi Arsitektural:**  
  Tambahkan view `#investigationsPageView` dengan dua panel:
  - **Panel Kiri:** Formulir inisiasi audit (Input Ticker IDX, selector horizon 30/60/90 hari, tombol "Mulai Audit Otonom") dan daftar riwayat investigasi terdahulu.
  - **Panel Kanan:** Dossier viewer interaktif yang menyajikan ringkasan eksekutif, tabel anomali statistik, timeline kronologis, dan kutipan berita resmi.

---

### 3.3 Kesenjangan 3: Ketiadaan Chart Candlestick & Anomali Volume (TradingView / Lightweight Charts)
* **Lokasi Spesifikasi:** `AGENTS.md` (Bagian 3) & `docs/TEAM_CLIENTS_INTEGRATION_GUIDE.md` (Bagian 2.E)
* **Analisis Data:**  
  Backend menghitung secara deterministik data kuantitatif:
  - Daily OHLCV dari Sectors Financial API v2 (`/v2/daily/{symbol}/`)
  - Volume Moving Average 20 hari ($MA_{20}$)
  - Z-Score volume ($V_z \ge 2.5$) dan abnormal returns
  - Memancarkan SSE event `anomaly_detected` berisi Z-Score, baseline value, dan nilai aktual.
* **Kondisi Web Saat Ini:**  
  Di `workspace.html` baris 2985–3015, fungsi `updateRightPanelAnomaly(payload)` hanya membuat elemen `div` teks sederhana di dalam kontainer `.signal-list`. **Sama sekali tidak ada grafik candlestick maupun bar volume visual.**
* **Dampak:** Produk terlihat seperti chatbot teks generik daripada platform intelijen pasar profesional setara Bloomberg Terminal / TradingView. Juri hackathon tidak dapat melihat secara visual lonjakan candlestick saat anomali $V_z \ge 2.5$ terjadi.
* **Solusi Arsitektural:**  
  Integrasikan library **TradingView Lightweight Charts** (berukuran ringan ~45 KB, bebas dependensi React) pada drawer kanan atau modal pop-up:
  - Render candlestick harian dari data Sectors v2.
  - Tampilkan histogram volume di sub-chart bawah.
  - Pasang marker lonjakan (panah merah / badge bulat) persis pada baris tanggal di mana $V_z \ge 2.5$ atau abnormal return terdeteksi.

---

### 3.4 Kesenjangan 4: Siklus Hidup & Manajemen Percakapan (Fork, Global Search, Rename, Pin)
* **Lokasi Backend:** `backend/core/server/server.go` (Baris 1102–1360)
* **Analisis Endpoint:**  
  Backend telah mengimplementasikan:
  1. `POST /api/chat/sessions/{id}/fork`: Menduplikasi percakapan hingga ID pesan tertentu (*branching exploration pattern*).
  2. `GET /api/chat/search?q={keyword}&limit=20`: Melakukan pencarian full-text pada tabel SQLite `chat_messages` dan mengembalikan cuplikan snippet pesan.
  3. `PATCH /api/chat/sessions/{id}`: Menerima payload `{"title": "Nama Baru", "is_pinned": true}`.
* **Kondisi Web Saat Ini:**  
  1. **Pencarian Palsu (Client-side Only):** Pada `workspace.html` baris 2731–2738, pencarian hanya menyaring judul pada DOM sidebar yang sedang tampil. Konten isi pesan tidak dicari ke backend SQLite!
  2. **Ketiadaan Tombol Fork:** Tidak ada tombol "Fork dari sini" pada bubble pesan asisten.
  3. **Ketiadaan Rename & Pin:** Pengguna tidak dapat menyematkan (*pin*) sesi riset penting di bagian teratas sidebar atau mengubah judul sesi yang digenerasi otomatis.
* **Solusi Arsitektural:**  
  - Hubungkan input pencarian dengan debounce 300ms ke endpoint `/api/chat/search?q=...` dan tampilkan dropdown hasil pencarian pesan dengan cuplikan kata kunci.
  - Tambahkan tombol menu dropdown (tiga titik `...`) pada setiap item riwayat di sidebar dengan aksi: *Pin*, *Rename*, *Export Markdown*, dan *Delete*.
  - Tambahkan tombol ikon *Fork* di samping tombol copy/export pada setiap bubble pesan asisten.

---

### 3.5 Kesenjangan 5: Dedicated 3-Tier Evidence Matrix & Timeline Component
* **Prinsip Non-Negotiable:** Law 2 (Strict Financial Non-Advisory Boundary) mewajibkan klasifikasi bukti:
  - `SUPPORTED`: Terbukti oleh laporan resmi Sectors v2 / IDXnet.
  - `UNCERTAIN`: Ada korelasi namun belum ada bukti kausalitas langsung (rumor pasar).
  - `CONTRADICTED`: Klaim atau rumor terbukti dibantah oleh data formal emiten.
* **Kondisi Web Saat Ini:**  
  Ketika backend memancarkan event `finding_emitted`, web hanya menyisipkan teks ke dalam laci *Thinking / Reasoning Steps*. Pengguna harus mengklik accordion pemikiran agen untuk melihat temuan ini, dan tampilannya bercampur dengan log eksekusi tool.
* **Solusi Arsitektural:**  
  Buat komponen **Evidence Matrix Card** khusus yang muncul di bawah jawaban akhir asisten:
  - Tiga kartu kolom atau badge berkode warna: Hijau (`SUPPORTED`), Kuning (`UNCERTAIN`), Merah (`CONTRADICTED`).
  - Dilengkapi skor keyakinan (*confidence score*), tautan sumber berita resmi (*source URL*), dan status temporal (*causality tag*: `LIKELY_CATALYST`, `PRECEDED_ANNOUNCEMENT`, `UNEXPLAINED_BY_NEWS`).

---

### 3.6 Kesenjangan 6: Manajemen Whitelist Telegram Bot & Heartbeat Monitor
* **Lokasi Backend:** `backend/core/server/server.go` (Baris 639–760)
* **Analisis Endpoint:**  
  Backend mendukung penyimpanan dan pembaruan array `allowed_users` (daftar ID atau username pengguna Telegram yang diizinkan berinteraksi dengan agen) serta status poller background.
* **Kondisi Web Saat Ini:**  
  Di modal Settings tab Telegram, hanya tersedia kolom `bot_token` dan tombol `Start/Stop/Test`. Tidak ada kolom input untuk menambahkan atau menghapus analis yang masuk ke daftar putih (*whitelist*).
* **Solusi Arsitektural:**  
  Tambahkan komponen input chip / tag pada tab Telegram di modal Settings untuk mengelola `allowed_users`, serta indikator visual heartbeat poller aktif (*Live Polling Status*).

---

### 3.7 Kesenjangan 7: Pelacakan Disiplin Kuota Kredit Sectors v2
* **Prinsip Non-Negotiable:** Law 5 (Credit Budget Discipline) mengharuskan perlindungan ketat terhadap alokasi 1.000 kredit Sectors grant.
* **Kondisi Web Saat Ini:**  
  Modal Settings tab Sectors hanya menampilkan total cache entries dan tombol "Bersihkan Cache Kadaluarsa".
* **Solusi Arsitektural:**  
  Tampilkan kartu statistik:
  - Total panggilan dihemat oleh cache lokal SQLite (`estimated_credit_saved`).
  - Pemisahan data: *Permanent Daily OHLCV* ($T < \text{today}$) vs *Expiring Fundamentals*.
  - Sisa estimasi kuota grant dari 1.000 kredit.

---

### 3.8 Kesenjangan 8: Resolusi Strategis Dualitas Web Workspace (`workspace.html` vs React SPA)
* **Dilema Arsitektur:**  
  - Skenario A: Mengembangkan React SPA dari awal di `clients/web/` (menginstall puluhan packages npm, setting Vite, Tailwind, router, dan bundling ke Go `//go:embed dist/*`). Risiko: butuh waktu setup build tooling dan rawan regresi sebelum deadline.
  - Skenario B: Menyempurnakan `backend/core/server/workspace.html` (arsitektur saat ini yang sudah jalan 100% tanpa npm, zero-dependency, styling Binance dark mode sudah sangat rapi, dan terintegrasi langsung dalam Go server).
* **Rekomendasi Tim Lead:**  
  **Prioritaskan Skenario B.** Sempurnakan `workspace.html` dengan menambahkan komponen-komponen yang kurang (Investigations view, Lightweight Charts, filter Ego-Network, pencarian global). Hal ini menjamin biner Go tetap beroperasi secara *single-binary zero-dependency* yang portabel dan siap didemokan kapan saja tanpa kendala build frontend eksternal. Kerangka React di `clients/web/` dapat disinkronkan secara terpisah untuk iterasi pasca-hackathon.

---

## 4. Spesifikasi Desain Antarmuka & Wireframe Komponen Baru

### 4.1 Wireframe Tab / Halaman Investigasi Formal (`#investigationsPageView`)

```text
┌────────────────────────────────────────────────────────────────────────────────────────┐
│ NISKAVA AGENT  [Terminal] [Memori Graf] [Investigasi Formal]               [Settings]  │
├──────────────────────────┬─────────────────────────────────────────────────────────────┤
│ 📋 DAFTAR INVESTIGASI    │ 🔍 DOSSIER INVESTIGASI: ANTM (Aneka Tambang Tbk)            │
│                          ├─────────────────────────────────────────────────────────────┤
│ [+ Mulai Audit Baru]     │ STATUS: [COMPLETED]  |  HORIZON: 30 HARI  |  TANGGAL: 24/09 │
│ ──────────────────────── │ ─────────────────────────────────────────────────────────── │
│ • INV-2026-ANTM (DONE)   │ 📊 RINGKASAN ANOMALI DETERMINISTIK (LAW 1)                  │
│   Volume Spike Z=+3.84σ  │ - Tanggal Anomali : 2026-09-20                              │
│                          │ - Metrik Terpicu  : VOLUME & PRICE SURGE                    │
│ • INV-2026-BBCA (DONE)   │ - Volume Aktual   : 85.0M lembar (Baseline MA20: 22.1M)    │
│   Normal Distribution    │ - Volume Z-Score  : +3.84σ (Ambang batas 2.50σ terlampaui) │
│                          │ ─────────────────────────────────────────────────────────── │
│ • INV-2026-BUMI (RUNNING)│ 🛡️ MATRIKS VERIFIKASI BUKTI 3-TIER (LAW 2)                  │
│   Stage 5: News Harvest  │ ┌─────────────────────────────────────────────────────────┐ │
│                          │ │ [SUPPORTED] Katalis Uji Coba Smelter Feronikel (95%)    │ │
│                          │ │ Sumber: Keterbukaan IDXnet | Status: LIKELY_CATALYST    │ │
│                          │ ├─────────────────────────────────────────────────────────┤ │
│                          │ │ [UNCERTAIN] Rumor Akuisisi Saham oleh Konsorsium (65%)  │ │
│                          │ │ Sumber: Media Sosial | Status: UNEXPLAINED_BY_NEWS      │ │
│                          │ └─────────────────────────────────────────────────────────┘ │
│                          │ ─────────────────────────────────────────────────────────── │
│                          │ 📈 VISUALISASI CANDLESTICK & ANOMALI VOLUME (TradingView)   │
│                          │ [==== Grafik Candlestick Harga & Baris Volume Tersemat ===] │
└──────────────────────────┴─────────────────────────────────────────────────────────────┘
```

### 4.2 Desain Komponen TradingView Lightweight Charts
Gunakan script tersemat dari CDN resmi atau static embed:
```html
<script src="https://unpkg.com/lightweight-charts@4.1.1/dist/lightweight-charts.standalone.production.js"></script>
```
Logika inisialisasi pada saat event `anomaly_detected` diterima:
```javascript
function renderCandlestickChart(containerId, candleData, anomalyDate) {
    const chart = LightweightCharts.createChart(document.getElementById(containerId), {
        width: 600,
        height: 280,
        layout: { background: { color: '#181A20' }, textColor: '#848E9C' },
        grid: { vertLines: { color: '#2B313A' }, horzLines: { color: '#2B313A' } },
    });
    const candleSeries = chart.addCandlestickSeries({
        upColor: '#0ECB81', downColor: '#F6465D', borderVisible: false,
    });
    candleSeries.setData(candleData);

    // Tandai tanggal anomali dengan marker merah
    if (anomalyDate) {
        candleSeries.setMarkers([{
            time: anomalyDate,
            position: 'aboveBar',
            color: '#F6465D',
            shape: 'arrowDown',
            text: 'Volume Anomaly (Z ≥ 2.5σ)',
        }]);
    }
}
```

---

## 5. Actionable Implementation Roadmap (Prioritas Eksekusi)

```
┌─────────────────────────────────────────────────────────────┐
│                 FASE 1: HIGH-IMPACT VISUAL                  │
│  - Hapus/dinamiskan data hardcoded statik di `#insightPanel`│
│  - Integrasi Lightweight Charts (candlestick & volume)      │
│  - Komponen Dedicated Evidence Matrix 3-Tier di chat view   │
│  - Hubungkan Global Search `/api/chat/search` di sidebar    │
│  - Tambahkan tombol OpenCode Forking pada pesan asisten     │
└──────────────────────────────┬──────────────────────────────┘
                               │
                               ▼
┌─────────────────────────────────────────────────────────────┐
│              FASE 2: INVESTIGATION PIPELINE UI              │
│  - Buat section `#investigationsPageView` di workspace.html │
│  - Hubungkan form "Mulai Audit" ke `POST /api/investigations│
│  - Render dossier temuan & timeline dari `GET /api/inv/{id}`│
│  - Tambahkan tombol Prune Mock Test Data di halaman Graf    │
└──────────────────────────────┬──────────────────────────────┘
                               │
                               ▼
┌─────────────────────────────────────────────────────────────┐
│             FASE 3: SETTINGS & POLISH WORKSPACE             │
│  - Input tag whitelist `allowed_users` di modal Telegram    │
│  - Kartu visual disiplin kredit Sectors v2 (Law 5)          │
│  - Aksi Pin & Rename sesi pada riwayat obrolan sidebar      │
│  - Sinkronisasi berkas `clients/web/index.html`             │
└─────────────────────────────────────────────────────────────┘
```

### Quality Gate Sebelum Presentasi Demo:
1. **Tidak Ada Data Hardcoded:** Panel samping kanan tidak lagi menampilkan data statik palsu ANTM/IHSG ketika sesi baru dibuka atau saat emiten lain sedang dianalisis.
2. **Verifikasi Jalur Investigasi:** Memasukkan ticker `ANTM` pada form audit $\to$ backend mengeksekusi pipeline 7-tahap $\to$ dossier temuan, metrik Z-Score, dan timeline tampil rapi dalam waktu < 20 detik.
3. **Uji Grafis Pasar:** Candlestick tampil interaktif dengan marker merah di atas baris tanggal anomali tanpa error konsol JavaScript.
4. **Pencarian Riwayat Pesan:** Mengetik kata kunci "dividen" atau "nikel" di input pencarian sidebar $\to$ menampilkan cuplikan pesan dari database SQLite via `/api/chat/search`.
5. **Kepatuhan Law 2:** Setiap temuan memiliki badge `SUPPORTED`, `UNCERTAIN`, atau `CONTRADICTED` dan footer menampilkan teks *Disclaimer Non-Advisory* resmi.
