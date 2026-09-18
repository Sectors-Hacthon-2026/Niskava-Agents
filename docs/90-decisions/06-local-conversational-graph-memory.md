# 06 — Engine Memori Percakapan & Lintas Sesi Berbasis Graf Lokal (Local Conversational Graph Memory)

**Status:** ACCEPTED  
**Tanggal:** 2026-09-16  
**Pengambil Keputusan:** Core Team  
**Kepatuhan Hackathon:** Track 1 · AI Agents & Assistants (*Memory or State Management Criterion*)  

---

## 1. Konteks & Permasalahan

Dalam skenario investigasi pasar modal nyata, analisis tidak pernah berhenti pada satu pertanyaan tunggal (*one-shot session*). Pengguna sering melakukan investigasi bertahap lintas hari:
* *Hari 1*: Pengguna menyelidiki lonjakan volume `ANTM`, mencatat harga 1450, dan mengidentifikasi katalis smelter.
* *Hari 4*: Pengguna bertanya: *"Bagaimana pergerakan saham tambang yang kemarin saya cek?"* atau *"Apakah anomali kemarin berhubungan dengan tren nikel pekan ini?"*

Model LLM standar mengalami **amnesia sesi** (*stateless*). Pendekatan naif berupa mengirim seluruh teks obrolan lama (*sliding window*) memiliki kelemahan fatal:
1. **Boros Token & Kuota**: Mengirim riwayat percakapan panjang menghabiskan context window dan menambah biaya token/latensi.
2. **Kehilangan Struktur Kausalitas**: Teks mentah tidak membedakan mana entitas inti (`TICKER: ANTM`), mana aksi pengguna (`WATCHING`, `HOLDING`), dan mana temuan resmi (`CATALYZED_BY`).
3. **Keterbatasan Evaluasi Hackathon**: Salah satu dari 6 kriteria wajib Track 1 Sectors Hackathon adalah **"Memory or state management"**. Agen harus memiliki manajemen memori yang cerdas dan terstruktur.

---

## 2. Keputusan

**Mengadopsi Arsitektur Local-First Conversational Graph Memory Engine yang memadukan persistensi relasional SQLite lokal dengan pemrosesan graf in-memory menggunakan NetworkX di Python.**

### Komponen Utama Engine:
1. **Persistensi Relasional Lokal di SQLite (`~/.niskava/niskava.db`)**:
   * Tabel `memory_nodes`: Menyimpan entitas unik (`USER`, `TICKER`, `SECTOR`, `PRICE_LEVEL`, `CATALYST_EVENT`).
   * Tabel `memory_edges`: Menyimpan relasi berarah antar-entitas beserta konteks, bobot kebaruan (*recency decay*), dan ID sesi investigasi terkait.
2. **Pemrosesan Graf In-Memory (NetworkX DiGraph)**:
   * Mengambil relasi aktif dari SQLite ke memori RAM dalam hitungan mikrodetik.
   * Melakukan ekstraksi subgraf berbasis tetangga (*Ego-Graph*) dengan radius $k \le 2$ di sekitar entitas yang sedang dibahas.
3. **Ekstraksi Memori Hibrida (Hybrid Ingestion)**:
   * **Deterministik (Zero-Token Cost)**: Sistem otomatis merekam simpul dan relasi setiap kali investigasi resmi selesai (misal: `User -[INVESTIGATED]-> ANTM`).
   * **Ekstraksi Ringan (Gemini Flash)**: Mengekstrak entitas dan preferensi dari pernyataan obrolan bebas pengguna (misal: *"Saya punya posisi di 1450"* $\to$ `User -[HOLDS_AT]-> 1450 -[TICKER]-> ANTM`).
4. **Prompt Augmentation Terbatas (<300 Token)**:
   * Subgraf yang terpilih diformat menjadi ringkasan bullet poin terstruktur dalam tag XML `<investigative_memory>...</investigative_memory>` dan disuntikkan ke prompt sistem agen.

---

## 3. Alternatif yang Dipertimbangkan

| Alternatif | Alasan Ditolak |
|---|---|
| **Sliding Window Chat History** | Boros token, latensi membengkak seiring bertambahnya sesi, dan gagal menghubungkan entitas lintas hari secara semantik. |
| **Vector Database Murni (Pinecone / Chroma)** | Pencarian vektor berbasis *cosine similarity* sangat baik untuk dokumen teks, namun tidak memahami hubungan struktural berarah (misal: *siapa membeli apa pada harga berapa*). |
| **Database Graf Eksternal (Neo4j / KuzuDB)** | Mengharuskan juri atau pengguna menginstal Docker container atau dependensi C++ biner tambahan. Melanggar prinsip *single executable / lightweight setup*. |
| **Layanan Cloud Agent Memory (Mem0 / Zep Cloud)** | Bergantung pada API eksternal pihak ketiga, menambah latensi jaringan, menimbulkan biaya langganan, dan berisiko membocorkan data portofolio pengguna ke cloud. |

---

## 4. Konsekuensi

### Positif
* **Memenuhi Syarat Mutlak Track 1**: Memberikan bukti nyata inovasi *State & Memory Management* tanpa rekayasa demo (*high technical depth*).
* **Ultra-Cepat & Ringan**: Traversal graf di NetworkX selesai dalam **<5 milidetik**, tidak menambah latensi pipa investigasi.
* **100% Offline & Privat**: Mengikuti kedaulatan data lokal ([`04-local-first-sqlite-storage.md`](04-local-first-sqlite-storage.md)); memori pengguna tersimpan murni di laptop sendiri.
* **Efisiensi Token**: Hanya menyuntikkan relasi yang relevan (100–300 token), menghemat kuota kredit dan context window.

### Negatif / Kompromi yang Diterima
* Membutuhkan logika normalisasi entitas (*entity resolution*) sederhana di Python agar `ANTM` dan `Aneka Tambang` merujuk pada simpul yang sama (diatasi dengan pemetaan ticker resmi BEI).
