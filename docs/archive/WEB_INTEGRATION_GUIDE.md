# Panduan Integrasi Web Workspace Dashboard — Niskava Agent

Dokumen teknis ini ditujukan bagi pengembang frontend (`clients/web/`) yang membangun antarmuka Web Workspace Dashboard Niskava Agent. Panduan ini menjelaskan pola integrasi REST API, penanganan streaming Server-Sent Events (SSE), visualisasi graf interaktif, serta sinkronisasi konfigurasi sistem dengan daemon Go Core (`http://localhost:20128`).

---

## 1. Arsitektur Komunikasi & Persiapan Klien

Web Workspace beroperasi sebagai Single Page Application (SPA) berbasis React, TypeScript, dan Vite yang berkomunikasi secara lokal dengan Go Core Daemon melalui protokol HTTP dan Server-Sent Events (SSE).

### 1.1 Penentuan Base URL Dinamis
Jangan melakukan *hardcoding* pada alamat backend. Gunakan resolusi dinamis agar aplikasi dapat berjalan mulus pada lingkungan pengembangan Vite (`localhost:5173`) maupun production bundle tersemat (`//go:embed`):

```typescript
// src/services/apiClient.ts
export const getBaseUrl = (): string => {
  if (import.meta.env.VITE_API_BASE_URL) {
    return import.meta.env.VITE_API_BASE_URL;
  }
  // Jika frontend dilayani langsung oleh Go server pada port yang sama
  if (window.location.port === '20128') {
    return window.location.origin;
  }
  // Default fallback saat local development via Vite dev server
  return 'http://localhost:20128';
};
```

### 1.2 Konfigurasi HTTP Client Standar
Gunakan instance HTTP client terpadu dengan timeout dan penanganan error standar:

```typescript
// src/services/apiClient.ts
import axios from 'axios';

export const apiClient = axios.create({
  baseURL: getBaseUrl(),
  timeout: 15000,
  headers: {
    'Content-Type': 'application/json',
    'Accept': 'application/json',
  },
});

apiClient.interceptors.response.use(
  (response) => response,
  (error) => {
    const message = error.response?.data?.error || error.message || 'Terjadi kesalahan koneksi ke backend';
    return Promise.reject(new Error(message));
  }
);
```

---

## 2. Integrasi Streaming Percakapan & ReAct Loop (`POST /api/chat`)

Analisis intelijen pasar Niskava melibatkan siklus ReAct (*Reasoning + Action*) yang memanggil tools deterministik secara dinamis. Frontend **wajib** menggunakan protokol Server-Sent Events (SSE) agar pengguna dapat melihat proses penalaran (*thinking*) dan pemanggilan tools secara transparan.

### 2.1 State Management Siklus Obrolan
Terapkan tiga status input utama:
1. `IDLE`: Input aktif, pengguna dapat mengetik prompt baru.
2. `STREAMING`: Agen sedang bernalar dan memanggil tool deterministik. Input text terkunci (*disabled*), tombol kirim berubah menjadi tombol **Abort**.
3. `ABORTING`: Pengguna menekan tombol batal, mengirimkan sinyal pembatalan ke `/api/chat/abort`.

### 2.2 Implementasi SSE Stream Reader
Gunakan antarmuka `fetch` standar dengan `ReadableStream` decoder untuk menangkap event secara real-time:

```typescript
// src/services/chatStream.ts
import { getBaseUrl } from './apiClient';

export interface ChatStreamHandlers {
  onSessionStart?: (sessionId: string) => void;
  onProgress?: (stage: string, message: string) => void;
  onToolStart?: (toolName: string, inputParams: Record<string, unknown>) => void;
  onToolResult?: (toolName: string, result: unknown) => void;
  onChunk?: (chunk: string) => void;
  onDone?: (sessionId: string) => void;
  onError?: (error: string) => void;
}

export async function sendChatMessage(
  prompt: string,
  sessionId: string | undefined,
  handlers: ChatStreamHandlers,
  abortSignal?: AbortSignal
): Promise<void> {
  const url = `${getBaseUrl()}/api/chat`;
  
  const response = await fetch(url, {
    method: 'POST',
    headers: {
      'Content-Type': 'application/json',
      'Accept': 'text/event-stream',
    },
    body: JSON.stringify({ prompt, session_id: sessionId }),
    signal: abortSignal,
  });

  if (!response.ok) {
    const errText = await response.text();
    throw new Error(`Server returned status ${response.status}: ${errText}`);
  }

  const reader = response.body?.getReader();
  if (!reader) throw new Error('ReadableStream tidak didukung oleh browser');

  const decoder = new TextDecoder('utf-8');
  let buffer = '';

  while (true) {
    const { done, value } = await reader.read();
    if (done) break;

    buffer += decoder.decode(value, { stream: true });
    const lines = buffer.split('\n');
    buffer = lines.pop() || '';

    let currentEvent = 'message';
    for (const line of lines) {
      const trimmed = line.trim();
      if (!trimmed) continue;

      if (trimmed.startsWith('event:')) {
        currentEvent = trimmed.replace('event:', '').trim();
      } else if (trimmed.startsWith('data:')) {
        const rawData = trimmed.replace('data:', '').trim();
        try {
          const parsed = JSON.parse(rawData);
          handleEventPayload(currentEvent, parsed, handlers);
        } catch {
          // Payload raw string
          handleEventPayload(currentEvent, rawData, handlers);
        }
      }
    }
  }
}

function handleEventPayload(event: string, data: any, handlers: ChatStreamHandlers) {
  switch (event) {
    case 'session_start':
      handlers.onSessionStart?.(data.session_id);
      break;
    case 'progress_step':
      handlers.onProgress?.(data.stage, data.message);
      break;
    case 'tool_call_start':
      handlers.onToolStart?.(data.tool, data.input);
      break;
    case 'tool_call_result':
      handlers.onToolResult?.(data.tool, data.result);
      break;
    case 'agent_message_chunk':
      handlers.onChunk?.(data.chunk);
      break;
    case 'done':
      handlers.onDone?.(data.session_id);
      break;
    case 'session_error':
      handlers.onError?.(data.error);
      break;
  }
}
```

### 2.3 Mekanisme Pembatalan Turn (Abort Action)
Jika pengguna membatalkan investigasi yang sedang berjalan:
1. Panggil abort controller lokal: `abortController.abort()`.
2. Kirim sinyal ke daemon server:
   ```typescript
   await apiClient.post('/api/chat/abort', { session_id: currentSessionId });
   ```

---

## 3. Integrasi Visualisasi Memory Graph

Modul graf memetakan hubungan antara emiten saham, broker transaksi, regulator, katalis komoditas, dan anomali pasar.

### 3.1 Skema Warna Tipe Entitas (UI Design System)
Untuk konsistensi tema Market Intelligence / Bloomberg Terminal:

| Tipe Node (`node_type`) | Label Tipe | Warna Utama Hex | Latar Belakang Node |
|---|---|---|---|
| `TICKER` | Emiten Saham | `#22C55E` (Matrix Green) | `rgba(34, 197, 94, 0.15)` |
| `CATALYST` | Katalis / Komoditas | `#38BDF8` (Sky Blue) | `rgba(56, 189, 248, 0.15)` |
| `BROKER` | Broker Sekuritas | `#F59E0B` (Amber Gold) | `rgba(245, 158, 11, 0.15)` |
| `REGULATOR` | Bursa / OJK | `#A855F7` (Purple) | `rgba(168, 85, 247, 0.15)` |
| `ANOMALY` | Lonjakan Anomali | `#EF4444` (Crimson Red) | `rgba(239, 68, 68, 0.20)` |
| `EVENT` | Aksi Korporasi | `#EC4899` (Pink) | `rgba(236, 72, 153, 0.15)` |

### 3.2 Pemanggilan Data Graf Terfilter
Gunakan parameter URL untuk mengaktifkan mode *Ego-Network*:

```typescript
// src/services/graphService.ts
import { apiClient } from './apiClient';
import { MemoryGraphResponse, MemoryGraphStats } from '../types/api';

export async function fetchMemoryGraph(params?: {
  sessionId?: string;
  ticker?: string;
  depth?: number;
  minWeight?: number;
  nodeTypes?: string[];
}): Promise<MemoryGraphResponse> {
  const query = new URLSearchParams();
  if (params?.sessionId) query.set('session_id', params.sessionId);
  if (params?.ticker) query.set('ticker', params.ticker);
  if (params?.depth) query.set('depth', params.depth.toString());
  if (params?.minWeight) query.set('min_weight', params.minWeight.toString());
  if (params?.nodeTypes && params.nodeTypes.length > 0) {
    query.set('node_types', params.nodeTypes.join(','));
  }

  const resp = await apiClient.get<MemoryGraphResponse>(`/api/graph?${query.toString()}`);
  return resp.data;
}

export async function fetchGraphStats(): Promise<MemoryGraphStats> {
  const resp = await apiClient.get<MemoryGraphStats>('/api/graph/stats');
  return resp.data;
}
```

### 3.3 Interaksi Klik Node (Ego-Network Expansion)
Saat pengguna mengklik sebuah node emiten (contoh: node `ANTM`):
1. Perbarui state visualisasi dengan parameter `ticker=ANTM` dan `depth=1`.
2. Lakukan re-fetch ke `/api/graph?ticker=ANTM&depth=1`.
3. Graf canvas akan memfokuskan simpul terpilih beserta tetangga langsungnya ($k \le 2$ hops sesuai Law 6).

---

## 4. Integrasi Pengaturan Sistem & Kunci API

Halaman pengaturan mengelola kredensial provider LLM, kunci Sectors API v2, dan preferensi aplikasi.

### 4.1 Aturan Pengamanan Form Kunci Rahasia
Backend menyensor kunci yang tersimpan menjadi format `sec_****1234` dan menyediakan flag `has_sectors_key: true`.

* **Aturan Frontend:**
  * Render placeholder berupa bullet sandi atau string bertanda `(Tersimpan: sec_****1234)`.
  * Biarkan input kosong secara default kecuali jika pengguna ingin mengganti kunci dengan yang baru.
  * Jangan pernah mengirimkan string sensor (`****`) kembali ke backend. Jika pengguna tidak mengganti kunci, hapus field tersebut dari payload `PATCH` atau kirim string kosong.

```typescript
// src/services/settingsService.ts
import { apiClient } from './apiClient';
import { ConfigView, TestConnectionRequest, TestConnectionResponse } from '../types/api';

export async function getSystemSettings(): Promise<ConfigView> {
  const resp = await apiClient.get<ConfigView>('/api/settings');
  return resp.data;
}

export async function updateSystemSettings(partialSettings: Partial<ConfigView>): Promise<ConfigView> {
  const resp = await apiClient.patch<ConfigView>('/api/settings', partialSettings);
  return resp.data;
}

export async function testProviderConnection(req: TestConnectionRequest): Promise<TestConnectionResponse> {
  const resp = await apiClient.post<TestConnectionResponse>('/api/settings/test-connection', req);
  return resp.data;
}
```

### 4.2 Tombol Uji Koneksi Live (*Test Connection*)
Sediakan tombol "Uji Koneksi" di samping masing-masing input provider (Sectors, Gemini, OpenAI, Ollama). Tampilkan animasi loading dan hasil latensi jaringan:
* Sukses: Indikator hijau dengan teks `"Terhubung (142 ms)"`.
* Gagal: Indikator merah dengan rincian pesan kesalahan teknis dari backend.

---

## 5. Integrasi Manajemen Telegram Bot Daemon

Dashboard web menyediakan kontrol operasional bot Telegram tanpa perlu mengakses terminal CLI.

### 5.1 Siklus Operasional Bot
Sediakan komponen status badge dan kontrol tombol aksi:
* Status `RUNNING`: Badge hijau dengan username `@NiskavaAgentBot`. Tombol: **Matikan Bot** (`POST /api/telegram/stop`).
* Status `STOPPED`: Badge kuning bertuliskan *Standby*. Tombol: **Nyalakan Bot** (`POST /api/telegram/start`).
* Status `NOT_CONFIGURED`: Badge abu-abu. Tampilkan tautan cepat menuju form pengisian `bot_token`.

```typescript
// src/services/telegramService.ts
import { apiClient } from './apiClient';
import { TelegramStatusResponse } from '../types/api';

export async function getTelegramStatus(): Promise<TelegramStatusResponse> {
  const resp = await apiClient.get<TelegramStatusResponse>('/api/settings/telegram');
  return resp.data;
}

export async function startTelegramBot(): Promise<TelegramStatusResponse> {
  const resp = await apiClient.post<TelegramStatusResponse>('/api/telegram/start');
  return resp.data;
}

export async function stopTelegramBot(): Promise<TelegramStatusResponse> {
  const resp = await apiClient.post<TelegramStatusResponse>('/api/telegram/stop');
  return resp.data;
}

export async function sendTelegramTestPing(chatId: string | number): Promise<{ status: string; message: string }> {
  const resp = await apiClient.post<{ status: string; message: string }>('/api/telegram/test', {
    chat_id: chatId,
  });
  return resp.data;
}
```

---

## 6. Integrasi Monitor Disiplin Kredit & Cache Sectors (Law 5)

Untuk memenuhi kriteria kompetisi Sectors Hackathon Track 1 mengenai efisiensi kuota 1.000 kredit:

### 6.1 Widget Metrik Cache
Panggil `GET /api/system/sectors-usage` untuk menampilkan:
* **Total Permintaan Terarsip:** `total_entries`.
* **Kredit Dihemat:** `estimated_credit_saved` (setiap hit pada cache SQLite lokal menghemat 1 kredit Sectors API).
* **Entri Kadaluarsa:** `expired_entries`.
* **Tombol Bersihkan Cache Kadaluarsa:** Pemicu `POST /api/system/cache/clean` untuk menjaga ukuran database tetap ringkas.

---

## 7. Referensi Lengkap Tipe Data TypeScript (`src/types/api.ts`)

Salin definisi antarmuka TypeScript berikut ke dalam proyek frontend Anda untuk memastikan *type safety* secara menyeluruh:

```typescript
// src/types/api.ts

export type VerificationStatus = 'SUPPORTED' | 'UNCERTAIN' | 'CONTRADICTED';
export type CausalityStatus = 'LIKELY_CATALYST' | 'PRECEDED_ANNOUNCEMENT' | 'UNEXPLAINED_BY_NEWS';
export type NodeType = 'TICKER' | 'CATALYST' | 'BROKER' | 'REGULATOR' | 'ANOMALY' | 'EVENT';

// --- Memory Graph ---
export interface MemoryNode {
  id: string;
  label: string;
  node_type: NodeType;
  metadata_json?: string | null;
  last_observed_at: string;
  created_at: string;
}

export interface MemoryEdge {
  source_id: string;
  target_id: string;
  relation: string;
  context_snippet?: string | null;
  session_id?: string | null;
  weight: number;
  confidence_score: number;
  last_observed_at: string;
  created_at: string;
}

export interface MemoryGraphResponse {
  session_id?: string;
  total_nodes: number;
  total_edges: number;
  nodes: MemoryNode[];
  edges: MemoryEdge[];
}

export interface HubNode {
  id: string;
  label: string;
  node_type: NodeType;
  degree: number;
}

export interface MemoryGraphStats {
  total_nodes: number;
  total_edges: number;
  node_types: Record<string, number>;
  top_hub_nodes: HubNode[];
}

// --- Settings & Diagnostics ---
export interface AuthView {
  ai_provider: string;
  sectors_api_key: string;
  has_sectors_key: boolean;
  sectors_base_url: string;
  gemini_api_key: string;
  has_gemini_key: boolean;
  gemini_model: string;
  openai_api_key: string;
  has_openai_key: boolean;
  openai_base_url: string;
  openai_model: string;
  anthropic_api_key: string;
  has_anthropic_key: boolean;
  ollama_base_url: string;
  ollama_model: string;
}

export interface ConfigView {
  auth: AuthView;
  storage: { db_path: string };
  engine: { python_bin: string; entrypoint: string };
  server: { port: number };
  preferences: {
    default_market: string;
    offline_mode: boolean;
    language: string;
    llm_timeout_secs?: number; // Batas waktu inferensi LLM dalam detik (10 - 300)
  };
  memory: {
    enabled: boolean;
    decay_lambda: number;
    ego_radius: number;
    max_context_tokens: number;
  };
  telegram: {
    bot_token: string;
    has_token: boolean;
    enabled: boolean;
    allowed_users: string[];
  };
}

export interface TestConnectionRequest {
  target: 'sectors' | 'gemini' | 'openai' | 'ollama' | 'anthropic';
  api_key?: string;
  base_url?: string;
  model?: string;
}

export interface TestConnectionResponse {
  target: string;
  success: boolean;
  message: string;
  latency_ms: number;
}

// --- Telegram Bot ---
export interface TelegramStatusResponse {
  status: 'RUNNING' | 'STOPPED' | 'NOT_CONFIGURED';
  bot_username: string;
  enabled: boolean;
  allowed_users: string[];
  has_token: boolean;
}

// --- System & Cache ---
export interface SectorsUsageStats {
  total_entries: number;
  expired_entries: number;
  permanent_entries: number;
  estimated_credit_saved: number;
  status: string;
}

export interface SystemDiagnostics {
  status: string;
  app: string;
  go_version: string;
  os: string;
  arch: string;
  num_cpu: number;
  database_path: string;
  database_size_bytes: number;
  total_sessions: number;
  timestamp: string;
}

// --- Sessions & Chat ---
export interface ChatMessage {
  id: string;
  session_id: string;
  role: 'user' | 'assistant' | 'system';
  content: string;
  status: 'COMPLETED' | 'STREAMING' | 'FAILED';
  created_at: string;
}

export interface ChatSession {
  id: string;
  title: string;
  model: string;
  status: 'IDLE' | 'BUSY';
  message_count: number;
  last_message_preview?: string;
  is_pinned: number;
  parent_session_id?: string | null;
  created_at: string;
  updated_at: string;
}
```

---

## 8. Kepatuhan Regulasi & Disclaimer Wajib

Sesuai aturan kompetisi **Law 2 (Strict Financial Non-Advisory Boundary)** dan **Law 3 (Zero Broker Execution)**:

1. **Dilarang Keras:** Menampilkan label, badge, atau tombol bertuliskan saran transaksi langsung seperti **BUY**, **SELL**, atau **TARGET PRICE**.
2. **Wajib Menampilkan Tag Bukti:** Setiap kartu temuan investigasi harus menyertakan badge verifikasi:
   * `[SUPPORTED]` (Warna Hijau `#22C55E`): Terbukti langsung oleh data resmi Sectors atau IDXnet.
   * `[UNCERTAIN]` (Warna Kuning `#FACC15`): Korelasi teramati namun kausalitas belum terkonfirmasi resmi.
   * `[CONTRADICTED]` (Warna Merah `#EF4444`): Dibantah oleh keterbukaan informasi emiten.
3. **Disclaimer Statis di Footer UI:**
   Cantumkan teks disclaimer resmi pada footer antarmuka:
   > *"Niskava Agent adalah platform riset intelijen pasar modal berbasis bukti untuk Bursa Efek Indonesia (IDX), BUKAN penasihat investasi berizin. Seluruh data dan sintesis disajikan untuk verifikasi informasi pasar dan BUKAN rekomendasi finansial."*
