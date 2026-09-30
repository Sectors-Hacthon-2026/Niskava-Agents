// Package telegram provides Telegram Bot integration for Niskava Agent.
package telegram

import (
	"regexp"
	"strings"
)

const (
	// StandardDisclaimer is the mandatory non-advisory disclaimer for Niskava Agent.
	StandardDisclaimer = "*Disclaimer: Niskava Agent is an autonomous IDX capital market research intelligence platform, NOT an investment advisor or registered broker. Analysis is presented solely for research and fact-verification.*"

	// DisclaimerSuffix is appended to responses that don't already contain the disclaimer.
	DisclaimerSuffix = "\n\n---\n" + StandardDisclaimer

	// DefaultMaxMessageLength is the default maximum character count for a Telegram message chunk.
	DefaultMaxMessageLength = 4000
)

// headingRegex matches Markdown ATX headings: # H1 / ## H2 / ### H3 etc.
var headingRegex = regexp.MustCompile(`(?m)^#{1,6}\s+(.+)$`)

// blockquoteLineRegex matches Markdown blockquote lines: > content
var blockquoteLineRegex = regexp.MustCompile(`(?m)^>\s?(.*)$`)

// hrRegex matches standalone horizontal rules: --- or *** or ___
var hrRegex = regexp.MustCompile(`^(\*{3,}|-{3,}|_{3,})\s*$`)

// boldDoubleAsterisk converts **bold** to *bold* for Telegram MarkdownV1.
var boldDoubleAsterisk = regexp.MustCompile(`\*\*(.+?)\*\*`)

// multiBlankLines collapses 3+ consecutive newlines to exactly 2.
var multiBlankLines = regexp.MustCompile(`\n{3,}`)

// separatorCellRegex matches a table separator cell like ---, :---, ---:
var separatorCellRegex = regexp.MustCompile(`^:?-+:?$`)

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

// FormatFinalResponse formats the final response string.
// Conversational chat bubbles omit repeated disclaimers for cleaner UX,
// while formal export documents and generated PDFs retain strict compliance disclaimers.
func FormatFinalResponse(content string, thought string) string {
	text := strings.TrimSpace(content)
	if text == "" && strings.TrimSpace(thought) != "" {
		text = strings.TrimSpace(thought)
	}
	return text
}

// isTableLine returns true if the line is a Markdown pipe-table row.
func isTableLine(line string) bool {
	trimmed := strings.TrimSpace(line)
	return strings.HasPrefix(trimmed, "|") && strings.HasSuffix(trimmed, "|")
}

// isSeparatorRow returns true if every cell in the row is a Markdown table separator (--- / :---: etc.).
func isSeparatorRow(row string) bool {
	trimmed := strings.TrimSpace(row)
	trimmed = strings.TrimPrefix(trimmed, "|")
	trimmed = strings.TrimSuffix(trimmed, "|")
	cells := strings.Split(trimmed, "|")
	for _, cell := range cells {
		if !separatorCellRegex.MatchString(strings.TrimSpace(cell)) {
			return false
		}
	}
	return true
}

// splitTableRow splits a pipe-table row into trimmed cell strings.
func splitTableRow(row string) []string {
	trimmed := strings.TrimSpace(row)
	trimmed = strings.TrimPrefix(trimmed, "|")
	trimmed = strings.TrimSuffix(trimmed, "|")
	parts := strings.Split(trimmed, "|")
	for i := range parts {
		parts[i] = strings.TrimSpace(parts[i])
	}
	return parts
}

// transpileTable converts a slice of Markdown pipe-table lines into a
// bullet-point key-value list readable in Telegram Markdown V1.
// Header row cells become bold labels; data cells follow each label.
func transpileTable(tableLines []string) string {
	if len(tableLines) < 1 {
		return ""
	}

	headerCells := splitTableRow(tableLines[0])

	var sb strings.Builder
	for _, row := range tableLines[1:] {
		if isSeparatorRow(row) {
			continue
		}
		cells := splitTableRow(row)
		sb.WriteString("• ")
		for i, cell := range cells {
			cell = strings.TrimSpace(cell)
			if cell == "" {
				continue
			}
			if i < len(headerCells) {
				header := strings.TrimSpace(headerCells[i])
				if header != "" {
					sb.WriteString("*")
					sb.WriteString(header)
					sb.WriteString(":* ")
				}
			}
			sb.WriteString(cell)
			if i < len(cells)-1 {
				sb.WriteString(" — ")
			}
		}
		sb.WriteString("\n")
	}
	return strings.TrimRight(sb.String(), "\n")
}

// TranspileMarkdownForTelegram converts standard Markdown syntax unsupported by
// Telegram Markdown V1 into Telegram-compatible equivalents:
//
//   - ## Heading / ### Heading  →  *Heading*  (bold line)
//   - > blockquote              →  _blockquote content_  (italic, > removed)
//   - --- horizontal rule       →  (empty line)
//   - **bold**                  →  *bold*  (Telegram V1 bold syntax)
//   - Pipe tables               →  bullet-point key-value list
func TranspileMarkdownForTelegram(text string) string {
	lines := strings.Split(text, "\n")
	result := make([]string, 0, len(lines))
	i := 0

	for i < len(lines) {
		line := lines[i]

		// Collect and convert contiguous table blocks into bullet lists.
		if isTableLine(line) {
			var tableLines []string
			for i < len(lines) && isTableLine(lines[i]) {
				tableLines = append(tableLines, lines[i])
				i++
			}
			converted := transpileTable(tableLines)
			if converted != "" {
				result = append(result, converted)
			}
			continue
		}

		// ATX Headings: ## Title → *Title*
		if match := headingRegex.FindStringSubmatch(line); match != nil {
			title := strings.TrimSpace(match[1])
			result = append(result, "*"+title+"*")
			i++
			continue
		}

		// Blockquotes: > content → _content_
		if blockquoteLineRegex.MatchString(line) {
			content := blockquoteLineRegex.ReplaceAllString(line, "$1")
			content = strings.TrimSpace(content)
			if content != "" {
				result = append(result, "_"+content+"_")
			} else {
				result = append(result, "")
			}
			i++
			continue
		}

		// Horizontal rules: --- / *** / ___ → blank line
		if hrRegex.MatchString(strings.TrimSpace(line)) {
			result = append(result, "")
			i++
			continue
		}

		result = append(result, line)
		i++
	}

	// Convert **bold** → *bold*
	output := strings.Join(result, "\n")
	output = boldDoubleAsterisk.ReplaceAllString(output, "*$1*")

	// Collapse 3+ consecutive blank lines to 2 for cleaner spacing.
	output = multiBlankLines.ReplaceAllString(output, "\n\n")

	return strings.TrimSpace(output)
}

// SanitizeTelegramMarkdown transpiles unsupported Markdown to Telegram-compatible
// format, then balances unclosed markdown tags to avoid Telegram parse entity errors.
func SanitizeTelegramMarkdown(text string) string {
	// Step 1: transpile heading/blockquote/table/HR to Telegram-safe format.
	sanitized := TranspileMarkdownForTelegram(text)

	// Step 2: Balance backtick count (odd count breaks inline code rendering).
	if strings.Count(sanitized, "`")%2 != 0 {
		sanitized += "`"
	}

	// Step 3: Balance single asterisks (Telegram V1 bold/italic marker).
	if strings.Count(sanitized, "*")%2 != 0 {
		sanitized += "*"
	}

	// Step 4: Balance underscores (Telegram V1 italic marker).
	if strings.Count(sanitized, "_")%2 != 0 {
		sanitized += "_"
	}

	return sanitized
}
