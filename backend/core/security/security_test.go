package security

import (
	"strings"
	"testing"
)

func TestValidateTicker(t *testing.T) {
	valid := []string{"ANTM", "BBCA", "TLKM", "GOTO", "BUMI", "admr"}
	for _, tk := range valid {
		res, err := ValidateTicker(tk)
		if err != nil {
			t.Errorf("expected '%s' to be valid, got err: %v", tk, err)
		}
		if res != strings.ToUpper(strings.TrimSpace(tk)) {
			t.Errorf("expected uppercase '%s', got '%s'", strings.ToUpper(tk), res)
		}
	}

	invalid := []string{"", "A", "AB", "ABC", "TOOLONGTICKER", "BBCA1", "ANTM; DROP TABLE", "AAPL$"}
	for _, tk := range invalid {
		_, err := ValidateTicker(tk)
		if err == nil {
			t.Errorf("expected '%s' to be invalid, got nil err", tk)
		}
	}
}

func TestScrubSecrets(t *testing.T) {
	raw := "Error connecting with key sec_live_98a7sd8f7a6sdf87asd6f876 and token AIzaSyD87as6df876asdf876 in log"
	scrubbed := ScrubSecrets(raw)
	if strings.Contains(scrubbed, "sec_live_98a7sd8f7a6sdf87asd6f876") {
		t.Errorf("failed to scrub Sectors API key")
	}
	if strings.Contains(scrubbed, "AIzaSyD87as6df876asdf876") {
		t.Errorf("failed to scrub Gemini API key")
	}
	if !strings.Contains(scrubbed, "[REDACTED_SECTORS_KEY]") {
		t.Errorf("missing replacement tag for sectors key")
	}
	if !strings.Contains(scrubbed, "[REDACTED_GEMINI_KEY]") {
		t.Errorf("missing replacement tag for gemini key")
	}
}

func TestNonAdvisoryDisclaimer(t *testing.T) {
	idDisclaimer := GetNonAdvisoryDisclaimer("id")
	if !strings.Contains(idDisclaimer, "penasihat investasi") {
		t.Errorf("Indonesian disclaimer must include 'penasihat investasi'")
	}
	enDisclaimer := GetNonAdvisoryDisclaimer("en")
	if !strings.Contains(enDisclaimer, "investment advice") {
		t.Errorf("English disclaimer must include 'investment advice'")
	}
}
