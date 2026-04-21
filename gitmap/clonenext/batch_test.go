package clonenext

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// TestLoadBatchFromCSV_HeaderlessSinglePath verifies the simplest case:
// one row, no header, one column.
func TestLoadBatchFromCSV_HeaderlessSinglePath(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "list.csv")
	writeFile(t, path, "./my-repo\n")

	got, err := LoadBatchFromCSV(path)
	if err != nil {
		t.Fatalf("LoadBatchFromCSV: %v", err)
	}
	if len(got) != 1 {
		t.Fatalf("got %d rows, want 1", len(got))
	}
	if !strings.HasSuffix(got[0], "my-repo") {
		t.Errorf("got %q, want suffix %q", got[0], "my-repo")
	}
}

// TestLoadBatchFromCSV_WithHeaderAndExtraColumns verifies the header row
// is skipped and ragged extra columns are ignored.
func TestLoadBatchFromCSV_WithHeaderAndExtraColumns(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "list.csv")
	writeFile(t, path, "repo,version,note\n./alpha,v++,first\n./beta,v3,second\n")

	got, err := LoadBatchFromCSV(path)
	if err != nil {
		t.Fatalf("LoadBatchFromCSV: %v", err)
	}
	if len(got) != 2 {
		t.Fatalf("got %d rows, want 2", len(got))
	}
}

// TestLoadBatchFromCSV_EmptyReturnsSentinel verifies the soft-fail sentinel.
func TestLoadBatchFromCSV_EmptyReturnsSentinel(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "list.csv")
	writeFile(t, path, "repo\n\n  \n")

	_, err := LoadBatchFromCSV(path)
	if !errors.Is(err, ErrBatchEmpty) {
		t.Fatalf("err = %v, want ErrBatchEmpty", err)
	}
}

// TestWalkBatchFromDir_OnlyGitDirsIncluded verifies non-git subdirs are
// filtered out and results are sorted.
func TestWalkBatchFromDir_OnlyGitDirsIncluded(t *testing.T) {
	root := t.TempDir()
	mkRepo(t, filepath.Join(root, "zeta"))
	mkRepo(t, filepath.Join(root, "alpha"))
	if err := os.MkdirAll(filepath.Join(root, "not-a-repo"), 0o755); err != nil {
		t.Fatalf("mkdir non-repo: %v", err)
	}

	got, err := WalkBatchFromDir(root)
	if err != nil {
		t.Fatalf("WalkBatchFromDir: %v", err)
	}
	if len(got) != 2 {
		t.Fatalf("got %d repos, want 2", len(got))
	}
	if !strings.HasSuffix(got[0], "alpha") || !strings.HasSuffix(got[1], "zeta") {
		t.Errorf("got %v, want sorted [alpha, zeta]", got)
	}
}

// TestWalkBatchFromDir_NoReposReturnsSentinel verifies the soft-fail path.
func TestWalkBatchFromDir_NoReposReturnsSentinel(t *testing.T) {
	root := t.TempDir()

	_, err := WalkBatchFromDir(root)
	if !errors.Is(err, ErrBatchEmpty) {
		t.Fatalf("err = %v, want ErrBatchEmpty", err)
	}
}

// writeFile writes content to path and fails the test on error.
func writeFile(t *testing.T, path, content string) {
	t.Helper()
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatalf("write %s: %v", path, err)
	}
}

// mkRepo creates a minimal git-repo-shaped directory at path (mkdir + .git).
func mkRepo(t *testing.T, path string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Join(path, ".git"), 0o755); err != nil {
		t.Fatalf("mkRepo %s: %v", path, err)
	}
}
