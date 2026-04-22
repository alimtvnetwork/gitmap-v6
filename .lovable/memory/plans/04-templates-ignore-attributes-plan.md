# Plan 04 — Templates folder, `add ignore`, `add attributes`, pretty markdown

Spec: [`spec/01-app/109-templates-ignore-attributes-pretty.md`](../../../spec/01-app/109-templates-ignore-attributes-pretty.md)
Memory: [`mem://features/templates-ignore-attributes`](../features/templates-ignore-attributes.md)

User-locked decisions (from current chat):

- **Templates location**: a new **`templates/`** folder beside `data/`
  (NOT inside `data/`), anchored to the binary's physical location
  (same rule as `mem://tech/database-location`).
- **Default behavior with no args**: `gitmap add ignore` (no language
  arg) uses the curated `common` template covering C#, Node.js,
  Python, Go, Rust, Java at minimum. Same for `add attributes`.
- **Multiple languages**: comma-separated, e.g. `go,rust` or
  `csharp,nodejs`.
- **Existing files**: respected. New content is inserted at the **top**
  inside a marker block; user content moves below under
  `# user entries`. Re-runs replace the marker block only.
- **Merge strategy**: concat → single-pass line dedupe (comments
  included). First writer wins; argument order preserved.
- **SVG removal from LFS defaults**: `*.svg` leaves the LFS list and
  becomes `*.svg text eol=lf linguist-detectable=true` in the common
  attributes template.
- **Code files in default attributes**: required (line-ending
  directives for known source extensions).
- **Pretty markdown renderer**: collapse redundant ``` fences to
  yellow one-liners; color `"quoted strings"` cyan; opt-out via
  `GITMAP_PRETTY=0`; auto-disable on non-TTY.

---

## Phasing

### Phase 0 — Scaffolding (no user-visible behavior)

- `gitmap/templates/` (Go package, NEW):
  - `embed.go` — `//go:embed all:assets/**` over the bundled defaults.
  - `assets/gitignore/common.gitignore` etc. — initial curated
    content.
  - `assets/gitattributes/common.gitattributes` etc.
- `gitmap/templates/locate.go` — `TemplatesDir() string` resolves
  `<binary-dir>/templates`. Uses the same `filepath.EvalSymlinks` rule
  as the DB locator. Materializes embedded assets on first call when
  the dir is missing. Never overwrites existing on-disk files.
- Constants: `constants_templates.go` with all paths, marker strings,
  and the language → filename map.
- Errors: `ErrTemplateUnknownLang`, `ErrTemplateReadFailed`,
  `ErrTemplateMergeFailed`, `ErrTemplateWriteFailed` in
  `constants_errors.go` (Code Red format, includes the offending
  language token / path).
- Tests: `TestTemplatesDir_Materializes`, `TestTemplatesDir_NoOverwrite`.

**Acceptance**: starting gitmap on a clean machine creates
`<binary-dir>/templates/` populated from embedded assets; second start
leaves the on-disk files untouched even if the embedded copy changed.

### Phase 1 — Merge / compact engine

- `gitmap/templates/merge.go`:
  - `Resolve(langs []string, kind Kind) ([]string, error)` — maps
    tokens to file paths under the resolved subdir; returns
    `ErrTemplateUnknownLang` on miss with the available list attached.
  - `MergeAndCompact(paths []string) (string, error)` — concat with
    `# --- <lang> ---` separators, then single-pass dedupe on trimmed
    line (comments included). First writer wins.
- Tests:
  - `TestMerge_DedupesCommentsAndRules`
  - `TestMerge_PreservesOrder`
  - `TestMerge_UnknownLangReturnsErrTemplateUnknownLang`
  - `TestMerge_LargeCorpus_FastEnough` (sanity: < 50ms for full
    github/gitignore-sized input).

### Phase 2 — Marker-block writer

- `gitmap/templates/markerwriter.go`:
  - `WriteWithMarker(repoFile, body, openMarker, closeMarker string)
    error` — atomic write via temp file + rename.
  - On existing file: locate marker pair (string match on full line);
    replace span; if no markers, prepend block + blank line +
    `# user entries` + existing content.
  - On absent file: write block alone.
- Tests:
  - `TestMarkerWriter_FreshFile`
  - `TestMarkerWriter_ExistingUserContentPreserved`
  - `TestMarkerWriter_RerunIsByteIdempotent`
  - `TestMarkerWriter_TornWriteRecovery` (kill between temp+rename).

### Phase 3 — `gitmap add ignore` command

- `gitmap/cmd/addignore.go`:
  - `runAddIgnore(args []string)` — parse comma-separated langs (or
    default `common`), `--dry-run` flag, call merge → markerwriter.
  - Print summary: `<N> rules added, <M> deduped, <K> bytes written`.
- `gitmap/helptext/add-ignore.md` (≤ 120 lines, with 3–8 line
  realistic simulation).
- Wire into root dispatch (`gitmap add …` subcommand router).
- Reserve aliases in `constants_cli.go` (final names TBD during PR).
- Marker comments per `mem://features/marker-comments` so completion
  generator picks it up; CI `generate-check` must stay green.
- Tests: `TestAddIgnore_DryRunWritesNothing`,
  `TestAddIgnore_MultiLangMergeOrderStable`,
  `TestAddIgnore_UnknownLangExitsNonZero`.

### Phase 4 — `gitmap add attributes` command

- `gitmap/cmd/addattributes.go` — mirror Phase 3 with the attributes
  subdir, marker strings, and target file `.gitattributes`.
- `gitmap/helptext/add-attributes.md`.
- **Migrate `lfs-common`** to read from
  `templates/gitattributes/lfs-common.gitattributes` instead of the
  hard-coded `lfsCommonPatterns` slice in
  [`gitmap/cmd/lfscommon.go`](../../../gitmap/cmd/lfscommon.go).
  Remove `*.svg` from the new template; existing repos keep working
  (we never untrack), only future runs change.
- Tests: parity tests against the existing `lfs-common` test suite,
  plus `TestAddAttributes_SvgIsTextNotLFS`.

### Phase 5 — Pretty-markdown renderer

- `gitmap/render/pretty.go` (≤ 200 lines):
  - `RenderMarkdown(src string, w io.Writer)` — line-based renderer
    extending the current help/changelog printer.
  - Fence collapse: scan paragraph + adjacent fenced block; if
    normalized bodies match (case-insensitive, whitespace-collapsed),
    emit one yellow line and skip the fence.
  - Quote highlighter: regex `"[^"\n]*"` colored cyan; precedence
    inline-code > quoted > plain.
  - TTY detection + `GITMAP_PRETTY=0` opt-out.
- Hook into `gitmap/cmd/help.go` and `gitmap/cmd/changelog.go`.
- Tests:
  - `TestPretty_CollapsesRedundantFenceToYellow`
  - `TestPretty_KeepsTrueCodeBlocks`
  - `TestPretty_HighlightsQuotedStrings`
  - `TestPretty_DisabledOnNonTTY`
  - `TestPretty_DisabledByEnvVar`.

### Phase 6 — Docs site + changelog

- `src/data/changelog.ts` — entry for the version that ships
  Phases 0–5 (likely batched).
- New docs page under `src/pages/` for `add-ignore` and
  `add-attributes` following the existing command-page pattern
  (CommandCard + alias chip).
- README.md — new row in the commands table.

---

## File touch-list (when implementing)

NEW:
- `gitmap/templates/embed.go`
- `gitmap/templates/locate.go`
- `gitmap/templates/merge.go`
- `gitmap/templates/markerwriter.go`
- `gitmap/templates/assets/gitignore/*.gitignore`
- `gitmap/templates/assets/gitattributes/*.gitattributes`
- `gitmap/cmd/addignore.go`
- `gitmap/cmd/addattributes.go`
- `gitmap/render/pretty.go`
- `gitmap/helptext/add-ignore.md`
- `gitmap/helptext/add-attributes.md`
- `gitmap/helptext/templates.md`
- `gitmap/constants/constants_templates.go`
- All `_test.go` siblings.

EDITED:
- `gitmap/cmd/lfscommon.go` — drop hard-coded slice, read from
  templates dir.
- `gitmap/cmd/help.go`, `gitmap/cmd/changelog.go` — switch to
  `render.RenderMarkdown`.
- `gitmap/cmd/setup.go` — `ensureGitignoreStep` to use the new
  marker-block convention so `setup` and `add ignore` agree.
- `gitmap/constants/constants_cli.go` — register `add ignore`,
  `add attributes` and aliases.
- `gitmap/constants/constants_errors.go` — new error constants.
- `src/data/changelog.ts`.
- `README.md`.

---

## Risks / open questions to resolve before coding

1. **Alias names** for `add ignore` / `add attributes`. Candidates:
   `ai` / `aa`, or `addig` / `addattr`. Defer until PR; uniqueness
   enforced by `.github/scripts/check-cmd-naming.sh`.
2. **`add` as a subcommand router** — gitmap doesn't have one yet.
   Either introduce a real `add` dispatcher or treat `add-ignore`
   as a single token under the hood. Spec is written for the
   readable form; implementation can choose.
3. **Curated `common` content** — needs a one-time review pass against
   github/gitignore. Tag the source revision in the file header so
   future re-syncs are auditable.
4. **First-run materialization on read-only install dirs** (e.g.,
   `/usr/local/bin`). Fallback: use `~/.gitmap/templates/` if the
   binary dir is not writable. Mirrors the same fallback the DB
   locator already uses.
5. **Pretty renderer false positives** — paragraph/fence equality must
   compare *normalized* text (lowercase, collapse whitespace, strip
   trailing punctuation). Need a small fixture corpus before shipping.
