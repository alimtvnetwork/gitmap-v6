package cmd

import "github.com/alimtvnetwork/gitmap-v5/gitmap/release"

// RenderChangelogEntryForSmoke is a tiny exported shim used only by the
// internal cmd/smokechangelog dev binary. It avoids exposing the rest of
// the cmd package and keeps the test surface explicit.
func RenderChangelogEntryForSmoke(entry release.ChangelogEntry) {
	renderChangelogEntry(entry)
}
