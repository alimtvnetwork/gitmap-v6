package templates

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/alimtvnetwork/gitmap-v6/gitmap/constants"
)

func TestMergeAttributesUsesAttributesMarkers(t *testing.T) {
	dir := t.TempDir()
	target := filepath.Join(dir, ".gitattributes")

	if _, err := Merge(MergeOptions{
		TargetPath: target,
		Kind:       constants.TemplateKindAttributes,
		Langs:      []string{"go"},
	}); err != nil {
		t.Fatalf("Merge attributes: %v", err)
	}
	out, _ := os.ReadFile(target)
	s := string(out)
	if !strings.Contains(s, constants.MarkerAttributesOpen) ||
		!strings.Contains(s, constants.MarkerAttributesClose) {
		t.Fatalf("expected attributes markers:\n%s", s)
	}
	if strings.Contains(s, constants.MarkerIgnoreOpen) {
		t.Fatalf("attributes file leaked ignore markers:\n%s", s)
	}
	if !strings.Contains(s, "*.go") {
		t.Errorf("expected go attributes content:\n%s", s)
	}
}

func TestMergeLFSWritesToAttributesAndUsesAttributeMarkers(t *testing.T) {
	dir := t.TempDir()
	target := filepath.Join(dir, ".gitattributes")

	res, err := Merge(MergeOptions{
		TargetPath: target,
		Kind:       constants.TemplateKindLFS,
		Langs:      nil, // common only
	})
	if err != nil {
		t.Fatalf("Merge lfs: %v", err)
	}
	if !res.Changed {
		t.Fatal("first lfs merge must report Changed=true")
	}

	out, _ := os.ReadFile(target)
	s := string(out)
	if !strings.Contains(s, "filter=lfs") {
		t.Fatalf("expected LFS filter directives:\n%s", s)
	}
	if strings.Contains(s, "*.svg filter=lfs") {
		t.Fatal("LFS template must NOT mark *.svg as LFS (it is text)")
	}
	if !strings.Contains(s, constants.MarkerAttributesOpen) {
		t.Fatalf("LFS merge must use the attributes markers:\n%s", s)
	}
}

func TestMergeAttributesIsIdempotent(t *testing.T) {
	dir := t.TempDir()
	target := filepath.Join(dir, ".gitattributes")
	opts := MergeOptions{
		TargetPath: target,
		Kind:       constants.TemplateKindAttributes,
		Langs:      []string{"go", "node"},
	}

	if _, err := Merge(opts); err != nil {
		t.Fatalf("first: %v", err)
	}
	first, _ := os.ReadFile(target)

	res, err := Merge(opts)
	if err != nil {
		t.Fatalf("second: %v", err)
	}
	if res.Changed {
		t.Fatal("second identical attributes merge must report Changed=false")
	}
	second, _ := os.ReadFile(target)
	if string(first) != string(second) {
		t.Fatalf("attributes idempotence violated:\n%s\n---\n%s", first, second)
	}
}

func TestExtensionMapping(t *testing.T) {
	cases := []struct {
		kind string
		want string
	}{
		{constants.TemplateKindIgnore, constants.TemplateExtIgnore},
		{constants.TemplateKindAttributes, constants.TemplateExtAttributes},
		{constants.TemplateKindLFS, constants.TemplateExtAttributes},
	}
	for _, c := range cases {
		if got := extFor(c.kind); got != c.want {
			t.Errorf("extFor(%q) = %q, want %q", c.kind, got, c.want)
		}
	}
}
