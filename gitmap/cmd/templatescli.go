package cmd

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/alimtvnetwork/gitmap-v6/gitmap/constants"
	"github.com/alimtvnetwork/gitmap-v6/gitmap/render"
	"github.com/alimtvnetwork/gitmap-v6/gitmap/templates"
)

const (
	cmdTemplatesList      = "list"
	cmdTemplatesListAlias = "tl"
	cmdTemplatesShow      = "show"
	cmdTemplatesShowAlias = "ts"
	usageTemplatesRoot    = `Usage: gitmap templates <subcommand>

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
	headerTemplatesList  = "KIND        LANG            SOURCE  PATH\n"
	fmtTemplatesListRow  = "%-10s  %-14s  %-6s  %s\n"
	labelTemplatesUser   = "user"
	labelTemplatesEmbed  = "embed"
	msgTemplatesEmpty    = "(no templates registered — embedded corpus is empty)\n"
	errTemplatesShowArgs = "templates show requires <kind> <lang>; e.g. 'templates show ignore go'\n"
	errTemplatesShowFail = "templates show: %v\n"
	errTemplatesListFail = "templates list: %v\n"
	errUnknownTemplatesSub = "unknown 'templates' subcommand: %s\n"
)

// dispatchTemplates routes `gitmap templates <subcommand>` calls.
func dispatchTemplates(command string) bool {
	if command != constants.CmdTemplates && command != constants.CmdTemplatesAlias {
		return false
	}
	if len(os.Args) < 3 {
		fmt.Fprint(os.Stderr, usageTemplatesRoot)
		os.Exit(1)
	}

	sub, rest := os.Args[2], os.Args[3:]
	switch sub {
	case cmdTemplatesList, cmdTemplatesListAlias:
		runTemplatesList()
	case cmdTemplatesShow, cmdTemplatesShowAlias:
		runTemplatesShow(rest)
	default:
		fmt.Fprintf(os.Stderr, errUnknownTemplatesSub, sub)
		fmt.Fprint(os.Stderr, usageTemplatesRoot)
		os.Exit(1)
	}

	return true
}

// runTemplatesList prints every available template grouped by kind.
func runTemplatesList() {
	entries, err := templates.List()
	if err != nil {
		fmt.Fprintf(os.Stderr, errTemplatesListFail, err)
		os.Exit(1)
	}
	if len(entries) == 0 {
		fmt.Print(msgTemplatesEmpty)

		return
	}

	fmt.Print(headerTemplatesList)
	for _, e := range entries {
		fmt.Printf(fmtTemplatesListRow, e.Kind, e.Lang, sourceLabel(e.Source), e.Path)
	}
}

// runTemplatesShow prints one template's raw bytes (audit header included)
// to stdout. Useful for diff-against-curated workflows.
func runTemplatesShow(args []string) {
	if len(args) < 2 {
		fmt.Fprint(os.Stderr, errTemplatesShowArgs)
		os.Exit(1)
	}
	kind, lang := args[0], args[1]
	r, err := templates.Resolve(kind, lang)
	if err != nil {
		fmt.Fprintf(os.Stderr, errTemplatesShowFail, err)
		os.Exit(1)
	}
	if _, err := os.Stdout.Write(r.Content); err != nil {
		fmt.Fprintf(os.Stderr, errTemplatesShowFail, err)
		os.Exit(1)
	}
}

// sourceLabel maps a templates.Source to the user-facing column value.
func sourceLabel(s templates.Source) string {
	if s == templates.SourceUser {
		return labelTemplatesUser
	}

	return labelTemplatesEmbed
}
