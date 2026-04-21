package clonenext

// Batch input handling for `gitmap cn` operating on multiple repos.
//
// Two entry points feed the same dispatcher:
//
//   - LoadBatchFromCSV: read a curated list of repo paths from a CSV file
//     (header optional; only the first column matters).
//   - WalkBatchFromDir: scan one level under a directory and return every
//     subdirectory that is itself a git repo.
//
// Both return absolute paths in deterministic (lexicographic) order so that
// re-runs and report rows stay stable.

import (
	"encoding/csv"
	"errors"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

// ErrBatchEmpty is returned when neither the CSV nor the cwd walk yielded
// any candidate repos. Callers surface this as a soft warning, not a crash.
var ErrBatchEmpty = errors.New("clonenext: batch input contained no repos")

// LoadBatchFromCSV reads a CSV file and returns one absolute repo path per
// non-empty data row. The first column is treated as the path; additional
// columns are ignored so users can keep notes/version overrides alongside.
//
// Header detection: if the first cell of row 0 (case-folded) equals "repo"
// or "path", row 0 is skipped. Otherwise it is treated as data.
func LoadBatchFromCSV(path string) ([]string, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	rows, err := readAllCSVRows(file)
	if err != nil {
		return nil, err
	}

	paths := extractFirstColumn(rows)
	if len(paths) == 0 {
		return nil, ErrBatchEmpty
	}

	return absoluteAndSorted(paths), nil
}

// readAllCSVRows reads every row from r using the standard CSV parser with
// FieldsPerRecord disabled so ragged rows are tolerated.
func readAllCSVRows(r io.Reader) ([][]string, error) {
	cr := csv.NewReader(r)
	cr.FieldsPerRecord = -1
	cr.TrimLeadingSpace = true

	return cr.ReadAll()
}

// extractFirstColumn pulls the first non-empty cell of each row, skipping
// a header row when one is detected.
func extractFirstColumn(rows [][]string) []string {
	if len(rows) == 0 {
		return nil
	}

	startIdx := 0
	if isHeaderRow(rows[0]) {
		startIdx = 1
	}

	out := make([]string, 0, len(rows)-startIdx)
	for _, row := range rows[startIdx:] {
		if len(row) == 0 {
			continue
		}
		cell := strings.TrimSpace(row[0])
		if len(cell) > 0 {
			out = append(out, cell)
		}
	}

	return out
}

// isHeaderRow returns true when the first cell looks like a column label
// rather than a real path.
func isHeaderRow(row []string) bool {
	if len(row) == 0 {
		return false
	}
	first := strings.ToLower(strings.TrimSpace(row[0]))

	return first == "repo" || first == "path" || first == "repo_path"
}

// WalkBatchFromDir returns every immediate subdirectory of root that is
// itself a git repository (i.e. contains a `.git` entry).
func WalkBatchFromDir(root string) ([]string, error) {
	entries, err := os.ReadDir(root)
	if err != nil {
		return nil, err
	}

	var repos []string
	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}
		candidate := filepath.Join(root, entry.Name())
		if isGitRepo(candidate) {
			repos = append(repos, candidate)
		}
	}

	if len(repos) == 0 {
		return nil, ErrBatchEmpty
	}

	return absoluteAndSorted(repos), nil
}

// isGitRepo reports whether path contains a .git entry (file or directory —
// `.git` files exist for git worktrees).
func isGitRepo(path string) bool {
	_, err := os.Stat(filepath.Join(path, ".git"))

	return err == nil
}

// absoluteAndSorted resolves each input path to an absolute form and
// returns the result sorted lexicographically. Paths that fail to resolve
// are kept as-is so the caller can surface a meaningful per-repo error
// later instead of dropping rows silently.
func absoluteAndSorted(paths []string) []string {
	out := make([]string, len(paths))
	for i, p := range paths {
		abs, err := filepath.Abs(p)
		if err == nil {
			out[i] = abs

			continue
		}
		out[i] = p
	}
	sort.Strings(out)

	return out
}
