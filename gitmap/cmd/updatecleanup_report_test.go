package cmd

import (
	"errors"
	"io/fs"
	"strings"
	"testing"
)

func TestClassifyRemoveError(t *testing.T) {
	tests := []struct {
		name       string
		err        error
		wantStatus cleanupArtifactStatus
		reasonHas  string
	}{
		{"nil", nil, cleanupStatusRemoved, "deleted"},
		{"missing", fs.ErrNotExist, cleanupStatusMissing, "already gone"},
		{"permission", fs.ErrPermission, cleanupStatusLocked, "permission denied"},
		{"generic", errors.New("sharing violation"), cleanupStatusLocked, "OS refused"},
	}
	for _, tt := range tests {
		gotStatus, gotReason := classifyRemoveError(tt.err)
		if gotStatus != tt.wantStatus {
			t.Errorf("%s: status = %q, want %q", tt.name, gotStatus, tt.wantStatus)
		}
		if !strings.Contains(gotReason, tt.reasonHas) {
			t.Errorf("%s: reason = %q, want substring %q", tt.name, gotReason, tt.reasonHas)
		}
	}
}

func TestCleanupReportSummaryCounts(t *testing.T) {
	r := newCleanupReport()
	// record bypasses printing in tests by using stdout — acceptable for unit scope.
	r.results = append(r.results,
		cleanupResult{Path: "a", Kind: cleanupKindTemp, Status: cleanupStatusRemoved},
		cleanupResult{Path: "b", Kind: cleanupKindBackup, Status: cleanupStatusLocked},
		cleanupResult{Path: "c", Kind: cleanupKindTemp, Status: cleanupStatusMissing},
		cleanupResult{Path: "d", Kind: cleanupKindSwapDir, Status: cleanupStatusGlobError},
	)
	if got := r.removedCount(); got != 1 {
		t.Errorf("removedCount = %d, want 1", got)
	}
	if got := r.errorCount(); got != 2 {
		t.Errorf("errorCount = %d, want 2 (locked + glob-error)", got)
	}
}

func TestCleanupStatusSortRankOrder(t *testing.T) {
	// Removed must come before Locked, which must come before Missing.
	if cleanupStatusSortRank(cleanupStatusRemoved) >= cleanupStatusSortRank(cleanupStatusLocked) {
		t.Fatal("removed should sort before locked")
	}
	if cleanupStatusSortRank(cleanupStatusLocked) >= cleanupStatusSortRank(cleanupStatusMissing) {
		t.Fatal("locked should sort before missing")
	}
}
