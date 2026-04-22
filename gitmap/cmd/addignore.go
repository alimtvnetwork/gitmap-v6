package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/alimtvnetwork/gitmap-v6/gitmap/constants"
	"github.com/alimtvnetwork/gitmap-v6/gitmap/templates"
)

// runAddIgnore handles `gitmap add ignore [langs...]`.
func runAddIgnore(args []string) {
	if len(args) == 0 {
		fmt.Fprint(os.Stderr, constants.ErrAddNoLangs)
		os.Exit(1)
	}

	cwd, err := os.Getwd()
	if err != nil {
		fmt.Fprintf(os.Stderr, constants.ErrAddMerge, err)
		os.Exit(1)
	}
	target := filepath.Join(cwd, ".gitignore")

	res, err := templates.Merge(templates.MergeOptions{
		TargetPath: target,
		Kind:       constants.TemplateKindIgnore,
		Langs:      args,
	})
	if err != nil {
		fmt.Fprintf(os.Stderr, constants.ErrAddMerge, err)
		os.Exit(1)
	}

	if !res.Changed {
		fmt.Printf(constants.MsgAddIgnoreUnchanged, res.WrittenPath)

		return
	}
	fmt.Printf(constants.MsgAddIgnoreWritten, res.WrittenPath, res.ManagedLines, res.UserLines)
}

// runAddAttributes is a Phase 3 stub so dispatchAdd compiles.
func runAddAttributes(_ []string) {
	fmt.Fprintln(os.Stderr, "add attributes: not yet implemented (Phase 3)")
	os.Exit(1)
}

// runAddLFSInstall is a Phase 3 stub so dispatchAdd compiles.
func runAddLFSInstall(_ []string) {
	fmt.Fprintln(os.Stderr, "add lfs-install: not yet implemented (Phase 3)")
	os.Exit(1)
}
