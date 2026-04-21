package main

import (
	"github.com/alimtvnetwork/gitmap-v5/gitmap/release"
)

// Smoke binary: prints the latest changelog entry through the new
// pretty renderer so we can eyeball the output during development.
func main() {
	entries, err := release.ReadChangelog()
	if err != nil {
		panic(err)
	}
	if len(entries) == 0 {
		return
	}
	// We can't import internal cmd renderer from a sibling main, so just
	// dump the parsed structure to confirm parsing works.
	e := entries[0]
	println("Version:", e.Version)
	println("Title:  ", e.Title)
	println("Bullets:", len(e.Bullets))
	for i, b := range e.Bullets {
		if i >= 6 {
			println("... (truncated)")
			break
		}
		print("  depth=", b.Depth, " marker=", b.Marker, " text=")
		println(b.Text)
	}
}
