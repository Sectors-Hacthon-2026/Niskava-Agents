# 05 — Batasan Regulasi & Larangan Nasihat Finansial Non-Advisory

**Status:** ACCEPTED  
**Tanggal:** 2026-09-16  
**Pengambil Keputusan:** Core Team  

---

## 1. Konteks & Permasalahan

Banyak produk AI finansial berupaya memikat pengguna dengan menghasilkan sinyal trading otomatis: *"REKOMENDASI: BUY saham ANTM sekarang dengan target harga 1.800"*.

Dalam lanskap regulasi pasar modal Indonesia (UU Pasar Modal No. 8 Tahun 1995 dan peraturan Otoritas Jasa Keuangan / OJK):
1. Memberikan rekomendasi investasi tanpa izin Penasihat Investasi atau Wakil Manajer Investasi (WMI) resmi adalah tindakan ilegal yang memiliki konsekuensi pidana dan perdata.
2. Pasar saham adalah domain stokastik berisiko tinggi; klaim kepastian harga oleh AI dapat menyebabkan kerugian modal masif bagi investor ritel.

---

## 2. Keputusan

**Mengunci batasan arsitektur Non-Advisory secara mutlak (*Strict Financial Non-Advisory Boundary*):**
1. **Larangan Sinyal Beli/Jual**: Agen dilarang keras menghasilkan rekomendasi trading langsung (`BUY`, `SELL`, `HOLD`, `TARGET PRICE`).
2. **Fokus Murni Intelijen & Investigasi Fakta**: Agen hanya diizinkan menyajikan data anomali kuantitatif, korelasi waktu, fakta keterbukaan informasi resmi, dan status validasi bukti (`SUPPORTED`, `UNCERTAIN`, `CONTRADICTED`).
3. **Penyertaan Klausul Disclaimer Wajib**: Setiap output laporan investigasi (CLI, Web Workspace, dan file ekspor) wajib menyertakan klausul disclaimer kepatuhan hukum standar.

---

## 3. Alternatif yang Dipertimbangkan

| Alternatif | Alasan Ditolak |
|---|---|
| **Memberikan Sinyal Buy/Sell dengan Disclaimer Lemah** | Melanggar regulasi OJK, membahayakan kredibilitas tim di mata dewan juri hackathon, dan mengekspos tim pada risiko hukum. |
| **Menyediakan Probabilitas Kenaikan Harga (Prediksi AI)** | Prediksi harga berbasis LLM bersifat spekulatif dan rawan halusinasi, bertentangan dengan prinsip dasar P1 (*Evidence over assumption*). |

---

## 4. Konsekuensi

### Positif
* **Kepatuhan Regulasi 100%**: Aman secara hukum di bawah yurisdiksi Republik Indonesia dan regulasi OJK.
* **Integritas Platform**: Sistem dipersepsikan oleh investor institusi dan dewan juri sebagai instrumen intelijen profesional yang kredibel, bukan bot trading spekulatif.
* **Toleransi Nol Terhadap Disinformasi**: Melindungi investor ritel dari manipulasi rumor gorengan saham.

### Negatif / Kompromi yang Diterima
* Pengguna ritel spekulatif yang mencari "contekan saham cepat" mungkin merasa kurang dimanjakan, namun pengguna profesional justru lebih menghargai objektivitas investigasi data.
