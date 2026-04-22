// Package constants — constants_add.go: CLI identifiers for the
// `gitmap add ...` subcommand router (templates feature).
package constants

// Top-level command. // gitmap:cmd top-level
const (
	CmdAdd      = "add"
	CmdAddAlias = "ad"
)

// `add` subcommands and their short aliases.
const (
	AddSubIgnore           = "ignore"
	AddSubIgnoreAlias      = "ai"
	AddSubAttributes       = "attributes"
	AddSubAttributesAlias  = "aa"
	AddSubLFSInstall       = "lfs-install"
	AddSubLFSInstallAlias  = "alfs"
)

// User-facing messages for the `add` family.
const (
	MsgAddIgnoreWritten     = "[add ignore] wrote %s (managed=%d, user=%d)\n"
	MsgAddIgnoreUnchanged   = "[add ignore] %s is already up to date\n"
	MsgAddAttributesWritten = "[add attributes] wrote %s (managed=%d, user=%d)\n"
	MsgAddAttributesNoop    = "[add attributes] %s is already up to date\n"
	MsgAddLFSDone           = "[add lfs-install] git lfs install + LFS attributes ready in %s\n"

	UsageAddRoot = `Usage: gitmap add <subcommand> [args]

Subcommands:
  ignore [langs...]      Merge curated .gitignore templates (alias: ai)
  attributes [langs...]  Merge curated .gitattributes templates (alias: aa)
  lfs-install            Run git lfs install + add LFS attributes (alias: alfs)

Examples:
  gitmap add ignore go node
  gitmap ai python rust
  gitmap add attributes go
  gitmap add lfs-install
`
)

// Errors for the `add` family.
const (
	ErrUnknownAddSubcommand = "unknown 'add' subcommand: %s\n"
	ErrAddNoLangs           = "at least one language is required (e.g. 'go', 'node'); see 'gitmap templates list'\n"
	ErrAddMerge             = "merge failed: %v\n"
	ErrAddLFSInstall        = "git lfs install failed: %v\n"
)
