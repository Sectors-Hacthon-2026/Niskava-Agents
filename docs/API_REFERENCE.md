# Niskava Daemon REST & SSE API Reference

Spesifikasi antarmuka Application Programming Interface (API) untuk Niskava Agent Daemon Server (`backend/core/server`). Daemon berjalan secara lokal sebagai HTTP server (default port `20128`) yang melayani REST endpoints dan Server-Sent Events (SSE) streaming untuk Web Workspace Dashboard, CLI, dan integrasi pihak ketiga.

---

## 1. Spesifikasi Umum

### 1.1 Base URL
```text
http://localhost:20128
```
Port dapat dikonfigurasi melalui:
1. CLI Flag: `niskava serve --port <PORT>`
2. Environment Variable: `NISKAVA_PORT=<PORT>`
3. Konfigurasi Lokal: `~/.niskava/config.yaml` (`server.port`)

### 1.2 Format Komunikasi
* **Request Header:**
  * `Content-Type: application/json`
  * `Accept: application/json` (REST) atau `text/event-stream` (SSE)
* **Response Header:**
  * `Content-Type: application/json; charset=utf-8` (REST)
  * `Content-Type: text/event-stream; charset=utf-8` (SSE)
* **CORS:**
  * Daemon mengizinkan Cross-Origin Resource Sharing (`Access-Control-Allow-Origin: *`) untuk mendukung komunikasi frontend Vite development server (`localhost:5173`) maupun production web bundles.
  * Mendukung HTTP Methods: `GET`, `POST`, `PATCH`, `PUT`, `DELETE`, `OPTIONS`.

### 1.3 Format Respon Error Standar
Setiap error non-2xx mengembalikan format JSON standar:
```json
{
  "error": "Pesan deskripsi kesalahan teknis atau validasi"
}
```

---

## 2. Ringkasan Endpoint

| Kategori | Method | Path | Deskripsi Singkat |
|---|---|---|---|
| **System & Health** | `GET` | `/api/health` | Health check status server dan pasar IDX |
| | `GET` | `/api/system/diagnostics` | Metrik lingkungan runtime, OS, CPU, dan database |
| | `GET` | `/api/system/sectors-usage` | Statistik disiplin kuota kredit & cache Sectors v2 |
| | `POST` | `/api/system/cache/clean` | Pembersihan record cache kadaluarsa |
| **Settings & Connectivity** | `GET` | `/api/settings` | Membaca konfigurasi aktif dengan kunci tersensor |
| | `PATCH` | `/api/settings` | Memperbarui konfigurasi sistem dan hot-reload |
| | `POST` | `/api/settings/test-connection` | Validasi koneksi langsung ke provider eksternal |
| **Telegram Bot Daemon** | `GET` | `/api/settings/telegram` | Status layanan dan konfigurasi Telegram Bot |
| | `PATCH` | `/api/settings/telegram` | Memperbarui token bot dan daftar whitelist user |
| | `POST` | `/api/telegram/start` | Menjalankan bot poller di background |
| | `POST` | `/api/telegram/stop` | Menghentikan bot poller secara graceful |
| | `POST` | `/api/telegram/test` | Mengirim notifikasi uji coba ke chat ID |
| **Memory Graph** | `GET` | `/api/graph` | Mengambil data node & edge terfilter untuk canvas |
| | `GET` | `/api/graph/stats` | Agregasi metrik entitas graf dan top hub nodes |
| | `GET` | `/graph` | Merender visualisasi graf HTML standalone (PyVis) |
| **Chat & Assistant** | `POST` | `/api/chat` | Streaming ReAct agent turn melalui SSE |
| | `GET` | `/api/chat/history` | Riwayat pesan obrolan per sesi |
| | `POST` | `/api/chat/abort` | Membatalkan streaming agen yang sedang berjalan |
| **Session Management** | `GET` | `/api/chat/sessions` | Mengambil daftar sesi percakapan dengan paginasi |
| | `POST` | `/api/chat/sessions` | Membuat sesi percakapan baru |
| | `GET` | `/api/chat/sessions/{id}` | Detail metadata sesi percakapan |
| | `PATCH` | `/api/chat/sessions/{id}` | Mengubah judul, pin status, atau status sesi |
| | `DELETE` | `/api/chat/sessions/{id}` | Menghapus sesi dan riwayat terkait |
| | `POST` | `/api/chat/sessions/{id}/fork` | Menduplikasi percakapan (OpenCode pattern) |
| | `GET` | `/api/chat/search` | Pencarian teks pesan lintas sesi |
| **Investigation Pipeline** | `GET` | `/api/investigations` | Mengambil daftar riwayat investigasi kuantitatif |
| | `POST` | `/api/investigations` | Menjalankan investigasi otonom 7-Stage pipeline |
| | `GET` | `/api/investigations/{id}` | Detail dossier temuan, anomali, dan bukti OSINT |

---

## 3. Detail Spesifikasi Endpoint

### 3.1 System & Health

#### `GET /api/health`
Mengembalikan status ketersediaan daemon server dan penanda pasar aktif.

* **Response Headers:** `200 OK`
* **Response Body:**
```json
{
  "status": "ok",
  "app": "Niskava Agent",
  "version": "1.0.0",
  "market": "IDX",
  "timestamp": "2026-09-24T12:00:00Z"
}
```

---

#### `GET /api/system/diagnostics`
Menyajikan informasi diagnostik infrastruktur lokal daemon, resource host, dan kondisi basis data SQLite Write-Ahead Logging (WAL).

* **Response Headers:** `200 OK`
* **Response Body:**
```json
{
  "status": "OK",
  "app": "Niskava Agent",
  "go_version": "go1.24.0",
  "os": "linux",
  "arch": "amd64",
  "num_cpu": 16,
  "database_path": "/home/user/.niskava/niskava.db",
  "database_size_bytes": 1048576,
  "total_sessions": 24,
  "timestamp": "2026-09-24T12:00:00Z"
}
```

---

#### `GET /api/system/sectors-usage`
Menampilkan efisiensi *credit budget* dan performa cache lokal pada tabel `sectors_cache` untuk kepatuhan terhadap Law 5 (disiplin kuota 1.000 kredit Sectors).

* **Response Headers:** `200 OK`
* **Response Body:**
```json
{
  "total_entries": 152,
  "expired_entries": 14,
  "permanent_entries": 138,
  "estimated_credit_saved": 152,
  "status": "HEALTHY"
}
```

---

#### `POST /api/system/cache/clean`
Menghapus seluruh rekaman cache Sectors yang masa berlakunya telah melewati waktu saat ini (`expires_at < CURRENT_TIMESTAMP`). Data historis daily OHLCV permanen (`expires_at IS NULL`) tidak akan terhapus.

* **Response Headers:** `200 OK`
* **Response Body:**
```json
{
  "status": "ok",
  "cleaned_entries": 14,
  "timestamp": "2026-09-24T12:00:00Z"
}
```

---

### 3.2 Settings & Connectivity

#### `GET /api/settings`
Membaca konfigurasi sistem saat ini. Seluruh token rahasia disensor secara otomatis demi keamanan data.

* **Response Headers:** `200 OK`
* **Response Body:**
```json
{
  "auth": {
    "ai_provider": "gemini",
    "sectors_api_key": "sec_****a1b2",
    "has_sectors_key": true,
    "sectors_base_url": "https://api.sectors.app/v2",
    "gemini_api_key": "AIza****3f9a",
    "has_gemini_key": true,
    "gemini_model": "gemini-2.0-flash",
    "openai_api_key": "",
    "has_openai_key": false,
    "openai_base_url": "http://localhost:20128/v1",
    "openai_model": "hermes",
    "anthropic_api_key": "",
    "has_anthropic_key": false,
    "ollama_base_url": "http://localhost:11434",
    "ollama_model": "qwen2.5:7b"
  },
  "storage": {
    "db_path": "/home/user/.niskava/niskava.db"
  },
  "engine": {
    "python_bin": "python3",
    "entrypoint": "backend.engine.runner"
  },
  "server": {
    "port": 20128
  },
  "preferences": {
    "default_market": "IDX",
    "offline_mode": false,
    "language": "id"
  },
  "memory": {
    "enabled": true,
    "decay_lambda": 0.05,
    "ego_radius": 2,
    "max_context_tokens": 4096
  },
  "telegram": {
    "bot_token": "7123****xyz9",
    "has_token": true,
    "enabled": false,
    "allowed_users": ["analyst_1", "12345678"]
  }
}
```

---

#### `PATCH /api/settings`
Memperbarui konfigurasi sistem secara parsial, menyimpannya ke berkas konfigurasi lokal (`~/.niskava/config.yaml`) dengan izin POSIX `0600`, dan me-reload state konfigurasi runtime pada daemon Go Core.

> [!NOTE]
> Jika field kunci API bernilai string sensor (mengandung `****`) atau string kosong, backend akan mengabaikan nilai tersebut dan mempertahankan kunci asli yang tersimpan.

* **Request Body (Semua field bersifat opsional):**
```json
{
  "auth": {
    "ai_provider": "gemini",
    "gemini_model": "gemini-2.5-flash",
    "sectors_api_key": "sec_live_000728ab231f7d90"
  },
  "preferences": {
    "language": "id",
    "offline_mode": false
  },
  "telegram": {
    "enabled": true
  }
}
```
* **Response Headers:** `200 OK`
* **Response Body:** Mengembalikan struktur `ConfigView` tersanitasi terbaru (sama seperti `GET /api/settings`).

---

#### `POST /api/settings/test-connection`
Melakukan pengujian konektivitas live ke endpoint layanan target untuk memvalidasi kredensial dan mengukur waktu latensi jaringan.

* **Request Body:**
```json
{
  "target": "sectors",
  "api_key": "sec_live_000728ab231f7d90",
  "base_url": "",
  "model": ""
}
```
*Field `target` valid:* `"sectors"`, `"gemini"`, `"openai"`, `"ollama"`, `"anthropic"`.
* **Response Headers:** `200 OK`
* **Response Body (Sukses):**
```json
{
  "target": "sectors",
  "success": true,
  "message": "Connected to Sectors v2 API successfully",
  "latency_ms": 142
}
```
* **Response Body (Gagal):**
```json
{
  "target": "sectors",
  "success": false,
  "message": "Invalid Sectors API key (Unauthorized)",
  "latency_ms": 95
}
```

---

### 3.3 Telegram Bot Daemon

#### `GET /api/settings/telegram`
Menampilkan status daemon bot Telegram, username bot, dan konfigurasi otorisasi.

* **Response Headers:** `200 OK`
* **Response Body:**
```json
{
  "status": "RUNNING",
  "bot_username": "NiskavaAgentBot",
  "enabled": true,
  "allowed_users": ["analyst_1", "12345678"],
  "has_token": true
}
```
*Status yang mungkin:* `"RUNNING"`, `"STOPPED"`, `"NOT_CONFIGURED"`.

---

#### `PATCH /api/settings/telegram`
Memperbarui konfigurasi bot Telegram.

* **Request Body:**
```json
{
  "bot_token": "7123456789:AAHxyz...",
  "enabled": true,
  "allowed_users": ["analyst_1", "98765432"]
}
```
* **Response Headers:** `200 OK`
* **Response Body:** Struktur status Telegram yang telah diperbarui.

---

#### `POST /api/telegram/start`
Menginisialisasi dan memulai *long-polling worker* bot Telegram di background goroutine.

* **Response Headers:** `200 OK`
* **Response Body:**
```json
{
  "status": "RUNNING",
  "message": "telegram bot started successfully",
  "username": "NiskavaAgentBot"
}
```

---

#### `POST /api/telegram/stop`
Menghentikan *long-polling worker* bot Telegram secara *graceful* melalui context cancellation.

* **Response Headers:** `200 OK`
* **Response Body:**
```json
{
  "status": "STOPPED",
  "message": "telegram bot stopped successfully"
}
```

---

#### `POST /api/telegram/test`
Mengirimkan pesan uji coba verifikasi ke akun/chat Telegram tertentu.

* **Request Body:**
```json
{
  "chat_id": 123456789
}
```
* **Response Headers:** `200 OK`
* **Response Body:**
```json
{
  "status": "ok",
  "chat_id": 123456789,
  "message": "test message sent successfully"
}
```

---

### 3.4 Memory Graph & Analytics

#### `GET /api/graph`
Mengambil data simpul (*nodes*) dan relasi (*edges*) memori asosiatif untuk visualisasi canvas graf (Cytoscape / React Flow).

* **Query Parameters:**
  * `session_id` (string, opsional): Memfilter graf khusus untuk sesi tertentu.
  * `ticker` (string, opsional): Mengaktifkan mode *Ego-Network* yang berpusat pada entitas emiten tertentu.
  * `depth` (integer, opsional, default `1`, max `2`): Kedalaman hop ekspansi *ego-network* sesuai batasan Law 6.
  * `min_weight` (float, opsional, default `0.0`): Bobot minimum relasi asosiasi.
  * `node_types` (string, opsional): Daftar tipe node dipisah koma (contoh: `TICKER,BROKER,CATALYST`).
* **Response Headers:** `200 OK`
* **Response Body:**
```json
{
  "session_id": "SES-UUID-42",
  "total_nodes": 3,
  "total_edges": 2,
  "nodes": [
    {
      "id": "node:ANTM",
      "label": "ANTM",
      "node_type": "TICKER",
      "metadata_json": null,
      "last_observed_at": "2026-09-24T12:00:00Z",
      "created_at": "2026-09-24T11:00:00Z"
    },
    {
      "id": "node:NICKEL",
      "label": "Nickel Commodity",
      "node_type": "CATALYST",
      "metadata_json": null,
      "last_observed_at": "2026-09-24T12:00:00Z",
      "created_at": "2026-09-24T11:00:00Z"
    }
  ],
  "edges": [
    {
      "source_id": "node:ANTM",
      "target_id": "node:NICKEL",
      "relation": "EXPOSED_TO",
      "context_snippet": "Kenaikan harga nikel global mendorong pendapatan emiten",
      "session_id": "SES-UUID-42",
      "weight": 2.5,
      "confidence_score": 0.95,
      "last_observed_at": "2026-09-24T12:00:00Z",
      "created_at": "2026-09-24T11:00:00Z"
    }
  ]
}
```

---

#### `GET /api/graph/stats`
Menghasilkan metrik agregasi seluruh memori asosiatif untuk widget ringkasan dan analisis sentralitas entitas pasar.

* **Response Headers:** `200 OK`
* **Response Body:**
```json
{
  "total_nodes": 45,
  "total_edges": 82,
  "node_types": {
    "TICKER": 12,
    "CATALYST": 18,
    "BROKER": 8,
    "REGULATOR": 4,
    "ANOMALY": 3
  },
  "top_hub_nodes": [
    {
      "id": "node:ANTM",
      "label": "ANTM",
      "node_type": "TICKER",
      "degree": 14
    },
    {
      "id": "node:NICKEL",
      "label": "Nickel Commodity",
      "node_type": "CATALYST",
      "degree": 9
    }
  ]
}
```

---

#### `GET /graph`
Merender antarmuka visualisasi graf interaktif standalone berbasis PyVis HTML untuk embedding pada `<iframe>` atau inspeksi cepat di browser.

* **Query Parameters:**
  * `session_id` (string, opsional): Filter visualisasi sesi.
* **Response Headers:** `Content-Type: text/html; charset=utf-8`

---

### 3.5 Conversational Assistant & SSE Streaming

#### `POST /api/chat`
Menjalankan turn percakapan investigasi ReAct. Respon dipancarkan secara real-time menggunakan format **Server-Sent Events (SSE)**.

* **Request Body:**
```json
{
  "prompt": "Analisis volume spike pada saham ANTM dalam 3 hari terakhir",
  "session_id": "SES-UUID-42"
}
```
*Jika `session_id` tidak disertakan, server akan membuat sesi baru secara otomatis.*

* **Response Headers:**
  * `Content-Type: text/event-stream`
  * `Cache-Control: no-cache`
  * `Connection: keep-alive`

* **Protokol Event SSE:**
Setiap event dikirimkan dalam format:
```text
event: <EVENT_TYPE>
data: <JSON_PAYLOAD>

```

##### Daftar Event SSE yang Dipancarkan:
1. `event: session_start`: Inisiasi sesi dan alokasi ID sesi.
2. `event: progress_step`: Indikator progres tahapan pipeline ReAct.
3. `event: tool_call_start`: Tanda bahwa agen memutuskan memanggil tool deterministik tertentu.
   ```json
   { "event": "tool_call_start", "tool": "quant_compute_anomalies", "input": {"ticker": "ANTM"} }
   ```
4. `event: tool_call_result`: Hasil eksekusi tool kuantitatif atau OSINT.
5. `event: agent_message_chunk`: Fragmen teks token narasi jawaban analisis (streaming Markdown).
   ```json
   { "event": "agent_message_chunk", "chunk": "Berdasarkan evaluasi statistik, " }
   ```
6. `event: done`: Sesi giliran ReAct selesai, response tersimpan permanen ke basis data SQLite.
   ```json
   { "session_id": "SES-UUID-42" }
   ```
7. `event: session_error`: Terjadi interupsi jaringan atau eksekusi.

---

#### `POST /api/chat/abort`
Membatalkan eksekusi streaming analisis ReAct yang sedang berjalan untuk sesi tertentu (*OpenCode abort pattern*).

* **Request Body:**
```json
{
  "session_id": "SES-UUID-42"
}
```
* **Response Headers:** `200 OK`
* **Response Body:**
```json
{
  "status": "aborted",
  "session_id": "SES-UUID-42"
}
```

---

#### `GET /api/chat/history`
Mengambil riwayat percakapan untuk sesi tertentu.

* **Query Parameters:**
  * `session_id` (string, wajib): ID sesi obrolan.
  * `limit` (integer, opsional, default `50`): Batas jumlah pesan yang diambil.
* **Response Headers:** `200 OK`
* **Response Body:**
```json
{
  "session_id": "SES-UUID-42",
  "total": 4,
  "messages": [
    {
      "id": "MSG-1727181001",
      "session_id": "SES-UUID-42",
      "role": "user",
      "content": "Analisis volume spike saham ANTM",
      "status": "COMPLETED",
      "created_at": "2026-09-24T12:00:00Z"
    },
    {
      "id": "MSG-1727181005",
      "session_id": "SES-UUID-42",
      "role": "assistant",
      "content": "Ditemukan anomali volume Z-Score sebesar 3.42...",
      "status": "COMPLETED",
      "created_at": "2026-09-24T12:00:05Z"
    }
  ]
}
```

---

### 3.6 Session Management

#### `GET /api/chat/sessions`
Mengambil daftar sesi obrolan dengan dukungan paginasi dan pencarian.

* **Query Parameters:**
  * `limit` (integer, default `50`)
  * `offset` (integer, default `0`)
  * `q` (string, opsional): Pencarian judul sesi.
* **Response Headers:** `200 OK`
* **Response Body:**
```json
{
  "total": 1,
  "sessions": [
    {
      "id": "SES-UUID-42",
      "title": "Investigasi ANTM & Komoditas Nikel",
      "model": "gemini-2.0-flash",
      "status": "IDLE",
      "message_count": 6,
      "last_message_preview": "Ditemukan anomali volume Z-Score...",
      "is_pinned": 1,
      "parent_session_id": null,
      "created_at": "2026-09-24T10:00:00Z",
      "updated_at": "2026-09-24T12:00:00Z"
    }
  ]
}
```

---

#### `POST /api/chat/sessions`
Membuat sesi percakapan kosong baru.

* **Request Body:**
```json
{
  "title": "Analisis Saham BBRI",
  "model": "gemini-2.0-flash"
}
```
* **Response Headers:** `201 Created`
* **Response Body:** Objek `ChatSession` baru.

---

#### `PATCH /api/chat/sessions/{id}`
Memperbarui atribut sesi (judul, status pin, atau status eksekusi).

* **Request Body:**
```json
{
  "title": "Judul Baru Sesi",
  "is_pinned": true
}
```
* **Response Headers:** `200 OK`

---

#### `DELETE /api/chat/sessions/{id}`
Menghapus sesi obrolan beserta seluruh relasi pesan `chat_messages` dan *scoped memory edges* (Cascade Delete).

* **Response Headers:** `200 OK`
* **Response Body:**
```json
{
  "status": "deleted",
  "session_id": "SES-UUID-42"
}
```

---

#### `POST /api/chat/sessions/{id}/fork`
Menduplikasi percakapan dari sesi sumber ke sesi cabang baru hingga ID pesan tertentu (*branching exploration pattern*).

* **Request Body:**
```json
{
  "title": "Eksplorasi Cabang ANTM",
  "up_to_message_id": "MSG-1727181005"
}
```
* **Response Headers:** `201 Created`
* **Response Body:** Objek `ChatSession` baru yang memiliki `parent_session_id`.

---

#### `GET /api/chat/search`
Mencari pesan tertentu berdasarkan kata kunci teks di seluruh riwayat obrolan.

* **Query Parameters:**
  * `q` (string, wajib): Kata kunci pencarian.
  * `limit` (integer, default `20`).
* **Response Headers:** `200 OK`

---

### 3.7 Investigation Pipeline (Headless Mode)

#### `GET /api/investigations`
Daftar seluruh riwayat investigasi terstruktur 7-Stage pipeline yang tersimpan di basis data SQLite.

* **Query Parameters:**
  * `limit` (integer, default `50`)
* **Response Headers:** `200 OK`

---

#### `POST /api/investigations`
Memicu eksekusi investigasi formal untuk ticker tertentu.

* **Request Body:**
```json
{
  "ticker": "BBCA",
  "timeframe_days": 30
}
```
* **Response Headers:** `202 Accepted`
* **Response Body:**
```json
{
  "session_id": "INV-2026-BBCA-01",
  "status": "RUNNING",
  "ticker": "BBCA"
}
```

---

#### `GET /api/investigations/{id}`
Mengambil dossier lengkap suatu sesi investigasi, mencakup:
* Metadata investigasi
* Daftar anomali statistik kuantitatif (`anomalies`)
* Temuan dan verifikasi 3-tier taxonomy (`findings`)
* Item bukti dan kutipan berita OSINT (`evidence_items`)
* Garis waktu kronologis (`timeline_events`)

* **Response Headers:** `200 OK`
* **Response Body:**
```json
{
  "investigation": {
    "id": "INV-2026-ANTM-01",
    "ticker": "ANTM",
    "status": "COMPLETED",
    "summary_text": "Anomali volume transaksi terkonfirmasi berkorelasi dengan kenaikan harga nikel LME..."
  },
  "anomalies": [
    {
      "metric_type": "VOLUME",
      "z_score": 3.42,
      "metric_value": 85000000,
      "baseline_value": 24000000,
      "anomaly_date": "2026-09-20"
    }
  ],
  "findings": [
    {
      "title": "Lonjakan Volume Didahului Kenaikan Komoditas Acuan",
      "verification_status": "SUPPORTED",
      "confidence_score": 0.95,
      "causality_status": "LIKELY_CATALYST"
    }
  ],
  "evidence_items": [
    {
      "source_name": "Kontan",
      "source_url": "https://investasi.kontan.co.id/news/...",
      "publication_date": "2026-09-19"
    }
  ],
  "timeline_events": [
    {
      "event_timestamp": "2026-09-19T10:00:00Z",
      "event_type": "NEWS",
      "headline": "Harga Nikel Melonjak 4.2% di Bursa London"
    }
  ]
}
```

---

## 4. Status Codes & Error Reference

| Kode HTTP | Makna | Kondisi Terjadi |
|---|---|---|
| `200 OK` | Berhasil | Permintaan GET, PATCH, atau DELETE berhasil diproses |
| `201 Created` | Berhasil Dibuat | Sesi baru atau fork sesi berhasil dibuat |
| `202 Accepted` | Diterima | Eksekusi pipeline panjang diterima di background |
| `400 Bad Request` | Kesalahan Input | Format JSON tidak valid, target provider tidak dikenal, atau parameter kurang |
| `404 Not Found` | Tidak Ditemukan | Sesi atau investigasi dengan ID bersangkutan tidak ditemukan |
| `405 Method Not Allowed` | Metode Ditolak | Mengirimkan `POST` ke endpoint yang hanya mendukung `GET` |
| `500 Internal Server Error` | Kesalahan Server | Kegagalan basis data SQLite lokal atau koneksi IPC engine |
