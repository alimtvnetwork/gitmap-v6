package formatter

import (
	"fmt"
	"io"
	"sort"

	"github.com/alimtvnetwork/gitmap-v6/gitmap/constants"
	"github.com/alimtvnetwork/gitmap-v6/gitmap/model"
)

// BranchSourceDebug renders a per-repo breakdown of how each repo's
// BranchSource was determined, plus a tally line summarizing how many
// repos fell into each bucket (HEAD, detached, remote-tracking,
// default, unknown). Output goes to w.
//
// This is a diagnostic view only — it is gated behind the
// --branch-source-debug flag and never appears in the default scan
// output, so the layout is optimized for readability rather than for
// machine parsing. Tally counts are emitted in deterministic order so
// snapshot-style tests stay stable.
func BranchSourceDebug(w io.Writer, records []model.ScanRecord) {
	fmt.Fprintln(w)
	fmt.Fprintf(w, constants.ColorYellow+constants.TermBranchSrcHeader+constants.ColorReset+"\n")
	fmt.Fprintf(w, constants.ColorDim+constants.TermSeparator+constants.ColorReset+"\n")

	tally := make(map[string]int, 5)
	for _, r := range records {
		src := normalizeBranchSource(r.BranchSource)
		tally[src]++
		printBranchSourceRow(w, r, src)
	}

	fmt.Fprintf(w, constants.ColorDim+constants.TermSeparator+constants.ColorReset+"\n")
	printBranchSourceTally(w, tally)
	fmt.Fprintln(w)
}

// normalizeBranchSource maps an empty source to "unknown" so the tally
// always uses a stable label. Any non-empty value is passed through
// untouched (it already matches one of the gitutil.BranchSource* constants).
func normalizeBranchSource(src string) string {
	if len(src) == 0 {
		return "unknown"
	}

	return src
}

// printBranchSourceRow writes one repo's debug line. The source label
// is colored to make scanning a long list easier:
//   - green  → HEAD              (the trustworthy case)
//   - cyan   → remote-tracking   (next best)
//   - yellow → default / detached (heuristic fallback)
//   - red    → unknown           (we couldn't tell — investigate)
func printBranchSourceRow(w io.Writer, r model.ScanRecord, src string) {
	color := branchSourceColor(src)
	fmt.Fprintf(w, constants.TermBranchSrcRowFmt,
		truncate(r.RepoName, 32),
		truncate(r.Branch, 10),
		color+src+constants.ColorReset,
	)
}

// branchSourceColor picks an ANSI color for a branch-source label.
func branchSourceColor(src string) string {
	switch src {
	case "HEAD":
		return constants.ColorGreen
	case "remote-tracking":
		return constants.ColorCyan
	case "default", "detached":
		return constants.ColorYellow
	default:
		return constants.ColorRed
	}
}

// printBranchSourceTally writes the totals line, e.g.
//   Totals: HEAD=12 remote-tracking=3 default=1 unknown=0
// The labels are sorted alphabetically for deterministic output.
func printBranchSourceTally(w io.Writer, tally map[string]int) {
	labels := make([]string, 0, len(tally))
	for k := range tally {
		labels = append(labels, k)
	}
	sort.Strings(labels)

	parts := make([]string, 0, len(labels))
	for _, l := range labels {
		parts = append(parts, fmt.Sprintf(constants.TermBranchSrcCountFmt, l, tally[l]))
	}
	fmt.Fprintf(w, constants.TermBranchSrcTallyFmt, joinSpace(parts))
}

// joinSpace joins parts with a single space. Kept inline to avoid
// importing strings just for one call site.
func joinSpace(parts []string) string {
	if len(parts) == 0 {
		return ""
	}
	out := parts[0]
	for i := 1; i < len(parts); i++ {
		out += " " + parts[i]
	}

	return out
}

// truncate clips s to at most n runes, appending no ellipsis (the
// column is right-padded by the format string anyway).
func truncate(s string, n int) string {
	if len(s) <= n {
		return s
	}

	return s[:n]
}
