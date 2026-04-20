package cmd

import (
	"flag"
	"fmt"
	"os"

	"github.com/user/gitmap/constants"
)

// runInstall handles the "install" command.
func runInstall(args []string) {
	checkHelp("install", args)

	fs := flag.NewFlagSet("install", flag.ExitOnError)

	var manager, version string
	var verbose, dryRun, check, list, yes bool

	fs.StringVar(&manager, constants.FlagInstallManager, "", constants.FlagDescInstallManager)
	fs.StringVar(&version, constants.FlagInstallVersion, "", constants.FlagDescInstallVersion)
	fs.BoolVar(&verbose, constants.FlagInstallVerbose, false, constants.FlagDescInstallVerbose)
	fs.BoolVar(&dryRun, constants.FlagInstallDryRun, false, constants.FlagDescInstallDryRun)
	fs.BoolVar(&check, constants.FlagInstallCheck, false, constants.FlagDescInstallCheck)
	fs.BoolVar(&list, constants.FlagInstallList, false, constants.FlagDescInstallList)
	fs.BoolVar(&yes, constants.FlagInstallYes, false, constants.FlagDescInstallYes)
	fs.BoolVar(&yes, "y", false, constants.FlagDescInstallYes)

	reordered := reorderFlagsBeforeArgs(args)
	fs.Parse(reordered)

	if list {
		printInstallList()

		return
	}

	tool := fs.Arg(0)
	if tool == "" {
		fmt.Fprint(os.Stderr, constants.ErrInstallToolRequired)
		os.Exit(1)
	}

	validateToolName(tool)

	opts := installOptions{
		Tool:    tool,
		Manager: manager,
		Version: version,
		Verbose: verbose,
		DryRun:  dryRun,
		Check:   check,
		Yes:     yes,
	}

	executeInstall(opts)
}

// installOptions holds parsed install flags.
type installOptions struct {
	Tool    string
	Manager string
	Version string
	Verbose bool
	DryRun  bool
	Check   bool
	Yes     bool
}

// printInstallList prints all supported tools.
func printInstallList() {
	fmt.Print(constants.MsgInstallListHeader)

	for tool, desc := range constants.InstallToolDescriptions {
		fmt.Printf(constants.MsgInstallListRow, tool, desc)
	}
}

// validateToolName checks if the tool is supported.
func validateToolName(tool string) {
	_, exists := constants.InstallToolDescriptions[tool]
	if exists {
		return
	}

	fmt.Fprintf(os.Stderr, constants.ErrInstallUnknownTool, tool)
	os.Exit(1)
}

// executeInstall runs the install flow for a tool.
func executeInstall(opts installOptions) {
	if opts.Tool == constants.ToolScripts {
		runInstallScripts()

		return
	}

	if opts.Tool == constants.ToolNppSettings {
		runNppSettingsOnly()

		return
	}

	if opts.Tool == constants.ToolVSCodeSync {
		runVSCodeSettingsOnly()

		return
	}

	if opts.Tool == constants.ToolOBSSync {
		runOBSSettingsOnly()

		return
	}

	if opts.Tool == constants.ToolWTSync {
		runWTSettingsOnly()

		return
	}

	if opts.Tool == constants.ToolVSCodeCtx {
		runVSCodeContextMenu()

		return
	}

	if opts.Tool == constants.ToolPwshCtx {
		runPwshContextMenu()

		return
	}

	if opts.Tool == constants.ToolAllDevTools {
		runAllDevTools(opts)

		return
	}

	originalTool := opts.Tool
	installName := resolveNppInstallName(opts.Tool)

	fmt.Printf(constants.MsgInstallChecking, installName)

	existingVersion := detectInstalledVersion(installName)
	if existingVersion != "" {
		fmt.Printf(constants.MsgInstallFound, installName, existingVersion)

		return
	}

	if opts.Check {
		fmt.Printf(constants.MsgInstallNotFound, installName)

		return
	}

	opts.Tool = installName
	manager := resolvePackageManager(opts.Manager)

	// Show version and manager info.
	if opts.Version != "" {
		fmt.Printf(constants.MsgInstallVersion, opts.Version)
	} else {
		fmt.Print(constants.MsgInstallVersionLabel)
	}

	fmt.Printf(constants.MsgInstallManager, manager)

	// Prompt for confirmation unless -y is set.
	if !opts.Yes && !opts.DryRun {
		if !confirmInstall(installName, opts.Version, manager) {
			fmt.Print(constants.MsgInstallAborted)

			return
		}
	}

	installTool(opts)

	// Sync settings for "npp" but not for "install-npp".
	if originalTool == constants.ToolNpp {
		runNppSettings()
	}
	if originalTool == constants.ToolNppInstall {
		fmt.Print(constants.MsgInstallNppSkipSet)
	}
}

// confirmInstall prompts the user for install confirmation.
func confirmInstall(tool, version, manager string) bool {
	if version != "" {
		fmt.Printf(constants.MsgInstallPrompt, tool, version, manager)
	} else {
		fmt.Printf(constants.MsgInstallPromptNoVer, tool, manager)
	}

	var answer string

	fmt.Scanln(&answer)

	return answer == "y" || answer == "Y"
}
