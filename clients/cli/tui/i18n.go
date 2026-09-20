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

// TUIStrings map holds English ("en") and Indonesian ("id") translations for core UI elements.
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
