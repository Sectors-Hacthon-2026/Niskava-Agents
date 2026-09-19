package tui

import (
	"fmt"

	"github.com/charmbracelet/lipgloss"
)

// PrintFullHelpGuide renders an interactive instruction manual for all surfaces, options, and commands.
func PrintFullHelpGuide() {
	headerStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("#00FF87")).
		Background(lipgloss.Color("#052E16")).
		Padding(0, 1)

	sectionStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("#4ADE80"))

	keyStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("#FACC15"))

	cmdStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("#38BDF8"))

	descStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#F8FAFC"))

	mutedStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#94A3B8"))

	fmt.Println()
	fmt.Println(RenderConstellationLine(85))
	fmt.Printf(" %s\n", headerStyle.Render("NISKAVA AGENT — PANDUAN LENGKAP PENGGUNAAN & INSTRUKSI SISTEM"))
	fmt.Println(RenderConstellationLine(85))

	fmt.Println()
	fmt.Println(sectionStyle.Render("1. PANDUAN NAVIGASI MENU UTAMA (LAUNCHER):"))
	fmt.Printf("   • %s : Berpindah kursor ke atas/bawah antar opsi menu.\n", keyStyle.Render("[Panah ↑/↓] atau [k/j]"))
	fmt.Printf("   • %s          : Eksekusi opsi menu terpilih.\n", keyStyle.Render("[Enter]"))
	fmt.Printf("   • %s        : Tekan tombol shortcut berikut untuk eksekusi cepat:\n", keyStyle.Render("Shortcut Direct Keys"))
	fmt.Printf("     - %s : Buka Web Workspace di browser default (REST & SSE Stream Visual)\n", keyStyle.Render("[W] / [1]"))
	fmt.Printf("     - %s : Masuk ke Live Terminal UI (REPL Interaktif & Anomali)\n", keyStyle.Render("[T] / [2]"))
	fmt.Printf("     - %s : Lihat Riwayat Sesi Investigasi & Audit Trail (SQLite)\n", keyStyle.Render("[S] / [3]"))
	fmt.Printf("     - %s : Tampilkan Panduan Lengkap & Instruksi Sistem ini\n", keyStyle.Render("[H] / [4]"))
	fmt.Printf("     - %s : Periksa Kesehatan Daemon, Koneksi DB & Provider AI\n", keyStyle.Render("[C] / [5]"))
	fmt.Printf("     - %s : Jalankan Quick Setup Wizard Konfigurasi (.env)\n", keyStyle.Render("[Q] / [6]"))
	fmt.Printf("     - %s : Hentikan daemon server & keluar dari Niskava Agent\n", keyStyle.Render("[E] / [7]"))

	fmt.Println()
	fmt.Println(sectionStyle.Render("2. INSTRUKSI FITUR & OP-SURFACES:"))
	fmt.Printf("   • %s:\n     Dashboard riset visual berbasis React SPA di browser. Menampilkan candlestick\n     TradingView/Recharts, marker anomali kuantitatif (MA20/Z-score), matriks bukti 3-tier,\n     dan graf kausalitas memori.\n", descStyle.Render("[W] Web UI Workspace"))
	fmt.Printf("   • %s:\n     Terminal interaktif ReAct inner-monologue.\n     - Ketik 4-5 huruf kode emiten (contoh: ANTM, BBCA, BUMI) untuk auto-investigasi 30 hari.\n     - Ketik pertanyaan bebas (contoh: 'Kenapa saham ANTM naik tajam kemarin?').\n", descStyle.Render("[T] Terminal UI (REPL)"))
	fmt.Printf("   • %s:\n     Membaca tabel SQLite lokal (~/.niskava/niskava.db) untuk melihat semua riwayat audit trail.\n", descStyle.Render("[S] Riwayat Sesi"))
	fmt.Printf("   • %s:\n     Menampilkan status daemon server, Sectors v2 API key, dan provider LLM (Gemini/OpenAI).\n", descStyle.Render("[C] Health Check"))
	fmt.Printf("   • %s:\n     Panduan setup interaktif untuk mengonfigurasi file .env secara otomatis.\n", descStyle.Render("[Q] Quick Setup Wizard"))

	fmt.Println()
	fmt.Println(sectionStyle.Render("3. DAFTAR PERINTAH SLASH DALAM SESI REPL:"))
	fmt.Printf("   • %-16s : Tampilkan daftar perintah bantuan REPL.\n", cmdStyle.Render("/help"))
	fmt.Printf("   • %-16s : Mulai sesi percakapan baru & bersihkan memory graph.\n", cmdStyle.Render("/reset"))
	fmt.Printf("   • %-16s : Bersihkan layar terminal & tampilkan banner HUD.\n", cmdStyle.Render("/clear"))
	fmt.Printf("   • %-16s : Membuka Web Workspace langsung di browser.\n", cmdStyle.Render("/web"))
	fmt.Printf("   • %-16s : Lihat riwayat investigasi tersimpan dari SQLite.\n", cmdStyle.Render("/sessions"))
	fmt.Printf("   • %-16s : Periksa status kesehatan provider AI & database.\n", cmdStyle.Render("/health"))
	fmt.Printf("   • %-16s : Keluar dari sesi REPL kembali ke menu utama.\n", cmdStyle.Render("/exit, quit"))

	fmt.Println()
	fmt.Println(sectionStyle.Render("4. COMMAND LINE CLI (PERINTAH DIRECT TERMINAL):"))
	fmt.Printf("   • %s : Investigasi 1 baris.\n", mutedStyle.Render("niskava investigate <TICKER> --days 30"))
	fmt.Printf("   • %s : Jalankan server daemon background.\n", mutedStyle.Render("niskava serve --port 8080"))
	fmt.Printf("   • %s : Tampilkan sesi dari CLI.\n", mutedStyle.Render("niskava sessions"))
	fmt.Printf("   • %s : Jalankan setup wizard langsung.\n", mutedStyle.Render("niskava setup"))

	fmt.Println()
	fmt.Println(RenderConstellationLine(85))
}
