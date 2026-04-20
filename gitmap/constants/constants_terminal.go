package constants

// ANSI color codes.
const (
	ColorReset  = "\033[0m"
	ColorGreen  = "\033[32m"
	ColorRed    = "\033[31m"
	ColorYellow = "\033[33m"
	ColorCyan   = "\033[36m"
	ColorWhite  = "\033[97m"
	ColorDim    = "\033[90m"
)

// Status banner box-drawing.
const (
	StatusBannerTop    = "╔══════════════════════════════════════╗"
	StatusBannerTitle  = "║         gitmap status                ║"
	StatusBannerBottom = "╚══════════════════════════════════════╝"
	StatusRepoCountFmt = "%d repos from .gitmap/output/gitmap.json"
)

// Status indicator strings.
const (
	StatusIconClean    = "✓ clean"
	StatusIconDirty    = "● dirty"
	StatusDash         = "—"
	StatusSyncDash     = "  —"
	StatusStashFmt     = "📦 %d"
	StatusSyncUpFmt    = "↑%d"
	StatusSyncDownFmt  = "↓%d"
	StatusSyncBothFmt  = "↑%d ↓%d"
	StatusStagedFmt    = "+%d"
	StatusModifiedFmt  = "~%d"
	StatusUntrackedFmt = "?%d"
)

// Status row format strings.
const (
	StatusRowFmt     = "  %-22s %s  %s  %s  %s  %s\n"
	StatusMissingFmt = "  %s%-22s %s⊘ not found%s\n"
	StatusHeaderFmt  = "  %s%-22s %-12s %-8s %-10s %-8s %-6s%s\n"
)

// Status table column headers.
var StatusTableColumns = []string{
	"REPO", "STATUS", "SYNC", "BRANCH", "STASH", "FILES",
}

// Summary format strings.
const (
	SummaryJoinSep      = " · "
	SummaryReposFmt     = "%d repos"
	SummaryCleanFmt     = "%d clean"
	SummaryDirtyFmt     = "%d dirty"
	SummaryAheadFmt     = "%d ahead"
	SummaryBehindFmt    = "%d behind"
	SummaryStashedFmt   = "%d stashed"
	SummaryMissingFmt   = "%d missing"
	SummarySucceededFmt = "%d succeeded"
	SummaryFailedFmt    = "%d failed"
	StatusFileCountSep  = " "
	TruncateEllipsis    = "…"
)

// Setup banner box-drawing.
const (
	SetupBannerTop     = "╔══════════════════════════════════════╗"
	SetupBannerTitle   = "║         gitmap setup                 ║"
	SetupBannerBottom  = "╚══════════════════════════════════════╝"
	SetupDryRunFmt     = "[DRY RUN] No changes will be made"
	SetupAppliedFmt    = "✓ %d settings applied"
	SetupSkippedFmt    = "⊘ %d settings unchanged"
	SetupFailedFmt     = "✗ %d settings failed"
	SetupErrorEntryFmt = "- %s"
)

// Changelog entry format strings.
const (
	ChangelogVersionFmt = "\n%s"
	ChangelogNoteFmt    = "  - %s"
)

// Exec banner box-drawing.
const (
	ExecBannerTop     = "╔══════════════════════════════════════╗"
	ExecBannerTitle   = "║           gitmap exec                ║"
	ExecBannerBottom  = "╚══════════════════════════════════════╝"
	ExecCommandFmt    = "Command: git %s"
	ExecRepoCountFmt  = "%d repos from .gitmap/output/gitmap.json"
	ExecSuccessFmt    = "  %s✓ %-22s%s\n"
	ExecFailFmt       = "  %s✗ %-22s%s\n"
	ExecMissingFmt    = "  %s⊘ %-22s %snot found%s\n"
	ExecOutputLineFmt = "    %s%s%s\n"
	ExecSummaryRule   = "──────────────────────────────────────────────────"
)

// Terminal output sections.
const (
	TermBannerTop    = "  ╔══════════════════════════════════════╗"
	TermBannerTitle  = "  ║            gitmap v%s               ║"
	TermBannerBottom = "  ╚══════════════════════════════════════╝"
	TermFoundFmt     = "  ✓ Found %d repositories"
	TermReposHeader  = "  ■ Repositories"
	TermTreeHeader   = "  ■ Folder Structure"
	TermCloneHeader  = "  ■ How to Clone on Another Machine"
	TermSeparator    = "  ──────────────────────────────────────────"
	TermTableRule    = "──────────────────────────────────────────────────────────────────────"
)

// Terminal repo entry formats.
const (
	TermRepoIcon  = "  📦 %s\n"
	TermPathLine  = "     Path:  %s\n"
	TermCloneLine = "     Clone: %s\n"
)

// Terminal clone help text.
const (
	TermCloneStep1     = "  1. Copy the output files to the target machine:"
	TermCloneCmd1      = "     .gitmap/output/gitmap.json  (or .csv / .txt)"
	TermCloneStep2     = "  2. Clone via JSON (shorthand):"
	TermCloneCmd2      = "     gitmap clone json --target-dir ./projects"
	TermCloneCmd2Alt   = "     gitmap c json               # alias"
	TermCloneStep3     = "  3. Clone via CSV (shorthand):"
	TermCloneCmd3      = "     gitmap clone csv --target-dir ./projects"
	TermCloneCmd3Alt   = "     gitmap c csv                # alias"
	TermCloneStep3t    = "  4. Clone via text (shorthand):"
	TermCloneCmd3t     = "     gitmap clone text --target-dir ./projects"
	TermCloneCmd3tAlt  = "     gitmap c text               # alias"
	TermCloneStep3b    = "  5. Or specify a file path directly:"
	TermCloneCmd3b     = "     gitmap clone .gitmap/output/gitmap.json --target-dir ./projects"
	TermCloneStep4     = "  6. Or run the PowerShell script directly:"
	TermCloneCmd4HTTPS = "     .\\direct-clone.ps1       # HTTPS clone commands"
	TermCloneCmd4SSH   = "     .\\direct-clone-ssh.ps1   # SSH clone commands"
	TermCloneStep5     = "  7. Full clone script with progress & error handling:"
	TermCloneCmd5      = "     .\\clone.ps1 -TargetDir .\\projects"
	TermCloneStep6     = "  8. Sync repos to GitHub Desktop:"
	TermCloneCmd6      = "     gitmap desktop-sync         # or: gitmap ds"
	TermCloneNote      = "  Note: safe-pull is auto-enabled when existing repos are detected."
)

// Folder structure Markdown.
const (
	StructureTitle       = "# Folder Structure"
	StructureDescription = "Git repositories discovered by gitmap."
	StructureRepoFmt     = "📦 **%s** (`%s`) — %s"
	TreeBranch           = "├──"
	TreeCorner           = "└──"
	TreePipe             = "│   "
	TreeSpace            = "    "
)

// CSV headers.
var ScanCSVHeaders = []string{
	"repoName", "httpsUrl", "sshUrl", "branch",
	"relativePath", "absolutePath", "cloneInstruction", "notes",
}

var LatestBranchCSVHeaders = []string{
	"branch", "remote", "sha", "commitDate", "subject", "ref",
}

// Latest-branch terminal display format strings.
const (
	LBTermLatestFmt  = "  Latest branch: %s\n"
	LBTermRemoteFmt  = "  Remote:        %s\n"
	LBTermSHAFmt     = "  SHA:           %s\n"
	LBTermDateFmt    = "  Commit date:   %s\n"
	LBTermSubjectFmt = "  Subject:       %s\n"
	LBTermRefFmt     = "  Ref:           %s\n"
	LBTermTopHdrFmt  = "  Top %d most recently updated remote branches (%s):\n"
	LBTermRowFmt     = "  %-30s %-30s %-9s %s\n"
)

// Latest-branch terminal table header columns.
var LatestBranchTableColumns = []string{
	"DATE", "BRANCH", "SHA", "SUBJECT",
}
