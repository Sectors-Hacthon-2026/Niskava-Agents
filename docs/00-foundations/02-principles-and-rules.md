# 02 — Prinsip Operasional & Batasan Kepatuhan (Principles & Rules)

**Status:** LOCKED  
**Versi Dokumen:** 1.0.0  
**Terakhir Diperbarui:** 2026-09-16  

Dokumen ini mendefinisikan prinsip-prinsip mutlak yang membatasi setiap keputusan arsitektur, desain modul, dan perilaku penalaran (*reasoning*) Niskava Agent. Fitur apa pun yang melanggar prinsip di bawah ini wajib ditolak.

---

## 1. Delapan Prinsip Operasional Inti (Operating Principles)

### P1 — Bukti di Atas Asumsi (Evidence Over Assumption)
Tidak ada temuan (*finding*) yang boleh berstatus `SUPPORTED` tanpa bukti terverifikasi dari data kuantitatif **Sectors API v2** atau dokumen resmi keterbukaan informasi emiten/BEI. Dugaan, rumor forum, atau simpulan tanpa bukti primer wajib dilabeli `UNCERTAIN` atau ditolak. Ketiadaan bukti dilaporkan sebagai *ketiadaan bukti*, tidak pernah diasumsikan.

### P2 — Deterministik Mendahului Generatif (Deterministic Before Generative)
Seluruh kalkulasi deret waktu (moving average, standar deviasi, volume $Z$-score, return harian, dan divergensi sektor) **wajib dihitung menggunakan script matematika deterministik Python** sebelum model LLM dipanggil. LLM digunakan untuk sintesis bahasa, perumusan query, dan korelasi konteks—bukan untuk kalkulasi aritmatika.

### P3 — Batasan Non-Penasihat Mutlak (Strict Non-Advisory Boundary)
Niskava Agent adalah platform intelijen data terbuka dan investigasi fakta, **bukan penasihat investasi berlisensi**. Sistem dilarang keras menghasilkan:
* Rekomendasi arah transaksi langsung (*BUY*, *SELL*, *HOLD*, *STRONG BUY*).
* Target harga pasti (*price target* spekulatif).
* Nasihat alokasi portofolio atau perencanaan keuangan pribadi.
Pelanggaran terhadap prinsip ini melanggar regulasi OJK dan integritas sistem.

### P4 — Kausalitas Sadar Waktu (Temporal-Aware Causality)
Hubungan sebab-akibat antara berita dan pergerakan pasar tidak boleh disimpulkan hanya dari kecocokan topik semata. Agen wajib memvalidasi urutan stempel waktu (*timestamp precedence*):
* Berita mendahului pergerakan volume $\rightarrow$ `LIKELY_CATALYST`.
* Volume melonjak mendahului publikasi berita resmi $\rightarrow$ `PRECEDED_ANNOUNCEMENT` (indikasi kebocoran informasi).
* Pergerakan volume tanpa ada berita relevan $\rightarrow$ `UNEXPLAINED_BY_NEWS`.

### P5 — Kedaulatan Data Lokal (Local-First Architecture)
Privasi analisis pengguna adalah prioritas. Seluruh sesi investigasi, log pemikiran agen, rekaman anomali, dan cache data tersimpan di mesin lokal pengguna (`~/.niskava/niskava.db`). Tidak ada data pengguna yang dikirim ke server pusat Niskava.

### P6 — Konservasi Kuota Kredit (Credit-Conscious Execution)
API kredit adalah sumber daya terbatas (grant 1.000 kredit pada hackathon). Sistem wajib menerapkan *caching layer* SQLite lokal untuk seluruh request data historis pasar. Request berulang pada data yang sama dilarang memotong kuota kredit. Sistem juga wajib menyediakan mode tiruan (*mock mode*) untuk pengujian otomatis.

### P7 — Jejak Audit Transparan (Auditable & Explainable Findings)
Setiap temuan yang disajikan kepada pengguna wajib memiliki jejak audit (*audit trail*):
1. Formula metrik yang memicu anomali.
2. Cuplikan teks rujukan asli (*evidence snippet*).
3. Tautan URL sumber atau ID dokumen resmi.
4. Nilai tingkat keyakinan (*confidence score* 0.0 – 1.0).

### P8 — Kontinuitas Akses Offline (Offline Continuity)
Sesi investigasi yang telah selesai dianalisis dan disimpan di database lokal harus dapat dibuka, dibaca, dan dieksplorasi kembali di CLI maupun Web Dashboard secara penuh tanpa memerlukan koneksi internet aktif.

### P9 — Pemisahan Lapisan Terstandarisasi (MCP Primitives & Modular Skills Separation)
Agen dilarang memadukan *I/O bursa*, *kalkulasi matematika*, dan *prosedur analisis* ke dalam satu fungsi serba bisa. Sistem wajib mematuhi pemisahan 4-layer:
1. **MCP Primitives**: Tool I/O standar data bursa (Sectors MCP) dan web scraping.
2. **Deterministic Compute Gate**: Firewall pemrosesan numerik (NumPy).
3. **Modular Domain Skills**: Standar Operasional Prosedur (SOP) analisis terisolasi dengan input/output contract yang ketat.
4. **Cognitive ReAct Agent**: Orkestrator pembuat keputusan, pemanggil tool, dan perangkum bukti.

### P10 — Larangan Dump Data Mentah (Zero Raw Prompt Stuffing)
Sistem dilarang memasukkan ribuan baris data candlestick atau JSON bursa mentah langsung ke dalam context window LLM. Data mentah wajib diringkas melalui Deterministic Compute Gate menjadi metrik statistik bernilai tinggi ($V_z$, $R_t$, $F_z$, rasio fundamental kunci) sebelum dikonsumsi oleh agen ReAct, guna menjaga integritas konteks dan efisiensi token.

---

## 2. Klausul Kepatuhan Hukum & Disclaimer Finansial Wajib

Setiap ekspor laporan, tampilan dashboard, dan ringkasan CLI wajib menyertakan klausul disclaimer berikut:

> **Pemberitahuan Kepatuhan Hukum & Risiko Pasar:**  
> *"Laporan investigasi ini dihasilkan secara otomatis oleh Niskava Agent untuk tujuan riset informasi, edukasi, dan intelijen pasar modal. Niskava bukan merupakan penasihat investasi, manajer investasi, atau pialang efek berizin. Seluruh temuan dan korelasi data yang disajikan bukan merupakan ajakan, tawaran, atau rekomendasi untuk membeli atau menjual efek tertentu. Seluruh keputusan investasi sepenuhnya merupakan tanggung jawab mandiri setiap pengguna."*
