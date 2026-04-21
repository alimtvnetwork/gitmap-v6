package constants

// Clone-next command messages.
const (
	MsgCloneNextCloning      = "Cloning %s into %s...\n"
	MsgCloneNextCreating     = "Creating GitHub repo %s...\n"
	MsgCloneNextCreated      = "✓ Created GitHub repo %s\n"
	MsgCloneNextDone         = "✓ Cloned %s\n"
	MsgCloneNextDesktop      = "✓ Registered %s with GitHub Desktop\n"
	MsgCloneNextRemovePrompt = "Remove current folder %s? [y/N] "
	MsgCloneNextRemoved      = "✓ Removed %s\n"
	MsgCloneNextMovedTo      = "→ Now in %s\n"
	MsgFlattenFallback       = "→ Falling back to versioned folder %s (current folder is locked by this shell)\n"
	MsgFlattenLockedHint     = "  Tip: 'cd ..' out of %s in your shell, then re-run to flatten.\n"
)

// Clone-next error and warning messages.
const (
	ErrCloneNextUsage         = "Usage: gitmap clone-next <v++|vN> [flags]"
	ErrCloneNextCwd           = "Error: cannot determine current directory: %v\n"
	ErrCloneNextNoRemote      = "Error: not a git repo or no remote origin: %v\n"
	ErrCloneNextBadVersion    = "Error: %v\n"
	ErrCloneNextExists        = "Error: target directory already exists: %s\nUse 'cd' to switch to it.\n"
	ErrCloneNextFailed        = "Error: clone failed for %s\n"
	ErrCloneNextRemoteParse   = "Error: cannot parse remote URL: %v\n"
	ErrCloneNextRepoCheck     = "Error: cannot check target repo: %v\n"
	ErrCloneNextRepoCreate    = "Error: cannot create GitHub repo %s: %v\n"
	WarnCloneNextRemoveFailed = "Warning: could not remove %s: %v\n"
)

// Clone-next flag descriptions.
const (
	FlagDescCloneNextDelete       = "Auto-remove current folder after clone"
	FlagDescCloneNextKeep         = "Keep current folder without prompting"
	FlagDescCloneNextNoDesktop    = "Skip GitHub Desktop registration"
	FlagDescCloneNextCreateRemote = "Create target GitHub repo if it does not exist (requires GITHUB_TOKEN)"
	FlagDescCloneNextCSV          = "Read repo paths from CSV file (one path per row, header optional)"
	FlagDescCloneNextAll          = "Walk current folder and run cn on every git repo found one level deep"
)

// Clone-next help strings for usage output.
const (
	HelpCloneNextFlags = "Clone-Next Flags:"
	HelpCNDelete       = "  --delete            Auto-remove current version folder after clone"
	HelpCNKeep         = "  --keep              Keep current folder without prompting for removal"
	HelpCNNoDesktop    = "  --no-desktop        Skip GitHub Desktop registration"
	HelpCNSSHKey       = "  --ssh-key, -K       SSH key name to use for clone"
	HelpCNVerbose      = "  --verbose           Show detailed clone-next output"
	HelpCNCreateRemote = "  --create-remote     Create target GitHub repo if missing (needs GITHUB_TOKEN)"
	HelpCNCSV          = "  --csv <path>        Batch mode: read repo list from CSV (one path per row)"
	HelpCNAll          = "  --all               Batch mode: cn every git repo one level under cwd"
)

// Clone-next batch mode messages and statuses (v3.42.0+).
const (
	MsgCloneNextBatchStart   = "→ Batch cn over %d repo(s)\n"
	MsgCloneNextBatchRepo    = "  • %s: %s -> %s\n"
	MsgCloneNextBatchSummary = "✓ Batch complete: %d ok, %d failed, %d skipped\n"
	MsgCloneNextBatchReport  = "  Report: %s\n"
	WarnCloneNextBatchReport = "Warning: could not write batch report: %v\n"
	ErrCloneNextBatchLoad    = "Error: could not load batch input: %v\n"

	BatchStatusOK      = "ok"
	BatchStatusFailed  = "failed"
	BatchStatusSkipped = "skipped"
)
