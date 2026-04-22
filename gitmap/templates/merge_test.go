package templates

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/alimtvnetwork/gitmap-v6/gitmap/constants"
)

func TestMergeIgnoreCreatesFileWithMarkers(t *testing.T) {
	dir := t.TempDir()
	target := filepath.Join(dir, ".gitignore")

	res, err := Merge(MergeOptions{
		TargetPath: target,
		Kind:       constants.TemplateKindIgnore,
		Langs:      []string{"go"},
	})
	if err != nil {
		t.Fatalf("Merge: %v", err)
	}
	if !res.Changed {
		t.Fatal("first merge must report Changed=true")
	}

	data, err := os.ReadFile(target)
	if err != nil {
		t.Fatalf("read: %v", err)
	}
	s := string(data)
	if !strings.Contains(s, constants.MarkerIgnoreOpen) ||
		!strings.Contains(s, constants.MarkerIgnoreClose) {
		t.Fatalf("missing markers:\n%s", s)
	}
	if !strings.Contains(s, "node_modules") == false {
		// presence of go content
	}
	if !strings.Contains(s, "*.exe") {
		t.Fatalf("expected go template content (*.exe) in:\n%s", s)
	}
}

func TestMergeIgnoreIsIdempotent(t *testing.T) {
	dir := t.TempDir()
	target := filepath.Join(dir, ".gitignore")
	opts := MergeOptions{
		TargetPath: target,
		Kind:       constants.TemplateKindIgnore,
		Langs:      []string{"go", "node"},
	}

	if _, err := Merge(opts); err != nil {
		t.Fatalf("first merge: %v", err)
	}
	first, _ := os.ReadFile(target)

	res, err := Merge(opts)
	if err != nil {
		t.Fatalf("second merge: %v", err)
	}
	if res.Changed {
		t.Fatal("second identical merge must report Changed=false")
	}
	second, _ := os.ReadFile(target)
	if !bytes.Equal(first, second) {
		t.Fatalf("idempotence violated:\n--- first ---\n%s\n--- second ---\n%s", first, second)
	}
}

func TestMergePreservesUserEntries(t *testing.T) {
	dir := t.TempDir()
	target := filepath.Join(dir, ".gitignore")

	preExisting := "# my custom rule\nsecret-notes/\n*.local.md\n"
	if err := os.WriteFile(target, []byte(preExisting), 0o644); err != nil {
		t.Fatalf("seed: %v", err)
	}

	if _, err := Merge(MergeOptions{
		TargetPath: target,
		Kind:       constants.TemplateKindIgnore,
		Langs:      []string{"go"},
	}); err != nil {
		t.Fatalf("Merge: %v", err)
	}

	out, _ := os.ReadFile(target)
	s := string(out)
	for _, want := range []string{"secret-notes/", "*.local.md", "# my custom rule"} {
		if !strings.Contains(s, want) {
			t.Errorf("user entry %q lost:\n%s", want, s)
		}
	}
	if !strings.Contains(s, constants.MarkerUserEntries) {
		t.Errorf("expected %q separator:\n%s", constants.MarkerUserEntries, s)
	}
}

func TestMergeDedupesAcrossLangs(t *testing.T) {
	dir := t.TempDir()
	target := filepath.Join(dir, ".gitignore")

	if _, err := Merge(MergeOptions{
		TargetPath: target,
		Kind:       constants.TemplateKindIgnore,
		Langs:      []string{"node", "python"}, // both touch dist/, build/
	}); err != nil {
		t.Fatalf("Merge: %v", err)
	}

	out, _ := os.ReadFile(target)
	if n := strings.Count(string(out), "\ndist/\n"); n != 1 {
		t.Fatalf("dist/ appears %d times, want 1:\n%s", n, out)
	}
}

func TestMergeAddingLangPreservesPrior(t *testing.T) {
	dir := t.TempDir()
	target := filepath.Join(dir, ".gitignore")

	if _, err := Merge(MergeOptions{
		TargetPath: target,
		Kind:       constants.TemplateKindIgnore,
		Langs:      []string{"go"},
	}); err != nil {
		t.Fatalf("first: %v", err)
	}
	if _, err := Merge(MergeOptions{
		TargetPath: target,
		Kind:       constants.TemplateKindIgnore,
		Langs:      []string{"go", "python"},
	}); err != nil {
		t.Fatalf("second: %v", err)
	}

	out, _ := os.ReadFile(target)
	s := string(out)
	if !strings.Contains(s, "*.exe") {
		t.Errorf("go content lost: %s", s)
	}
	if !strings.Contains(s, "__pycache__/") {
		t.Errorf("python content not added: %s", s)
	}
}
