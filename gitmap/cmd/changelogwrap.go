package cmd

import (
	"os"
	"strconv"
	"strings"
	"unicode/utf8"

	"github.com/alimtvnetwork/gitmap-v5/gitmap/constants"
)

// changelogWrapWidth resolves the terminal column width used for wrapping
// changelog bullets. Honours $COLUMNS, then falls back to the constant
// default. Always clamped to [WrapMin, WrapMax] so a tiny terminal still
// renders readably and a giant terminal doesn't produce ridiculous lines.
func changelogWrapWidth() int {
	width := constants.ChangelogPrettyWrapDefault
	if cols := os.Getenv(constants.ChangelogPrettyEnvColumns); len(cols) > 0 {
		if parsed, err := strconv.Atoi(strings.TrimSpace(cols)); err == nil && parsed > 0 {
			width = parsed
		}
	}
	if width < constants.ChangelogPrettyWrapMin {
		return constants.ChangelogPrettyWrapMin
	}
	if width > constants.ChangelogPrettyWrapMax {
		return constants.ChangelogPrettyWrapMax
	}

	return width
}

// renderInlineMarkdown converts a small subset of Markdown to ANSI-styled
// text: **bold** and `code`. Other markdown is passed through unchanged.
// Depth is reserved for future depth-aware tweaks (e.g. dimming nested
// code spans); currently unused.
func renderInlineMarkdown(text string, _ int) string {
	out := convertInlineSpans(text, "**", constants.ChangelogPrettyBoldOpen, constants.ChangelogPrettyBoldClose)

	return convertInlineSpans(out, "`", constants.ChangelogPrettyCodeOpen, constants.ChangelogPrettyCodeClose)
}

// convertInlineSpans replaces matched delim pairs with ANSI open/close.
// Unmatched trailing delimiters are left in place.
func convertInlineSpans(text, delim, open, close string) string {
	var b strings.Builder
	rest := text
	for {
		start := strings.Index(rest, delim)
		if start < 0 {
			b.WriteString(rest)

			return b.String()
		}
		end := strings.Index(rest[start+len(delim):], delim)
		if end < 0 {
			b.WriteString(rest)

			return b.String()
		}
		b.WriteString(rest[:start])
		b.WriteString(open)
		b.WriteString(rest[start+len(delim) : start+len(delim)+end])
		b.WriteString(close)
		rest = rest[start+len(delim)+end+len(delim):]
	}
}

// wrapWithHangingIndent wraps body to fit within wrapWidth, prefixing the
// first line with prefix and every subsequent line with hanging.
// Returns a single string ending in "\n".
func wrapWithHangingIndent(body, prefix, hanging string, wrapWidth int) string {
	limit := wrapWidth - visibleLen(prefix)
	if limit < 10 {
		limit = 10
	}

	words := strings.Fields(body)
	lines := packWordsIntoLines(words, limit)

	return joinWrappedLines(lines, prefix, hanging)
}

// packWordsIntoLines groups words into lines no wider than limit.
func packWordsIntoLines(words []string, limit int) []string {
	var lines []string
	var current strings.Builder
	used := 0
	for _, word := range words {
		w := visibleLen(word)
		if used == 0 {
			current.WriteString(word)
			used = w
			continue
		}
		if used+1+w > limit {
			lines = append(lines, current.String())
			current.Reset()
			current.WriteString(word)
			used = w
			continue
		}
		current.WriteByte(' ')
		current.WriteString(word)
		used += 1 + w
	}
	if current.Len() > 0 {
		lines = append(lines, current.String())
	}

	return lines
}

// joinWrappedLines stitches packed lines with the right indents.
func joinWrappedLines(lines []string, prefix, hanging string) string {
	if len(lines) == 0 {
		return prefix + "\n"
	}

	var b strings.Builder
	for i, line := range lines {
		if i == 0 {
			b.WriteString(prefix)
		} else {
			b.WriteString(hanging)
		}
		b.WriteString(line)
		b.WriteByte('\n')
	}

	return b.String()
}

// visibleLen returns the rune count of s, ignoring ANSI escape sequences
// so wrapping math doesn't count colour codes as printable characters.
func visibleLen(s string) int {
	count := 0
	i := 0
	for i < len(s) {
		if s[i] == 0x1b && i+1 < len(s) && s[i+1] == '[' {
			i = skipAnsiSequence(s, i)
			continue
		}
		_, size := utf8.DecodeRuneInString(s[i:])
		count++
		i += size
	}

	return count
}

// skipAnsiSequence advances past a CSI escape sequence (ESC [ … letter).
func skipAnsiSequence(s string, start int) int {
	i := start + 2
	for i < len(s) {
		c := s[i]
		i++
		if (c >= 'A' && c <= 'Z') || (c >= 'a' && c <= 'z') {
			return i
		}
	}

	return i
}

// repeatSpace returns a string of n spaces.
func repeatSpace(n int) string {
	if n <= 0 {
		return ""
	}

	return strings.Repeat(" ", n)
}
