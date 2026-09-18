# 05 — Formula Deteksi Anomali Kuantitatif Deterministik

**Status:** LOCKED  
**Versi Dokumen:** 1.0.0  
**Terakhir Diperbarui:** 2026-09-16  

Modul ini mengeksekusi perhitungan statistik deret waktu (*time-series*) menggunakan Python murni (`numpy`/`math`) secara deterministik sebelum memanggil model bahasa (LLM). Hal ini menjamin konsistensi matematis 100%, menghilangkan halusinasi numerik, dan menghemat biaya token prompt.

---

## 1. Formula Matematis & Ambang Batas (Thresholds)

### A. Volume Anomaly $Z$-Score ($V_z$)
Mengukur signifikansi statistik volume perdagangan pada hari $t$ dibandingkan dengan distribusi volume 20 hari bursa sebelumnya:

$$\mu_{20} = \frac{1}{20} \sum_{i=1}^{20} V_{t-i}$$

$$\sigma_{20} = \sqrt{\frac{1}{20} \sum_{i=1}^{20} (V_{t-i} - \mu_{20})^2}$$

$$V_z = \frac{V_t - \mu_{20}}{\sigma_{20}}$$

* **Ambang Batas Trigger**: $V_z \ge 2.5$ diklasifikasikan sebagai `VOLUME_SPIKE` (kemungkinan terjadi secara acak $< 0.6\%$).

---

### B. Abnormal Return ($R_t$)
Mengukur perubahan persentase harga penutupan saham harian:

$$R_t = \frac{P_{\text{close}, t} - P_{\text{close}, t-1}}{P_{\text{close}, t-1}} \times 100\%$$

* **Ambang Batas Trigger**: $|R_t| \ge 5.0\%$ diklasifikasikan sebagai `PRICE_BREAKOUT`.

---

### C. Divergensi Sektoral ($D_t$)
Memisahkan apakah anomali disebabkan oleh tren umum industri (misal: seluruh sektor tambang naik karena harga komoditas global) atau spesifik emiten (*idiosyncratic*):

$$D_t = R_{\text{stock}, t} - R_{\text{sector}, t}$$

* **Ambang Batas Trigger**: $|D_t| \ge 4.0\%$ menandakan pergerakan bersifat **Idiosyncratic Catalyst** (spesifik aksi korporasi atau berita emiten bersangkutan).

---

## 2. Klasifikasi Matriks Anomali

| Kondisi $V_z$ | Kondisi $|R_t|$ | Kondisi $|D_t|$ | Klasifikasi Sistem | Tindakan Agen |
|---|---|---|---|---|
| $\ge 2.5$ | $\ge 5.0\%$ | $\ge 4.0\%$ | `IDIOSYNCRATIC_CATALYST` | Panggil Skill `event-correlation` dengan prioritas tinggi. |
| $\ge 2.5$ | $< 5.0\%$ | Bebas | `VOLUME_ACCUMULATION` | Periksa data *Broker Summary* (Bandarmology) untuk pihak pembeli. |
| $< 2.5$ | $\ge 5.0\%$ | $< 4.0\%$ | `SECTOR_BETA_RALLY` | Tandai sebagai pergerakan makro sektor, bukan katalis internal. |

---

## 3. Contoh Implementasi Python

```python
import numpy as np

def detect_anomalies(daily_data: list[dict], sector_return: float) -> list[dict]:
    """Menghitung anomali volume dan return secara deterministik."""
    if len(daily_data) < 21:
        return []
    
    anomalies = []
    # Evaluasi hari bursa terbaru terhadap jendela 20 hari sebelumnya
    volumes = [d["volume"] for d in daily_data[:-1]]
    mu_20 = float(np.mean(volumes[-20:]))
    sigma_20 = float(np.std(volumes[-20:]))
    
    current_day = daily_data[-1]
    v_t = current_day["volume"]
    v_z = (v_t - mu_20) / sigma_20 if sigma_20 > 0 else 0.0
    
    prev_close = daily_data[-2]["close"]
    curr_close = current_day["close"]
    r_t = ((curr_close - prev_close) / prev_close) * 100.0
    divergence = r_t - sector_return
    
    if v_z >= 2.5 or abs(r_t) >= 5.0:
        anomalies.append({
            "date": current_day["date"],
            "z_score": round(v_z, 2),
            "price_change_pct": round(r_t, 2),
            "divergence_pct": round(divergence, 2),
            "classification": "IDIOSYNCRATIC_CATALYST" if abs(divergence) >= 4.0 else "GENERAL_MARKET"
        })
    return anomalies
```
