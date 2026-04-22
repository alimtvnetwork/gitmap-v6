---
name: Templates folder + add ignore / add attributes / pretty markdown
description: PLANNED — external templates/ folder beside data/ holds .gitignore and .gitattributes templates; new `gitmap add ignore` and `gitmap add attributes` commands merge+compact selected language templates into repo files under marker blocks; pretty-markdown CLI renderer collapses redundant fenced blocks to yellow and colors quoted strings cyan
type: feature
---

# Templates: Ignore / Attributes / Pretty Markdown

**Status:** Planned (spec only). Spec:
[`spec/01-app/109-templates-ignore-attributes-pretty.md`](../../../spec/01-app/109-templates-ignore-attributes-pretty.md).
Plan: [`.lovable/memory/plans/04-templates-ignore-attributes-plan.md`](../plans/04-templates-ignore-attributes-plan.md).

## Three features, one foundation

A new **`templates/`** folder ships beside `data/` (same `<binary-dir>`
anchoring rule as the SQLite DB). It holds editable `.gitignore` and
`.gitattributes` templates per language. Embedded via `go:embed` and
materialized on first run; user-owned thereafter.

### Commands

- `gitmap add ignore` — writes/updates `.gitignore` from
  `templates/gitignore/common.gitignore` (curated multi-language
  baseline: C#, Node, Python, Go, Rust, Java, …).
- `gitmap add ignore go,rust` — merges multiple language templates.
- `gitmap add attributes` — same model for `.gitattributes`. Default
  template includes LFS lines for binaries, `text eol=…` directives
  for known source extensions, and linguist hints.
- `--dry-run` on both prints the final file to stdout, writes nothing.

### Merge → compact algorithm

Concat selected templates in arg order, then single-pass dedupe by
trimmed line (comments dedupe too). First writer wins; order preserved.
O(n), runs in ms.

### Marker block

Generated content sits at the **top** of the target file inside:

```
# >>> gitmap-ignore (do not edit between markers) >>>
…
# <<< gitmap-ignore <<<
```

Existing user content is preserved verbatim **below** the block under
a `# user entries` comment. Re-runs replace the block in place; user
content untouched.

## SVG removal from LFS defaults

The hard-coded `lfs-common` slice in
[`gitmap/cmd/lfscommon.go`](../../../gitmap/cmd/lfscommon.go)
currently includes `*.svg`. SVGs are XML text — LFS bloats the store
and kills meaningful diffs. The new
`templates/gitattributes/common.gitattributes` emits
`*.svg text eol=lf linguist-detectable=true` instead. `lfs-common`
will be migrated to read from
`templates/gitattributes/lfs-common.gitattributes` (no hard-coded
list).

## Code files MUST be in default attributes

Beyond binary LFS lines, the default template MUST cover line-ending
normalization for known source extensions so cross-platform checkouts
are consistent (`*.go text eol=lf`, `*.cs text eol=crlf`,
`*.ps1 text eol=crlf`, etc.).

## Pretty-markdown renderer

Companion polish for any place gitmap prints embedded markdown
(help text, changelog):

1. **Collapse redundant fences.** If a fenced ` ``` ` block's body
   equals the adjacent paragraph (case-insensitive, whitespace-
   collapsed), drop the fence and print the paragraph once in
   `ColorYellow`.
2. **Color `"quoted strings"`** with `ColorCyan`. Inline code stays
   bold-dim. Precedence: inline-code > quoted > plain.

Lives in `gitmap/render/pretty.go` (new), opt-out via
`GITMAP_PRETTY=0`. Auto-disabled when stdout is not a TTY.

## Why external templates (not embedded const slices)

User can edit on disk without rebuilding. Syncs cleanly via dotfiles
repo. Embedded copy is the source-of-truth shipped with each release;
on-disk copy is never overwritten after first materialization (a
future `gitmap templates reset` will be the explicit re-seed).

## Out of scope

- `gitmap templates list/reset/diff` (future spec).
- Auto-detecting repo languages and selecting templates without args.
- Removing existing entries from user files (we only **add**).
- `git lfs migrate import` for already-committed binaries.
- New markdown parser library — the pretty renderer extends the
  existing line-based printer.
