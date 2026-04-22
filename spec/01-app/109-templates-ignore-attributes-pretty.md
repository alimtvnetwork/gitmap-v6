# Spec 109 — Templates: `.gitignore`, `.gitattributes`, and Pretty-Markdown CLI Renderer

> Status: **Planned** (not implemented). This spec captures three related
> features that share one foundation: an external **`templates/`** folder
> shipped beside the `data/` folder, owned by the user, and consumed by
> several `gitmap` commands.

---

## 1. Motivation

Today gitmap can already (a) seed a curated set of Git LFS patterns via
`gitmap lfs-common` and (b) ensure release-related lines are appended to
`.gitignore` during `gitmap setup`. Both features hard-code their content
inside Go source. Users can't customize the lists without rebuilding.

This spec generalizes the pattern. We introduce:

1. **`gitmap add ignore [<langs>]`** — generate or augment `.gitignore`
   from one or more language templates (default: a curated "common"
   bundle covering C#, Node.js, Python, Go, Rust, Java).
2. **`gitmap add attributes [<langs>]`** — generate or augment
   `.gitattributes` from a curated baseline that includes:
   - `git lfs install` setup (canonical LFS lines for binary types),
   - **code-file line-ending / linguist hints** for known source
     extensions (C#, Node/TS, Python, Go, Rust, Java, etc.),
   - removes `*.svg` from the LFS default set (SVGs are text-ish and
     should be tracked normally — see §4.3).
3. **Pretty-Markdown CLI renderer** — when gitmap prints embedded
   help/changelog markdown to a terminal, collapse triple-backtick
   fenced blocks whose body equals the surrounding paragraph to a
   single yellow line, and color any `"quoted string"` distinctly.

All three load their content from the new external **`templates/`**
folder so users can edit, add, or override entries without recompiling.

---

## 2. Templates folder layout

The folder lives **beside `data/`**, anchored to the binary's physical
location (same resolution rule as `mem://tech/database-location`):

```
<binary-dir>/
  data/                  # SQLite DB + profile JSON (existing)
  templates/             # NEW — user-editable text templates
    gitignore/
      common.gitignore           # curated multi-language baseline
      csharp.gitignore
      nodejs.gitignore
      python.gitignore
      go.gitignore
      rust.gitignore
      java.gitignore
      cpp.gitignore
      … (one file per supported language)
    gitattributes/
      common.gitattributes       # curated baseline (LFS + linguist)
      csharp.gitattributes
      nodejs.gitattributes
      python.gitattributes
      go.gitattributes
      rust.gitattributes
      java.gitattributes
      … (one file per supported language)
    README.md            # explains the folder, override rules, format
```

### 2.1. First-run seeding

Templates are **embedded** in the binary via `go:embed templates/**`
(source-of-truth shipped with each release) and **materialized** to
`<binary-dir>/templates/` on first use if the folder is missing. After
that, the user owns the on-disk copy. Re-deploys never overwrite
existing files unless the user runs `gitmap templates reset`
(future command, out of scope here).

### 2.2. Why beside `data/` and not inside it

`data/` is gitmap's mutable runtime state (DB, lock files). `templates/`
is **declarative configuration** the user edits like a dotfile. Keeping
them as siblings makes that distinction obvious and lets users sync
`templates/` across machines (e.g., via dotfiles repo) without dragging
the SQLite DB along.

---

## 3. `gitmap add ignore [<langs>]`

### 3.1. CLI

```
gitmap add ignore                       # → uses templates/gitignore/common.gitignore
gitmap add ignore go                    # → uses templates/gitignore/go.gitignore
gitmap add ignore go,rust               # → merges go + rust
gitmap add ignore csharp,nodejs,python  # → merges all three
gitmap add ignore --dry-run go,rust     # → prints final file to stdout, writes nothing
```

Aliases: `gitmap add ig`, `gitmap ai` (TBD during impl; reserve in
`constants_cli.go`).

### 3.2. Behavior

1. Resolve each language token to `<templates>/gitignore/<lang>.gitignore`.
   Unknown token → exit non-zero with `ErrTemplateUnknownLang` listing
   available languages (sorted, from disk).
2. Read all selected templates in argument order.
3. **Merge → compact** (see §5).
4. If `<repo-root>/.gitignore` exists:
   - Insert the merged template at the **top** of the file under a
     marker block:
     ```
     # >>> gitmap-ignore (do not edit between markers) >>>
     <merged template>
     # <<< gitmap-ignore <<<
     ```
   - Existing user content is preserved verbatim **below** the marker
     block, separated by a single blank line and a `# user entries`
     comment.
   - On re-run, the marker block is replaced in place; user content
     untouched.
5. If `.gitignore` does not exist: write the merged template as the
   whole file (no marker block needed yet, but include it so future
   runs are uniform).

### 3.3. Default "common" template

`templates/gitignore/common.gitignore` is the union of the most-used
ignores across major ecosystems, deduped. Source: distilled from
[github/gitignore](https://github.com/github/gitignore) (MIT-licensed).
Includes at minimum:

- OS noise: `.DS_Store`, `Thumbs.db`, `desktop.ini`
- Editor: `.idea/`, `.vscode/`, `*.swp`, `*.swo`
- Logs / temp: `*.log`, `*.tmp`, `*.bak`
- **Per language** (compact subset, full version in dedicated files):
  C# `bin/ obj/ *.user`, Node `node_modules/ dist/ .env`,
  Python `__pycache__/ *.pyc .venv/`, Go `vendor/ *.exe`,
  Rust `target/`, Java `target/ *.class`, C++ `*.o *.obj`.

---

## 4. `gitmap add attributes [<langs>]`

### 4.1. CLI

Mirrors §3.1 — same flags, same merge semantics, different target file
(`.gitattributes`) and different template folder (`templates/gitattributes/`).

```
gitmap add attributes                  # common (LFS + linguist + code linends)
gitmap add attributes go,rust
gitmap add attributes --dry-run csharp
```

### 4.2. Default "common" template

Each line is one of three kinds:

- **LFS tracking** (canonical Git LFS form):
  `*.psd filter=lfs diff=lfs merge=lfs -text`
- **Line-ending normalization** for source code:
  `*.go text eol=lf`, `*.cs text eol=crlf`, `*.ps1 text eol=crlf`
- **Linguist hints** for repo language stats:
  `*.min.js linguist-generated=true`, `vendor/** linguist-vendored=true`

### 4.3. SVG removal from LFS defaults

The current hard-coded `lfs-common` list in
[`gitmap/cmd/lfscommon.go`](../../gitmap/cmd/lfscommon.go) includes
`*.svg`. SVGs are **XML text** in practice (icons, illustrations) —
tracking them in LFS bloats the LFS store and breaks meaningful diffs.

When this spec ships:

- `templates/gitattributes/common.gitattributes` does **not** track
  `*.svg` via LFS. Instead it emits:
  ```
  *.svg text eol=lf linguist-detectable=true
  ```
- `gitmap lfs-common` is updated to read from
  `templates/gitattributes/lfs-common.gitattributes` instead of the
  hard-coded slice. Existing repos that already track `*.svg` via LFS
  keep working (we never *untrack*); only future runs change.

### 4.4. Code-file coverage in defaults

In addition to binary LFS lines, the common template **must** ship
`text eol=…` directives for known source extensions so cross-platform
checkouts get consistent line endings:

| Extension | Directive |
|---|---|
| `*.go`, `*.rs`, `*.py`, `*.rb`, `*.sh`, `*.yml`, `*.yaml`, `*.json`, `*.md` | `text eol=lf` |
| `*.cs`, `*.ps1`, `*.psm1`, `*.bat`, `*.cmd` | `text eol=crlf` |
| `*.ts`, `*.tsx`, `*.js`, `*.jsx`, `*.css`, `*.html` | `text eol=lf` |
| `*.java`, `*.kt`, `*.gradle` | `text eol=lf` |
| `*.c`, `*.h`, `*.cpp`, `*.hpp` | `text eol=lf` |

Per-language template files extend or override these.

### 4.5. Marker block

Same as §3.2 step 4, with marker comments:

```
# >>> gitmap-attributes (do not edit between markers) >>>
…
# <<< gitmap-attributes <<<
```

---

## 5. Merge → compact algorithm

When the user passes multiple languages (e.g., `go,rust`), we
concatenate the templates and then **compact line-by-line**.

```
input:  ["templates/gitignore/go.gitignore",
         "templates/gitignore/rust.gitignore"]

step 1 (merge):
  read each file in order, append to a string builder, separated
  by a single "\n# --- <lang> ---\n" header comment.

step 2 (compact, single pass):
  seen := map[string]struct{}{}
  for each line in the merged buffer:
    trimmed := strings.TrimRight(line, " \t\r\n")
    key := trimmed                  # comments and blanks ARE keyed
    if _, dup := seen[key]; dup {
      continue                      # drop duplicate (incl. duplicate comments)
    }
    seen[key] = struct{}{}
    out.WriteString(trimmed + "\n")
```

Properties:

- O(n) in total bytes; runs in milliseconds even for the full
  github/gitignore corpus.
- **Comments dedupe too** — if both `go.gitignore` and `rust.gitignore`
  start with `# Compiled binaries`, only the first occurrence survives.
- **Order is preserved** — first writer wins. Argument order matters
  (`go,rust` ≠ `rust,go` only in section ordering).
- The first-merge result MAY be cached to
  `<binary-dir>/templates/.cache/<sha>.txt` so a second invocation
  with the same arg set is a single file read. Optional optimization;
  ship without it first.

The same algorithm applies to `.gitattributes`.

---

## 6. Pretty-Markdown CLI renderer (display polish)

### 6.1. Problem

Today, when gitmap renders embedded markdown (help text, changelog,
`gitmap help <cmd>`) to the terminal, it prints fenced code blocks
verbatim. Often the fenced block's body is **identical** to the
surrounding paragraph text — duplicating the same sentence as both
prose and a code block. That's noise.

### 6.2. Rule

In the markdown→ANSI renderer:

1. **Collapse redundant fences.** If a fenced block (` ``` … ``` `) is
   immediately preceded or followed by a paragraph whose normalized
   text equals the block's normalized body (case-insensitive,
   whitespace-collapsed), **drop the fence** and render the paragraph
   once, in **`ColorYellow`**, to signal "this is the canonical phrasing".
2. **Highlight quoted strings.** Any run matching `"…"` (no embedded
   quotes) inside a paragraph renders in a distinct color
   (`ColorCyan`). Inline code (` `…` `) keeps its existing styling
   (`ColorDim` + bold). Order of precedence: inline-code > quoted >
   plain.
3. Untouched: headings, lists, links, tables, true code blocks (where
   body ≠ surrounding text).

### 6.3. Where it lives

- New file `gitmap/render/pretty.go` (≤ 200 lines).
- Hooked into the existing markdown printer used by
  `gitmap/helptext/` and `gitmap/cmd/changelog.go`. We do **not**
  introduce a new markdown library — extend the current line-based
  renderer.
- Behavior is opt-out via env var `GITMAP_PRETTY=0` for users who
  pipe output to non-TTY consumers (the renderer also auto-disables
  ANSI when `!isatty(stdout)`).

---

## 7. Constants & errors

Add to `gitmap/constants/`:

- `constants_cli.go`:
  - `CmdAddIgnore       = "add ignore"`
  - `CmdAddIgnoreShort  = "ai"`           (TBD)
  - `CmdAddAttributes   = "add attributes"`
  - `CmdAddAttrShort    = "aa"`           (TBD)
- `constants_templates.go` (NEW):
  - `TemplatesDir       = "templates"`
  - `IgnoreSubdir       = "gitignore"`
  - `AttrSubdir         = "gitattributes"`
  - `MarkerIgnoreOpen   = "# >>> gitmap-ignore (do not edit between markers) >>>"`
  - `MarkerIgnoreClose  = "# <<< gitmap-ignore <<<"`
  - `MarkerAttrOpen     = "# >>> gitmap-attributes (do not edit between markers) >>>"`
  - `MarkerAttrClose    = "# <<< gitmap-attributes <<<"`
- `constants_errors.go`:
  - `ErrTemplateUnknownLang`
  - `ErrTemplateReadFailed`
  - `ErrTemplateMergeFailed`
  - `ErrTemplateWriteFailed`

All follow the Code Red zero-swallow rule and include the offending
template path / language token in the message.

---

## 8. Help text

Three new files under `gitmap/helptext/` (≤ 120 lines each, per
`mem://features/command-help-system`):

- `add-ignore.md`
- `add-attributes.md`
- `templates.md` (overview of the folder, layout, override rules)

Each includes a 3–8 line realistic terminal simulation block.

---

## 9. Out of scope (deliberate non-goals)

- `gitmap templates reset` / `gitmap templates list` — separate spec.
- Auto-detecting languages in the repo and selecting templates without
  args — separate spec (`gitmap add ignore --auto`).
- Removing existing entries from a user's `.gitignore`. We only **add**.
- LFS migration of already-committed binaries — out of scope; user
  must still run `git lfs migrate import` per
  `gitmap/helptext/lfs-common.md`.
- A new markdown parser. The pretty renderer extends the existing
  line-based printer.

---

## 10. Acceptance criteria (for future implementation PR)

1. Fresh repo, no `.gitignore`:
   `gitmap add ignore go,rust` writes a marker-blocked `.gitignore`
   containing the deduped union of both templates.
2. Existing `.gitignore` with user lines:
   `gitmap add ignore python` inserts the marker block at the top,
   user lines moved below `# user entries`, no user line lost.
3. Re-running `gitmap add ignore python` is a no-op (marker block
   replaced byte-identical).
4. `gitmap add attributes` no longer LFS-tracks `*.svg`; it emits
   `*.svg text eol=lf linguist-detectable=true` instead.
5. `gitmap lfs-common` continues to work and reads its pattern list
   from `templates/gitattributes/lfs-common.gitattributes` (no
   functional regression).
6. Help text `gitmap help add-ignore` renders, with one fenced block
   collapsed to a yellow line where applicable, and `"quoted strings"`
   shown in cyan.
7. `gitmap add ignore go,rust --dry-run` prints the final merged file
   to stdout and exits 0 without modifying disk.
8. Unknown language: `gitmap add ignore lolcode` exits non-zero with
   `ErrTemplateUnknownLang` and lists available languages from disk.

---

## 11. Related specs / memory

- `mem://tech/database-location` — anchoring rule for `<binary-dir>`.
- `mem://features/command-help-system` — help text constraints.
- `mem://tech/code-red-error-management` — error reporting format.
- `gitmap/cmd/lfscommon.go` — current hard-coded LFS list to migrate.
- `gitmap/cmd/setup.go` — current `ensureGitignoreStep` to align with
  new marker-block convention.
