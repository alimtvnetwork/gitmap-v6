package constants

// Go release asset constants.
const (
	AssetsStagingDir = ".gitmap/release-assets"
	GitHubTokenEnv   = "GITHUB_TOKEN"
)

// Asset flag descriptions.
const (
	FlagDescBin         = "Cross-compile Go binaries and include in release assets"
	FlagDescTargets     = "Comma-separated cross-compile targets (e.g. windows/amd64,linux/arm64)"
	FlagDescListTargets = "Print resolved target matrix and exit"
)

// Asset help text.
const (
	HelpBin         = "  --bin, -b           Cross-compile Go binaries and include in release assets"
	HelpTargets     = "  --targets <list>    Cross-compile targets: windows/amd64,linux/arm64"
	HelpListTargets = "  --list-targets      Print resolved target matrix and exit"
)

// List-targets messages.
const (
	MsgListTargetsHeader = "Release targets (%d):\n"
	MsgListTargetsSource = "Source: %s\n\n"
	MsgListTargetsRow    = "  %s/%s\n"
)

// Asset messages.
const (
	MsgAssetDetected     = "  → Detected Go project: %s\n"
	MsgAssetCrossCompile = "\n  Cross-compiling %d target(s)...\n"
	MsgAssetBuilt        = "  ✓ Built %s (%s/%s)\n"
	MsgAssetBuildSummary = "  → Built %d/%d binaries successfully\n"
	MsgAssetUploaded     = "  ✓ Uploaded %s\n"
	MsgAssetUploadStart  = "\n  Uploading %d asset(s) to GitHub...\n"
	MsgAssetSkipped      = ""
	MsgAssetNoMain       = "  → No buildable main package found, skipping binaries\n"
	MsgAssetNoGoProject  = ""
	MsgAssetStagingClean = "  ✓ Cleaned up staging directory\n"
)

// Asset dry-run messages.
const (
	MsgAssetDryRunHeader = "  [dry-run] Would cross-compile %d binaries:\n"
	MsgAssetDryRunBinary = "    → %s\n"
	MsgAssetDryRunUpload = "  [dry-run] Would upload %d assets\n"
)

// Asset error messages — Code Red: all file errors include exact path and reason.
const (
	ErrAssetBuildFailed = "Error: build failed for %s/%s: %s (operation: compile)\n"
	ErrAssetUploadFinal = "Error: upload failed for asset %s: %v (operation: upload)\n"
	ErrAssetNoToken     = "Error: GITHUB_TOKEN not set — skipping asset upload (reason: environment variable not set)\n"
	ErrAssetRemoteParse = "Error: could not parse remote origin: %v (operation: resolve)\n"
)

// Retry constants.
const (
	RetryMaxAttempts   = 3
	RetryBaseDelayMs   = 1000
	RetryBackoffFactor = 2
)

// Retry HTTP status codes.
const (
	HTTPTooManyRequests = 429
	HTTPServerErrorMin  = 500
)

// Retry messages.
const (
	MsgRetryAttempt = "  ⟳ Retry %d/%d for %s (waiting %s)...\n"
	MsgRetrySuccess = "  ✓ Uploaded %s (attempt %d)\n"
)
