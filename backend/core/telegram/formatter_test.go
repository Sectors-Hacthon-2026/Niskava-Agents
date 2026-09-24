package telegram

import (
	"strings"
	"testing"
)

func TestSplitMessageShort(t *testing.T) {
	short := "Halo, ini pesan singkat dari Niskava Agent."
	chunks := SplitMessage(short, 4000)
	if len(chunks) != 1 {
		t.Fatalf("expected 1 chunk, got %d", len(chunks))
	}
	if chunks[0] != short {
		t.Errorf("expected chunk to equal '%s', got '%s'", short, chunks[0])
	}
}

func TestSplitMessageEmpty(t *testing.T) {
	chunks := SplitMessage("", 4000)
	if len(chunks) != 1 || chunks[0] != "" {
		t.Fatalf("expected 1 empty chunk, got %v", chunks)
	}
}

func TestSplitMessageDefaultMaxLen(t *testing.T) {
	// If maxLen <= 0, defaults to 4000
	short := "Pesan dengan maxLen <= 0"
	chunks := SplitMessage(short, 0)
	if len(chunks) != 1 || chunks[0] != short {
		t.Fatalf("expected default maxLen to handle short message properly")
	}

	chunksNegative := SplitMessage(short, -10)
	if len(chunksNegative) != 1 || chunksNegative[0] != short {
		t.Fatalf("expected negative maxLen to default to 4000")
	}
}

func TestSplitMessageLongWithParagraphs(t *testing.T) {
	p1 := strings.Repeat("A", 2500)
	p2 := strings.Repeat("B", 2500)
	longText := p1 + "\n\n" + p2

	chunks := SplitMessage(longText, 4000)
	if len(chunks) != 2 {
		t.Fatalf("expected 2 chunks, got %d", len(chunks))
	}

	if len(chunks[0]) > 4000 {
		t.Errorf("chunk 0 exceeds maxLen: length=%d", len(chunks[0]))
	}
	if len(chunks[1]) > 4000 {
		t.Errorf("chunk 1 exceeds maxLen: length=%d", len(chunks[1]))
	}

	if chunks[0] != p1 {
		t.Errorf("expected chunk 0 to be paragraph 1")
	}
	if chunks[1] != p2 {
		t.Errorf("expected chunk 1 to be paragraph 2")
	}
}

func TestSplitMessageLongWithSingleNewlines(t *testing.T) {
	line1 := strings.Repeat("X", 2500)
	line2 := strings.Repeat("Y", 2500)
	longText := line1 + "\n" + line2

	chunks := SplitMessage(longText, 4000)
	if len(chunks) != 2 {
		t.Fatalf("expected 2 chunks, got %d", len(chunks))
	}

	if chunks[0] != line1 {
		t.Errorf("expected chunk 0 to be line 1")
	}
	if chunks[1] != line2 {
		t.Errorf("expected chunk 1 to be line 2")
	}
}

func TestSplitMessageLongWithSpaces(t *testing.T) {
	word1 := strings.Repeat("W", 2500)
	word2 := strings.Repeat("Z", 2500)
	longText := word1 + " " + word2

	chunks := SplitMessage(longText, 4000)
	if len(chunks) != 2 {
		t.Fatalf("expected 2 chunks, got %d", len(chunks))
	}

	if chunks[0] != word1 {
		t.Errorf("expected chunk 0 to split cleanly at space")
	}
	if chunks[1] != word2 {
		t.Errorf("expected chunk 1 to be word 2")
	}
}

func TestSplitMessageLongContinuousNoBreak(t *testing.T) {
	longText := strings.Repeat("C", 8500)

	chunks := SplitMessage(longText, 4000)
	if len(chunks) != 3 {
		t.Fatalf("expected 3 chunks for 8500 chars (4000+4000+500), got %d", len(chunks))
	}

	for i, chunk := range chunks {
		if len(chunk) > 4000 {
			t.Errorf("chunk %d exceeded limit: len=%d", i, len(chunk))
		}
	}

	combined := strings.Join(chunks, "")
	if combined != longText {
		t.Errorf("joined chunks do not match original continuous string")
	}
}

func TestFormatFinalResponse(t *testing.T) {
	content := "Analisis saham ANTM menunjukkan anomali volume Z-Score = 3.2."
	thought := "Proses berpikir ReAct engine..."

	// 1. Without disclaimer present
	formatted := FormatFinalResponse(content, thought)
	if !strings.Contains(formatted, content) {
		t.Errorf("expected formatted text to contain content")
	}
	if !strings.Contains(formatted, StandardDisclaimer) {
		t.Errorf("expected formatted text to contain standard disclaimer")
	}

	// 2. Calling FormatFinalResponse again should not duplicate disclaimer
	doubleFormatted := FormatFinalResponse(formatted, "")
	count := strings.Count(doubleFormatted, StandardDisclaimer)
	if count != 1 {
		t.Errorf("expected exactly 1 disclaimer, got %d", count)
	}

	// 3. If content is empty, falls back to thought
	formattedFromThought := FormatFinalResponse("", thought)
	if !strings.Contains(formattedFromThought, thought) {
		t.Errorf("expected formatted text to use thought when content is empty")
	}
	if !strings.Contains(formattedFromThought, StandardDisclaimer) {
		t.Errorf("expected disclaimer attached to thought fallback")
	}
}

func TestWrapMarkdownTables(t *testing.T) {
	input := `Berikut adalah rincian metrik anomali:

| Tanggal | Metrik | Nilai | Baseline | Z-Score |
|---|---|---|---|---|
| 2026-09-08 | VOLUME_SURGE | 125M | 24M | 35.71 |
| 2026-09-09 | PRICE_SURGE | +8.2% | +0.5% | 4.12 |

Hasil analisis membuktikan anomali signifikan.`

	output := WrapMarkdownTables(input)
	if !strings.Contains(output, "```") {
		t.Fatalf("expected output to contain code block for table wrapping")
	}
	if !strings.Contains(output, "```\n| Tanggal | Metrik |") {
		t.Errorf("expected table header to start right after ```, got:\n%s", output)
	}
	if !strings.Contains(output, "| 2026-09-09 | PRICE_SURGE | +8.2% | +0.5% | 4.12 |\n```") {
		t.Errorf("expected table to end with ```, got:\n%s", output)
	}
}

func TestSanitizeTelegramMarkdown(t *testing.T) {
	// Unclosed formatting tag like lone asterisk or underscore
	input := "Saham BBCA *memiliki volume tinggi tapi _tidak ada konfirmasi"
	output := SanitizeTelegramMarkdown(input)
	if output == "" {
		t.Fatalf("expected non-empty output")
	}
}
