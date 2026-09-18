# 03 — Validasi Bukti, Kausalitas & Model Keyakinan

**Status:** LOCKED  
**Versi Dokumen:** 1.0.0  
**Terakhir Diperbarui:** 2026-09-16  

---

## 1. Taksonomi Status Verifikasi Bukti (3 Tingkat)

Setiap temuan (*finding*) yang diproduksi oleh Niskava Agent wajib diklasifikasikan ke dalam salah satu dari tiga label status verifikasi berikut:

```text
┌─────────────────┐      Didukung data kuantitatif Sectors v2 atau
│    SUPPORTED    │ ───▶ dokumen resmi keterbukaan informasi BEI / OJK.
└─────────────────┘      Tingkat kepastian tinggi (High Confidence).

┌─────────────────┐      Terdapat kedekatan waktu dengan berita media pihak
│    UNCERTAIN    │ ───▶ ketiga atau rumor pasar, namun belum ada konfirmasi resmi.
└─────────────────┘      Tingkat kepastian tentatif (Medium-Low Confidence).

┌─────────────────┐      Klaim atau narasi pasar bertentangan secara diametral
│  CONTRADICTED   │ ───▶ dengan fakta empiris laporan keuangan resmi.
└─────────────────┘      Membantah rumor / disinformasi pasar secara objektif.
```

### Contoh Kasus Nyata:
1. **`SUPPORTED`**: *"Lonjakan volume transaksi ANTM pada 12 September didukung oleh keterbukaan informasi penyelesaian commissioning fasilitas smelter feronikel Halmahera Timur."*
2. **`UNCERTAIN`**: *"Beredar rumor di forum saham mengenai potensi akuisisi tambang emas di Sumbawa, namun belum ditemukan pengumuman resmi di IDXnet."*
3. **`CONTRADICTED`**: *"Klaim bahwa emiten mengalami krisis likuiditas dibantah oleh data neraca Sectors v2 yang mencatat kas bersih IDR 4.2 Triliun dan Quick Ratio 2.1x."*

---

## 2. Penilaian Hubungan Kausalitas vs Korelasi Waktu

Agen dilarang menyimpulkan hubungan sebab-akibat murni dari kemiripan narasi. Agen wajib memvalidasi urutan stempel waktu (*temporal precedence*):

| Urutan Stempel Waktu Kejadian | Analisis Logika Agen | Status Kausalitas |
|---|---|---|
| Berita resmi rilis pukul 08:30 WIB $\rightarrow$ Volume melonjak tajam pukul 10:00 WIB. | Berita terbit sebelum lonjakan $\rightarrow$ Reaksi wajar pasar terhadap informasi material baru. | `LIKELY_CATALYST` |
| Volume melonjak tajam hari Selasa $\rightarrow$ Pengumuman resmi baru dirilis hari Kamis. | Volume mendahului informasi $\rightarrow$ Indikasi potensi asimetri informasi atau kebocoran berita. | `PRECEDED_ANNOUNCEMENT` |
| Volume perdagangan melonjak drastis, namun tidak ditemukan pengumuman atau berita relevan. | Anomali murni pergerakan modal / transaksi negosiasi antar broker (*Bandarmology*). | `UNEXPLAINED_BY_NEWS` |

---

## 3. Formula Skor Keyakinan (Confidence Score)

Skor keyakinan ($C \in [0.00, 1.00]$) dihitung secara algoritmik:

$$C = w_{\text{source}} \times w_{\text{temporal}} \times w_{\text{consistency}}$$

* **$w_{\text{source}}$**: Bobot otoritas sumber (Tier 1 Resmi = $1.0$, Tier 2 Media Kredibel = $0.8$, Tier 3 Komunitas = $0.4$).
* **$w_{\text{temporal}}$**: Kedekatan waktu kejadian dengan lonjakan anomali ($\Delta t \le 24\text{ jam} = 1.0$, $\Delta t \le 48\text{ jam} = 0.8$, $\Delta t > 48\text{ jam} = 0.5$).
* **$w_{\text{consistency}}$**: Konsistensi data broker/volume Sectors yang mendukung narasi ($1.0$ jika volume melonjak bersamaan, $0.7$ jika tidak ada lonjakan volume).
