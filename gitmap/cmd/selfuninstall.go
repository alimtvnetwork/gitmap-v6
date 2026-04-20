package cmd

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/user/gitmap/constants"
)

// selfUninstallOpts holds parsed flags for self-uninstall.
type selfUninstallOpts struct {
	Confirm     bool
	KeepData    bool
	KeepSnippet bool
}

// runSelfUninstall is the entry point for `gitmap self-uninstall`.
// On Windows the running .exe is locked, so we copy ourselves to a temp
// path and re-exec the hidden self-uninstall-runner from there.
func runSelfUninstall(args []string) {
	checkHelp(constants.CmdSelfUninstall, args)
	opts := parseSelfUninstallFlags(args)
	if !opts.Confirm && !confirmSelfUninstall() {
		fmt.Fprint(os.Stderr, constants.ErrSelfUninstallNoConfirm)
		os.Exit(1)
	}
	if shouldHandoffSelfUninstall() {
		handoffSelfUninstall(opts, args)
		return
	}
	executeSelfUninstall(opts)
}

// parseSelfUninstallFlags reads --confirm / --keep-data / --keep-snippet.
func parseSelfUninstallFlags(args []string) selfUninstallOpts {
	fs := flag.NewFlagSet(constants.CmdSelfUninstall, flag.ExitOnError)
	opts := selfUninstallOpts{}
	fs.BoolVar(&opts.Confirm, "confirm", false, constants.FlagDescSelfConfirm)
	fs.BoolVar(&opts.KeepData, "keep-data", false, constants.FlagDescSelfKeepData)
	fs.BoolVar(&opts.KeepSnippet, "keep-snippet", false, constants.FlagDescSelfKeepSnippet)
	fs.Parse(reorderFlagsBeforeArgs(args))

	return opts
}

// confirmSelfUninstall prints the target list and prompts for "yes".
func confirmSelfUninstall() bool {
	printSelfUninstallTargets()
	fmt.Print(constants.MsgSelfUninstallConfirmPrompt)
	var answer string
	if _, err := fmt.Scanln(&answer); err != nil {
		return false
	}

	return answer == "yes"
}

// printSelfUninstallTargets prints what self-uninstall will remove.
func printSelfUninstallTargets() {
	fmt.Print(constants.MsgSelfUninstallHeader)
	fmt.Print(constants.MsgSelfUninstallTargets)
	fmt.Printf(constants.MsgSelfUninstallTargetBin, selfDeployDir())
	fmt.Printf(constants.MsgSelfUninstallTargetData, selfDataDir())
	fmt.Printf(constants.MsgSelfUninstallTargetSnippet, defaultProfileForOS())
	fmt.Printf(constants.MsgSelfUninstallTargetCompl, selfDeployDir())
}

// executeSelfUninstall removes each target the user did not opt out of.
func executeSelfUninstall(opts selfUninstallOpts) {
	if !opts.KeepSnippet {
		removeProfileSnippet(defaultProfileForOS())
	}
	removeCompletionFiles(selfDeployDir())
	if !opts.KeepData {
		removePathBestEffort(selfDataDir())
	}
	removeDeployArtifacts(selfDeployDir())
	fmt.Print(constants.MsgSelfUninstallDone)
}

// shouldHandoffSelfUninstall reports whether the running binary lives
// inside the directory we are about to delete (Windows only — on Unix
// we can unlink an open file safely).
func shouldHandoffSelfUninstall() bool {
	if runtime.GOOS != "windows" {
		return false
	}
	self, err := os.Executable()
	if err != nil {
		return false
	}
	deploy := selfDeployDir()
	if len(deploy) == 0 {
		return false
	}

	return strings.HasPrefix(filepath.Clean(self), filepath.Clean(deploy))
}

// runSelfUninstallRunner is the hidden command run by the temp-copy
// handoff. It performs the actual removal then deletes itself.
func runSelfUninstallRunner() {
	opts := parseSelfUninstallFlags(os.Args[2:])
	executeSelfUninstall(opts)
	scheduleSelfDelete()
}
