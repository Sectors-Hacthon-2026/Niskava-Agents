package tui

import (
	"testing"

	tea "github.com/charmbracelet/bubbletea"
)

func TestI18nLanguageSwitching(t *testing.T) {
	// 1. Default should be "en"
	SetLanguage("en")
	if ActiveLanguage != "en" {
		t.Errorf("expected ActiveLanguage 'en', got %s", ActiveLanguage)
	}

	headerEn := T("header_title")
	if headerEn != " [●] NISKAVA AGENT — AUTONOMOUS MARKET INTELLIGENCE " {
		t.Errorf("unexpected English header string: %s", headerEn)
	}

	// 2. Switch to Indonesian "id"
	SetLanguage("id")
	if ActiveLanguage != "id" {
		t.Errorf("expected ActiveLanguage 'id', got %s", ActiveLanguage)
	}

	headerId := T("header_title")
	if headerId != " [●] NISKAVA AGENT — INTELIJEN PASAR OTONOM " {
		t.Errorf("unexpected Indonesian header string: %s", headerId)
	}

	// Reset to English
	SetLanguage("en")
}

func TestLangSelectorModel(t *testing.T) {
	model := NewLangSelectorModel()
	if len(model.Languages) < 2 {
		t.Fatalf("expected at least 2 supported languages")
	}

	// Test pressing '2' to select Indonesian "id"
	updated, _ := model.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'2'}})
	m := updated.(LangSelectorModel)

	if m.Selected != "id" {
		t.Errorf("expected selected language 'id', got %s", m.Selected)
	}
}
