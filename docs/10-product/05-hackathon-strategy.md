# 05 — Strategi Sectors Hackathon Indonesia 2026

**Status:** LOCKED  
**Versi Dokumen:** 1.1.0  
**Terakhir Diperbarui:** 2026-09-16  
**Target Lintasan:** Track 1 · AI Agents & Assistants  

Panduan operasional dan strategi terfokus tim untuk memenangkan **Sectors Hackathon Indonesia 2026** ([hackathon.sectors.app](https://hackathon.sectors.app/)).

---

## 1. Profil Kompetisi & Target Kategori

* **Nama Event**: Sectors Hackathon Indonesia 2026
* **Penyelenggara**: Supertype, Sectors, dan Algoritma
* **Total Hadiah**: IDR 50.000.000 (Tunai IDR 30.000.000 + Langganan Sectors Insider + Sectors API Credits)
  * Juara 1: IDR 15.000.000 + 6 bulan Sectors Insider + 20.000 API credits
  * Runner-up: IDR 9.000.000 + 4 bulan Sectors Insider + 15.000 API credits
  * 3 Finalis: Masing-masing IDR 2.000.000 + 2 bulan Sectors Insider + 10.000 API credits
* **API Credit Grant**: 1.000 Sectors API Credits per tim (diklaim via portal tim setelah onboarding)
* **Kategori Terpilih**: **Track 1 · AI Agents & Assistants**
  > *"Build conversational or autonomous AI products with custom agent logic or orchestration at their core."*

### Timeline Kritis Submisi
* **19 Agustus 2026**: Pembukaan pendaftaran & periode pengembangan (*build period*).
* **22 September 2026 (23:59 WIB)**: Batas akhir pendaftaran tim & penutupan onboarding.
* **30 September 2026 (23:59 WIB)**: **Batas Akhir Submisi & Pembekuan Kode (Hard Deadline)**.
* **1–8 Oktober 2026**: Periode Penjurian Asinkronus (berdasarkan video demo & repo GitHub).
* **9 Oktober 2026**: Pengumuman Pemenang Resmi di website kompetisi dan Instagram (@sectors.app & @algoritma).

---

## 2. Gate 0: Syarat Kelayakan Mutlak (Eligibility & Onboarding)

Sesuai aturan resmi kompetisi (Rule 03 & 04):
1. **Verifikasi Onboarding Akun Sectors:** Setiap anggota tim (1–4 orang) **wajib membuat akun Sectors dan menyelesaikan proses onboarding di sectors.app sebelum tim menulis satu baris pun kode proyek**.
2. **Klaim Grant 1.000 Kredit:** Setelah seluruh anggota tervalidasi onboarding-nya, perwakilan tim mengklaim 1.000 kredit dari halaman tim di portal kompetisi.
3. **Penguncian Roster (Roster Lock):** Begitu kredit diklaim, susunan anggota tim terkunci permanen—tidak dapat menambah atau mengubah anggota.
4. **Larangan Akun Ganda:** Dilarang mendaftarkan akun ganda untuk memanipulasi kredit tambahan (pelanggaran berakibat diskualifikasi langsung).

---

## 3. Rubrik Penjurian Resmi & Strategi Nilai Maksimal (40-30-30)

Penjurian bersifat 100% asinkronus tanpa sesi presentasi live. Evaluasi dilakukan oleh tim internal Sectors & Supertype berdasarkan 3 kriteria berbobot:

| Kriteria Penilaian | Bobot | Deskripsi Resmi Juri | Strategi Niskava untuk Meraih Skor Maksimal |
|---|:---:|---|---|
| **Real-World Usability** | **40%** | Seberapa baik proyek menyelesaikan masalah dunia nyata? Dapatkah seseorang menggunakannya hari ini dan memperoleh manfaat riil? | • **Siap Pakai Hari Ini**: CLI `niskava investigate <ticker>` langsung menghasilkan laporan bukti terstruktur dalam <6 detik.<br>• **Dual Surfaces**: Analis terminal dan pengguna desktop terlayani sekaligus (CLI + Web UI).<br>• **Menghapus 40 Menit Kerja Manual**: Mengotomasi korelasi anomali volume dengan berita pasar modal. |
| **Video Demo & Storytelling** | **30%** | Seberapa menarik, memikat, dan jelas video diproduksi? Apakah mengomunikasikan masalah secara efektif bagi audiens sasaran? | • **Teaser 1 Menit yang Memikat**: Menampilkan hook "kontras antara rumor vs fakta bukti" secara dramatis.<br>• **Judging Video 3 Menit Solid**: Alur storytelling jernih: Persona Analis $\to$ Masalah Information Asymmetry $\to$ Investigasi Live ANTM $\to$ Bukti Audit Trail di Web UI. |
| **Technical Depth & Execution** | **30%** | Diverifikasi langsung lewat repositori GitHub: Seberapa inovatif pemanfaatan Sectors API? Apakah proyek riil, berfungsi nyata, berarsitektur rapi, dan bukan tipuan demo? | • **Custom Hybrid Architecture**: Single binary Go (Zero-CGO SQLite + SSE) + Python Deterministic Quant pre-LLM.<br>• **Bukan Wrapper Chatbot**: 7-Stage Pipeline dengan model kausalitas temporal nyata.<br>• **Ground Truth Benchmarks**: Disertai suite pengujian benchmark kasus historis nyata BEI (`ANTM`, `BUMI`, `GOTO`, `BBRI`). |

---

## 4. Kualifikasi Resmi Track 1 (The Qualifying Test Compliance)

Panitia menetapkan uji kualifikasi ketat untuk Track 1: *Proyek harus memiliki custom-built agent logic or orchestration sendiri di sekeliling model—bukan sekadar menghubungkan prompt ke client off-the-shelf seperti Claude Desktop, OpenClaw, atau Hermes.*

Niskava memenuhi seluruh 6 kualifikasi resmi Track 1:

| Kriteria Kualifikasi Resmi Track 1 | Bukti Implementasi Arsitektur Niskava Agent | Lokasi Dokumen Teknis |
|---|---|---|
| **1. Multi-step reasoning flows** | Pipeline investigasi 7 tahap otonom: Initiation $\to$ Baseline $\to$ Quant Anomaly $\to$ Gap Detection $\to$ OSINT Harvest $\to$ Evidence Correlation $\to$ Synthesis. | [`30-agent/01-investigation-pipeline.md`](../30-agent/01-investigation-pipeline.md) |
| **2. Custom tool-use pipelines** | Orkestrasi alat dinamis: Pemanggilan deterministik Sectors v2 API, pencarian web OSINT terarah, dan korelasi temporal. | [`30-agent/02-skills-catalog.md`](../30-agent/02-skills-catalog.md) |
| **3. Routing between data sources** | Pemisahan fakta kuantitatif (*Sectors Ground Truth*) dan narasi publik (*IDXnet, Corporate News, Web Signals*). | [`20-architecture/01-system-overview.md`](../20-architecture/01-system-overview.md) |
| **4. Memory or state management** | **Local Conversational Graph Memory Engine (06-local-conversational-graph-memory)**: Menyimpan relasi entitas, catatan harga posisi, dan riwayat investigasi dalam graf asosiatif SQLite + NetworkX lintas sesi, sehingga agen bebas dari amnesia konteks. | [`30-agent/05-conversational-memory-engine.md`](../30-agent/05-conversational-memory-engine.md) |
| **5. Autonomous task execution** | Agen secara mandiri merumuskan kueri pencarian berita berdasarkan tanggal anomali yang ditemukan oleh mesin NumPy tanpa campur tangan manusia. | [`20-architecture/06-osint-engine.md`](../20-architecture/06-osint-engine.md) |
| **6. Purpose-built interface** | Antarmuka khusus analis pasar modal: Terminal TUI (`Bubbletea`) untuk kecepatan eksekusi dan Web Dashboard (*Cyber-OSINT / Bloomberg Terminal style*) untuk visualisasi bukti. | [`10-product/03-product-scope-and-surfaces.md`](03-product-scope-and-surfaces.md) |

---

## 5. Checklist Lengkap Submisi & Aturan Pembekuan (*Submission Freeze*)

Submisi dilakukan melalui portal hackathon sebelum **30 September 2026 pukul 23:59 WIB**.

### Checklist Berkas Submisi
- [ ] **1-Sentence Problem Statement**:
  > *"Pelaku pasar modal Indonesia kehilangan momentum dan modal karena lambat memverifikasi penyebab lonjakan saham tidak wajar akibat terpisahnya data kuantitatif bursa (Sectors API) dengan konteks keterbukaan informasi dan berita pasar modal (OSINT)."*
- [ ] **Public GitHub Repository**:
  - Dibuat dalam rentang waktu kompetisi (19 Agustus – 30 September 2026).
  - Wajib tetap berstatus **publik selama minimal 90 hari** setelah pemenang diumumkan (hingga Januari 2027).
  - Bersih dari *API Key* atau kredensial rahasia apa pun.
  - Memiliki `README.md` yang mudah dijalankan oleh juri secara lokal.
- [ ] **1-Minute Video Teaser**:
  - Format rekaman layar produk berjalan nyata (YouTube/Medsos publik).
  - Fokus pada *hook*: Pergerakan anomali ANTM terdeteksi $\to$ CLI beraksi $\to$ Dashboard memetakan bukti.
- [ ] **3-Minute Judging Video**:
  - Walkthrough lengkap: Penjelasan masalah, target audiens, arsitektur hybrid, demo eksekusi live, dan pembuktian jejak audit bukti.
  - Link YouTube/Vimeo (Public atau Unlisted), Google Drive (akses publik aktif), atau Loom.
- [ ] **Postingan Media Sosial Wajib**:
  - Dipublikasikan di **Instagram, LinkedIn, Threads, atau TikTok**.
  - **Wajib menandai (*tag*) akun resmi Sectors**.
  - **Wajib menggunakan template thumbnail resmi** yang disediakan oleh panitia di portal.
- [ ] **Team Snapshot**: Foto seluruh anggota tim terlampir pada form submisi.

### Aturan Pembekuan (*Code Freeze Rule*)
* Repositori dan aplikasi **langsung membeku permanen saat tombol submit ditekan** atau pada 30 September 23:59 WIB (mana yang tercapai lebih dulu).
* **DILARANG MELAKUKAN COMMIT, PUSH, ATAU EDIT APAPUN SETELAH FREEZE**, termasuk perbaikan bug kecil. Pelanggaran mengakibatkan diskualifikasi otomatis.
* Satu-satunya pengecualian darurat adalah kebocoran API Key: lapor panitia di Slack `#support`, cabut/rotasi key, lalu push 1 commit khusus menghapus key tersebut.

---

## 6. Skenario Demo Unggulan: Kasus ANTM (Aneka Tambang Tbk)

Kasus saham `ANTM` dipilih sebagai skenario evaluasi utama (*Golden Path*):
1. **Likuiditas & Volatilitas Nyata**: ANTM memiliki perputaran transaksi besar di BEI dan sering mengalami lonjakan volume mendadak akibat katalis komoditas global maupun kebijakan lokal.
2. **Keterkaitan Data Sektoral & Komoditas**: Memperlihatkan perbandingan anomali ANTM terhadap pergerakan subsektor tambang logam di Sectors API (`/v2/subsector/`).
3. **Kausalitas Temporal yang Gamblang**: Lonjakan volume transaksi di Sectors mendahului atau bersamaan dengan rilis pengumuman resmi ekspansi smelter dan perjanjian ekspor di media finansial/IDXnet, membuktikan keandalan logika korelasi bukti Niskava.
4. **Retensi Memori Graf Lintas Sesi (*Cross-Session Recall*)**: Di segmen penutup video 3 menit, pengguna membuka sesi baru dan bertanya: *"Bagaimana kelanjutan saham tambang yang kemarin saya cek?"*. Agen secara instan memanggil graf memori lokal (Ego-Graph ANTM) dan melanjutkan analisis tanpa amnesia konteks.

---

## 7. Alokasi Kuota 1.000 Kredit API Sectors

Untuk menjamin kuota kredit bertahan hingga evaluasi juri tuntas tanpa risiko *over-quota*:

| Tahapan Operasional | Alokasi Kredit | Strategi Mitigasi & Penghematan |
|---|:---:|---|
| **Development & Logic Tuning** | 200 credits | *Local SQLite Caching*: Data historis tanggal lampau di-cache permanen di `sectors_cache`. |
| **Testing & Benchmark CI/CD** | 200 credits | Mode *Offline/Mock Runner* (`MOCK_SECTORS=1`) menggunakan fixture JSON tersimpan. |
| **Perekaman Video Demo** | 200 credits | Eksekusi live terencana untuk rekaman video 1 menit teaser dan 3 menit judging. |
| **Cadangan Evaluasi Langsung Juri** | 400 credits | Disimpan murni untuk pengujian lokal oleh dewan juri saat memeriksa repositori. |
| **Total Anggaran** | **1.000 credits** | **100% Terukur & Terkelola** |
