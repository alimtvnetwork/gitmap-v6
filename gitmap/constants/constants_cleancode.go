package constants

// Clean-code / coding-guidelines installer.
// One-liner published at the URL below installs the alimtvnetwork
// coding-guidelines (v15) into the current directory via PowerShell IRM | IEX.
//
// The four CLI aliases (clean-code, code-guide, cg, cc) all dispatch to the
// same flow — see gitmap/cmd/installcleancode.go.
const (
	DefaultCleanCodeURL = "https://raw.githubusercontent.com/alimtvnetwork/coding-guidelines-v15/main/install.ps1"
)

// Clean-code installer messages.
const (
	MsgCleanCodeRunning  = "  Installing coding guidelines from %s\n"
	MsgCleanCodeDone     = "  OK Coding guidelines installed.\n"
	MsgCleanCodeNoPwsh   = "  ✗ PowerShell not found on PATH. Install PowerShell 7+ or run the one-liner manually:\n      irm %s | iex\n"
	ErrCleanCodeFailed   = "  ✗ Coding guidelines install failed: %v\n"
	MsgCleanCodeNonWin   = "  Note: this installer is PowerShell-based; on non-Windows it requires PowerShell 7+ (pwsh).\n"
)
