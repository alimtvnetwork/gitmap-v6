// Package cmd — extra cleanup passes for update-cleanup.
//
// These complement the pattern-based pass in updatecleanup_remove.go by
// targeting two artifact classes that don't fit the simple-glob model:
//  1. The obsolete v2.90.0 drive-root forwarding shim
//     (e.g. E:\gitmap.exe sitting at the literal drive root, NOT
//     inside a gitmap\ subfolder).
//  2. *.gitmap-tmp-* swap directories left by interrupted clones.
//
// Both passes record per-artifact results into the shared cleanupReport so
// users see the same structured "[<kind>] <path> — <status>: <reason>"
// output as the temp/backup passes.
package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"
)

const (
	driveRootShimMaxBytes = 5 * 1024 * 1024
	cloneSwapDirGlob      = "*.gitmap-tmp-*"
)

// cleanupDriveRootShim removes the obsolete drive-root forwarding shim if present.
// Safe-by-default: only deletes when the candidate sits at the literal drive root
// AND is under the size cap AND is not the active binary.
func cleanupDriveRootShim(ctx updateCleanupContext, report *cleanupReport) {
	if runtime.GOOS != "windows" {
		return
	}

	shimPath := resolveDriveRootShimPath(ctx.selfPath)
	if len(shimPath) == 0 {
		return
	}

	res, ok := evaluateDriveRootShim(shimPath, ctx.selfPath)
	if !ok {
		return
	}
	if res.Status != "" {
		report.record(res)

		return
	}

	report.record(removeCleanupCandidate(shimPath, filepath.Clean(shimPath), cleanupKindDriveShim))
}

// evaluateDriveRootShim runs the safety checks. Returns (result, true) when
// the candidate exists and warrants a recorded outcome (skipped or pending
// removal). Returns (zero, false) when the candidate doesn't exist at all
// — nothing to record because nothing to clean.
func evaluateDriveRootShim(shimPath, selfPath string) (cleanupResult, bool) {
	if normalizeCleanupPath(shimPath) == normalizeCleanupPath(selfPath) {
		return cleanupResult{}, false
	}

	parent := filepath.Dir(shimPath)
	if !isLiteralDriveRoot(parent) {
		return cleanupResult{
			Path:   shimPath,
			Kind:   cleanupKindDriveShim,
			Status: cleanupStatusSkippedNotInRoot,
			Reason: fmt.Sprintf("parent %q is not a literal drive root", parent),
		}, true
	}

	info, err := os.Stat(shimPath)
	if err != nil || info.IsDir() {
		return cleanupResult{}, false
	}

	if info.Size() > driveRootShimMaxBytes {
		return cleanupResult{
			Path:   shimPath,
			Kind:   cleanupKindDriveShim,
			Status: cleanupStatusSkippedTooLarge,
			Reason: fmt.Sprintf("size %d bytes exceeds %d-byte safety cap", info.Size(), driveRootShimMaxBytes),
		}, true
	}

	// Caller will perform the removal so the result reflects real OS state.
	return cleanupResult{}, true
}

// resolveDriveRootShimPath returns <drive>:\<binaryName> derived from selfPath.
func resolveDriveRootShimPath(selfPath string) string {
	if len(selfPath) == 0 {
		return ""
	}

	drive := filepath.VolumeName(selfPath)
	if len(drive) == 0 {
		return ""
	}

	binaryName := filepath.Base(selfPath)

	return filepath.Join(drive+`\`, binaryName)
}

// isLiteralDriveRoot returns true when path is the literal drive root (e.g. "E:\").
func isLiteralDriveRoot(path string) bool {
	clean := strings.TrimRight(path, `\/`)

	return len(clean) == 2 && clean[1] == ':'
}

// cleanupCloneSwapDirs removes *.gitmap-tmp-* directories left by interrupted clones.
// Scans every cleanup directory we already resolved.
func cleanupCloneSwapDirs(ctx updateCleanupContext, report *cleanupReport) {
	dirs := uniqueParentDirs(ctx.tempPatterns, ctx.backupPatterns)
	for _, dir := range dirs {
		removeCloneSwapDirsIn(dir, report)
	}
}

// uniqueParentDirs extracts the unique parent directories from glob patterns.
func uniqueParentDirs(patternGroups ...[]string) []string {
	seen := map[string]bool{}
	out := make([]string, 0)
	for _, group := range patternGroups {
		for _, pattern := range group {
			dir := filepath.Dir(pattern)
			key := normalizeCleanupPath(dir)
			if seen[key] {
				continue
			}
			seen[key] = true
			out = append(out, dir)
		}
	}

	return out
}

// removeCloneSwapDirsIn removes every *.gitmap-tmp-* dir directly under base
// and records each outcome. Glob errors and per-dir RemoveAll failures are
// classified the same way regular cleanup candidates are.
func removeCloneSwapDirsIn(base string, report *cleanupReport) {
	pattern := filepath.Join(base, cloneSwapDirGlob)
	matches, err := filepath.Glob(pattern)
	if err != nil {
		report.record(cleanupResult{
			Path:   pattern,
			Kind:   cleanupKindSwapDir,
			Status: cleanupStatusGlobError,
			Reason: "invalid swap-dir glob pattern",
			Err:    err,
		})

		return
	}

	for _, match := range matches {
		recordSwapDirAttempt(match, report)
	}
}

// recordSwapDirAttempt stats the candidate, attempts RemoveAll, and records
// the result in the shared report.
func recordSwapDirAttempt(match string, report *cleanupReport) {
	info, statErr := os.Stat(match)
	if statErr != nil || !info.IsDir() {
		return
	}

	cleanPath := filepath.Clean(match)
	if err := os.RemoveAll(match); err != nil {
		status, reason := classifyRemoveError(err)
		report.record(cleanupResult{
			Path:   cleanPath,
			Kind:   cleanupKindSwapDir,
			Status: status,
			Reason: reason,
			Err:    err,
		})

		return
	}

	report.record(cleanupResult{
		Path:   cleanPath,
		Kind:   cleanupKindSwapDir,
		Status: cleanupStatusRemoved,
		Reason: "directory tree deleted successfully",
	})
}
