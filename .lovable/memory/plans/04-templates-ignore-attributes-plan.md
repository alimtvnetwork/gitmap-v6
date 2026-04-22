# Plan 04: Templates (ignore / attributes) + Pretty Renderer

Spec: `spec/01-app/109-templates-ignore-attributes-pretty.md`
Memory: `mem://features/templates-ignore-attributes`

## Resolved open questions

| Question                                  | Decision                                      |
|-------------------------------------------|-----------------------------------------------|
| Alias names                               | `ai` / `aa` / `alfs` / `tl` / `ts`            |
| Real `add` subcommand router              | Yes — `gitmap/cmd/rootadd.go` + dispatchAdd   |
| Audit trail for curated common template   | Per-file `# source:` header + version int     |
| Read-only install fallback                | `~/.gitmap/templates/` overlay materialized   |
| Pretty renderer fixture corpus            | `gitmap/render/testdata/pretty/*.{in.md,want.txt}` |

## Phases

### Phase 0 — Scaffolding (THIS PR)

- [x] Spec, memory, plan written
- [ ] `gitmap/templates/embed.go` — `//go:embed assets/**`
- [ ] `gitmap/templates/resolver.go` — overlay > embed lookup
- [ ] `gitmap/templates/materialize.go` — first-run extract to `~/.gitmap/templates/`
- [ ] `gitmap/templates/paths.go` — user-templates dir per OS
- [ ] `gitmap/templates/assets/.keep` placeholder (Phase 1 fills it)
- [ ] `gitmap/constants/constants_templates.go` — kind/lang/marker constants
- [ ] No CLI wiring yet (next phase)

### Phase 1 — Seed corpus

- common + go + node + python + rust + csharp for both `ignore` and `attributes`
- `lfs/common.gitattributes` (binary patterns, NO `*.svg`)
- Each file has `# source: ... # version: 1` header

### Phase 2 — `add` router + `add ignore`

- `gitmap/cmd/rootadd.go` with `dispatchAdd`
- `gitmap/cmd/addignore.go` + merge engine in `gitmap/templates/merge.go`
- Marker-block aware, single-pass dedupe
- Idempotence test: run twice → byte-identical

### Phase 3 — `add attributes` + `add lfs-install`

- Mirror Phase 2 for attributes
- `add lfs-install` runs `git lfs install --local` then merges `lfs/common.gitattributes`

### Phase 4 — discovery commands

- `templates list` (groups by kind, shows overlay vs embed source)
- `templates show <kind> <lang>` (prints to stdout)

### Phase 5 — Pretty renderer

- `gitmap/render/pretty.go` with the 4 rules
- `gitmap/render/testdata/pretty/case-001..NNN`
- Table-driven test that loops fixtures

### Phase 6 — Wire pretty into CLI

- Help output (`gitmap help <cmd>`)
- Changelog display (`gitmap changelog`)

## Non-goals

- No download-on-demand. Templates are embedded, audit-trailed, versioned.
- No yaml/toml config of templates. Filesystem overlay only.
- No editor integration. Files on disk only.
