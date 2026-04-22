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

Flags (show):
  --raw                      Disable pretty markdown rendering even on a TTY

Examples:
  gitmap templates list
  gitmap tpl tl
  gitmap templates show ignore go
  gitmap tpl ts attributes node
  gitmap templates show ignore go --raw   # bypass pretty renderer
`
	headerTemplatesList    = "KIND        LANG            SOURCE  PATH\n"
	fmtTemplatesListRow    = "%-10s  %-14s  %-6s  %s\n"
	labelTemplatesUser     = "user"
	labelTemplatesEmbed    = "embed"
	msgTemplatesEmpty      = "(no templates registered — embedded corpus is empty)\n"
	errTemplatesShowArgs   = "templates show requires <kind> <lang>; e.g. 'templates show ignore go'\n"
	errTemplatesShowFail   = "templates show: %v\n"
	errTemplatesListFail   = "templates list: %v\n"
	errUnknownTemplatesSub = "unknown 'templates' subcommand: %s\n"
	flagTemplatesShowRaw   = "raw"
	flagDescTemplatesRaw   = "Print template bytes verbatim, skipping the pretty markdown renderer"
	envTemplatesNoPretty   = "GITMAP_NO_PRETTY"
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

// runTemplatesShow prints one template to stdout. Markdown templates
// (.md / .markdown) are routed through render.RenderANSI when the user
// is on a real TTY and hasn't opted out via --raw or GITMAP_NO_PRETTY,
// matching the help-text rendering contract. Non-markdown templates
// (.gitignore, .gitattributes, …) are always written byte-for-byte so
// the output stays diff-friendly and safe to redirect into a file.
func runTemplatesShow(args []string) {
	rest, raw := parseTemplatesShowFlags(args)
	if len(rest) < 2 {
		fmt.Fprint(os.Stderr, errTemplatesShowArgs)
		os.Exit(1)
	}
	kind, lang := rest[0], rest[1]
	r, err := templates.Resolve(kind, lang)
	if err != nil {
		fmt.Fprintf(os.Stderr, errTemplatesShowFail, err)
		os.Exit(1)
	}

	out := r.Content
	if shouldPrettyRenderTemplate(r.Path, raw) {
		out = []byte(render.RenderANSI(string(r.Content)))
	}

	if _, err := os.Stdout.Write(out); err != nil {
		fmt.Fprintf(os.Stderr, errTemplatesShowFail, err)
		os.Exit(1)
	}
}

// parseTemplatesShowFlags pulls --raw out of the arg list and returns the
// remaining positional args + the flag value. Uses a dedicated FlagSet so
// flags can appear before or after the <kind> <lang> positionals.
func parseTemplatesShowFlags(args []string) ([]string, bool) {
	fs := flag.NewFlagSet(cmdTemplatesShow, flag.ExitOnError)
	rawFlag := fs.Bool(flagTemplatesShowRaw, false, flagDescTemplatesRaw)
	reordered := reorderFlagsBeforeArgs(args)
	_ = fs.Parse(reordered)

	return fs.Args(), *rawFlag
}

// shouldPrettyRenderTemplate decides whether to route template bytes
// through the pretty markdown renderer. All gates must pass:
//
//   - the template path has a markdown extension (.md / .markdown);
//   - the user did not pass --raw;
//   - GITMAP_NO_PRETTY is unset (shared opt-out env, mirrors helptext);
//   - stdout is connected to a real TTY (so pipes / redirects stay clean).
func shouldPrettyRenderTemplate(path string, raw bool) bool {
	if raw {
		return false
	}
	if !isMarkdownTemplatePath(path) {
		return false
	}
	if os.Getenv(envTemplatesNoPretty) != "" {
		return false
	}

	return templatesStdoutIsTerminal()
}

// isMarkdownTemplatePath returns true for .md / .markdown files
// (case-insensitive). Templates today are .gitignore / .gitattributes —
// this guard future-proofs the renderer for markdown overlays
// (e.g. ~/.gitmap/templates/notes/*.md) without changing existing UX.
func isMarkdownTemplatePath(path string) bool {
	ext := strings.ToLower(filepath.Ext(path))

	return ext == ".md" || ext == ".markdown"
}

// templatesStdoutIsTerminal mirrors helptext.stdoutIsTerminal — kept
// local to avoid exporting helptext internals just for one caller.
func templatesStdoutIsTerminal() bool {
	info, err := os.Stdout.Stat()
	if err != nil {
		return false
	}

	return (info.Mode() & os.ModeCharDevice) != 0
}

// sourceLabel maps a templates.Source to the user-facing column value.
func sourceLabel(s templates.Source) string {
	if s == templates.SourceUser {
		return labelTemplatesUser
	}

	return labelTemplatesEmbed
}
