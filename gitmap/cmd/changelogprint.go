package cmd

import (
	"fmt"

	"github.com/alimtvnetwork/gitmap-v5/gitmap/constants"
	"github.com/alimtvnetwork/gitmap-v5/gitmap/release"
)

// renderChangelogEntry pretty-prints a single changelog entry.
//
// Layout:
//
//	  ──────────────────────────────────────────────────────────────────────
//	  v3.32.1  •  Fix `gitmap status` looking at legacy bare `output/` path
//	  ──────────────────────────────────────────────────────────────────────
//	  • Bullet text wraps at 100 cols with hanging indent.
//	      ◦ Nested bullets are dimmer and indented under their parent.
//
// Bold (**x**) and inline code (`x`) markdown are rendered with ANSI; all
// other markdown is left as-is.
func renderChangelogEntry(entry release.ChangelogEntry) {
	bullets := selectChangelogBullets(entry)
	printChangelogHeader(entry)
	printChangelogBullets(bullets)
	fmt.Println()
}

// selectChangelogBullets prefers the structured Bullets slice, falling
// back to the flat Notes slice for legacy callers / tests.
func selectChangelogBullets(entry release.ChangelogEntry) []release.ChangelogBullet {
	if len(entry.Bullets) > 0 {
		return entry.Bullets
	}

	out := make([]release.ChangelogBullet, 0, len(entry.Notes))
	for _, note := range entry.Notes {
		out = append(out, release.ChangelogBullet{Depth: 0, Marker: "-", Text: note})
	}

	return out
}

// printChangelogHeader prints the rule + version + title block.
func printChangelogHeader(entry release.ChangelogEntry) {
	fmt.Println()
	fmt.Printf("  %s%s%s\n", constants.ColorDim, constants.ChangelogPrettyRule, constants.ColorReset)
	if len(entry.Title) > 0 {
		fmt.Printf(constants.ChangelogPrettyHeaderFmt,
			constants.ColorCyan, entry.Version, constants.ColorReset,
			constants.ColorDim+"  •  "+constants.ColorReset,
			constants.ColorWhite, entry.Title, constants.ColorReset)
	} else {
		fmt.Printf(constants.ChangelogPrettyHeaderBare,
			constants.ColorCyan, entry.Version, constants.ColorReset)
	}
	fmt.Printf("  %s%s%s\n", constants.ColorDim, constants.ChangelogPrettyRule, constants.ColorReset)
}

// printChangelogBullets renders each bullet with depth-aware styling.
func printChangelogBullets(bullets []release.ChangelogBullet) {
	width := changelogWrapWidth()
	for i := range bullets {
		printChangelogBullet(bullets[i], width)
	}
}

// printChangelogBullet renders a single bullet with hanging indent.
func printChangelogBullet(bullet release.ChangelogBullet, wrapWidth int) {
	indent := changelogIndent(bullet.Depth)
	marker := changelogMarker(bullet)
	color := changelogMarkerColor(bullet.Depth)
	prefix := fmt.Sprintf("  %s%s%s%s ", indent, color, marker, constants.ColorReset)
	hanging := "  " + indent + repeatSpace(visibleLen(marker)+1)

	body := renderInlineMarkdown(bullet.Text, bullet.Depth)
	wrapped := wrapWithHangingIndent(body, prefix, hanging, wrapWidth)
	fmt.Print(wrapped)
}

// changelogIndent returns the leading indent string for the given depth.
func changelogIndent(depth int) string {
	out := ""
	for i := 0; i < depth; i++ {
		out += constants.ChangelogPrettyIndentUnit
	}

	return out
}

// changelogMarker returns the bullet glyph or ordered-list marker.
func changelogMarker(bullet release.ChangelogBullet) string {
	if bullet.Ordered {
		return bullet.Marker
	}
	if bullet.Depth == 0 {
		return constants.ChangelogPrettyMarkerL0
	}
	if bullet.Depth == 1 {
		return constants.ChangelogPrettyMarkerL1
	}

	return constants.ChangelogPrettyMarkerLN
}

// changelogMarkerColor selects a colour based on bullet depth.
func changelogMarkerColor(depth int) string {
	if depth == 0 {
		return constants.ColorGreen
	}
	if depth == 1 {
		return constants.ColorCyan
	}

	return constants.ColorDim
}
