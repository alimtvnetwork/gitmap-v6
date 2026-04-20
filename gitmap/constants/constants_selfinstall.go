package constants

// Self-install / self-uninstall command messages, errors, and defaults.
//
// These constants back the `gitmap self-install` and `gitmap self-uninstall`
// commands, which manage the gitmap binary itself (as opposed to the
// `install`/`uninstall` commands that manage third-party tools).
//
// Spec: spec/01-app/90-self-install-uninstall.md

// Default install directories per platform.
const (
	SelfInstallDefaultWindows = "D:\\gitmap"
	SelfInstallDefaultUnix    = ".local/bin/gitmap" // joined under $HOME at runtime
)

// Embedded script names.
const (
	SelfInstallScriptPwsh = "install.ps1"
	SelfInstallScriptBash = "install.sh"
)

// Remote installer URLs (fallback when embedded scripts are missing).
const (
	SelfInstallRemotePwsh = "https://raw.githubusercontent.com/" +
		"alimtvnetwork/gitmap-v4/main/gitmap/scripts/install.ps1"
	SelfInstallRemoteBash = "https://raw.githubusercontent.com/" +
		"alimtvnetwork/gitmap-v4/main/gitmap/scripts/install.sh"
)

// Self-install messages.
const (
	MsgSelfInstallHeader     = "\n  gitmap self-install\n\n"
	MsgSelfInstallPrompt     = "  Install directory [%s]: "
	MsgSelfInstallUsing      = "  Using install directory: %s\n"
	MsgSelfInstallEmbedded   = "  Running embedded installer (%s)...\n"
	MsgSelfInstallRemote     = "  Embedded installer unavailable; downloading from %s\n"
	MsgSelfInstallDone       = "  ✓ Install completed.\n"
	MsgSelfInstallReminder   = "  Open a new terminal (or reload your profile) to pick up PATH changes.\n"
)

// Self-install errors.
const (
	ErrSelfInstallScriptWrite  = "Error: write installer to temp: %v\n"
	ErrSelfInstallScriptRun    = "Error: run installer: %v\n"
	ErrSelfInstallDownload     = "Error: download installer from %s: %v\n"
	ErrSelfInstallNoShell      = "Error: no supported shell found (need PowerShell on Windows or bash on Unix)\n"
	ErrSelfInstallReadStdin    = "Error: read install dir from stdin: %v\n"
)

// Self-uninstall messages.
const (
	MsgSelfUninstallHeader        = "\n  gitmap self-uninstall\n\n"
	MsgSelfUninstallTargets       = "  The following will be removed:\n"
	MsgSelfUninstallTargetBin     = "    - Binary + deploy dir: %s\n"
	MsgSelfUninstallTargetData    = "    - Data dir:            %s\n"
	MsgSelfUninstallTargetSnippet = "    - PATH snippet from:   %s\n"
	MsgSelfUninstallTargetCompl   = "    - Completion files in: %s\n"
	MsgSelfUninstallConfirmPrompt = "\n  Type 'yes' to proceed: "
	MsgSelfUninstallSkipBin       = "  ⚠ Could not resolve own binary location: %v\n"
	MsgSelfUninstallRemovedBin    = "  ✓ Removed binary: %s\n"
	MsgSelfUninstallRemovedDir    = "  ✓ Removed dir:    %s\n"
	MsgSelfUninstallSnippetGone   = "  ✓ PATH snippet removed from %s\n"
	MsgSelfUninstallSnippetMiss   = "  - No PATH snippet found in %s\n"
	MsgSelfUninstallDone          = "\n  ✓ gitmap has been uninstalled. Restart your terminal to clear $env:Path.\n\n"
	MsgSelfUninstallHandoffActive = "  Handing off to %s so the original binary can self-delete...\n"
)

// Self-uninstall errors.
const (
	ErrSelfUninstallNoConfirm   = "Error: refusing to run without --confirm or interactive 'yes'.\n"
	ErrSelfUninstallRemove      = "Error: remove %s: %v\n"
	ErrSelfUninstallSnippetRead = "Error: read profile %s: %v\n"
	ErrSelfUninstallSnippetWrite = "Error: rewrite profile %s: %v\n"
	ErrSelfUninstallHandoffCopy = "Error: create handoff copy: %v\n"
)

// Hidden runner subcommand for the self-uninstall handoff (lets the temp
// copy delete the original .exe on Windows where the running file is
// locked).
const CmdSelfUninstallRunner = "self-uninstall-runner" // gitmap:cmd skip

// Flag names shared by self-install / self-uninstall.
const (
	FlagSelfDir          = "--dir"
	FlagSelfYes          = "--yes"
	FlagSelfConfirm      = "--confirm"
	FlagSelfKeepData     = "--keep-data"
	FlagSelfKeepSnippet  = "--keep-snippet"
	FlagSelfFromVersion  = "--version"
)

// Flag descriptions.
const (
	FlagDescSelfDir         = "Install directory (prompted with default if omitted)"
	FlagDescSelfYes         = "Skip the install-directory prompt and accept the default"
	FlagDescSelfConfirm     = "Required for self-uninstall to actually remove files"
	FlagDescSelfKeepData    = "Preserve the .gitmap data dir during self-uninstall"
	FlagDescSelfKeepSnippet = "Leave the PATH snippet in shell profile during self-uninstall"
	FlagDescSelfFromVersion = "Pin a specific gitmap version to install (e.g. v3.0.0)"
)
