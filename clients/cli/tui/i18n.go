// Package tui provides interactive terminal interfaces for Niskava Agent.
package tui

import "strings"

// ActiveLanguage determines the current active locale (defaults to "en").
var ActiveLanguage = "en"

// SetLanguage updates the active locale.
func SetLanguage(lang string) {
	if strings.ToLower(lang) == "id" || strings.ToLower(lang) == "indonesian" {
		ActiveLanguage = "id"
	} else {
		ActiveLanguage = "en"
	}
}

// TUIStrings map holds English ("en") default and Indonesian ("id") translations for all UI elements.
var TUIStrings = map[string]map[string]string{
	"header_title": {
		"en": " [●] NISKAVA AGENT — AUTONOMOUS MARKET INTELLIGENCE ",
		"id": " [●] NISKAVA AGENT — INTELIJEN PASAR OTONOM ",
	},
	"target_label": {
		"en": " Target: ",
		"id": " Target: ",
	},
	"observation_horizon": {
		"en": " (%d Days Observation Horizon)\n",
		"id": " (%d Hari Pengamatan)\n",
	},
	"agent_reasoning": {
		"en": "💭 Agent Reasoning (ReAct Monologue):",
		"id": "💭 Penalaran Agen (Monolog ReAct):",
	},
	"tool_activity": {
		"en": "🛠️  Tool Execution Activity:",
		"id": "🛠️  Aktivitas Eksekusi Alat:",
	},
	"anomaly_detected": {
		"en": "🚨 QUANTITATIVE ANOMALY DETECTED (NUMPY LAW 1):",
		"id": "🚨 ANOMALI KUANTITATIF TERDETEKSI (NUMPY LAW 1):",
	},
	"audit_trail_summary": {
		"en": "AUDIT TRAIL SUMMARY (3-TIER TAXONOMY):",
		"id": "RINGKASAN TEMUAN (AUDIT TRAIL 3-TIER):",
	},
	"causality_label": {
		"en": "Causality",
		"id": "Kausalitas",
	},
	"confidence_score_label": {
		"en": "Confidence Score",
		"id": "Skor Keyakinan",
	},
	"financial_disclaimer": {
		"en": "FINANCIAL DISCLAIMER (NON-ADVISORY - LAW 2 & RULE 12):\nNiskava Agent is an autonomous market intelligence and OSINT platform, NOT an investment advisor.\nThe system NEVER provides BUY/SELL recommendations or security price targets.",
		"id": "DISCLAIMER FINANSIAL (NON-ADVISORY - LAW 2 & ATURAN 12):\nNiskava Agent adalah platform intelijen pasar dan OSINT otonom, BUKAN penasihat investasi.\nSistem TIDAK PERNAH memberikan rekomendasi BELI/JUAL atau target harga sekuritas apa pun.",
	},
	"session_saved_hint": {
		"en": "Session saved: %s (%s)\nType 'niskava serve --open' to open interactive web workspace in browser.\n",
		"id": "Sesi tersimpan: %s (%s)\nKetik 'niskava serve --open' untuk membuka visual workspace interaktif di browser.\n",
	},
	"prompt_placeholder": {
		"en": "Type market research question or / for commands...",
		"id": "Ketik pertanyaan riset pasar atau / untuk perintah...",
	},
	"launcher_web_title": {
		"en": "Web UI (Open in Browser)",
		"id": "Web UI (Buka di Browser)",
	},
	"launcher_web_desc": {
		"en": "Start web daemon server & auto-open in default browser",
		"id": "Jalankan server web & buka otomatis di browser default",
	},
	"launcher_term_title": {
		"en": "Terminal UI (Interactive Live CLI)",
		"id": "Terminal UI (CLI Interaktif)",
	},
	"launcher_term_desc": {
		"en": "Interactive research REPL session with live anomaly reasoning",
		"id": "Sesi REPL interaktif berbasis perintah riset & anomali",
	},
	"launcher_sessions_title": {
		"en": "Session History & Audit Trail (SQLite)",
		"id": "Riwayat Sesi & Audit Trail (SQLite)",
	},
	"launcher_sessions_desc": {
		"en": "Inspect past investigation sessions & verified evidence from local database",
		"id": "Inspeksi riwayat investigasi & bukti terverifikasi dari database",
	},
	"launcher_help_title": {
		"en": "Help Guide & Usage Instructions",
		"id": "Panduan & Instruksi Penggunaan (Help Guide)",
	},
	"launcher_help_desc": {
		"en": "Complete guide on navigation, slash commands, and system architecture",
		"id": "Instruksi lengkap navigasi, opsi menu, dan perintah slash",
	},
	"launcher_health_title": {
		"en": "System & API Key Health Check",
		"id": "Pemeriksaan Kesehatan Sistem & API Key",
	},
	"launcher_health_desc": {
		"en": "Check status of daemon server, database connection, & AI providers",
		"id": "Periksa status daemon server, koneksi database, dan provider AI",
	},
	"launcher_setup_title": {
		"en": "Quick Setup Wizard (.env)",
		"id": "Quick Setup Wizard (.env)",
	},
	"launcher_setup_desc": {
		"en": "Quick setup wizard for Sectors, Gemini, or OpenAI API keys",
		"id": "Konfigurasi cepat API key Sectors, Gemini, atau OpenAI",
	},
	"launcher_lang_title": {
		"en": "Language Preference (Active: English)",
		"id": "Preferensi Bahasa (Aktif: Bahasa Indonesia)",
	},
	"launcher_lang_desc": {
		"en": "Toggle active interface language between English (Default) and Bahasa Indonesia",
		"id": "Ubah preferensi bahasa antarmuka antara Bahasa Inggris dan Bahasa Indonesia",
	},
	"launcher_exit_title": {
		"en": "Exit",
		"id": "Keluar",
	},
	"launcher_exit_desc": {
		"en": "Stop daemon server and exit Niskava Agent",
		"id": "Hentikan daemon server dan keluar dari Niskava Agent",
	},
	"launcher_quitting_msg": {
		"en": "\nExiting Niskava Agent. Goodbye!\n",
		"id": "\nKeluar dari Niskava Agent. Sampai jumpa!\n",
	},
	"slash_help_desc": {
		"en": "Complete guide to commands & system instructions",
		"id": "Panduan lengkap perintah & instruksi sistem",
	},
	"slash_reset_desc": {
		"en": "Start new chat session & clear memory graph",
		"id": "Mulai sesi obrolan baru & bersihkan memory graph",
	},
	"slash_graph_desc": {
		"en": "Open visual Cyber-OSINT Knowledge Graph in browser",
		"id": "Buka visualisasi Cyber-OSINT Knowledge Graph di browser",
	},
	"slash_clear_desc": {
		"en": "Clear terminal screen & redraw HUD banner",
		"id": "Bersihkan layar terminal & tampilkan ulang banner HUD",
	},
	"slash_web_desc": {
		"en": "Open Web Workspace visual dashboard in browser",
		"id": "Buka dashboard visual Web Workspace di browser",
	},
	"slash_sessions_desc": {
		"en": "Inspect investigation session history & audit trail from SQLite",
		"id": "Inspeksi riwayat sesi investigasi & audit trail dari SQLite",
	},
	"slash_health_desc": {
		"en": "Check status of daemon server, database, & AI providers",
		"id": "Periksa status daemon server, database, & provider AI",
	},
	"slash_lang_desc": {
		"en": "Switch active language preference (/lang en | /lang id)",
		"id": "Ubah preferensi bahasa aktif (/lang en | /lang id)",
	},
	"slash_exit_desc": {
		"en": "Exit Live REPL session back to main menu",
		"id": "Keluar dari sesi Live REPL kembali ke menu utama",
	},
	"help_header": {
		"en": "\nCOMMAND LIST FOR NISKAVA LIVE ASSISTANT:",
		"id": "\nDAFTAR PERINTAH NISKAVA LIVE ASSISTANT:",
	},
	"help_prompt_desc": {
		"en": "  <FREE PROMPT>        Ask stock market research questions (e.g. 'Why did ANTM stock surge yesterday?')",
		"id": "  <PROMPT BEBAS>       Tanyakan pertanyaan riset pasar saham (contoh: 'Kenapa saham ANTM naik kemarin?')",
	},
	"help_ticker_desc": {
		"en": "  <STOCK TICKER>       Type 4-letter stock ticker directly for quick analysis (e.g. ANTM, BBCA, BUMI)",
		"id": "  <KODE EMITEN>        Ketik langsung 4 huruf kode emiten untuk analisis cepat (contoh: ANTM, BBCA, BUMI)",
	},
	"help_graph_desc": {
		"en": "  /graph               Open visual Cyber-OSINT Knowledge Graph in browser",
		"id": "  /graph               Buka visualisasi Cyber-OSINT Knowledge Graph di browser",
	},
	"help_reset_desc": {
		"en": "  /reset               Start new conversation session & clear memory graph",
		"id": "  /reset               Mulai sesi percakapan baru & bersihkan memory graph",
	},
	"help_sessions_desc": {
		"en": "  /sessions            Inspect investigation session history & audit trail from local SQLite",
		"id": "  /sessions            Lihat riwayat sesi investigasi & audit trail dari SQLite lokal",
	},
	"help_web_desc": {
		"en": "  /web                 Open Web Workspace visual dashboard in browser",
		"id": "  /web                 Buka dashboard visual Web Workspace di browser",
	},
	"help_health_desc": {
		"en": "  /health              Check status of database, API keys, and AI providers",
		"id": "  /health              Periksa status database, API keys, dan provider AI",
	},
	"help_lang_desc": {
		"en": "  /lang [en|id]        Switch active interface language preference",
		"id": "  /lang [en|id]        Ubah preferensi bahasa aktif",
	},
	"help_clear_desc": {
		"en": "  /clear               Clear terminal screen",
		"id": "  /clear               Bersihkan layar terminal",
	},
	"help_exit_desc": {
		"en": "  /exit, quit          Exit REPL session",
		"id": "  /exit, quit          Keluar dari sesi REPL",
	},
	"health_header": {
		"en": "\nSYSTEM HEALTH DIAGNOSTICS STATUS:",
		"id": "\nSTATUS KESEHATAN SISTEM:",
	},
	"health_installed": {
		"en": "Configured (Live Ready)",
		"id": "Terpasang (Live Ready)",
	},
	"health_not_installed": {
		"en": "Not Configured (Offline Mode Active)",
		"id": "Belum Terpasang (Mode Offline Aktif)",
	},
	"sessions_header": {
		"en": "\nSAVED INVESTIGATION SESSION HISTORY (SQLITE):",
		"id": "\nRIWAYAT SESI INVESTIGASI TERSIMPAN (SQLITE):",
	},
	"sessions_empty": {
		"en": "No saved investigation sessions found.",
		"id": "Belum ada sesi investigasi tersimpan.",
	},
}

// T retrieves localized string for ActiveLanguage, falling back to "en".
func T(key string) string {
	if translations, exists := TUIStrings[key]; exists {
		if text, hasLang := translations[ActiveLanguage]; hasLang {
			return text
		}
		if text, hasDefault := translations["en"]; hasDefault {
			return text
		}
	}
	return key
}

// GetLocalizedLauncherItems returns menu items localized according to ActiveLanguage.
func GetLocalizedLauncherItems() []LauncherItem {
	return []LauncherItem{
		{
			ShortcutKey: "W",
			Title:       T("launcher_web_title"),
			Description: T("launcher_web_desc"),
			ActionID:    "web",
		},
		{
			ShortcutKey: "T",
			Title:       T("launcher_term_title"),
			Description: T("launcher_term_desc"),
			ActionID:    "terminal",
		},
		{
			ShortcutKey: "S",
			Title:       T("launcher_sessions_title"),
			Description: T("launcher_sessions_desc"),
			ActionID:    "sessions",
		},
		{
			ShortcutKey: "H",
			Title:       T("launcher_help_title"),
			Description: T("launcher_help_desc"),
			ActionID:    "help",
		},
		{
			ShortcutKey: "C",
			Title:       T("launcher_health_title"),
			Description: T("launcher_health_desc"),
			ActionID:    "health",
		},
		{
			ShortcutKey: "L",
			Title:       T("launcher_lang_title"),
			Description: T("launcher_lang_desc"),
			ActionID:    "lang",
		},
		{
			ShortcutKey: "Q",
			Title:       T("launcher_setup_title"),
			Description: T("launcher_setup_desc"),
			ActionID:    "setup",
		},
		{
			ShortcutKey: "E",
			Title:       T("launcher_exit_title"),
			Description: T("launcher_exit_desc"),
			ActionID:    "exit",
		},
	}
}

// GetLocalizedSlashCommands returns slash commands localized according to ActiveLanguage.
func GetLocalizedSlashCommands() []SlashCommand {
	return []SlashCommand{
		{Command: "/help", Description: T("slash_help_desc")},
		{Command: "/reset", Description: T("slash_reset_desc")},
		{Command: "/graph", Description: T("slash_graph_desc")},
		{Command: "/clear", Description: T("slash_clear_desc")},
		{Command: "/web", Description: T("slash_web_desc")},
		{Command: "/sessions", Description: T("slash_sessions_desc")},
		{Command: "/health", Description: T("slash_health_desc")},
		{Command: "/lang", Description: T("slash_lang_desc")},
		{Command: "/exit", Description: T("slash_exit_desc")},
	}
}
