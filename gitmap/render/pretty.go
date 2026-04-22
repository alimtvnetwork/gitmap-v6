// Package render contains terminal-output helpers shared across CLI
// commands. The pretty markdown renderer (pretty.go) is the primary
// caller-facing API.
package render

import (
	"bufio"
	"strings"

	"github.com/alimtvnetwork/gitmap-v6/gitmap/constants"
)

// Token sentinels used by Render() so unit tests can assert on a stable,
// ANSI-free string. RenderANSI swaps these for real escape codes.
const (
	TokYellowOpen  = "[Y]"
	TokYellowClose = "[/Y]"
	TokCyanOpen    = "[C]"
	TokCyanClose   = "[/C]"
	TokMutedOpen   = "[M]"
	TokMutedClose  = "[/M]"

	collapseArrow = "→ "
	bodyIndent    = "  "
)

// Render applies the four pretty-print rules to markdown input and returns
// a token-annotated string (no ANSI escape codes). Use RenderANSI when
// emitting to a terminal.
//
// Rules:
//  1. A fenced code block whose body matches the immediately-preceding
//     paragraph's text collapses to a single yellow "→ <content>" line and
//     the fence is dropped.
//  2. Double-quoted strings ("like this") become cyan; single quotes are
//     left alone (apostrophes).
//  3. An italic line directly under a heading renders as a muted subtitle.
//  4. Body content under a heading is indented by two spaces.
func Render(md string) string {
	lines := splitLines(md)
	doc := parse(lines)

	var out strings.Builder
	for _, b := range doc {
		emitBlock(&out, b)
	}

	return strings.TrimRight(out.String(), "\n") + "\n"
}

// RenderANSI is Render with ANSI codes substituted for the token sentinels.
func RenderANSI(md string) string {
	r := strings.NewReplacer(
		TokYellowOpen, constants.ColorYellow,
		TokYellowClose, constants.ColorReset,
		TokCyanOpen, constants.ColorCyan,
		TokCyanClose, constants.ColorReset,
		TokMutedOpen, constants.ColorDim,
		TokMutedClose, constants.ColorReset,
	)

	return r.Replace(Render(md))
}

// splitLines splits on \n and drops the trailing empty token from a final
// newline. Preserves blank lines in between.
func splitLines(s string) []string {
	var out []string
	scanner := bufio.NewScanner(strings.NewReader(s))
	scanner.Buffer(make([]byte, 0, 64*1024), 1024*1024)
	for scanner.Scan() {
		out = append(out, scanner.Text())
	}

	return out
}
