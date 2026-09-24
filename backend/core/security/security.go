package security

import (
	"fmt"
	"regexp"
	"strings"
)

var (
	tickerRegex     = regexp.MustCompile(`^[A-Z]{4,5}$`)
	sectorsKeyRegex = regexp.MustCompile(`sec_live_[a-zA-Z0-9_-]+`)
	geminiKeyRegex  = regexp.MustCompile(`AIzaSy[a-zA-Z0-9_-]+`)
)

// ValidateTicker sanitizes and validates an IDX stock ticker (4-5 uppercase letters).
func ValidateTicker(ticker string) (string, error) {
	cleaned := strings.ToUpper(strings.TrimSpace(ticker))
	if !tickerRegex.MatchString(cleaned) {
		return "", fmt.Errorf("invalid IDX ticker '%s': tickers must be 4 to 5 alphabetical characters (e.g. ANTM, BBCA)", ticker)
	}
	return cleaned, nil
}

// ScrubSecrets redacts sensitive API keys from strings, error messages, and logs.
func ScrubSecrets(input string) string {
	scrubbed := sectorsKeyRegex.ReplaceAllString(input, "[REDACTED_SECTORS_KEY]")
	scrubbed = geminiKeyRegex.ReplaceAllString(scrubbed, "[REDACTED_GEMINI_KEY]")
	return scrubbed
}

// GetNonAdvisoryDisclaimer returns the official compliance disclaimer per Law 2 and Hackathon Rule 12.
func GetNonAdvisoryDisclaimer(lang string) string {
	if strings.ToLower(lang) == "en" {
		return "DISCLAIMER: Niskava Agent is an autonomous market intelligence research platform for the Indonesia Stock Exchange (IDX), NOT a licensed investment advice or broker. All findings, anomaly scores, and evidence items are presented descriptively for factual research and DO NOT constitute buy/sell recommendations or price targets."
	}
	return "DISCLAIMER: Niskava Agent adalah platform intelijen dan riset pasar modal otonom untuk Bursa Efek Indonesia (IDX), BUKAN penasihat investasi atau broker berizin. Seluruh temuan, skor anomali, dan korelasi bukti disajikan secara deskriptif untuk tujuan riset verifikasi fakta dan BUKAN merupakan rekomendasi beli/jual atau target harga investasi."
}
