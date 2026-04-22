package render

import "strings"

// emitBlock writes one parsed block in pretty form.
func emitBlock(out *strings.Builder, b block) {
	switch b.kind {
	case bkHeading:
		out.WriteString(b.text)
		out.WriteByte('\n')
	case bkSubtitle:
		out.WriteString(bodyIndent)
		out.WriteString(TokMutedOpen)
		out.WriteString(b.text)
		out.WriteString(TokMutedClose)
		out.WriteByte('\n')
	case bkParagraph:
		out.WriteString(bodyIndent)
		out.WriteString(highlightQuotes(b.text))
		out.WriteByte('\n')
	case bkFence:
		for _, l := range b.lines {
			out.WriteString(bodyIndent)
			out.WriteString(l)
			out.WriteByte('\n')
		}
	case bkBlank:
		out.WriteByte('\n')
	}
}

// highlightQuotes wraps every "double-quoted span" in cyan tokens.
// Single quotes are left untouched (they're usually apostrophes).
func highlightQuotes(s string) string {
	var b strings.Builder
	inQuote := false
	for i := 0; i < len(s); i++ {
		c := s[i]
		if c != '"' {
			b.WriteByte(c)

			continue
		}
		if !inQuote {
			b.WriteString(TokCyanOpen)
			b.WriteByte('"')
			inQuote = true

			continue
		}
		b.WriteByte('"')
		b.WriteString(TokCyanClose)
		inQuote = false
	}
	if inQuote { // unterminated quote: close the token defensively
		b.WriteString(TokCyanClose)
	}

	return b.String()
}
