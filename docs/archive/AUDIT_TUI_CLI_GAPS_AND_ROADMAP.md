# Audit Komprehensif Kesenjangan Fitur & Roadmap Implementasi CLI / TUI — Niskava Agent

> **Single Source of Truth (SSoT) Kesiapan Antarmuka Terminal & Command-Line Interface**  
> **Target:** Pengembang CLI/TUI (Agung dkk.), Core Backend Engineer, & Evaluasi Hackathon Indonesia 2026.  
> **Status Basis Data & Engine Backend:** 100% Siap (272/272 Pytest Green, 100% Go Test Suite Passing).

---

> [!CAUTION]
> ## 🚨 ARAHAN MUTLAK TIM: PEMBEKUAN BACKEND (BE CODE FREEZE)
> 1. **DILARANG KERAS MENYENTUH KODE BACKEND:**  
>    Seluruh direktori backend (`backend/core/`, `backend/engine/`, `backend/core/db/`, `backend/core/server/`, `backend/core/ipc/`) berstatus **STABIL, TERUJI, & LENGKAP**. Tim CLI/TUI dilarang melakukan modifikasi, penambahan logic, atau refactoring pada berkas backend Go maupun Python Engine secara mandiri/sepihak.
> 2. **JALUR KOMUNIKASI & KOORDINASI WAJIB VIA IKHSAN:**  
>    Jika terdapat kendala kontrak data, kebutuhan penyesuaian event IPC/SSE, atau bug pada layer backend, **ANGGOTA TIM WAJIB MENGABARI DAN MENGHUBUNGI IKHSAN** terlebih dahulu sebelum melakukan tindakan apa pun. Penyesuaian backend hanya boleh dilakukan melalui koordinasi langsung bersama Ikhsan untuk mencegah regresi arsitektur, kerusakan 272 suite uji coba otomatis, atau pelanggaran terhadap 6 Hukum Niskava & Aturan Resmi Hackathon.

---

## 1. Executive Summary & Audit Scope

Dokumen ini merupakan hasil audit teknis mendalam terhadap implementasi antarmuka terminal (*Command Line Interface* / *Terminal User Interface*) pada repositori **Niskava Agent** (`clients/cli/` dan `clients/cli/tui/`). Audit ini membandingkan secara komprehensif kapabilitas yang telah disediakan dan diekspos oleh **Backend (BE)**—baik pada Go Core Daemon (`backend/core/`), lapisan basis data SQLite WAL (`backend/core/db/`), subprocess IPC (`backend/core/ipc/`), maupun Python Agent Engine (`backend/engine/`)—dengan fitur yang saat ini benar-benar telah terintegrasi di terminal.

### Ringkasan Temuan Audit:
1. **Ketidaksinkronan Subcommand Publik:** Terdapat referensi dalam dokumentasi resmi (`docs/CLI_INTEGRATION_GUIDE.md`) mengenai perintah `niskava terminal`, namun perintah tersebut belum didaftarkan sebagai subcommand Cobra di `clients/cli/root.go`, sehingga eksekusi langsung via shell menghasilkan error `unknown command`.
2. **Dead Flag pada Pipeline Investigasi:** Flag `-i` / `--interactive` pada `niskava investigate [TICKER]` telah dideklarasikan di level Cobra flag, namun tidak dievaluasi sama sekali pada fungsi eksekutor `RunE` (`clients/cli/investigate.go`), menyebabkan alur selalu jatuh ke headless pipeline tanpa opsi transisi ke REPL interaktif.
3. **Ketiadaan Transisi Interaktif Pasca-Headless Audit:** Setelah `niskava investigate [TICKER]` selesai mengeksekusi pipeline 7-tahap, program langsung keluar begitu saja ke prompt shell tanpa memberikan prompt konfirmasi (*call to action*) seperti: `[Enter] Lanjutkan diskusi interaktif untuk emiten ini?`.
4. **Kelemahan Penanganan Abort ReAct saat Daemon Mode:** Ketika REPL berjalan terhubung ke Go Server Daemon (`StreamChatViaSSE`), penekanan `Ctrl + C` oleh pengguna hanya membatalkan *context* HTTP lokal di CLI, tetapi **tidak mengirimkan sinyal pembatalan ke endpoint backend** (`POST /api/chat/sessions/{id}/abort`). Akibatnya, server tetap mengeksekusi LLM loop di background, membuang kuota kredit AI, dan membiarkan status sesi terkunci pada `RUNNING` (memicu `HTTP 409 Conflict` jika pengguna segera mengirim prompt baru).
5. **Kekosongan Metadata Sesi pada Standalone Subprocess:** Pada mode mandiri tanpa daemon, `repl.go` hanya menyimpan pesan langsung ke tabel `chat_messages` tanpa memanggil `CreateChatSession` atau `TouchChatSession` pada tabel `chat_sessions`. Hal ini mengakibatkan sesi yang dibuat via REPL mandiri memiliki judul kosong atau pratinjau pesan kosong saat dibuka kembali.
6. **Disinkronisasi Diagnostik `/health` vs `niskava doctor`:** Di dalam REPL, perintah `/health` hanya menampilkan ping latensi ringkas, dan tidak mengintegrasikan kapabilitas diagnostik komprehensif yang telah dibangun pada `clients/cli/doctor.go` (pemeriksaan izin WAL SQLite, dependensi Python, versi runtime, dan status kuota Sectors).
7. **Fitur Lanjutan Backend yang Belum Tersedia di Terminal:** Fitur-fitur matang seperti percabangan sesi (*session forking* / OpenCode pattern), ekspor laporan ke Markdown/JSON, pencarian global riwayat chat (*full-text search*), pengelolaan whitelist Telegram bot, serta inspeksi/pembersihan cache Sectors v2 belum memiliki antarmuka perintah CLI ataupun *slash command* di REPL.

---

## 2. Matrix Kesenjangan Fitur: Backend vs CLI / TUI

Berikut adalah matriks perbandingan antara kapabilitas yang telah dibangun di Backend dengan status implementasinya pada antarmuka CLI/TUI:

| Modul / Domain Backend | Kapabilitas yang Disediakan BE | Status di CLI / TUI | Tingkat Urgensi | Dampak Teknis / UX |
|---|---|---|---|---|
| **Cobra CLI Routing** | Subcommand eksplisit `niskava terminal` (alias `chat`, `repl`) | ❌ **Belum Ada** | **P0 (Kritis)** | Inkonsistensi dokumentasi panduan integrasi; pengguna shell gagal membuka REPL secara langsung tanpa launcher. |
| **Pipeline CLI (`investigate`)** | Flag `--interactive / -i` untuk transisi ke REPL | ⚠️ **Dead Code** | **P1 (Tinggi)** | Flag terdaftar di Cobra namun diabaikan di `RunE`; pengguna tidak bisa masuk ke mode interaktif emiten. |
| **Pasca-Pipeline UX** | Transisi ke REPL setelah audit headless selesai | ❌ **Belum Ada** | **P1 (Tinggi)** | Pengguna harus mengetik ulang perintah terpisah untuk menanyakan temuan audit emiten yang baru selesai diproses. |
| **SSE ReAct Stream Abort** | Endpoint `POST /api/chat/sessions/{id}/abort` | ❌ **Belum Terhubung** | **P0 (Kritis)** | `Ctrl+C` saat daemon aktif meninggalkan proses Python berjalan di background & sesi terkunci `RUNNING` (409 Conflict). |
| **Session Metadata Lifecycle** | `db.CreateChatSession`, `db.TouchChatSession` | ⚠️ **Parsial** | **P1 (Tinggi)** | Mode standalone REPL tidak menginisialisasi baris di `chat_sessions`; judul sesi dan preview pesan menjadi `-`. |
| **OpenCode Session Forking** | `db.ForkChatSession`, `POST /api/chat/sessions/{id}/fork` | ❌ **Belum Ada** | **P2 (Sedang)** | Pengguna terminal tidak dapat mencabangkan alur analisa hipotesis tanpa merusak riwayat utama. |
| **Report Exporting** | Format generator Markdown & JSON di core server | ❌ **Belum Ada** | **P1 (Tinggi)** | Analis tidak dapat mengekspor transkrip investigasi terminal ke berkas `.md` atau `.json` lokal. |
| **Global Message Search** | `db.SearchChatMessages`, `GET /api/chat/search` | ❌ **Belum Ada** | **P2 (Sedang)** | Tidak ada cara mencari kembali temuan masa lalu berbasis kata kunci dari terminal. |
| **Session Deletion & Pruning** | `db.DeleteChatSession`, `db.ClearSessionHistory` | ❌ **Belum Ada** | **P2 (Sedang)** | Pengguna tidak dapat membersihkan atau menghapus sesi lama melalui perintah `niskava sessions`. |
| **Sectors Cache Management** | `db.GetSectorsCacheStats`, `db.CleanExpiredCache` | ❌ **Belum Ada** | **P2 (Sedang)** | Disiplin kuota kredit 1.000 panggilan Sectors v2 tidak dapat dipantau dari REPL atau CLI tanpa REST client eksternal. |
| **Telegram Bot Management** | Start, Stop, Test, Whitelist (`allowed_users`) | ⚠️ **Parsial** | **P2 (Sedang)** | CLI hanya menjalankan bot foreground; tidak ada perintah `niskava telegram test` atau `whitelist`. |
| **REPL Slash Commands** | `/fork`, `/export`, `/search`, `/anomalies`, `/skills`, `/doctor` | ❌ **Belum Ada** | **P1 (Tinggi)** | Popup menu slash di TUI belum mencakup fitur investigasi tingkat lanjut yang didukung backend. |
| **REPL Help Synchronization** | Panduan perintah `/timeout` | ⚠️ **Parsial** | **P3 (Rendah)** | Perintah `/timeout` sudah berfungsi di kode, namun tidak tertera pada fungsi `printHelp()`. |

---

## 3. Analisis Teknis Detail Kesenjangan (Root Cause & Code Audit)

### 3.1 Kesenjangan 1: Ketidakhadiran Subcommand `niskava terminal`
* **Lokasi Berkas:** `clients/cli/root.go`, `docs/CLI_INTEGRATION_GUIDE.md`
* **Analisis Kode:**  
  Di `docs/CLI_INTEGRATION_GUIDE.md` baris 29 tertulis:
  ```text
  2. Live Conversational REPL (niskava terminal)
  ```
  Namun pada `clients/cli/root.go`, hanya terdapat perintah dasar `niskava` (yang membuka launcher menu interaktif) dan flag `-s / --session`. Tidak ada `terminalCmd` yang didaftarkan ke `RootCmd`.
* **Dampak:** Pengguna atau skrip automasi yang memanggil `niskava terminal` akan langsung menerima pesan error:
  ```text
  Error: unknown command "terminal" for "niskava"
  ```
* **Solusi Arsitektural:**  
  Tambahkan `terminalCmd` di `clients/cli/terminal.go` yang langsung memanggil `tui.RunLiveREPL(cfg, appDB, srv.URL, sessionFlag)`.

---

### 3.2 Kesenjangan 2: Dead Flag `--interactive / -i` pada `niskava investigate`
* **Lokasi Berkas:** `clients/cli/investigate.go` (Baris 19, 99)
* **Analisis Kode:**  
  Pada fungsi `init()`, flag dideklarasikan:
  ```go
  investigateCmd.Flags().BoolVarP(&interactiveFlag, "interactive", "i", false, "run in interactive conversational investigation mode")
  ```
  Tetapi pada `RunE` (baris 29–93), variabel `interactiveFlag` **sama sekali tidak dievaluasi**. Alur selalu langsung menginisialisasi sesi pipeline headless, memanggil `ipc.RunSubprocess()`, dan merender Bubbletea model headless (`tui.NewModel`).
* **Dampak:** Pengguna yang mengeksekusi `niskava investigate ANTM -i` dengan ekspektasi membuka terminal riset interaktif langsung terfokus pada emiten `ANTM` tetap dipaksa melihat visualisasi headless 7-tahap tanpa interaksi.
* **Solusi Arsitektural:**  
  Pada awal fungsi `RunE` di `investigate.go`, tambahkan percabangan:
  ```go
  if interactiveFlag {
      srv, _ := server.Start(ctx, cfg.Server.Port, appDB, cfg)
      initialPrompt := fmt.Sprintf("Lakukan investigasi anomali volume dan verifikasi bukti untuk saham %s", ticker)
      return tui.RunLiveREPLWithInitialPrompt(cfg, appDB, srv.URL, sessionID, initialPrompt)
  }
  ```

---

### 3.3 Kesenjangan 3: Peniadaan Sinyal Abort HTTP pada Mode REPL Daemon
* **Lokasi Berkas:** `clients/cli/tui/repl.go` (Baris 680–703) dan `clients/cli/tui/sse_client.go`
* **Analisis Kode:**  
  Ketika daemon aktif (`usingDaemon := IsDaemonAlive(serverURL)` bernilai `true`), alur eksekusi memanggil:
  ```go
  eventsChan, errChan = StreamChatViaSSE(ctx, serverURL, sessionID, prompt)
  ```
  Ketika sinyal interupsi `Ctrl + C` tertangkap, goroutine interupsi memanggil `cancel()`. Hal ini menutup koneksi HTTP socket di sisi klien Go. Namun, `server.go` menangani request streaming chat melalui goroutine internal yang **memerlukan sinyal pembatalan eksplisit** melalui endpoint:
  ```http
  POST /api/chat/sessions/{id}/abort
  ```
  CLI tidak pernah menembakkan request ini saat pengguna menekan `Ctrl + C`.
* **Dampak:** Status sesi di database SQLite tetap berstatus `RUNNING`. Jika pengguna mengetik prompt berikutnya 1 detik kemudian, Go Core Server akan mengembalikan `HTTP 409 Conflict` karena sesi dianggap masih sibuk memproses inferensi sebelumnya.
* **Solusi Arsitektural:**  
  Dalam blok pembatalan di `repl.go`, jika `usingDaemon` aktif, jalankan panggilan non-blocking:
  ```go
  if usingDaemon {
      go func(sURL, sID string) {
          req, _ := http.NewRequest("POST", fmt.Sprintf("%s/api/chat/sessions/%s/abort", sURL, sID), nil)
          client := &http.Client{Timeout: 1500 * time.Millisecond}
          _, _ = client.Do(req)
      }(serverURL, sessionID)
  }
  ```

---

### 3.4 Kesenjangan 4: Sinkronisasi Lifecycle Metadata Sesi Chat pada Subprocess Mode
* **Lokasi Berkas:** `clients/cli/tui/repl.go` (Baris 662–672, 790–801)
* **Analisis Kode:**  
  Saat berjalan mandiri (*standalone mode* tanpa daemon), `repl.go` merekam `ChatMessage` ke SQLite:
  ```go
  if !usingDaemon && appDB != nil {
      userMsg := &db.ChatMessage{...}
      _ = appDB.SaveChatMessage(userMsg)
  }
  ```
  Namun, tabel `chat_sessions` tidak pernah diinisialisasi atau diperbarui:
  1. `appDB.CreateChatSession(&db.ChatSession{...})` tidak dipanggil di awal sesi.
  2. `appDB.TouchChatSession(sessionID, preview)` tidak dipanggil setelah asisten selesai merespons.
* **Dampak:** Ketika pengguna menjalankan `niskava sessions -t chat`, sesi-sesi yang baru dibuat melalui REPL mandiri tidak memiliki baris di tabel `chat_sessions`, atau judul dan pratinjaunya bernilai kosong.
* **Solusi Arsitektural:**  
  Pastikan `appDB.CreateChatSession` dipanggil jika sesi belum ada, dan perbarui pratinjau pesan terakhir dengan memanggil `appDB.TouchChatSession(sessionID, previewText)` setelah response selesai disimpan.

---

### 3.5 Kesenjangan 5: Fitur Manajemen Sesi yang Kurang Lengkap pada `niskava sessions`
* **Lokasi Berkas:** `clients/cli/sessions.go`
* **Analisis Kode:**  
  Saat ini `sessionsCmd` hanya mendukung pencetakan daftar tabel (`printFormattedSessions`) dengan filter tipe `chat` atau `investigation`.
  Sementara itu, basis data Go (`backend/core/db/db.go`) telah memiliki fungsi:
  - `DeleteChatSession(id string) error`
  - `ClearSessionHistory(sessionID string) error`
  - `ForkChatSession(sourceID, newID, newTitle, upToMessageID string) error`
  - `SearchChatMessages(query string, limit int) ([]ChatSearchResult, error)`
* **Dampak:** Pengguna terminal yang ingin menghapus riwayat investigasi lama, mencabangkan alur analisis, atau mencari pembahasan emiten tertentu harus memanipulasi database secara manual via biner `sqlite3`.
* **Solusi Arsitektural:**  
  Tambahkan subcommand di bawah `sessionsCmd`:
  - `niskava sessions list` (default)
  - `niskava sessions delete <SESSION_ID>` (menghapus sesi beserta riwayat percakapannya)
  - `niskava sessions search <KEYWORD>` (mencari pesan di seluruh sesi dengan cuplikan konteks)
  - `niskava sessions export <SESSION_ID> [--format md|json] [--out file]` (menghasilkan berkas laporan audit)

---

### 3.6 Kesenjangan 6: Penambahan Slash Commands Produktivitas pada Interactive REPL
* **Lokasi Berkas:** `clients/cli/tui/repl.go` (Baris 440–615), `clients/cli/tui/i18n.go`
* **Analisis Kode:**  
  REPL saat ini mengenali 13 slash commands: `/help`, `/chats`, `/resume`, `/timeout`, `/reset`, `/graph`, `/clear`, `/web`, `/sessions`, `/health`, `/lang`, `/back`, `/exit`.
  Namun terdapat kapabilitas backend penting yang belum dapat diakses melalui REPL:
  1. **/export [md|json]**: Mengekspor transkrip sesi aktif ke berkas laporan markdown berstandar institutional research.
  2. **/fork [title]**: Menduplikasi sesi saat ini menjadi sesi cabang baru (OpenCode pattern) untuk menguji skenario pasar alternatif tanpa mengubah sesi lama.
  3. **/search <keyword>**: Mencari teks riwayat percakapan masa lalu langsung dari REPL prompt.
  4. **/anomalies**: Menampilkan tabel anomali kuantitatif yang terdeteksi selama sesi aktif berlangsung.
  5. **/skills**: Menampilkan daftar 6 domain SOP terdaftar (`market_anomaly_recon`, `event_causality_audit`, `insider_bandarmology_forensic`, `financial_health_stress_test`, `mining_commodity_divergence`, `peer_valuation_benchmark`).
  6. **/cache [stats|clean]**: Menampilkan efisiensi kuota Sectors v2 API dan membersihkan entri cache yang telah kadaluarsa.
  7. **/doctor**: Menjalankan evaluasi diagnostik sistem lengkap (sama seperti `niskava doctor`) langsung di dalam REPL.
* **Solusi Arsitektural:**  
  Daftarkan perintah-perintah tersebut pada `ReplSlashCommands` di `clients/cli/tui/repl.go` dan tambahkan handler percabangannya di fungsi `RunLiveREPL`.

---

### 3.7 Kesenjangan 7: Kelengkapan Operasional Telegram Bot pada CLI
* **Lokasi Berkas:** `clients/cli/telegram.go`
* **Analisis Kode:**  
  `telegramCmd` hanya memanggil `botSvc.Start()` secara blocking di latar depan. Padahal, daemon backend (`backend/core/server/server.go`) menyediakan endpoint:
  - `GET /api/settings/telegram` (Status poller, username bot, dan whitelist)
  - `POST /api/telegram/test` (Mengirim pesan tes verifikasi ke Chat ID)
  - Pengelolaan `allowed_users` di konfigurasi
* **Solusi Arsitektural:**  
  Sediakan subcommand atau flags:
  - `niskava telegram test --chat-id <ID>`: Menguji konektivitas bot dengan mengirim pesan ping.
  - `niskava telegram status`: Mengecek apakah daemon bot sedang berjalan di background server.
  - `niskava telegram whitelist add/remove <USER>`: Mengelola daftar analis yang diizinkan berinteraksi dengan bot.

---

### 3.8 Kesenjangan 8: Sinkronisasi Bantuan `/timeout` pada `printHelp()`
* **Lokasi Berkas:** `clients/cli/tui/repl.go` (Baris 923–937)
* **Analisis Kode:**  
  Fungsi `printHelp()` di `repl.go`:
  ```go
  func printHelp() {
      fmt.Println(T("help_header"))
      fmt.Println(T("help_prompt_desc"))
      fmt.Println(T("help_ticker_desc"))
      fmt.Println(T("help_graph_desc"))
      fmt.Println(T("help_reset_desc"))
      fmt.Println("  • /chats        : " + T("slash_chats_desc"))
      fmt.Println("  • /resume <id>  : " + T("slash_resume_desc"))
      fmt.Println(T("help_sessions_desc"))
      fmt.Println(T("help_web_desc"))
      fmt.Println(T("help_health_desc"))
      fmt.Println(T("help_lang_desc"))
      fmt.Println(T("help_clear_desc"))
      fmt.Println(T("help_exit_desc"))
  }
  ```
  Baris untuk `/timeout` tertinggal, padahal implementasi `/timeout` telah selesai di `repl.go` baris 516–549.
* **Solusi:** Tambahkan baris bantuan untuk `/timeout` (`fast`, `balanced`, `deep`, `local`, atau custom detik) pada `printHelp()`.

---

## 4. Blueprint Spesifikasi Teknis Implementasi (Hanya di Layer Klien)

> [!IMPORTANT]
> Seluruh blueprint berikut **HANYA DITULIS DI LAPISAN `clients/cli/`**. Jangan menyentuh direktori `backend/`!

### 4.1 Registrasi Subcommand `niskava terminal`
```go
// clients/cli/terminal.go
package cli

import (
	"context"
	"fmt"

	"github.com/Sectors-Hacthon-2026/Niskava-Agents/backend/core/server"
	"github.com/Sectors-Hacthon-2026/Niskava-Agents/clients/cli/tui"
	"github.com/spf13/cobra"
)

var terminalCmd = &cobra.Command{
	Use:     "terminal",
	Aliases: []string{"repl", "chat"},
	Short:   "Launch interactive conversational intelligence REPL directly",
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		srv, err := server.Start(ctx, cfg.Server.Port, appDB, cfg)
		if err != nil {
			return fmt.Errorf("failed to start background daemon: %w", err)
		}
		srv.ConfigPath = cfgFile

		_ = tui.RunLiveREPL(cfg, appDB, srv.URL, sessionFlag)
		return nil
	},
}

func init() {
	RootCmd.AddCommand(terminalCmd)
}
```

### 4.2 Integrasi Abort Request pada `repl.go`
```go
// Di dalam goroutine penanganan sinyal SIGINT di executeChatTurn:
case <-sigChan:
    now := time.Now()
    if !lastSignalTime.IsZero() && now.Sub(lastSignalTime) <= 2*time.Second {
        interrupted.Store(true)
        fmt.Println("\n" + lipgloss.NewStyle().Foreground(ColorDanger).Bold(true).Render(strings.TrimSpace(T("repl_execution_cancelled"))))
        
        // Kirim HTTP POST abort jika terhubung ke daemon
        if usingDaemon {
            go func(baseURL, sessID string) {
                abortURL := fmt.Sprintf("%s/api/chat/sessions/%s/abort", baseURL, sessID)
                req, _ := http.NewRequest("POST", abortURL, nil)
                req.Header.Set("Content-Type", "application/json")
                client := &http.Client{Timeout: 1500 * time.Millisecond}
                _, _ = client.Do(req)
            }(serverURL, sessionID)
        }
        
        cancel()
        return
    }
```

### 4.3 Implementasi Slash Command `/export` pada REPL
```go
if strings.HasPrefix(lower, "/export") {
    parts := strings.Fields(input)
    format := "md"
    if len(parts) > 1 && strings.ToLower(parts[1]) == "json" {
        format = "json"
    }
    
    filename := fmt.Sprintf("niskava_report_%s.%s", sessionID, format)
    errExport := exportSessionToFile(appDB, sessionID, format, filename)
    if errExport != nil {
        fmt.Println(lipgloss.NewStyle().Foreground(ColorDanger).Render(fmt.Sprintf("Gagal mengekspor laporan: %v", errExport)))
    } else {
        fmt.Println(lipgloss.NewStyle().Foreground(ColorSuccess).Bold(true).Render(fmt.Sprintf("✓ Laporan berhasil diekspor ke: %s", filename)))
    }
    continue
}
```

---

## 5. Actionable Implementation Roadmap (Prioritas Eksekusi)

```
┌─────────────────────────────────────────────────────────────┐
│                 FASE 1: INTEGRITY & STABILITY               │
│  - Registrasi subcommand `niskava terminal` & alias `chat`  │
│  - Perbaikan flag `-i / --interactive` di `investigate.go`  │
│  - Call-to-action interaktif pasca headless pipeline        │
│  - Penanganan sinyal Abort HTTP `/abort` saat `Ctrl + C`    │
│  - Sinkronisasi metadata `chat_sessions` di mode standalone │
└──────────────────────────────┬──────────────────────────────┘
                               │
                               ▼
┌─────────────────────────────────────────────────────────────┐
│                FASE 2: PRODUCTIVITY & RESEARCH              │
│  - Penambahan slash commands REPL: `/export`, `/fork`       │
│  - Penambahan slash commands: `/search`, `/anomalies`       │
│  - Penambahan slash command `/doctor` (integrasi diagnosa)  │
│  - Sinkronisasi `/timeout` di `printHelp()`                 │
│  - Tambah perintah `niskava sessions export & delete`       │
└──────────────────────────────┬──────────────────────────────┘
                               │
                               ▼
┌─────────────────────────────────────────────────────────────┐
│             FASE 3: SYSTEM HYGIENE & TELEGRAM               │
│  - Penambahan slash command `/cache [stats|clean]` di REPL  │
│  - Penambahan slash command `/skills` (katalog SOP)         │
│  - Perluasan subcommand `niskava telegram test --chat-id`   │
│  - Verifikasi akhir end-to-end TUI suite                    │
└─────────────────────────────────────────────────────────────┘
```

### Quality Gate Sebelum Merge:
1. `go test -v -race ./clients/cli/...` lulus tanpa error persaingan thread (*data race*).
2. `go build -o bin/niskava ./cmd/niskava` menghasilkan biner bersih tanpa CGO dependency.
3. Menjalankan skenario uji pembatalan: Buka REPL $\to$ Kirim prompt investigasi $\to$ Tekan `Ctrl + C` 2x $\to$ Pastikan tidak terjadi hanging dan prompt baru dapat langsung dikirim tanpa `HTTP 409 Conflict`.
4. Menjalankan skenario ekspor: Ketik `/export md` $\to$ Berkas Markdown terbentuk di direktori kerja aktif dengan format terstruktur rapi.
