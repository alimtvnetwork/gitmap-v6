// Package constants — constants_templates_cli.go: CLI identifiers for the
// `gitmap templates ...` discovery command.
package constants

// Top-level command. // gitmap:cmd top-level
const (
	CmdTemplates      = "templates"
	CmdTemplatesAlias = "tpl"
)

// `templates` subcommands and their short aliases.
const (
	TemplatesSubList      = "list"
	TemplatesSubListAlias = "tl"
	TemplatesSubShow      = "show"
	TemplatesSubShowAlias = "ts"
)

// User-facing strings.
const (
	UsageTemplatesRoot = `Usage: gitmap templates <subcommand>

Subcommands:
  list                       List every available template (alias: tl)
  show <kind> <lang>         Print a single template to stdout (alias: ts)

Kinds:
  ignore | attributes | lfs

Examples:
  gitmap templates list
  gitmap tpl tl
  gitmap templates show ignore go
  gitmap tpl ts attributes node
`
	HeaderTemplatesList    = "KIND        LANG            SOURCE  PATH\n"
	FmtTemplatesListRow    = "%-10s  %-14s  %-6s  %s\n"
	LabelTemplatesUser     = "user"
	LabelTemplatesEmbed    = "embed"
	MsgTemplatesEmpty      = "(no templates registered — embedded corpus is empty)\n"
	ErrTemplatesShowArgs   = "templates show requires <kind> <lang>; e.g. 'templates show ignore go'\n"
	ErrTemplatesShowFailed = "templates show: %v\n"
	ErrTemplatesListFailed = "templates list: %v\n"
	ErrUnknownTemplatesSub = "unknown 'templates' subcommand: %s\n"
)
