# 04 — Spesifikasi Kontrak IPC & REST/SSE API

**Status:** ACCEPTED  
**Versi Dokumen:** 1.0.0  
**Terakhir Diperbarui:** 2026-09-16  

Dokumen ini mendefinisikan kontrak komunikasi data antar komponen sistem:
1. **Subprocess IPC Contract**: Komunikasi antara Go Core Daemon dan Python Agent Engine.
2. **Local REST & SSE Server Surface**: Komunikasi antara Go Core Server dan Web Workspace (React SPA).

---

## 1. Subprocess IPC Protocol (Go ↔ Python)

Go Core mengeksekusi Python Agent Engine sebagai child process melalui pemanggilan dinamis:
```bash
# Default invocation (relatif terhadap direktori kerja):
python3 -m engine.runner --ticker ANTM --days 30 --session INV-2026-0042

# Atau via path modul yang dikonfigurasi secara dinamis:
${NISKAVA_PYTHON_BIN:-python3} -m ${NISKAVA_ENGINE_MODULE:-engine.runner} --ticker ANTM --days 30 --session INV-2026-0042
```

> [!TIP]
> **Resolusi Biner & Modul Dinamis:** Parameter eksekusi di atas dapat disesuaikan melalui CLI flag `--python-bin` dan `--engine-path`, environment variable (`NISKAVA_PYTHON_BIN`, `NISKAVA_ENGINE_PATH`), atau berkas konfigurasi `~/.niskava/config.yaml`. Jika modul dipindahkan atau direfaktor, Go Core tetap dapat memanggil engine tanpa perlu kompilasi ulang kode Go.

### Format Pesan: Streaming JSON Lines (JSONL via STDOUT)
Setiap baris yang dicetak Python ke STDOUT merupakan objek JSON mandiri yang valid dan diakhiri dengan karakter newline `\n`. Go membaca STDOUT menggunakan buffer scanner baris demi baris.

### Skema Event IPC

#### A. Event: `session_start`
```json
{
  "event": "session_start",
  "session_id": "INV-2026-0042",
  "ticker": "ANTM",
  "timeframe_days": 30,
  "timestamp": "2026-09-16T08:00:00Z"
}
```

#### B. Event: `progress_step`
```json
{
  "event": "progress_step",
  "session_id": "INV-2026-0042",
  "stage": "SECTORS_BASELINE",
  "step_index": 1,
  "total_steps": 4,
  "message": "Mengambil 30 hari data candlestick dari Sectors v2 API..."
}
```

#### C. Event: `anomaly_detected`
```json
{
  "event": "anomaly_detected",
  "session_id": "INV-2026-0042",
  "ticker": "ANTM",
  "anomaly_date": "2026-09-12",
  "metric_type": "VOLUME_SPIKE",
  "z_score": 3.84,
  "metric_value": 184500000.0,
  "baseline_value": 48200000.0,
  "price_change_pct": 8.25,
  "sector_change_pct": 0.45,
  "description": "Lonjakan volume perdagangan 3.84x di atas rata-rata bergerak 20 hari."
}
```

#### D. Event: `finding_emitted`
```json
{
  "event": "finding_emitted",
  "session_id": "INV-2026-0042",
  "id": "FND-01",
  "title": "Katalis Ekspansi Smelter Nikel Terkonfirmasi",
  "claim_text": "Kenaikan volume didukung keterbukaan informasi peresmian fasilitas pengolahan nikel.",
  "verification_status": "SUPPORTED",
  "confidence_score": 0.94,
  "causality_status": "LIKELY_CATALYST",
  "evidence": [
    {
      "source_type": "OFFICIAL_DISCLOSURE",
      "source_name": "IDXnet / Keterbukaan Informasi BEI",
      "source_url": "https://www.idx.co.id/...",
      "publication_date": "2026-09-12T08:30:00Z",
      "snippet_text": "Perseroan telah menyelesaikan commissioning unit smelter feronikel."
    }
  ]
}
```

#### E. Event: `session_complete` & `session_error`
```json
{
  "event": "session_complete",
  "session_id": "INV-2026-0042",
  "status": "COMPLETED",
  "total_anomalies": 1,
  "total_findings": 3,
  "duration_ms": 4250,
  "summary": "Investigasi selesai. 2 temuan SUPPORTED, 1 temuan UNCERTAIN."
}
```

---

## 2. Local REST & SSE Surface (Go Server ↔ React SPA)

Go Core Daemon menjalankan server HTTP lokal pada port default `8080`.

### Katalog Endpoint REST
| Method | Path | Kegunaan | Payload Request |
|---|---|---|---|
| `GET` | `/api/v1/health` | Status server dan verifikasi SQLite lokal | `-` |
| `GET` | `/api/v1/sessions` | Mengambil daftar riwayat seluruh investigasi | Query: `?limit=20&page=1` |
| `GET` | `/api/v1/sessions/:id` | Detail lengkap sesi investigasi (anomali, temuan, timeline) | `-` |
| `POST` | `/api/v1/investigate` | Memulai investigasi baru secara asynchronous | `{"ticker": "ANTM", "days": 30}` |

### Server-Sent Events (SSE) Endpoint
* **Path**: `GET /api/v1/investigations/:id/stream`
* **Header**: `Content-Type: text/event-stream`, `Cache-Control: no-cache`
* **Format**:
```text
event: progress
data: {"step": 2, "message": "Deteksi anomali kuantitatif selesai..."}

event: anomaly
data: {"ticker": "ANTM", "date": "2026-09-12", "z_score": 3.84}

event: finding
data: {"title": "Smelter Commissioning", "status": "SUPPORTED", "confidence": 0.94}

event: done
data: {"status": "COMPLETED", "session_id": "INV-2026-0042"}
```
React Web Workspace mendengarkan stream ini menggunakan objek `EventSource` bawaan browser.
