# Daftar Pertanyaan Terbuka & Keputusan Tertunda (Open Questions)

**Status:** ACTIVE REGISTER  
**Versi Dokumen:** 1.1.0  
**Terakhir Diperbarui:** 2026-09-16  

Daftar isu desain, pertimbangan arsitektur, dan pertanyaan produk yang masih dalam tahap evaluasi. Dokumen ini diperbarui secara berkala hingga keputusan final dikunci ke dalam Architecture Decision Record (ADR).

---

## 1. Register Pertanyaan Terbuka

| ID | Topik & Pertanyaan | Status | Keputusan / Dampak Arsitektur | Target Keputusan |
|---|---|:---:|---|---|
| **OQ-01** | **Pilihan LLM Default untuk MVP**: Apakah mengunci Gemini 2.0 Flash (Cloud) sebagai default, atau menyediakan opsi fallback lokal via Ollama? | `RESOLVED` | **Keputusan:** Dikunci pada **Gemini 2.0 Flash** via API. Menjamin target latensi end-to-end < 6 detik (P95), kestabilan video demo, dan kemudahan evaluasi repositori oleh dewan juri tanpa perlu setup GPU lokal. | 16 Sep 2026 |
| **OQ-02** | **Parameter Ambang Batas Saham Lapis Tiga**: Apakah threshold $V_z \ge 2.5$ perlu dinaikkan untuk saham dengan market cap kecil (< IDR 1T) yang sering memiliki volatilitas liar harian? | `PROPOSED` | Modul deteksi anomali kuantitatif di `20-architecture/05`. Rencana usulan: $V_z \ge 3.5$ untuk saham non-LQ45. | Tahap Tuning (23 Sep 2026) |
| **OQ-03** | **Integrasi Langsung PDF Keterbukaan Informasi BEI**: Apakah parser PDF langsung untuk dokumen pengumuman IDXnet perlu masuk ke scope MVP, atau cukup mengandalkan teks agregasi Sectors v2 News? | `RESOLVED` | **Keputusan:** Scope MVP difokuskan pada **Sectors v2 News API + Tavily/DuckDuckGo** untuk menjamin MVP berfungsi mulus (*working prototype*) sebelum deadline pembekuan kode 30 September. Parsing raw PDF dijadwalkan untuk v2.0. | 16 Sep 2026 |
| **OQ-04** | **Watchlist & Background Daemon Mode**: Apakah agen perlu menjalankan background loop untuk memonitor watchlist saham secara otomatis setiap penutupan bursa (16:00 WIB)? | `DEFERRED` | Ditaruh di roadmap pasca-hackathon (v2.0) agar tim fokus menyempurnakan alur investigasi on-demand per emiten di v1.0. | Pasca-Hackathon |

---

## 2. Prosedur Penyelesaian

Ketika sebuah pertanyaan terbuka di atas disepakati solusinya:
1. Status diubah menjadi `RESOLVED`.
2. Keputusan dicatat ke dalam dokumen spesifikasi yang relevan atau dibuatkan ADR baru di `docs/90-decisions/`.
3. Tautan ADR dicantumkan pada tabel di atas sebagai riwayat keputusan.
