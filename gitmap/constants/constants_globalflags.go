package constants

// Cross-command CLI flag values.
//
// Each command parses its own `flag.NewFlagSet`, so these strings are
// reused safely (e.g. --json appears for `list-versions`, `list-releases`,
// `amend-list`, etc.). Centralised here so renames stay consistent.
const (
	FlagOpenValue = "--open"
	FlagJSON      = "--json"
	FlagLimit     = "--limit"
	FlagSource    = "--source"
	FlagCompact   = "--compact"
	FlagGroups    = "--groups"
)