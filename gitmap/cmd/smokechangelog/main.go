// Package main is a smoke binary that invokes the changelog pretty
// renderer end-to-end so we can eyeball the output visually.
package main

import (
	"github.com/alimtvnetwork/gitmap-v5/gitmap/cmd"
	"github.com/alimtvnetwork/gitmap-v5/gitmap/release"
)

func main() {
	entries, err := release.ReadChangelog()
	if err != nil {
		panic(err)
	}
	if len(entries) == 0 {
		return
	}
	cmd.RenderChangelogEntryForSmoke(entries[0])
}
