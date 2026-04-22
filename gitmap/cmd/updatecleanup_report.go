package cmd

import (
	"errors"
	"fmt"
	"io/fs"
	"os"
	"sort"

	"github.com/alimtvnetwork/gitmap-v6/gitmap/constants"
)

// cleanupArtifactKind labels the *category* of artifact that was processed,
// so the per-artifact and summary output groups results in a useful way.
type cleanupArtifactKind string

const (
	cleanupKindTemp      cleanupArtifactKind = "handoff-copy"
	cleanupKindBackup    cleanupArtifactKind = "old-backup"
	cleanupKindDriveShim cleanupArtifactKind = "drive-root-shim"
	cleanupKindSwapDir   cleanupArtifactKind = "clone-swap-dir"
)

// cleanupArtifactStatus describes the *outcome* of a single cleanup attempt.
// Every status maps to a stable reason code shown to the user so logs are
// grep-able and downstream tools can parse them.
type cleanupArtifactStatus string

const (
	cleanupStatusRemoved          cleanupArtifactStatus = "removed"
	cleanupStatusLocked           cleanupArtifactStatus = "locked"
	cleanupStatusMissing          cleanupArtifactStatus = "missing"
	cleanupStatusSkippedActive    cleanupArtifactStatus = "skipped-active"
	cleanupStatusSkippedDuplicate cleanupArtifactStatus = "skipped-duplicate"
	cleanupStatusSkippedTooLarge  cleanupArtifactStatus = "skipped-too-large"
	cleanupStatusSkippedNotInRoot cleanupArtifactStatus = "skipped-not-in-root"
	cleanupStatusGlobError        cleanupArtifactStatus = "glob-error"
)

// cleanupResult is one row in the cleanup report.
//
// Reason is a short, human-readable explanation tied to the status (e.g.
// "file held by another process" for `locked`). Err carries the underlying
// OS error when present, so verbose mode can surface the raw `*PathError`
// for support diagnostics.
type cleanupResult struct {
	Path   string
	Kind   cleanupArtifactKind
	Status cleanupArtifactStatus
	Reason string
	Err    error
}

// cleanupReport collects every result and prints a structured per-artifact
// log followed by a status-grouped summary. It is intentionally not
// goroutine-safe — cleanup runs serially.
type cleanupReport struct {
	results []cleanupResult
}

// newCleanupReport returns an empty report.
func newCleanupReport() *cleanupReport {
	return &cleanupReport{results: make([]cleanupResult, 0, 8)}
}

// record stores a result and prints its per-artifact line immediately so
// users see progress on slow filesystems instead of a delayed summary.
func (r *cleanupReport) record(res cleanupResult) {
	r.results = append(r.results, res)
	printCleanupResultLine(res)
}

// removedCount returns the number of artifacts actually deleted.
func (r *cleanupReport) removedCount() int {
	count := 0
	for _, res := range r.results {
		if res.Status == cleanupStatusRemoved {
			count++
		}
	}

	return count
}

// errorCount returns the number of results that the user should know about
// (locked / glob-error). Skipped/missing results are NOT errors.
func (r *cleanupReport) errorCount() int {
	count := 0
	for _, res := range r.results {
		if res.Status == cleanupStatusLocked || res.Status == cleanupStatusGlobError {
			count++
		}
	}

	return count
}

// printSummary prints a status-grouped table with counts and reason codes.
// Empty reports just emit MsgUpdateCleanNone.
func (r *cleanupReport) printSummary() {
	if len(r.results) == 0 {
		fmt.Println(constants.MsgUpdateCleanNone)

		return
	}

	groups := groupCleanupResultsByStatus(r.results)
	statuses := sortedCleanupStatuses(groups)

	fmt.Println(constants.MsgUpdateCleanSummaryHeader)
	for _, status := range statuses {
		bucket := groups[status]
		fmt.Printf(constants.MsgUpdateCleanSummaryRow, status, len(bucket))
	}

	removed := r.removedCount()
	failed := r.errorCount()
	fmt.Printf(constants.MsgUpdateCleanSummaryTotal, removed, failed, len(r.results))
}

// groupCleanupResultsByStatus buckets results by their status code so the
// summary can show counts per outcome.
func groupCleanupResultsByStatus(results []cleanupResult) map[cleanupArtifactStatus][]cleanupResult {
	groups := map[cleanupArtifactStatus][]cleanupResult{}
	for _, res := range results {
		groups[res.Status] = append(groups[res.Status], res)
	}

	return groups
}

// sortedCleanupStatuses returns the status keys in a stable display order:
// successes first, then actionable problems, then informational skips.
func sortedCleanupStatuses(groups map[cleanupArtifactStatus][]cleanupResult) []cleanupArtifactStatus {
	keys := make([]cleanupArtifactStatus, 0, len(groups))
	for k := range groups {
		keys = append(keys, k)
	}
	sort.SliceStable(keys, func(i, j int) bool {
		return cleanupStatusSortRank(keys[i]) < cleanupStatusSortRank(keys[j])
	})

	return keys
}

// cleanupStatusSortRank assigns a sort weight to each status. Lower = earlier.
func cleanupStatusSortRank(s cleanupArtifactStatus) int {
	switch s {
	case cleanupStatusRemoved:
		return 0
	case cleanupStatusLocked:
		return 1
	case cleanupStatusGlobError:
		return 2
	case cleanupStatusMissing:
		return 3
	case cleanupStatusSkippedActive:
		return 4
	case cleanupStatusSkippedDuplicate:
		return 5
	case cleanupStatusSkippedTooLarge:
		return 6
	case cleanupStatusSkippedNotInRoot:
		return 7
	default:
		return 99
	}
}

// printCleanupResultLine emits one per-artifact line in a fixed shape:
//
//	  <symbol> [<kind>] <path> — <status>: <reason>
//
// The symbol gives a quick visual scan; the bracketed kind and trailing
// status:reason pair are stable tokens for grep / log parsing.
func printCleanupResultLine(res cleanupResult) {
	symbol := cleanupStatusSymbol(res.Status)
	stream := cleanupStatusStream(res.Status)
	fmt.Fprintf(stream, constants.MsgUpdateCleanLine,
		symbol, res.Kind, res.Path, res.Status, res.Reason,
	)
}

// cleanupStatusSymbol picks the leading glyph for a result line.
func cleanupStatusSymbol(s cleanupArtifactStatus) string {
	switch s {
	case cleanupStatusRemoved:
		return "✓"
	case cleanupStatusLocked, cleanupStatusGlobError:
		return "✗"
	case cleanupStatusMissing:
		return "·"
	default:
		return "→"
	}
}

// cleanupStatusStream picks stdout vs stderr based on whether the result
// is an actual problem the user must act on.
func cleanupStatusStream(s cleanupArtifactStatus) *os.File {
	if s == cleanupStatusLocked || s == cleanupStatusGlobError {
		return os.Stderr
	}

	return os.Stdout
}

// classifyRemoveError maps an os.Remove error to a (status, reason) pair.
// Distinguishes "file just isn't there" from "OS refused to delete it" so
// users can tell harmless leftovers from real lock conflicts.
func classifyRemoveError(err error) (cleanupArtifactStatus, string) {
	if err == nil {
		return cleanupStatusRemoved, "deleted successfully"
	}
	if errors.Is(err, fs.ErrNotExist) {
		return cleanupStatusMissing, "already gone (no cleanup needed)"
	}
	if errors.Is(err, fs.ErrPermission) {
		return cleanupStatusLocked, "permission denied — file may be held by another process"
	}

	// Fall through: treat as "locked" because that's overwhelmingly the
	// real-world cause on Windows (sharing violation surfaces as a
	// generic *PathError, not fs.ErrPermission).
	return cleanupStatusLocked, fmt.Sprintf("OS refused removal: %v", err)
}

// logUpdateCleanupExecutableError reports os.Executable failures during
// path resolution (before the cleanup report is built).
func logUpdateCleanupExecutableError(err error) {
	fmt.Fprintf(os.Stderr, constants.ErrUpdateCleanupExecPath, err)
}

// logUpdateCleanupConfigReadError reports powershell.json read failures
// during path resolution (before the cleanup report is built).
func logUpdateCleanupConfigReadError(path string, err error) {
	fmt.Fprintf(os.Stderr, constants.ErrUpdateCleanupConfigRead, path, err)
}
