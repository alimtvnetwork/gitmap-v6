package cloner

import (
	"testing"

	"github.com/alimtvnetwork/gitmap-v6/gitmap/gitutil"
	"github.com/alimtvnetwork/gitmap-v6/gitmap/model"
)

// TestPickCloneStrategy locks in the BranchSource → strategy mapping.
// If you intentionally change a row, update the table here in the same
// PR — silent behavior changes in clone selection are exactly the kind
// of regression this test exists to catch.
func TestPickCloneStrategy(t *testing.T) {
	cases := []struct {
		name          string
		branch        string
		source        string
		wantUseBranch bool
		wantBranch    string
	}{
		{"head trusted", "main", gitutil.BranchSourceHEAD, true, "main"},
		{"remote-tracking trusted", "develop", gitutil.BranchSourceRemoteTracking, true, "develop"},
		{"default trusted", "main", gitutil.BranchSourceDefault, true, "main"},
		{"detached untrusted", "HEAD", gitutil.BranchSourceDetached, false, ""},
		{"detached with sha untrusted", "abc123", gitutil.BranchSourceDetached, false, ""},
		{"unknown untrusted", "main", gitutil.BranchSourceUnknown, false, ""},
		{"empty source untrusted", "main", "", false, ""},
		{"unfamiliar source untrusted", "main", "future-source", false, ""},
		{"missing branch overrides trusted source", "", gitutil.BranchSourceHEAD, false, ""},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			got := pickCloneStrategy(model.ScanRecord{
				Branch: tc.branch, BranchSource: tc.source,
			})
			if got.useBranchFlag != tc.wantUseBranch {
				t.Errorf("useBranchFlag = %v, want %v (reason: %s)",
					got.useBranchFlag, tc.wantUseBranch, got.reason)
			}
			if got.branch != tc.wantBranch {
				t.Errorf("branch = %q, want %q", got.branch, tc.wantBranch)
			}
			if len(got.reason) == 0 {
				t.Errorf("reason should never be empty")
			}
		})
	}
}
