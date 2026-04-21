package cloner

import (
	"github.com/alimtvnetwork/gitmap-v6/gitmap/gitutil"
	"github.com/alimtvnetwork/gitmap-v6/gitmap/model"
)

// cloneStrategy describes how a single repo should be cloned given the
// confidence we have in its recorded branch. The strategy is a pure
// function of the ScanRecord's Branch + BranchSource fields — no I/O,
// so it is trivially unit-testable.
type cloneStrategy struct {
	// useBranchFlag controls whether `git clone -b <branch>` is used.
	// When false we let the remote's HEAD pick the initial branch,
	// which is the only safe choice when our recorded branch is
	// missing or untrustworthy.
	useBranchFlag bool
	// branch is the branch name to pass to `-b` when useBranchFlag is
	// true. Empty otherwise.
	branch string
	// reason is a short human-readable label describing why this
	// strategy was selected. Surfaced in verbose / debug output and in
	// CloneResult.Notes so users can audit decisions after the fact.
	reason string
}

// pickCloneStrategy maps (branch, branchSource) → cloneStrategy.
//
// Decision matrix:
//
//	BranchSource     | Branch present? | Strategy
//	-----------------+-----------------+------------------------------------
//	HEAD             | yes             | -b <branch>           (trusted)
//	remote-tracking  | yes             | -b <branch>           (trusted)
//	default          | yes             | -b <branch>           (trusted)
//	detached         | *               | no -b, follow remote HEAD
//	unknown          | *               | no -b, follow remote HEAD
//	(any)            | no              | no -b, follow remote HEAD
//
// "Detached" is treated as untrustworthy because the recorded value is
// often a commit SHA or the literal string "HEAD", neither of which is
// a valid argument for `git clone -b`. Letting the remote's HEAD pick
// the initial branch is always safe — the user can switch later.
func pickCloneStrategy(rec model.ScanRecord) cloneStrategy {
	if len(rec.Branch) == 0 {
		return cloneStrategy{
			useBranchFlag: false,
			reason:        "no recorded branch — using remote HEAD",
		}
	}

	switch rec.BranchSource {
	case gitutil.BranchSourceHEAD,
		gitutil.BranchSourceRemoteTracking,
		gitutil.BranchSourceDefault:
		return cloneStrategy{
			useBranchFlag: true,
			branch:        rec.Branch,
			reason:        "trusted source: " + rec.BranchSource,
		}
	case gitutil.BranchSourceDetached:
		return cloneStrategy{
			useBranchFlag: false,
			reason:        "detached HEAD — using remote HEAD",
		}
	case gitutil.BranchSourceUnknown, "":
		return cloneStrategy{
			useBranchFlag: false,
			reason:        "unknown branch source — using remote HEAD",
		}
	default:
		// Forward-compat: an unfamiliar source is treated as
		// untrusted. Better to land on remote HEAD than to fail
		// with `Remote branch X not found`.
		return cloneStrategy{
			useBranchFlag: false,
			reason:        "unrecognized branch source: " + rec.BranchSource,
		}
	}
}
