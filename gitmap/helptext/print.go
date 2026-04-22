package helptext

import (
	"embed"
	"fmt"
	"os"

	"github.com/alimtvnetwork/gitmap-v6/gitmap/render"
)

//go:embed *.md
var files embed.FS

// envNoPretty disables ANSI rendering when set to a non-empty value.
// Mirrors the convention of NO_COLOR but scoped to gitmap's pretty
// markdown pipeline so users can opt out without losing other coloring.
const envNoPretty = "GITMAP_NO_PRETTY"

// Print reads and prints the help file for the given command. When
// stdout is a TTY and pretty rendering is not disabled via env, the
// markdown is routed through render.RenderANSI for collapsed code
// blocks, cyan quotes, muted subtitles, and indented bodies. Otherwise
// the raw markdown is printed unchanged so pipes / redirects stay clean.
func Print(command string) {
	data, err := files.ReadFile(command + ".md")
	if err != nil {
		fmt.Fprintf(os.Stderr, "No help available for '%s'\n", command)
		os.Exit(1)
	}

	if shouldPretty() {
		fmt.Print(render.RenderANSI(string(data)))

		return
	}

	fmt.Print(string(data))
}

// PrintRaw bypasses the pretty renderer and prints the embedded
// markdown verbatim. Useful for callers that pipe help into a pager
// or other tooling that handles its own formatting.
func PrintRaw(command string) {
	data, err := files.ReadFile(command + ".md")
	if err != nil {
		fmt.Fprintf(os.Stderr, "No help available for '%s'\n", command)
		os.Exit(1)
	}
	fmt.Print(string(data))
}

// shouldPretty reports whether the pretty renderer should be applied
// to help-text output. We require both a TTY on stdout and the absence
// of an opt-out environment variable.
func shouldPretty() bool {
	if os.Getenv(envNoPretty) != "" {
		return false
	}

	return stdoutIsTerminal()
}

// stdoutIsTerminal reports whether stdout is connected to a real TTY.
// Dependency-free check using ModeCharDevice — matches the approach
// used by gitmap/cmd/scanprogress.go for stderr.
func stdoutIsTerminal() bool {
	info, err := os.Stdout.Stat()
	if err != nil {
		return false
	}

	return (info.Mode() & os.ModeCharDevice) != 0
}
