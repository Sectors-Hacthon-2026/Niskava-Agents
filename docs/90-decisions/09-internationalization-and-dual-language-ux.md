# 09 — Standardisasi Internasionalisasi (i18n) Dwibahasa & Dual-Language UX

**Status:** ACCEPTED  
**Tanggal:** 2026-09-22  
**Pengambil Keputusan:** Core Team  
**Dokumen Terkait:** [`../10-product/03-product-scope-and-surfaces.md`](../10-product/03-product-scope-and-surfaces.md), [`../10-product/05-hackathon-strategy.md`](../10-product/05-hackathon-strategy.md), [`01-hybrid-stack-go-python-react.md`](01-hybrid-stack-go-python-react.md)

---

## 1. Konteks & Permasalahan

Niskava Agent dibangun untuk pasar modal Indonesia (Bursa Efek Indonesia / IDX). Namun, partisipasi dalam Sectors Hackathon Indonesia 2026 melibatkan dewan juri gabungan (Supertype, Sectors, Algoritma) yang mengevaluasi repositori GitHub, video demo, dan kegunaan produk baik dari perspektif lokal maupun standar internasional:

* **Kendala Bahasa Tunggal**: Jika terminal dan laporan investigasi hanya dalam Bahasa Indonesia, daya tarik bagi evaluasi teknis global dan integrasi AI agents internasional (misalnya Claude Desktop / Cursor via MCP) berkurang.
* **Kebutuhan Analis Domestik**: Sebagian besar pelaku pasar modal ritel dan analis junior di Jakarta membutuhkan istilah dan narasi dalam Bahasa Indonesia yang akrab dengan regulasi BEI, OJK, dan media lokal (Kontan, Bisnis).
* **Konsistensi UI**: TUI Bubbletea sebelumnya mencampurkan teks bahasa Inggris dan Indonesia pada beberapa komponen layar (search bar placeholder, quick setup cards, filter status).

---

## 2. Keputusan Arsitektur

Niskava Agent menetapkan **Standardisasi i18n Dwibahasa Penuh (Bahasa Indonesia & English)** pada seluruh permukaan sistem:

### A. Dynamic TUI Localization Engine (`clients/cli/tui/i18n.go`)
1. **Penyedia Terjemahan Terpusat**: Seluruh string antarmuka TUI (launcher menu, HUD status, session history selector, help screen, quick setup wizard) dikelola dalam kamus pasangan kunci terpusat `translations[lang][key]`.
2. **Pengalihan Bahasa Instan (Instant Language Switching)**:
   * Pengguna dapat memilih preferensi bahasa melalui CLI flag global `--lang en` atau `--lang id`.
   * Di dalam TUI Launcher, pengguna dapat menekan tombol `L` kapan saja untuk beralih bahasa secara instan (*on-the-fly*).
3. **Bahasa Default**: Default disetel ke **English (`en`)** untuk kepatuhan evaluasi juri internasional, dengan transisi 1-tombol ke **Bahasa Indonesia (`id`)**.

### B. Propagasi Bahasa ke Python Agent Engine (`backend/core/ipc/`)
* Bahasa aktif di Go Core diteruskan ke child process Python melalui parameter flag `--language` dan environment variable `NISKAVA_LANG`.
* Python ReAct Agent menyesuaikan bahasa instruksi sintesis narasi akhir sesuai preferensi pengguna, sementara istilah finansial baku (*Z-score, Abnormal Return, Supported, Uncertain, Contradicted*) tetap dipertahankan sesuai taksonomi bukti.

---

## 3. Alternatif yang Dipertimbangkan

| Alternatif | Alasan Ditolak |
|---|---|
| **Hardcoded Bahasa Indonesia Saja** | Mengurangi skor usability dan daya tarik bagi penilai internasional serta integrasi tool LLM global. |
| **Hardcoded English Saja** | Mengurangi relevansi lokal terhadap istilah regulasi IDXnet, surat pengumuman suspensi BEI, dan artikel berita domestik. |
| **Pustaka Eksternal i18n yang Kompleks (Gettext/PO files)** | Menambah dependensi runtime eksternal yang memperberat binary Go dan menyulitkan kompilasi zero-CGO. |

---

## 4. Konsekuensi

### Positif
* **Skor Usability Maksimal (Rubrik 40%)**: Memuaskan analis internasional maupun lokal secara fleksibel.
* **Tampilan Terminal Bersih & Seragam**: Menghilangkan inkonsistensi teks campur aduk (*Indoglish*) pada antarmuka TUI.
* **Performa Nol Overhead**: Pemetaan string Go murni berbasis `map[string]string` yang sangat cepat tanpa operasi I/O disk tambahan.

### Negatif / Kompromi
* Pengembang wajib mendaftarkan string baru ke kamus `i18n.go` saat menambahkan fitur tampilan baru.
