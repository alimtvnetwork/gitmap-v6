package cmd

import (
	"fmt"
	"os"

	"github.com/alimtvnetwork/gitmap-v6/gitmap/constants"
	"github.com/alimtvnetwork/gitmap-v6/gitmap/templates"
)

// dispatchTemplates routes `gitmap templates <subcommand>` calls.
func dispatchTemplates(command string) bool {
	if command != constants.CmdTemplates && command != constants.CmdTemplatesAlias {
		return false
	}
	if len(os.Args) < 3 {
		fmt.Fprint(os.Stderr, constants.UsageTemplatesRoot)
		os.Exit(1)
	}

	sub, rest := os.Args[2], os.Args[3:]
	switch sub {
	case constants.TemplatesSubList, constants.TemplatesSubListAlias:
		runTemplatesList()
	case constants.TemplatesSubShow, constants.TemplatesSubShowAlias:
		runTemplatesShow(rest)
	default:
		fmt.Fprintf(os.Stderr, constants.ErrUnknownTemplatesSub, sub)
		fmt.Fprint(os.Stderr, constants.UsageTemplatesRoot)
		os.Exit(1)
	}

	return true
}

// runTemplatesList prints every available template grouped by kind.
func runTemplatesList() {
	entries, err := templates.List()
	if err != nil {
		fmt.Fprintf(os.Stderr, constants.ErrTemplatesListFailed, err)
		os.Exit(1)
	}
	if len(entries) == 0 {
		fmt.Print(constants.MsgTemplatesEmpty)

		return
	}

	fmt.Print(constants.HeaderTemplatesList)
	for _, e := range entries {
		fmt.Printf(constants.FmtTemplatesListRow, e.Kind, e.Lang, sourceLabel(e.Source), e.Path)
	}
}

// runTemplatesShow prints one template's raw bytes (audit header included)
// to stdout. Useful for diff-against-curated workflows.
func runTemplatesShow(args []string) {
	if len(args) < 2 {
		fmt.Fprint(os.Stderr, constants.ErrTemplatesShowArgs)
		os.Exit(1)
	}
	kind, lang := args[0], args[1]
	r, err := templates.Resolve(kind, lang)
	if err != nil {
		fmt.Fprintf(os.Stderr, constants.ErrTemplatesShowFailed, err)
		os.Exit(1)
	}
	if _, err := os.Stdout.Write(r.Content); err != nil {
		fmt.Fprintf(os.Stderr, constants.ErrTemplatesShowFailed, err)
		os.Exit(1)
	}
}

// sourceLabel maps a templates.Source to the user-facing column value.
func sourceLabel(s templates.Source) string {
	if s == templates.SourceUser {
		return constants.LabelTemplatesUser
	}

	return constants.LabelTemplatesEmbed
}
