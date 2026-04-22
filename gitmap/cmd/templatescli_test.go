package cmd

import "testing"

// TestIsMarkdownTemplatePathRecognizesMarkdownExtensions locks the file-
// extension allow-list. Adding a new markdown extension is a deliberate
// choice — this test will scream if someone broadens it accidentally.
func TestIsMarkdownTemplatePathRecognizesMarkdownExtensions(t *testing.T) {
	cases := map[string]bool{
		"assets/notes/intro.md":           true,
		"assets/notes/intro.MD":           true,
		"assets/notes/intro.markdown":     true,
		"assets/notes/intro.MARKDOWN":     true,
		"assets/ignore/go.gitignore":      false,
		"assets/attributes/go.gitattributes": false,
		"assets/lfs/common.gitattributes": false,
		"plain":                           false,
		"":                                false,
	}
	for path, want := range cases {
		if got := isMarkdownTemplatePath(path); got != want {
			t.Errorf("isMarkdownTemplatePath(%q) = %v, want %v", path, got, want)
		}
	}
}

// TestShouldPrettyRenderTemplateRespectsRawFlag verifies that --raw
// short-circuits the pretty pipeline regardless of file extension or env.
// This is the primary user-facing escape hatch for diff-against-curated
// workflows ('templates show … --raw > /tmp/x && diff …').
func TestShouldPrettyRenderTemplateRespectsRawFlag(t *testing.T) {
	t.Setenv(envTemplatesNoPretty, "")
	if shouldPrettyRenderTemplate("notes/intro.md", true) {
		t.Fatal("--raw must disable pretty rendering for markdown templates")
	}
}

// TestShouldPrettyRenderTemplateRespectsEnvOptOut verifies the shared
// GITMAP_NO_PRETTY env var also short-circuits this code path, keeping
// the opt-out contract consistent with `gitmap help`.
func TestShouldPrettyRenderTemplateRespectsEnvOptOut(t *testing.T) {
	t.Setenv(envTemplatesNoPretty, "1")
	if shouldPrettyRenderTemplate("notes/intro.md", false) {
		t.Fatal("GITMAP_NO_PRETTY must disable pretty rendering")
	}
}

// TestShouldPrettyRenderTemplateSkipsNonMarkdown is the safety guard for
// the dominant case today: .gitignore / .gitattributes templates must
// always be byte-faithful so users can pipe them straight into a file.
func TestShouldPrettyRenderTemplateSkipsNonMarkdown(t *testing.T) {
	t.Setenv(envTemplatesNoPretty, "")
	if shouldPrettyRenderTemplate("ignore/go.gitignore", false) {
		t.Fatal(".gitignore templates must never be pretty-rendered")
	}
	if shouldPrettyRenderTemplate("attributes/go.gitattributes", false) {
		t.Fatal(".gitattributes templates must never be pretty-rendered")
	}
	if shouldPrettyRenderTemplate("lfs/common.gitattributes", false) {
		t.Fatal("lfs templates must never be pretty-rendered")
	}
}

// TestParseTemplatesShowFlagsExtractsRaw verifies the flag splitter
// correctly separates the boolean from the positional <kind> <lang>,
// regardless of argument order (relies on reorderFlagsBeforeArgs).
func TestParseTemplatesShowFlagsExtractsRaw(t *testing.T) {
	cases := []struct {
		name string
		args []string
		want bool
	}{
		{"flag after positionals", []string{"ignore", "go", "--raw"}, true},
		{"flag before positionals", []string{"--raw", "ignore", "go"}, true},
		{"no flag", []string{"ignore", "go"}, false},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			rest, raw := parseTemplatesShowFlags(tc.args)
			if raw != tc.want {
				t.Errorf("raw = %v, want %v", raw, tc.want)
			}
			if len(rest) != 2 || rest[0] != "ignore" || rest[1] != "go" {
				t.Errorf("rest = %v, want [ignore go]", rest)
			}
		})
	}
}
