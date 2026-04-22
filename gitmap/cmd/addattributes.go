package cmd

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/alimtvnetwork/gitmap-v6/gitmap/constants"
	"github.com/alimtvnetwork/gitmap-v6/gitmap/templates"
)

// runAddAttributes handles `gitmap add attributes [langs...]`.
func runAddAttributes(args []string) {
	if len(args) == 0 {
		fmt.Fprint(os.Stderr, constants.ErrAddNoLangs)
		os.Exit(1)
	}

	target, err := targetInCwd(constants.TemplateExtAttributes)
	if err != nil {
		fmt.Fprintf(os.Stderr, constants.ErrAddMerge, err)
		os.Exit(1)
	}

	res, err := templates.Merge(templates.MergeOptions{
		TargetPath: target,
		Kind:       constants.TemplateKindAttributes,
		Langs:      args,
	})
	if err != nil {
		fmt.Fprintf(os.Stderr, constants.ErrAddMerge, err)
		os.Exit(1)
	}

	if !res.Changed {
		fmt.Printf(constants.MsgAddAttributesNoop, res.WrittenPath)

		return
	}
	fmt.Printf(constants.MsgAddAttributesWritten, res.WrittenPath, res.ManagedLines, res.UserLines)
}

// runAddLFSInstall runs `git lfs install --local` and then merges the curated
// lfs/common.gitattributes into ./.gitattributes.
func runAddLFSInstall(_ []string) {
	if err := runGitLFSInstall(); err != nil {
		fmt.Fprintf(os.Stderr, constants.ErrAddLFSInstall, err)
		os.Exit(1)
	}

	target, err := targetInCwd(constants.TemplateExtAttributes)
	if err != nil {
		fmt.Fprintf(os.Stderr, constants.ErrAddMerge, err)
		os.Exit(1)
	}

	if _, err := templates.Merge(templates.MergeOptions{
		TargetPath: target,
		Kind:       constants.TemplateKindLFS,
		Langs:      nil, // common is implicit
	}); err != nil {
		fmt.Fprintf(os.Stderr, constants.ErrAddMerge, err)
		os.Exit(1)
	}

	fmt.Printf(constants.MsgAddLFSDone, target)
}

// targetInCwd returns the absolute path of `<cwd>/<basename><ext>`. The
// basename is empty so callers get e.g. ".gitattributes".
func targetInCwd(ext string) (string, error) {
	cwd, err := os.Getwd()
	if err != nil {
		return "", err
	}

	return filepath.Join(cwd, ext), nil
}

// runGitLFSInstall executes `git lfs install --local` in the current dir.
// Stdout/stderr are streamed so the user sees the LFS hook messages.
func runGitLFSInstall() error {
	cmd := exec.Command("git", "lfs", "install", "--local")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	return cmd.Run()
}
