# gitmap templates

Discover and inspect the embedded `.gitignore` / `.gitattributes` template
corpus that powers `gitmap add ignore`, `gitmap add attributes`, and
`gitmap add lfs-install`. Two read-only subcommands.

## Alias

tpl

## Subcommands

| Subcommand | Alias | Purpose |
|------------|-------|---------|
| `templates list` | `tl` | Print every available template with its KIND, LANG, SOURCE, and PATH |
| `templates show <kind> <lang>` | `ts` | Write a single resolved template (overlay > embed) to stdout |

## Kinds

| Kind | File extension | Used by |
|------|----------------|---------|
| `ignore` | `.gitignore` | `gitmap add ignore` |
| `attributes` | `.gitattributes` | `gitmap add attributes` |
| `lfs` | `.gitattributes` (LFS lines) | `gitmap add lfs-install` |

## Source resolution

Every template lookup checks two locations, in order:

1. **User overlay** — `~/.gitmap/templates/<kind>/<lang>.<ext>`
   Source label: **`user`**.
2. **Embedded corpus** — bundled into the gitmap binary via `go:embed`.
   Source label: **`embed`**.

The first hit wins. `templates list` shows which one each entry resolves
to so you can tell at a glance which templates you've forked.

## Examples

### Example 1: List every template

    gitmap templates list

**Output:**

    KIND        LANG            SOURCE  PATH
    ignore      common          embed   assets/ignore/common.gitignore
    ignore      go              embed   assets/ignore/go.gitignore
    ignore      node            user    /home/me/.gitmap/templates/ignore/node.gitignore
    ignore      python          embed   assets/ignore/python.gitignore
    attributes  common          embed   assets/attributes/common.gitattributes
    attributes  go              embed   assets/attributes/go.gitattributes
    lfs         common          embed   assets/lfs/common.gitattributes

The `node` row above shows what a forked template looks like: SOURCE flips
from `embed` to `user` and PATH points at the absolute overlay file.

### Example 2: Print a single template to stdout

    gitmap templates show ignore go

**Output:** the raw bytes of `ignore/go.gitignore` (overlay if present,
otherwise embed), audit-trail header included:

    # source: github/gitignore
    # kind: ignore
    # lang: go
    # version: 1
    *.exe
    *.test
    *.out
    ...

### Example 3: Diff your overlay against the curated embed

    gitmap templates show ignore node > /tmp/curated-node.gitignore
    diff ~/.gitmap/templates/ignore/node.gitignore /tmp/curated-node.gitignore

`templates show` always resolves overlay-first — but **once your overlay
file exists**, the embed copy is the only way to recover the curated
bytes. Pipe `templates show` through `diff` to audit your fork before
re-syncing.

### Example 4: Use the short aliases

    gitmap tpl tl
    gitmap tpl ts attributes common

Both `tpl` (the umbrella alias) and `tl` / `ts` (the per-subcommand
aliases) round-trip identically with their long forms.

## How forking works

To customize a template, copy the embedded version to the overlay path
and edit it:

    mkdir -p ~/.gitmap/templates/ignore
    gitmap templates show ignore python > ~/.gitmap/templates/ignore/python.gitignore
    $EDITOR ~/.gitmap/templates/ignore/python.gitignore

Subsequent `gitmap add ignore python` calls (and any future `add` flow
that resolves `ignore/python`) will pick up your overlay automatically.
`gitmap templates list` will report SOURCE=`user` for that row.

To revert a fork, just delete the overlay file — the next resolve falls
back to `embed`.

## Pretty rendering

`templates show` writes raw bytes by default — perfect for the diff and
redirect workflows above. When the resolved template is **markdown**
(`.md` / `.markdown`) **and** stdout is a real TTY, the output is routed
through the same pretty markdown renderer used by `gitmap help`:

- Cyan `"double quotes"` for emphasized terms.
- Yellow `→ collapsed` lines when a fenced block restates the
  preceding paragraph.
- Muted subtitles under headings, indented bodies.

Today the embedded corpus is `.gitignore` / `.gitattributes` only, so
this kicks in for **markdown overlays** you drop into
`~/.gitmap/templates/<kind>/<lang>.md` and for any future markdown
templates added to the embed.

Two opt-outs, both honored even on a TTY:

- `--raw` — per-invocation flag. Use when you need byte-faithful output
  inside a TTY session (e.g. `templates show notes intro.md --raw | sha256sum`).
- `GITMAP_NO_PRETTY=1` — environment opt-out, shared with `gitmap help`.
  Set it once in your shell profile to disable pretty rendering across
  the whole CLI.

Pipes and redirects automatically bypass the renderer (stdout is no
longer a TTY), so `templates show foo bar > out.md` always writes the
unmodified bytes.

## Notes

- `templates list` and `templates show` are pure reads. They never
  write to disk and never invoke git.
- Unknown `<kind>` or `<lang>` arguments to `templates show` exit 1
  with `template not found: kind=… lang=…`.
- The embedded corpus is versioned via the `# version: N` header on
  each file. When that integer bumps, gitmap re-resolves cleanly — but
  user overlays are **never** auto-upgraded; you decide when to refresh.

## See Also

- [add lfs-install](add-lfs-install.md) — Install Git LFS hooks + merge `lfs/common.gitattributes`
- [lfs-common](lfs-common.md) — Per-pattern `git lfs track` flow (no template)
- [setup](setup.md) — Configure Git global settings
- [doctor](doctor.md) — Diagnose binary, PATH, and config issues
