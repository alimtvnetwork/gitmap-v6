package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/user/gitmap/constants"
)

// cleanupTempArtifacts removes update handoff copies and generated scripts.
func cleanupTempArtifacts(ctx updateCleanupContext) int {
	return removeCleanupPatterns(ctx.tempPatterns, ctx.selfPath, constants.MsgUpdateTempRemoved)
}

// cleanupBackupArtifacts removes .old binaries left by deploy and PATH sync.
func cleanupBackupArtifacts(ctx updateCleanupContext) int {
	return removeCleanupPatterns(ctx.backupPatterns, ctx.selfPath, constants.MsgUpdateOldRemoved)
}

// removeCleanupPatterns removes every file matched by the provided glob list.
func removeCleanupPatterns(patterns []string, selfPath, successMsg string) int {
	seen := map[string]bool{}
	cleaned := 0
	for _, pattern := range patterns {
		cleaned += removeCleanupPattern(pattern, selfPath, seen, successMsg)
	}

	return cleaned
}

// removeCleanupPattern removes files for a single glob pattern.
func removeCleanupPattern(pattern, selfPath string, seen map[string]bool, successMsg string) int {
	matches, err := filepath.Glob(pattern)
	if err != nil {
		logUpdateCleanupGlobError(pattern, err)

		return 0
	}

	cleaned := 0
	for _, match := range matches {
		if removeCleanupMatch(match, selfPath, seen, successMsg) {
			cleaned++
		}
	}

	return cleaned
}

// removeCleanupMatch removes a single cleanup candidate once.
func removeCleanupMatch(match, selfPath string, seen map[string]bool, successMsg string) bool {
	cleanPath := filepath.Clean(match)
	normalizedPath := normalizeCleanupPath(cleanPath)
	if hasSeenCleanupPath(seen, normalizedPath) {
		return false
	}
	if isActiveCleanupPath(normalizedPath, selfPath) {
		return false
	}

	return removeCleanupFile(match, cleanPath, successMsg)
}

// hasSeenCleanupPath reports whether this cleanup path was already processed.
func hasSeenCleanupPath(seen map[string]bool, normalizedPath string) bool {
	if seen[normalizedPath] {
		return true
	}

	seen[normalizedPath] = true

	return false
}

// isActiveCleanupPath reports whether the candidate points to the active binary.
func isActiveCleanupPath(normalizedPath, selfPath string) bool {
	return len(selfPath) > 0 && normalizedPath == normalizeCleanupPath(selfPath)
}

// removeCleanupFile removes a cleanup candidate and prints the success message.
func removeCleanupFile(match, cleanPath, successMsg string) bool {
	err := os.Remove(match)
	if err != nil {
		logUpdateCleanupRemoveError(cleanPath, err)

		return false
	}

	fmt.Printf(successMsg, filepath.Base(match))

	return true
}

// logUpdateCleanupExecutableError reports os.Executable failures.
func logUpdateCleanupExecutableError(err error) {
	fmt.Fprintf(os.Stderr, constants.ErrUpdateCleanupExecPath, err)
}

// logUpdateCleanupConfigReadError reports powershell.json read failures.
func logUpdateCleanupConfigReadError(path string, err error) {
	fmt.Fprintf(os.Stderr, constants.ErrUpdateCleanupConfigRead, path, err)
}

// logUpdateCleanupGlobError reports filepath.Glob failures.
func logUpdateCleanupGlobError(path string, err error) {
	fmt.Fprintf(os.Stderr, constants.ErrUpdateCleanupGlob, path, err)
}

// logUpdateCleanupRemoveError reports os.Remove failures.
func logUpdateCleanupRemoveError(path string, err error) {
	fmt.Fprintf(os.Stderr, constants.ErrUpdateCleanupRemove, path, err)
}
