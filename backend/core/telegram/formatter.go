// Package telegram provides Telegram Bot integration for Niskava Agent.
package telegram

import (
	"strings"
)

const (
	// StandardDisclaimer is the mandatory non-advisory disclaimer for Niskava Agent.
	StandardDisclaimer = "*Disclaimer: Niskava Agent adalah platform intelijen dan OSINT pasar modal otonom IDX, BUKAN penasihat investasi atau broker. Analisis disajikan untuk riset dan verifikasi fakta.*"

	// DisclaimerSuffix is appended to responses that don't already contain the disclaimer.
	DisclaimerSuffix = "\n\n---\n" + StandardDisclaimer

	// DefaultMaxMessageLength is the default maximum character count for a Telegram message chunk.
	DefaultMaxMessageLength = 4000
)

// SplitMessage safely breaks a long message into chunks under maxLen characters.
// It prioritizes splitting at double-newlines (paragraphs), then single-newlines,
// and finally word/space boundaries to avoid breaking markdown entities or words.
// If maxLen <= 0, it defaults to 4000.
func SplitMessage(text string, maxLen int) []string {
	if maxLen <= 0 {
		maxLen = DefaultMaxMessageLength
	}

	if len(text) <= maxLen {
		return []string{text}
	}

	var chunks []string
	remaining := text

	for len(remaining) > maxLen {
		sub := remaining[:maxLen]
		splitIdx := -1
		delimiterLen := 0

		// 1. Try double-newline (paragraph break)
		if idx := strings.LastIndex(sub, "\n\n"); idx > 0 {
			splitIdx = idx
			delimiterLen = 2
		} else if idx := strings.LastIndex(sub, "\n"); idx > 0 {
			// 2. Try single-newline
			splitIdx = idx
			delimiterLen = 1
		} else if idx := strings.LastIndex(sub, " "); idx > 0 {
			// 3. Try space boundary
			splitIdx = idx
			delimiterLen = 1
		} else {
			// 4. Hard break at maxLen
			splitIdx = maxLen
			delimiterLen = 0
		}

		chunk := strings.TrimRight(remaining[:splitIdx], " ")
		if chunk != "" {
			chunks = append(chunks, chunk)
		}
		remaining = strings.TrimLeft(remaining[splitIdx+delimiterLen:], "\n")
	}

	if len(remaining) > 0 {
		chunks = append(chunks, remaining)
	}

	return chunks
}

// FormatFinalResponse formats the final response string and appends the standard
// non-advisory disclaimer if not already present.
func FormatFinalResponse(content string, thought string) string {
	text := content
	if text == "" && thought != "" {
		text = thought
	}

	if strings.Contains(text, StandardDisclaimer) {
		return text
	}

	return text + DisclaimerSuffix
}

// WrapMarkdownTables detects Markdown tables and wraps them inside preformatted blocks (```)
// so that table columns don't break or collapse irregularly on mobile devices.
func WrapMarkdownTables(text string) string {
	lines := strings.Split(text, "\n")
	var result []string
	inTable := false

	isTableLine := func(line string) bool {
		trimmed := strings.TrimSpace(line)
		return strings.HasPrefix(trimmed, "|") && strings.HasSuffix(trimmed, "|")
	}

	for _, line := range lines {
		if isTableLine(line) {
			if !inTable {
				inTable = true
				result = append(result, "```")
			}
			result = append(result, line)
		} else {
			if inTable {
				inTable = false
				result = append(result, "```")
			}
			result = append(result, line)
		}
	}

	if inTable {
		result = append(result, "```")
	}

	return strings.Join(result, "\n")
}

// SanitizeTelegramMarkdown balances unclosed markdown tags (like unclosed * or _)
// to avoid Telegram parse entity errors.
func SanitizeTelegramMarkdown(text string) string {
	// First wrap tables to protect tabular pipes
	sanitized := WrapMarkdownTables(text)

	// Balance backticks
	backtickCount := strings.Count(sanitized, "`")
	if backtickCount%2 != 0 {
		sanitized += "`"
	}

	// Balance bold asterisks
	boldCount := strings.Count(sanitized, "**")
	if boldCount%2 != 0 {
		sanitized += "**"
	}

	// Balance single asterisks
	asteriskCount := strings.Count(sanitized, "*")
	if asteriskCount%2 != 0 {
		sanitized += "*"
	}

	// Balance underscores
	underscoreCount := strings.Count(sanitized, "_")
	if underscoreCount%2 != 0 {
		sanitized += "_"
	}

	return sanitized
}
