package cmd

import (
	"flag"

	"github.com/alimtvnetwork/gitmap-v6/gitmap/constants"
)

// ScanFlags bundles every parsed scan-command flag so callers don't have
// to thread a long positional return list through helper functions.
type ScanFlags struct {
	Dir               string
	ConfigPath        string
	Mode              string
	Output            string
	OutFile           string
	OutputPath        string
	GHDesktop         bool
	OpenFolder        bool
	Quiet             bool
	NoVSCodeSync      bool
	NoAutoTags        bool
	Workers           int
	BranchSourceDebug bool
}

// ParseScanFlags parses flags for the scan command into a ScanFlags
// struct. New flags are added as fields here rather than as additional
// positional return values, keeping the call sites stable as the flag
// surface grows.
func ParseScanFlags(args []string) ScanFlags {
	fs := flag.NewFlagSet(constants.CmdScan, flag.ExitOnError)
	cfgFlag := fs.String("config", constants.DefaultConfigPath, constants.FlagDescConfig)
	modeFlag := fs.String("mode", "", constants.FlagDescMode)
	outputFlag := fs.String("output", "", constants.FlagDescOutput)
	outFileFlag := fs.String("out-file", "", constants.FlagDescOutFile)
	outputPathFlag := fs.String("output-path", "", constants.FlagDescOutputPath)
	ghDesktopFlag, openFlag, quietFlag := registerScanBoolFlags(fs)
	noVSCodeSyncFlag := fs.Bool(constants.FlagNoVSCodeSync, false, constants.FlagDescNoVSCodeSync)
	noAutoTagsFlag := fs.Bool(constants.FlagNoAutoTags, false, constants.FlagDescNoAutoTags)
	workersFlag := fs.Int(constants.FlagScanWorkers, constants.DefaultScanWorkers, constants.FlagDescScanWorkers)
	branchSrcDebugFlag := fs.Bool(constants.FlagBranchSourceDebug, false, constants.FlagDescBranchSourceDebug)
	fs.Parse(args)

	return ScanFlags{
		Dir:               resolveScanDir(fs),
		ConfigPath:        *cfgFlag,
		Mode:              *modeFlag,
		Output:            *outputFlag,
		OutFile:           *outFileFlag,
		OutputPath:        *outputPathFlag,
		GHDesktop:         *ghDesktopFlag,
		OpenFolder:        *openFlag,
		Quiet:             *quietFlag,
		NoVSCodeSync:      *noVSCodeSyncFlag,
		NoAutoTags:        *noAutoTagsFlag,
		Workers:           *workersFlag,
		BranchSourceDebug: *branchSrcDebugFlag,
	}
}

// registerScanBoolFlags registers boolean flags for the scan command.
func registerScanBoolFlags(fs *flag.FlagSet) (*bool, *bool, *bool) {
	ghDesktopFlag := fs.Bool("github-desktop", false, constants.FlagDescGHDesktop)
	openFlag := fs.Bool("open", false, constants.FlagDescOpen)
	quietFlag := fs.Bool("quiet", false, constants.FlagDescQuiet)

	return ghDesktopFlag, openFlag, quietFlag
}

// resolveScanDir returns the scan directory from positional args or default.
func resolveScanDir(fs *flag.FlagSet) string {
	if fs.NArg() > 0 {
		return fs.Arg(0)
	}

	return constants.DefaultDir
}

// parseCloneFlags parses flags for the clone command.
func parseCloneFlags(args []string) (source, folderName, targetDir, sshKeyName string, safePull, ghDesktop, noReplace, verbose bool) {
	fs := flag.NewFlagSet(constants.CmdClone, flag.ExitOnError)
	targetFlag := fs.String("target-dir", constants.DefaultDir, constants.FlagDescTargetDir)
	safePullFlag := fs.Bool("safe-pull", false, constants.FlagDescSafePull)
	ghDesktopFlag := fs.Bool("github-desktop", false, constants.FlagDescGHDesktop)
	verboseFlag := fs.Bool("verbose", false, constants.FlagDescVerbose)
	noReplaceFlag := fs.Bool("no-replace", false, constants.FlagDescCloneNoReplace)
	sshKeyFlag := fs.String("ssh-key", "", "SSH key name for clone")
	fs.StringVar(sshKeyFlag, "K", "", "SSH key name (short)")
	fs.Parse(args)

	source = resolveCloneSource(fs)
	folderName = resolveCloneFolderName(fs)

	return source, folderName, *targetFlag, *sshKeyFlag, *safePullFlag, *ghDesktopFlag, *noReplaceFlag, *verboseFlag
}

// resolveCloneSource returns the clone source from positional args.
func resolveCloneSource(fs *flag.FlagSet) string {
	if fs.NArg() > 0 {
		return fs.Arg(0)
	}

	return ""
}

// resolveCloneFolderName returns the optional folder name (second positional arg).
func resolveCloneFolderName(fs *flag.FlagSet) string {
	if fs.NArg() > 1 {
		return fs.Arg(1)
	}

	return ""
}
