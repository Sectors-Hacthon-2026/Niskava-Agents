# 06 — Naskah & Storyboard Video Demo Penjurian (Demo Video Script)

**Status:** ACCEPTED  
**Versi Dokumen:** 1.0.0  
**Terakhir Diperbarui:** 2026-09-22  
**Target Kompetisi:** Sectors Hackathon Indonesia 2026 — Track 1: AI Agents & Assistants  
**Bobot Penjurian:** **30% Video Demo & Storytelling** + Mendukung **40% Usability**

---

## 📌 Ringkasan Eksekutif Video

Sesuai aturan resmi kompetisi (Rule 05 & Rule 10), tim wajib menyerahkan dua format video:
1. **Video Teaser 1 Menit**: Hook dramatis untuk postingan media sosial wajib (Instagram/LinkedIn/TikTok) dengan template thumbnail resmi Sectors.
2. **Video Penjurian 3 Menit (Judging Walkthrough)**: Panduan demonstrasi komprehensif bagi dewan juri untuk membuktikan bahwa Niskava adalah produk nyata, bukan sekadar chatbot wrapper.

---

## Bagian A: Video Teaser 1 Menit (Social Media Hook)

* **Durasi Target**: Tepat 60 detik.
* **Tujuan**: Menarik perhatian pemirsa dengan kontras dramatis antara "Kepanikan Rumor Saham" vs "Kebenaran Faktual Bukti Finansial".
* **Audio**: Musik latar bertempo cepat bernuansa synthwave/cyberpunk dengan sulih suara (*voiceover*) tegas dan percaya diri.

### Storyboard Detik-per-Detik (1 Menit):

| Detik | Visual di Layar | Teks di Layar / Callout | Narasi Voiceover (Bahasa Indonesia / Subtitle EN) |
|---|---|---|---|
| **00:00 - 00:08** | Cuplikan tangkapan layar forum grup saham ritel penuh rumor panik ("Saham X digoreng! Ada kebocoran info!"). Layar berkedip merah. | **"Rumor atau Fakta?"** | *"Berapa kali Anda panik atau tertipu rumor saat saham tiba-tiba melonjak gila-gilaan di bursa?"* |
| **00:08 - 00:18** | Transisi cepat ke terminal gelap Matrix/Market Intelligence. Perintah diketik: `niskava investigate ANTM`. Angka Z-Score volume melesat deterministik. | **"Hukum 1: Matematika Dulu, Baru Generatif"** | *"Jangan tanya chatbot biasa yang jago berhalusinasi. Perkenalkan Niskava Agent: platform intelijen pasar otonom pertama untuk Bursa Efek Indonesia."* |
| **00:18 - 00:35** | Layar menampilkan HUD real-time. ReAct thinking steps bergulir: memanggil Sectors v2 API $\to$ NumPy Compute Gate $\to$ Sectors News & Disclosure Engine. | **"7-Stage Autonomous Investigation"** | *"Dalam 5 detik, Niskava menarik data transaksi resmi Sectors API, menghitung Z-Score deterministik, memburu dokumen keterbukaan BEI, dan memvalidasi urutan waktu."* |
| **00:35 - 00:50** | Kamera beralih ke Web Workspace (`localhost:20128`). Tampil grafik candlestick dengan pin anomali merah dan kartu bukti hijau `[SUPPORTED]`. | **"Taksonomi Bukti 3-Tier"** | *"Hasilnya? Bukti terverifikasi, bukan tebak-tebakan. Rumor dibantah, fakta keterbukaan diuji, jejak audit tersimpan lokal."* |
| **00:50 - 01:00** | Logo Niskava Agent menyala emas bersama logo Sectors Hackathon 2026. Teks slogan muncul. | **"Don't just answer questions. Investigate them."** | *"Niskava Agent. Jangan cuma menjawab pertanyaan. Selidiki faktanya. Coba gratis di GitHub kami sekarang!"* |

---

## Bagian B: Video Penjurian 3 Menit (Judging Walkthrough)

* **Durasi Target**: 2 menit 55 detik s/d 3 menit 00 detik (Maksimum toleransi 180 detik).
* **Format**: Rekaman layar bersih (1080p 60fps) dengan picture-in-picture presenter di sudut kanan bawah.
* **Struktur 4 Babak**: Masalah (30s) $\to$ Arsitektur Hybrid (45s) $\to$ Live Demo ANTM (80s) $\to$ Kepatuhan Regulasi & Penutup (25s).

---

### Babak 1: Latar Belakang Masalah & Persona (00:00 – 00:30)

* **Visual**:
  * Slide pembuka bertuliskan judul resmi proyek, nomor tim, dan track: **Track 1: AI Agents & Assistants**.
  * Diagram grafis kontras: Data kuantitatif terpisah dari dokumen pengumuman bursa.
* **Narasi Voiceover**:
  > *"Halo Dewan Juri Sectors Hackathon 2026. Saya [Nama Presenter] dari Tim Niskava. Di pasar modal Indonesia, setiap kali sebuah saham melonjak atau anjlok drastis, analis ekuitas dan investor ritel menghabiskan 30 hingga 45 menit secara manual membuka laporan volume, mencari PDF keterbukaan informasi di IDXnet, dan menyisir portal berita untuk mencari tahu: apa katalis sebenarnya?*
  > *Chatbot AI biasa gagal karena dua hal: pertama, mereka berhalusinasi saat menghitung angka. Kedua, mereka tidak memahami kausalitas waktu rilis berita vs waktu pergerakan volume. Inilah mengapa kami membangun Niskava Agent."*

---

### Babak 2: Arsitektur Unik & Nilai Tambah Teknis (00:30 – 01:15)

* **Visual**:
  * Menampilkan diagram Arsitektur Hybrid (Go Core + Python Engine + Embedded/React UI).
  * Menyoroti 4-Layer Agentic Hierarchy dan 6 Hukum Arsitektur Niskava.
* **Narasi Voiceover**:
  > *"Niskava bukan sekadar AI wrapper. Kami membangun arsitektur hybrid dengan 4-Layer Agentic Hierarchy yang ketat:*
  > *Pertama, Layer Primitives: Kami mengekspos Sectors Financial API v2 sebagai Model Context Protocol (MCP) server lokal dan Sectors News Engine.*
  > *Kedua, Deterministic Compute Gate: Sesuai Hukum 1 kami, LLM dilarang keras menghitung matematika. Volume Z-Score, abnormal return, dan divergensi sektor dihitung deterministik via NumPy sebelum AI diikutsertakan.*
  > *Ketiga, 6 Domain Skills Registry: Prosedur analitik standar industri untuk audit kausalitas, bandarmologi, dan stress-test neraca.*
  > *Keempat, ReAct Cognitive Loop dengan Local Graph Memory di SQLite dan NetworkX untuk mengingat konteks emiten lintas sesi tanpa dependensi cloud berbayar."*

---

### Babak 3: Live Demo Skenario Nyata ANTM (01:15 – 02:35)

* **Visual**:
  * **Tampilan 1 (CLI TUI)**:
    1. Buka terminal: jalankan `niskava`.
    2. Tampilkan HUD Market Intelligence dan tekan `L` untuk menunjukkan fitur i18n dwibahasa instan (EN/ID).
    3. Masukkan prompt: *"Apakah ada anomali lonjakan volume pada saham ANTM dan apa penyebab faktualnya?"*
    4. Sorot terminal saat ReAct thinking steps mengalir secara real-time:
       - Memanggil skill `market-anomaly-recon` $\to$ Z-Score volume terdeteksi +3.84σ.
       - Memanggil skill `event-causality-audit` $\to$ Menemukan pengumuman smelter Halmahera Timur.
       - Memanggil taksonomi bukti $\to$ Mengklasifikasikan pengumuman sebagai `[SUPPORTED]` (Confidence: 0.95).
  * **Tampilan 2 (Web Workspace `localhost:20128`)**:
    1. Buka browser: jalankan `niskava serve --open`.
    2. Tunjukkan bahwa server Go langsung menyajikan Web Canvas interaktif secara mandiri tanpa kompilasi node/npm.
    3. Buka visualisasi graf memori via `niskava graph`: tampilkan simpul relasi ANTM terhubung dengan nikel, smelter, dan foreign inflow.
* **Narasi Voiceover**:
  > *"Mari kita lihat aksinya secara langsung. Di terminal, saya menjalankan `niskava`. Perhatikan antarmuka TUI kami yang responsif, dilengkapi dukungan multi-bahasa instan.*
  > *Saya menanyakan pergerakan anomali ANTM. Dalam hitungan detik, agen ReAct tidak langsung mengarang jawaban. Sistem memanggil skill deteksi kuantitatif kami, menemukan lonjakan volume 184,5 juta lembar atau 3,84 standar deviasi di atas rata-rata 20 hari.*
  > *Agen kemudian secara otonom meluncurkan penelusuran berita bursa dan pengumuman resmi BEI dalam jendela waktu anomali. Agen menemukan keterbukaan informasi peresmian smelter baru, mencocokkan stempel waktu publikasi yang mendahului lonjakan bursa, dan menetapkan status SUPPORTED dengan confidence score 0,95.*
  > *Di browser, seluruh hasil ini tersinkronisasi live via Server-Sent Events, lengkap dengan visualisasi knowledge graph memori lokal."*

---

### Babak 4: Kepatuhan Regulasi & Penutup (02:35 – 03:00)

* **Visual**:
  * Menampilkan teks disclaimer non-advisory di footer terminal dan web.
  * Menampilkan repository GitHub publik yang bersih dari API keys, dengan status 150 unit test passing.
* **Narasi Voiceover**:
  > *"Terakhir, Niskava patuh 100% pada Aturan Hackathon 06 dan 12: Kami murni platform riset bukti dan TIDAK PERNAH memberikan rekomendasi beli/jual spekulatif atau eksekusi otomatis ke broker.*
  > *Seluruh sesi disimpan secara berdaulat di SQLite lokal mesin pengguna.*
  > *Niskava Agent: mengubah spekulasi pasar menjadi kepastian intelijen terverifikasi. Terima kasih."*

---

## 📋 Checklist Teknis Perekaman Video

- [ ] Resolusi rekaman 1920x1080 (16:9), framerate konstan 60 fps.
- [ ] Font terminal menggunakan `JetBrains Mono` ukuran 14-16pt agar terbaca jelas di ponsel.
- [ ] Mode gelap konsisten (`#090D16`) antara terminal TUI dan Web Canvas.
- [ ] Mikrofon jernih tanpa gema; level audio voiceover dinormalisasi ke -14 LUFS.
- [ ] Tidak menampilkan API key asli (gunakan mock atau key bertopeng).
- [ ] Menyertakan subtitle Bahasa Inggris (closed captions .srt) untuk aksesibilitas juri internasional.
- [ ] Durasi akhir judging video tidak melebihi 180 detik (diskualifikasi jika melanggar batas panitia).
