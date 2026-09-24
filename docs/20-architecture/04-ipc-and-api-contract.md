# 04 — Spesifikasi Kontrak IPC & REST/SSE API

**Status:** ACCEPTED  
**Versi Dokumen:** 1.2.0  
**Terakhir Diperbarui:** 2026-09-20  

Dokumen ini mendefinisikan kontrak komunikasi data antar komponen sistem Niskava:
1. **Subprocess IPC Contract**: Protokol pertukaran data JSON Lines antara Go Core Daemon (`backend/core/ipc/`) dan Python Agent Engine (`backend/engine/`).
2. **Local REST & SSE Server Surface**: Spesifikasi antarmuka API HTTP dan streaming Server-Sent Events antara Go Core Server (`backend/core/server/`) dan Web Workspace / Klien eksternal.

---

## 1. Subprocess IPC Protocol (Go ↔ Python)

Go Core mengeksekusi Python Agent Engine sebagai child process melalui pemanggilan dinamis:

### Dua Mode Eksekusi IPC:
1. **Headless Investigation Pipeline (7-Stage Sequential SOP):**
   ```bash
   ${NISKAVA_PYTHON_BIN:-python3} -m ${NISKAVA_ENGINE_MODULE:-backend.engine.runner} \
     --ticker ANTM --days 30 --session INV-2026-0042
   ```
2. **Interactive Conversational ReAct Agent Turn:**
   ```bash
   ${NISKAVA_PYTHON_BIN:-python3} -m backend.engine.agent.react_agent \
     --prompt "Investigasi pergerakan anomali ANTM" --session-id "SES-UUID-42"
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

#### E. Event ReAct Loop: `agent_thought`
Menyampaikan penalaran internal model secara transparan ke UI:
```json
{
  "event": "agent_thought",
  "session_id": "SES-UUID-42",
  "thought": "Pengguna menanyakan anomali ANTM. Saya harus memanggil compute_quant_anomalies untuk memeriksa volume Z-score."
}
```

#### F. Event ReAct Loop: `agent_tool_call`
Menyampaikan eksekusi pemanggilan tool deterministik:
```json
{
  "event": "agent_tool_call",
  "session_id": "SES-UUID-42",
  "tool_name": "compute_quant_anomalies",
  "tool_args": {
    "symbol": "ANTM",
    "days": 30
  }
}
```

#### G. Event ReAct Loop: `agent_observation`
Menyampaikan hasil evaluasi dari tool deterministik:
```json
{
  "event": "agent_observation",
  "session_id": "SES-UUID-42",
  "tool_name": "compute_quant_anomalies",
  "result": {
    "anomalies_count": 1,
    "top_z_score": 3.84,
    "date": "2026-09-12"
  }
}
```

#### H. Event ReAct Loop: `agent_message_chunk`
Token teks streaming inkremental untuk rendering halus di klien:
```json
{
  "event": "agent_message_chunk",
  "session_id": "SES-UUID-42",
  "chunk": "Berdasarkan evaluasi kuantitatif deterministik pada data 30 hari..."
}
```

#### I. Event ReAct Loop: `agent_message_complete`
Penanda akhir pesan giliran percakapan:
```json
{
  "event": "agent_message_complete",
  "session_id": "SES-UUID-42",
  "content": "Berdasarkan evaluasi kuantitatif...",
  "thought": "Penalaran lengkap...",
  "tool_calls": []
}
```

#### J. Event: `session_complete` & `session_error`
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

## 2. Local REST & SSE Surface (Go Server ↔ React SPA / Clients)

Go Core Daemon menjalankan HTTP REST & Server-Sent Events (SSE) server lokal pada port default `8080` (dapat dikonfigurasi via flag `--port` atau env var `NISKAVA_PORT`).

### Katalog Endpoint REST API

| Method | Endpoint Path | Kegunaan | Request Payload / Query Params | Format Response |
|---|---|---|---|---|
| `GET` | `/api/health` | Health check & verifikasi runtime daemon | `-` | JSON status, version, market, timestamp |
| `GET` | `/api/chat/sessions` | Mengambil daftar sesi percakapan | `?limit=50&offset=0&q={search}` | JSON `{sessions: [...], total: N, limit: 50, offset: 0}` |
| `POST` | `/api/chat/sessions` | Membuat sesi percakapan baru | `{"title": "Analisis ANTM", "model": "hermes"}` | JSON `{session: {...}}` (HTTP 201) |
| `GET` | `/api/chat/sessions/{id}` | Detail metadata sesi tertentu | URL Param `{id}` | JSON `{session: {...}}` |
| `PATCH` | `/api/chat/sessions/{id}` | Update metadata sesi (rename, pin, status) | `{"title": "...", "is_pinned": true}` | JSON `{session: {...}}` |
| `DELETE` | `/api/chat/sessions/{id}` | Hapus sesi beserta seluruh pesan (cascade) | URL Param `{id}` | JSON `{"deleted": true, "id": "..."}` |
| `GET` | `/api/chat/sessions/{id}/messages` | Ambil riwayat seluruh pesan dalam sesi | URL Param `{id}` | JSON `{messages: [...]}` (termasuk thought & tool calls) |
| `POST` | `/api/chat/sessions/{id}/fork` | Cabangkan percakapan ke sesi baru | `{"title": "Fork Sesi", "up_to_message_id": "..."}` | JSON `{session: {...}}` (HTTP 201) |
| `POST` | `/api/chat/sessions/{id}/reset` | Kosongkan seluruh pesan dalam sesi | URL Param `{id}` | JSON `{"reset": true, "id": "..."}` |
| `POST` | `/api/chat/sessions/{id}/abort` | Batalkan eksekusi ReAct yang sedang streaming | URL Param `{id}` | JSON `{"aborted": true, "id": "..."}` |
| `GET` | `/api/chat/sessions/{id}/export` | Ekspor transkrip percakapan | `?format=markdown` atau `?format=json` | Markdown plain text / JSON payload terstruktur |
| `GET` | `/api/chat/search` | Pencarian pesan global lintas seluruh sesi | `?q={kata_kunci}` | JSON `{query: "...", count: N, results: [...]}` |
| `POST` | `/api/chat` | Eksekusi turn percakapan dengan streaming SSE | `{"prompt": "...", "session_id": "..."}` | `text/event-stream` (Server-Sent Events) |
| `GET` | `/api/graph/data` | Ekspor data node & edge graf memori lokal | `?session_id={id}&radius=2` | JSON `{nodes: [...], edges: [...]}` |
| `GET` | `/graph` | Halaman visualisasi graf interaktif Vis.js di browser | `-` | HTML interaktif Vis.js Network |
| `GET` | `/api/sessions` | Riwayat sesi investigasi pipeline headless | `-` | JSON daftar sesi `investigations` |
| `GET` | `/api/investigations/{id}` | Detail lengkap hasil anomali & bukti investigasi | URL Param `{id}` | JSON objek investigasi, anomalies, findings, timeline |
| `GET` | `/` | Web Workspace AI Assistant Canvas (Market Intelligence) | `-` | HTML/CSS/JS Single-Page Web Dashboard |

---

## 3. Protokol Streaming SSE (`POST /api/chat`)

Endpoint `POST /api/chat` menerima prompt pengguna dan mengalirkan respons ReAct secara real-time via Server-Sent Events (SSE).

### Header Response:
```http
HTTP/1.1 200 OK
Content-Type: text/event-stream
Cache-Control: no-cache
Connection: keep-alive
Access-Control-Allow-Origin: *
```

### Urutan Event SSE yang Dialirkan:

1. **`event: thought`**  
   Dialirkan ketika agen sedang menyusun hipotesis atau rencana pemanggilan tool.
   ```text
   event: thought
   data: {"thought": "Mendeteksi indikasi anomali volume pada ANTM..."}
   ```

2. **`event: tool_call`**  
   Dialirkan saat tool deterministik mulai dipanggil.
   ```text
   event: tool_call
   data: {"tool": "compute_quant_anomalies", "args": {"symbol": "ANTM", "days": 30}}
   ```

3. **`event: observation`**  
   Dialirkan ketika tool selesai menghasilkan observasi numerik/data.
   ```text
   event: observation
   data: {"tool": "compute_quant_anomalies", "result": {"z_score": 3.84}}
   ```

4. **`event: message_chunk`**  
   Dialirkan per token teks untuk respons narasi agent.
   ```text
   event: message_chunk
   data: {"chunk": "Terdeteksi "}
   ```

5. **`event: message_complete`**  
   Dialirkan ketika generasi pesan selesai dan seluruh pesan telah disimpan di SQLite.
   ```text
   event: message_complete
   data: {"content": "Terdeteksi lonjakan volume 3.84σ...", "thought": "...", "tool_calls": []}
   ```

6. **`event: done`**  
   Menandai akhir giliran percakapan. Sesi bertransisi dari `BUSY` kembali ke `IDLE`.
   ```text
   event: done
   data: {"status": "COMPLETED", "session_id": "SES-UUID-42"}
   ```

7. **`event: error`** (Jika terjadi kegagalan)  
   ```text
   event: error
   data: {"error": "Timeout saat memanggil upstream LLM"}
   ```

### Protokol Ketahanan Streaming SSE (SSE Streaming Resilience):
1. **Zero Write Deadline (`WriteTimeout: 0`)**: Server Go menonaktifkan deadline penulisan global pada `http.Server` dan menggunakan `http.NewResponseController(w).SetWriteDeadline(time.Time{})` pada endpoint `/api/chat`, mencegah terputusnya koneksi streaming saat LLM melakukan investigasi multi-tool berdurasi panjang (>60 detik).
2. **Periodic Heartbeat (`: keep-alive\n\n`)**: Server mengirim komentar SSE `: keep-alive\n\n` setiap 15 detik selama eksekusi tool atau inferensi model berlangsung. Komentar ini diabaikan oleh parser event SSE standar namun menjaga koneksi soket TCP dan proxy/gateway tetap aktif (*anti-idle*).
3. **1 MB Scanner Buffer**: Klien TUI dan web menggunakan buffer `bufio.NewScanner` berkapasitas 1 MB (`1024 * 1024`) untuk menjamin payload observasi berita dan tabel keuangan berukuran besar dapat dibaca tanpa memicu `bufio.ErrTooLong`.

### Manajemen Pembatalan Klien (Client Abort):
Klien web atau terminal dapat menghentikan streaming kapan saja dengan mengirimkan request `POST /api/chat/sessions/{id}/abort`. Go Core akan secara instan membatalkan context eksekusi runner Python dan mengembalikan sesi ke status `IDLE`.

