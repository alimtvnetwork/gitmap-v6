package cmd

import (
	"strings"
	"testing"

	"github.com/alimtvnetwork/gitmap-v6/gitmap/constants"
)

// TestRenderInlineMarkdownHighlightsDoubleQuotes locks in that changelog
// bullet bodies route their text through the shared pretty-markdown
// quote-highlight rule (render.HighlightQuotesANSI) so the formatting
// matches `gitmap help` output. Regression guard: previously bullets only
// rendered **bold** / `code` and ignored "double quotes" entirely.
func TestRenderInlineMarkdownHighlightsDoubleQuotes(t *testing.T) {
	got := renderInlineMarkdown(`Renamed "old" to "new"`, 0)

	if !strings.Contains(got, constants.ColorCyan) {
		t.Fatalf("expected cyan ANSI for double-quoted spans, got %q", got)
	}
	if !strings.Contains(got, constants.ColorReset) {
		t.Fatalf("expected ColorReset to close cyan span, got %q", got)
	}
	// Sentinel tokens must not leak into terminal output.
	if strings.Contains(got, "[C]") || strings.Contains(got, "[/C]") {
		t.Fatalf("token sentinels leaked into ANSI output: %q", got)
	}
}

// TestRenderInlineMarkdownLeavesApostrophesAlone protects the single-quote
// passthrough rule — bullets like "user's repo" must not gain stray ANSI.
func TestRenderInlineMarkdownLeavesApostrophesAlone(t *testing.T) {
	got := renderInlineMarkdown(`user's repo`, 0)

	if strings.Contains(got, constants.ColorCyan) {
		t.Fatalf("apostrophes must not trigger cyan styling: %q", got)
	}
}

// TestRenderInlineMarkdownPreservesBoldAndCode keeps the existing
// bold/code behavior intact after wiring in the quote-highlight pass.
func TestRenderInlineMarkdownPreservesBoldAndCode(t *testing.T) {
	got := renderInlineMarkdown("Use **bold** and `code` here", 0)

	if !strings.Contains(got, constants.ChangelogPrettyBoldOpen) {
		t.Fatalf("bold open marker missing: %q", got)
	}
	if !strings.Contains(got, constants.ChangelogPrettyCodeOpen) {
		t.Fatalf("code open marker missing: %q", got)
	}
}
