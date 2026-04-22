package templates

import (
	"bufio"
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/alimtvnetwork/gitmap-v6/gitmap/constants"
)

// MergeOptions configures a Merge call.
type MergeOptions struct {
	TargetPath string   // absolute path to .gitignore or .gitattributes
	Kind       string   // constants.TemplateKindIgnore | TemplateKindAttributes | TemplateKindLFS
	Langs      []string // language template names; "common" is always prepended
}

// MergeResult summarizes a merge.
type MergeResult struct {
	WrittenPath  string
	ManagedLines int
	UserLines    int
	Changed      bool // false when re-run produced byte-identical output
}

// Merge composes the requested templates and writes the target file with a
// stable marker block. Idempotent: a second call with identical args yields
// a byte-identical file (Changed=false).
func Merge(opts MergeOptions) (MergeResult, error) {
	open, closeT := markersFor(opts.Kind)

	managed, err := composeManaged(opts.Kind, opts.Langs)
	if err != nil {
		return MergeResult{}, err
	}

	existing, _ := os.ReadFile(opts.TargetPath) // missing file is fine
	user := extractUser(existing, open, closeT)

	rendered := render(open, closeT, managed, user)

	res := MergeResult{
		WrittenPath:  opts.TargetPath,
		ManagedLines: len(managed),
		UserLines:    len(user),
		Changed:      !bytes.Equal(existing, rendered),
	}
	if !res.Changed {
		return res, nil
	}
	if err := atomicWrite(opts.TargetPath, rendered); err != nil {
		return res, err
	}

	return res, nil
}

// markersFor returns the (open, close) marker tokens for a template kind.
// LFS shares the attributes markers because both write to .gitattributes.
func markersFor(kind string) (string, string) {
	if kind == constants.TemplateKindIgnore {
		return constants.MarkerIgnoreOpen, constants.MarkerIgnoreClose
	}

	return constants.MarkerAttributesOpen, constants.MarkerAttributesClose
}

// composeManaged loads `common` then each requested lang, dedupes, and
// returns the deduped line slice (audit headers stripped).
func composeManaged(kind string, langs []string) ([]string, error) {
	ordered := append([]string{"common"}, dedupeStrings(langs)...)

	var lines []string
	for _, lang := range ordered {
		r, err := Resolve(kind, lang)
		if err != nil {
			return nil, fmt.Errorf("compose %s/%s: %w", kind, lang, err)
		}
		lines = append(lines, stripHeader(r.Content)...)
	}

	return dedupeLines(lines), nil
}

// dedupeStrings removes duplicates while preserving first-seen order.
func dedupeStrings(in []string) []string {
	seen := make(map[string]struct{}, len(in))
	out := in[:0:0]
	for _, s := range in {
		if _, ok := seen[s]; ok {
			continue
		}
		seen[s] = struct{}{}
		out = append(out, s)
	}

	return out
}

// stripHeader drops the leading audit-trail block (lines starting with
// `# source:` / `# kind:` / `# lang:` / `# version:`) and any blank lines
// immediately following it.
func stripHeader(data []byte) []string {
	var out []string
	scanner := bufio.NewScanner(bytes.NewReader(data))
	headerDone := false
	for scanner.Scan() {
		line := scanner.Text()
		if !headerDone && isHeaderLine(line) {
			continue
		}
		if !headerDone && strings.TrimSpace(line) == "" {
			continue
		}
		headerDone = true
		out = append(out, line)
	}

	return out
}

func isHeaderLine(line string) bool {
	return strings.HasPrefix(line, constants.TemplateHeaderSource) ||
		strings.HasPrefix(line, constants.TemplateHeaderKind) ||
		strings.HasPrefix(line, constants.TemplateHeaderLang) ||
		strings.HasPrefix(line, constants.TemplateHeaderVersion)
}

// dedupeLines removes duplicate lines (trimmed key) while preserving order.
// Blank lines are collapsed: at most one consecutive blank survives.
func dedupeLines(in []string) []string {
	seen := make(map[string]struct{}, len(in))
	out := make([]string, 0, len(in))
	prevBlank := false
	for _, line := range in {
		key := strings.TrimSpace(line)
		if key == "" {
			if prevBlank {
				continue
			}
			prevBlank = true
			out = append(out, "")

			continue
		}
		prevBlank = false
		if _, ok := seen[key]; ok {
			continue
		}
		seen[key] = struct{}{}
		out = append(out, line)
	}

	return trimTrailingBlanks(out)
}

func trimTrailingBlanks(in []string) []string {
	end := len(in)
	for end > 0 && strings.TrimSpace(in[end-1]) == "" {
		end--
	}

	return in[:end]
}

// extractUser returns lines from existing file content that live OUTSIDE
// the managed marker block. The `# user entries` separator (if present) is
// dropped — render() re-emits it.
func extractUser(existing []byte, open, closeT string) []string {
	if len(existing) == 0 {
		return nil
	}
	var out []string
	scanner := bufio.NewScanner(bytes.NewReader(existing))
	inManaged := false
	for scanner.Scan() {
		line := scanner.Text()
		switch {
		case strings.TrimSpace(line) == open:
			inManaged = true
		case strings.TrimSpace(line) == closeT:
			inManaged = false
		case inManaged:
			// drop managed content — we regenerate it
		case strings.TrimSpace(line) == constants.MarkerUserEntries:
			// drop the separator — render() re-adds it
		default:
			out = append(out, line)
		}
	}

	return trimLeadingAndTrailingBlanks(out)
}

func trimLeadingAndTrailingBlanks(in []string) []string {
	start := 0
	for start < len(in) && strings.TrimSpace(in[start]) == "" {
		start++
	}
	end := len(in)
	for end > start && strings.TrimSpace(in[end-1]) == "" {
		end--
	}

	return in[start:end]
}

// render produces the final file bytes: managed block on top, user entries
// below the separator. Always ends with a single trailing newline.
func render(open, closeT string, managed, user []string) []byte {
	var buf bytes.Buffer
	buf.WriteString(open)
	buf.WriteByte('\n')
	for _, l := range managed {
		buf.WriteString(l)
		buf.WriteByte('\n')
	}
	buf.WriteString(closeT)
	buf.WriteByte('\n')

	if len(user) > 0 {
		buf.WriteByte('\n')
		buf.WriteString(constants.MarkerUserEntries)
		buf.WriteByte('\n')
		for _, l := range user {
			buf.WriteString(l)
			buf.WriteByte('\n')
		}
	}

	return buf.Bytes()
}

// atomicWrite writes via temp file + rename to avoid torn writes on crash.
func atomicWrite(path string, data []byte) error {
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return fmt.Errorf("mkdir %s: %w", dir, err)
	}
	tmp, err := os.CreateTemp(dir, ".gitmap-tmp-*")
	if err != nil {
		return fmt.Errorf("tempfile in %s: %w", dir, err)
	}
	tmpName := tmp.Name()
	if _, err := tmp.Write(data); err != nil {
		_ = tmp.Close()
		_ = os.Remove(tmpName)

		return fmt.Errorf("write %s: %w", tmpName, err)
	}
	if err := tmp.Close(); err != nil {
		_ = os.Remove(tmpName)

		return fmt.Errorf("close %s: %w", tmpName, err)
	}
	if err := os.Rename(tmpName, path); err != nil {
		_ = os.Remove(tmpName)

		return fmt.Errorf("rename %s -> %s: %w", tmpName, path, err)
	}

	return nil
}
