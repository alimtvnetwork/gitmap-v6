package cmd

import (
	"os"
	"path/filepath"
	"time"
)

// cleanupRemoveMaxAttempts caps the retry loop for transient Windows file locks
// (e.g. a freshly-renamed .old still settling, or AV scanners holding a handle).
const (
	cleanupRemoveMaxAttempts = 5
	cleanupRemoveRetryDelay  = 200 * time.Millisecond
)

// cleanupTempArtifacts processes every gitmap-update-* / *.ps1 candidate
// and records each outcome in the shared report.
func cleanupTempArtifacts(ctx updateCleanupContext, report *cleanupReport) {
	processCleanupPatterns(ctx.tempPatterns, ctx.selfPath, cleanupKindTemp, report)
}

// cleanupBackupArtifacts processes every *.old candidate and records each
// outcome in the shared report.
func cleanupBackupArtifacts(ctx updateCleanupContext, report *cleanupReport) {
	processCleanupPatterns(ctx.backupPatterns, ctx.selfPath, cleanupKindBackup, report)
}

// processCleanupPatterns walks every glob pattern in the group, deduplicates
// matches across overlapping pattern lists, and records one cleanupResult
// per unique candidate.
func processCleanupPatterns(patterns []string, selfPath string, kind cleanupArtifactKind, report *cleanupReport) {
	seen := map[string]bool{}
	for _, pattern := range patterns {
		processCleanupPattern(pattern, selfPath, kind, seen, report)
	}
}

// processCleanupPattern runs a single glob and records one result per match.
// A glob error itself is recorded so the user sees which pattern misfired.
func processCleanupPattern(pattern, selfPath string, kind cleanupArtifactKind, seen map[string]bool, report *cleanupReport) {
	matches, err := filepath.Glob(pattern)
	if err != nil {
		report.record(cleanupResult{
			Path:   pattern,
			Kind:   kind,
			Status: cleanupStatusGlobError,
			Reason: "invalid cleanup glob pattern",
			Err:    err,
		})

		return
	}

	for _, match := range matches {
		processCleanupMatch(match, selfPath, kind, seen, report)
	}
}

// processCleanupMatch evaluates one candidate against the dedupe set and
// the active-binary guard, then attempts removal.
func processCleanupMatch(match, selfPath string, kind cleanupArtifactKind, seen map[string]bool, report *cleanupReport) {
	cleanPath := filepath.Clean(match)
	normalizedPath := normalizeCleanupPath(cleanPath)

	if seen[normalizedPath] {
		report.record(cleanupResult{
			Path:   cleanPath,
			Kind:   kind,
			Status: cleanupStatusSkippedDuplicate,
			Reason: "already processed via another pattern",
		})

		return
	}
	seen[normalizedPath] = true

	if isActiveCleanupPath(normalizedPath, selfPath) {
		report.record(cleanupResult{
			Path:   cleanPath,
			Kind:   kind,
			Status: cleanupStatusSkippedActive,
			Reason: "candidate is the currently-running gitmap binary",
		})

		return
	}

	report.record(removeCleanupCandidate(match, cleanPath, kind))
}

// isActiveCleanupPath reports whether the candidate points to the active binary.
func isActiveCleanupPath(normalizedPath, selfPath string) bool {
	return len(selfPath) > 0 && normalizedPath == normalizeCleanupPath(selfPath)
}

// removeCleanupCandidate attempts os.Remove with a small retry loop and
// returns a fully-classified cleanupResult. Retries exist because Windows
// frequently holds a brief sharing lock on a freshly-renamed .old file or
// on a binary the parent handoff process is still releasing.
func removeCleanupCandidate(match, cleanPath string, kind cleanupArtifactKind) cleanupResult {
	var lastErr error
	for attempt := 1; attempt <= cleanupRemoveMaxAttempts; attempt++ {
		err := os.Remove(match)
		if err == nil {
			return cleanupResult{
				Path:   cleanPath,
				Kind:   kind,
				Status: cleanupStatusRemoved,
				Reason: "deleted successfully",
			}
		}

		lastErr = err
		if attempt < cleanupRemoveMaxAttempts {
			time.Sleep(cleanupRemoveRetryDelay)
		}
	}

	status, reason := classifyRemoveError(lastErr)

	return cleanupResult{
		Path:   cleanPath,
		Kind:   kind,
		Status: status,
		Reason: reason,
		Err:    lastErr,
	}
}
