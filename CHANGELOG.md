# Changelog

## v3.12.1 — (2026-04-20) — AST registry parity + spec cross-links + legacy-field test cleanup

### Added

- **AST-derived `topLevelCmds()` registry parity test** — `gitmap/constants/cmd_constants_parity_test.go` adds `TestTopLevelCmdRegistryMatchesAST`, which uses `go/parser` to walk every `gitmap/constants/constants_*.go`, collects every `Cmd*` constant declared inside a `// gitmap:cmd top-level` block (minus those tagged `// gitmap:cmd skip`), and asserts the resulting set is exactly equal to the manual `topLevelCmds()` registry consumed by `TestTopLevelCmdConstantsAreUnique` / `TestTopLevelCmdAliasesAreUnique`. The registry can no longer drift silently — adding a new top-level `Cmd*` without registering it (or vice versa) fails CI with a clear "missing from registry" / "registered but not declared" diff.
- **Spec cross-links from CLI overview** — `spec/01-app/02-cli-interface.md` and `spec/01-app/38-command-help.md` gained a `> **Related:**` callout under the H1 pointing at `spec/01-app/99-cli-cmd-uniqueness-ci-guard.md`, so future contributors discover the uniqueness contract and the 6-step handoff checklist directly from the CLI overview and the help-system spec.
- **Spec §5 implementation note** — `spec/01-app/99-cli-cmd-uniqueness-ci-guard.md` updated to mark the AST parity test as implemented (no longer "future hardening") with the file path and v3.12.1 history entry.

### Fixed

- **Stale `Draft` / `PreRelease` `ReleaseMeta` / `Options` field references in tests** — `gitmap/release/metadata_test.go` and `gitmap/tests/release_test/skipmeta_test.go` still constructed `ReleaseMeta{Draft: …, PreRelease: …}` and `release.Options{Draft: …}` using the pre-v15 field names, breaking `go vet` / `go build` with `unknown field Draft in struct literal`. Renamed both to the v15 `IsDraft` / `IsPreRelease` form, matching every production caller. The legacy-JSON compat shim in `release/metadata.go::ReadReleaseMeta` (which still reads the old `draft` / `preRelease` JSON keys) is intentionally untouched and remains the supported migration path for v3.4.x metadata files on disk.
- **`go vet` `non-constant format string`** in `gitmap/cmd/probe.go:127` — `fmt.Fprintf(os.Stderr, result.Error+"\n")` triggered the printf-check because the format string was constructed at runtime from a struct field. Reshaped the call to `fmt.Fprintf(os.Stderr, "%s\n", result.Error)` so the format string is a compile-time constant.

### Verified

- Full-repo audit for residual legacy-field callers: every `\.(Draft|PreRelease)\b` and `^\s*(Draft|PreRelease)\s*:` match outside of (a) `release.Version.PreRelease` (semver suffix — different struct), (b) `store/migrate_v15phase5.go` (the rename migration itself), (c) `release/metadata.go::ReadReleaseMeta` (the JSON backward-compat overlay), and (d) `--draft` user-facing CLI flag strings was confirmed to be either intentional or already migrated. No further call sites need updating.

## v3.12.0 — (2026-04-20) — Pinned-version release snippet + gitmap-v4 rename

### Added

- **Pinned-version install snippet on the GitHub release page** — the release publisher (`gitmap/release/installsnippet.go`, wired into `workflowgithub.go::uploadToGitHub`) now auto-appends a markdown block containing PowerShell + bash one-liners that hard-code the just-published tag. Idempotent via a hidden `<!-- gitmap-pinned-install-snippet:<tag> -->` HTML marker. Anyone copying the snippet from `…/releases/tag/v3.12.0` installs exactly v3.12.0 — never "latest", never a `-v<N+1>` sibling repo. Template lives in `constants_release.go` as `ReleaseSnippetTemplate` / `ReleaseSnippetMarker`.
- **Pinned-version short-circuit in installer scripts** — `gitmap/scripts/install.ps1` and `install.sh` gained a new branch in their discovery prelude: when `-Version <tag>` (PowerShell) or `--version <tag>` (bash) is supplied, the installer now skips both the `releases/latest` API call **and** the versioned-repo `-v<N>` discovery probe, downloading `…/releases/download/<tag>/…` directly. Closes the gap where a snippet copied from a v3.x release page could silently jump to the v4 repo's latest tag.
- **Spec doc** `spec/07-generic-release/08-pinned-version-install-snippet.md` — full NEA/AI handoff contract: rendered snippets, installer-side flag matrix, release-cutting checklist, and a CI test contract for future work.

### Changed

- **Repo rename `gitmap-v3` → `gitmap-v4` across the entire codebase** — every Go constant (`SourceRepoCloneURL`, `SelfInstallRemotePwsh/Bash`, `GitmapRepoPrefix`, install hint URLs), every install/uninstall script (`install.ps1`, `install.sh`, `install-quick.ps1`, `install-quick.sh`, `uninstall-quick.*`), every spec doc under `spec/01-app/` and `spec/07-generic-release/`, every helptext markdown, the README, the React `src/data/*.ts` files, GitHub workflows, and historical CHANGELOG entries were rewritten via `sed -i 's/gitmap-v3/gitmap-v4/g'`. The only remaining `gitmap-v3` references are inside `.gitmap/` artifacts, which are immutable per project policy.

## v3.11.1 — (2026-04-20) — Alias-collision CI guard

### Added

- **Alias-collision uniqueness test** — extended `gitmap/constants/cmd_constants_test.go` with `TestTopLevelCmdAliasesAreUnique`, which iterates every top-level `Cmd*` constant and fails when two distinct identifiers share the same short-form value (string length ≤ 2). Catches future regressions like a hypothetical `CmdFooAlias = "ls"` shadowing the existing `CmdListAlias`, before they reach the build phase. Companion `TestTopLevelCmdConstantsAreUnique` covers full-length command-name collisions. Manual `topLevelCmds()` registry is the source of truth and excludes anything marked `// gitmap:cmd skip`.

## v3.11.0 — (2026-04-19) — Constants hygiene + Phase 1.4 migration fix

### Fixed

- **v15 Phase 1.4 migration** — `GoProjectMetadata` and `PendingTask` rebuilds failed on databases first created at v3.5.0+ with `SQL logic error: no such column: Id`. Both tables were already singular before v15, so the canonical `CREATE TABLE IF NOT EXISTS` pass produced the v15-shaped table (with `{Table}Id` PK) before the rebuild ran, leaving no `Id` column to SELECT. Added `adaptOldColumnList()` in `gitmap/store/migrate_v15rebuild.go` that detects the existing PK shape via `columnExists()` and rewrites the leading `Id` token in `OldColumnList` to `{Table}Id` when needed. Idempotent and a no-op for genuine legacy → v15 paths.
- **`go vet` `non-constant format string`** in `gitmap/movemerge/finalize.go:50` — `logErr` was inferred as a printf-style wrapper. Reshaped `logErr(prefix, msg string)` to accept a pre-formatted message and moved `fmt.Sprintf(constants.ErrMMPushFailFmt, sha)` to the call site so the printf-check never triggers.
- **Unused-import build break** in `gitmap/store/migrations.go` — removed orphaned `"github.com/user/gitmap/constants"` import left over from a prior refactor.
- **`CmdReleaseAlias` Go redeclaration** — same name was bound to `"r"` (in `constants_cli.go`) and `"release-alias"` (in `constants_releasealias.go`). Renamed the `constants_cli.go` constant to `CmdReleaseShort` so the `release-alias` family owns `CmdReleaseAlias` exclusively.
- **`cd` / `go` constant collision** — `CmdCDCmd` (`"cd"`) and `CmdCDCmdAlias` (`"go"`) in `constants_cli.go` shadowed `CmdCD` / `CmdCDAlias` in `constants_cd.go`. Removed the duplicates and repointed `gitmap/cmd/rootdata.go` dispatch at the canonical constants.

### Added

- **CI uniqueness test** — `gitmap/cmd/cmdconstants_unique_test.go` (+ helpers in `cmdconstants_unique_helpers_test.go`) parses every `gitmap/constants/constants_*.go`, applies the same `gitmap:cmd top-level` / `gitmap:cmd skip` markers used by `completion/internal/gencommands`, and fails the test suite when two distinct `Cmd*` identifiers claim the same string value. Catches future redeclarations and dispatch shadowing at CI time before they reach the build phase.
- **Parallel pull worker pool** (`gitmap/cmd/pullparallel.go`) — buffered-channel pool with `sync.WaitGroup` and a mutex around the non-thread-safe `BatchProgress` tracker. Opt-in via `--parallel <N>`.
- **`--only-available` pull pre-filter** (`gitmap/cmd/pullfilter.go`) — intersects the target repo list with `FindNext` results so `gitmap pull --only-available` skips repos that have no new tags. Fail-open: falls back to a full pull if the database is inaccessible.
- **`gitmap probe` and `gitmap sf` help docs** — `gitmap/helptext/probe.md` and `gitmap/helptext/sf.md` (synopsis, flags, examples, 3–8 line realistic terminal simulation), discoverable via `gitmap help probe` / `gitmap help sf`.

### Changed

- **`constants_cli.go` size reduction** — extracted the `Shorthand*` group into `gitmap/constants/constants_clone.go` and the cross-command `Flag*` values into a new `gitmap/constants/constants_globalflags.go`. `constants_cli.go` is now 188 lines (under the 200-line guideline).

## v3.5.0 — (2026-04-19) — v15 Database Naming Alignment (Phase 1 complete)

### Changed

- **Phase 1 of the v15 database naming migration is complete.** All 22 SQLite tables now follow the strict v15 convention from <https://github.com/alimtvnetwork/coding-guidelines-v15/blob/main/spec/04-database-conventions/01-naming-conventions.md>: PascalCase + **singular** table names, `{TableName}Id` primary keys, foreign keys that match the referenced PK name, `IsX` prefix for booleans, and abbreviations treated as words (`SshKey` not `SSHKey`, `CsharpProjectMetadata` not `CSharpProjectMetadata`).
- **Renamed tables** (legacy → v15): `Repos`→`Repo`, `Groups`→`Group`, `GroupRepos`→`GroupRepo`, `Releases`→`Release`, `Aliases`→`Alias`, `Bookmarks`→`Bookmark`, `Amendments`→`Amendment`, `CommitTemplates`→`CommitTemplate`, `Settings`→`Setting`, `SSHKeys`→`SshKey`, `InstalledTools`→`InstalledTool`, `TempReleases`→`TempRelease`, `ZipGroups`→`ZipGroup`, `ZipGroupItems`→`ZipGroupItem`, `ProjectTypes`→`ProjectType`, `DetectedProjects`→`DetectedProject`, `GoProjectMetadata` (kept), `GoRunnableFiles`→`GoRunnableFile`, `CSharpProjectMeta`→`CsharpProjectMetadata`, `CSharpProjectFiles`→`CsharpProjectFile`, `CSharpKeyFiles`→`CsharpKeyFile`. `RepoVersionHistory`, `CommandHistory`, `TaskType`, `PendingTask`, `CompletedTask` were already singular and only got `{TableName}Id` PK renames.
- **Renamed columns**: every legacy `Id` PK is now `{TableName}Id` (e.g., `Repo.RepoId`, `Release.ReleaseId`, `CsharpProjectMetadata.CsharpProjectMetadataId`). Foreign keys updated to match (e.g., `GoRunnableFile.GoProjectMetadataId`, `CsharpProjectFile.CsharpProjectMetadataId`). `Release.Draft` → `Release.IsDraft` and `Release.PreRelease` → `Release.IsPreRelease` complete the IsX boolean-prefix consistency (`IsLatest` was already correct).
- **Migration safety contract** (applies to every Phase 1.1–1.5 rebuild):
  1. Detect-then-act on every legacy plural — fresh installs are no-ops.
  2. `PRAGMA foreign_keys=OFF` for the duration of each table rebuild.
  3. Row-count parity check between old and new on every rebuild — abort + return on mismatch.
  4. Legacy plural names retained as `LegacyTable*` constants and listed in `Reset()` so cleanup works at any migration state.
  5. SQLite-reserved word `Group` is double-quoted in every DDL/DML occurrence.
- **Go-side propagation**: `model.ReleaseRecord.Draft/PreRelease` → `IsDraft/IsPreRelease` (with JSON tags `isDraft`/`isPreRelease`); `release.Options.Draft` → `release.Options.IsDraft`; `release.ReleaseMeta.Draft/PreRelease` → `IsDraft/IsPreRelease`. `ReadReleaseMeta` includes a JSON overlay that accepts the legacy `"draft"`/`"preRelease"` keys so on-disk `.gitmap/release/*.json` files from v3.4.x and earlier still load.
- **CLI flag `--draft`** is intentionally retained (user-facing). Internal struct fields use the v15 `IsX` naming.

### Added

- New shared migration infrastructure in `gitmap/store/migrate_v15rebuild.go` — generic `runV15Rebuild(spec)` helper using a `v15RebuildSpec` struct (OldTable, NewTable, NewCreateSQL, OldColumnList, NewColumnList, StartMsg, DoneMsg). Drives all 22 table rebuilds.
- New phase migrators wired into `store.Migrate()` in dependency-safe order:
  - `migrate_v15phase2.go` — Group, Release, Alias, Bookmark + GroupRepo FK-text rebuild.
  - `migrate_v15phase3.go` — Amendment, CommitTemplate, Setting, SshKey, InstalledTool, TempRelease.
  - `migrate_v15phase4.go` — ZipGroup family, Project family (incl. CSharp→Csharp), Task family, History tables.
  - `migrate_v15phase5.go` — `Release.Draft`→`IsDraft`, `Release.PreRelease`→`IsPreRelease` (column rename via the same rebuild infrastructure).
- Pre-rename column patches for very old installs: `preV15Phase2EnsureReleaseColumns()` (Source/Notes on legacy `Releases`), `migrateZipGroupItemPaths()` and `migrateTRCommitSha()` already targeted legacy plurals before the v15 rebuilds copied the data.
- Regenerated `spec/01-app/gitmap-database-erd.mmd` to reflect every v15 table name, PK, FK, and `IsDraft`/`IsPreRelease` boolean.
- Updated `spec/12-consolidated-guidelines/11-database.md` with the v15 naming conventions table (singular + `{TableName}Id` + `IsX` boolean prefix + reserved-word quoting + abbreviation rules), with a link to the upstream v15 spec.

### Notes

- This release is purely a naming alignment — no new commands, no behavior changes for end users beyond the schema. Existing databases upgrade in place via the idempotent rebuild migrators; rollback is via `gitmap db-migrate` against an older binary's CREATE statements after restoring a DB backup.
- Phase 2 (ScanFolder, VersionProbe, `gitmap find-next`) and Phase 3 (parallel `pull`, bulk `cn next all`) remain on the roadmap.

## v3.0.0 — (2026-04-19)

### Added

- `gitmap as [alias-name] [--force|-f]` (alias `s-alias`) — tag the **current** Git repository with a short alias and persist it in the active-profile SQLite database. Resolves the repo top-level via `git rev-parse --show-toplevel`, builds a single-repo `ScanRecord` through the existing `mapper.BuildRecords()` pipeline (so the upserted row matches the schema other commands use), upserts into `Repos`, then maps `alias-name → Repos.Id` in the alias store. When `alias-name` is omitted the repo folder basename is used. Refuses to clobber an existing alias unless `--force` is passed. Exits 1 with a CWD-aware message when invoked outside a Git repo.
- `gitmap release-alias <alias> <version>` (alias `ra`) — release a previously-aliased repo from **any** working directory. Resolves alias → absolute path via the alias store, `os.Chdir`s into the repo, runs the existing `runRelease` pipeline (lint → test → tag → push → assets), then restores the original CWD via `defer`. Forwards `--dry-run` to `runRelease` for safe previews.
- `gitmap release-alias-pull <alias> <version>` (alias `rap`) — thin sugar for `release-alias --pull`. Runs `git pull --ff-only` in the resolved repo before releasing; hard-fails on non-fast-forward (never tags on top of a divergent tree). The flag remains canonical, the verb is sugar.
- **Auto-stash semantics for `release-alias`**: dirty working trees are auto-stashed (`git stash push --include-untracked -m "gitmap-release-alias autostash <alias>-<version>-<unix-ts>"`) before the release runs and popped on exit via `defer`, so the stash always fires — including when `runRelease` aborts. The pop locates the stash by **label match** against `git stash list` (not by `stash@{0}`), so a concurrent `git stash` from another process never causes us to pop the wrong entry. A failed pop warns only — the user's tree is still recoverable via `git stash list` / `git stash apply`. Bypass with `--no-stash` (intended for CI runners that always start clean and want to fail loudly on unexpected dirt).
- `gitmap db-migrate` (alias `dbm`) — explicit, idempotent schema migration command. Re-runs every `CREATE TABLE IF NOT EXISTS` and column-migration step on the active profile DB. Now invoked automatically at the end of `gitmap update` so a freshly-updated binary never has to repair the database on its first real run. `--verbose` prints extra context.
- New shared migration helpers in `gitmap/store/migrations.go`: `columnExists(table, column)`, `tableExists(table)`, `isBenignAlterError(err)`, and `logMigrationFailure(table, column, action, err, stmt)` — every warning now names the table, column, and action so issues can be diagnosed without trial-and-error.
- New files: `gitmap/cmd/{as.go, asops.go, releasealias.go, releasealias_git.go, dbmigrate.go}`, `gitmap/constants/{constants_as.go, constants_releasealias.go, constants_dbmigrate.go}`, `gitmap/store/migrations.go`, `gitmap/helptext/{as.md, release-alias.md, release-alias-pull.md, db-migrate.md}`, `spec/01-app/98-as-and-release-alias.md`.

### Changed

- **`migrateTRCommitSha` switched to detect-then-act.** Previously the migration always tried `ALTER TABLE TempReleases RENAME COLUMN "Commit" TO CommitSha` and only suppressed errors via brittle string-matching on `"no such column"`. On Unix builds where the SQLite driver formats the error slightly differently (or the table is fresh and only has `CommitSha`), the warning leaked through with the cosmetic `no such column: ""Commit""` message. The migration now uses `PRAGMA table_info(TempReleases)` to check whether `Commit` actually exists before attempting the rename, eliminating the spurious warning entirely on every OS regardless of driver wording.
- **Generator switched from explicit allowlist to marker-comment opt-in.** `gitmap/completion/internal/gencommands/main.go` no longer maintains a `sourceFiles` list or a `skipNames` map. Instead it scans every `../constants/*.go` automatically and includes only `const (...)` blocks whose doc comment contains `// gitmap:cmd top-level`. Individual specs inside an opted-in block can be excluded with a trailing `// gitmap:cmd skip` line comment (used for subcommand IDs like `"create"` / `"add"` shared across `gitmap group`). Domain owners now control inclusion locally without ever editing the generator. Added markers across 40 const blocks in 34 constants files (52 skip annotations mirror the previous policy exactly); `allcommands_generated.go` regenerates byte-for-byte identically (143 entries).
- `gitmap/cmd/update.go::runUpdateRunner` now calls `runPostUpdateMigrate()` after the binary swap completes, so every `gitmap update` finishes by running migrations. Best-effort: failures warn but do not block (the user may have an in-flight DB lock or a read-only environment).
- `gitmap/completion/completion.go::manualExtras` is now empty with an updated doc comment pointing future contributors at the marker convention instead of the old `sourceFiles` + `skipNames` instructions.
- All migration warnings (`addColumnIfNotExists`, `migrateZipGroupItemPaths` data-copy step, `migrateTRCommitSha`) now route through `isBenignAlterError` for a uniform suppression policy: `no such column`, `no such table`, `duplicate column`, and `already exists` are all benign on fresh installs.

### CI

- Added a `generate-check` job to `.github/workflows/ci.yml` that runs `go generate ./...` in `gitmap/` and fails with `git diff --exit-code` (printing the drifted file list and the fix command) if any generated file is out of sync with the constants. Wired into `test-summary`'s `needs` so the SHA-passthrough cache won't mark a run green unless the drift check also passed.

### Notes

- The original task description asked for a bump to `v2.97.0`; we are already at `v3.0.0` from the preceding `db-migrate` and marker-comment work, so the version was kept and the changelog rolled into a single v3.0.0 entry covering `as`, `release-alias`, `release-alias-pull`, `db-migrate`, the migration hardening, the generator refactor, and the CI drift check.

---

## Migration guide — v2.x → v3.0.0 (constants contributors)

If you maintain a custom `constants_*.go` file in `gitmap/constants/` that exposes command IDs for shell tab-completion, you must opt-in explicitly using marker comments.

### What changed
- **Old (v2.x):** The generator (`internal/gencommands/main.go`) relied on a hard-coded `sourceFiles` list and a `skipNames` map. Adding a new command required editing the generator.
- **New (v3.0.0):** The generator scans every `constants/*.go` file automatically. Inclusion is controlled locally via comments.

### What you need to do

1. Open your `constants_*.go` file.
2. Locate the `const (...)` block containing your `Cmd*` string constants.
3. Add `// gitmap:cmd top-level` to the block's **doc comment** (the comment immediately above `const`).
4. If any constant in that block is a *subcommand* (e.g., `"create"` or `"add"` used only inside `gitmap group`), add a trailing line comment `// gitmap:cmd skip` to that specific spec.

**Example:**

```go
// gitmap:cmd top-level
// Bookmark commands.
const (
    CmdBookmarkAdd    = "add"    // gitmap:cmd skip
    CmdBookmarkList   = "list"
    CmdBookmarkRemove = "remove"
)
```

5. Re-run `go generate ./...` in `gitmap/` to regenerate `allcommands_generated.go`.
6. Verify with `git diff` — only your new command values should appear; no manual edits to the generator needed.

### Verification
- CI now runs a `generate-check` job that fails if `allcommands_generated.go` drifts from the constants. If your PR fails this check, the error message prints the exact command to fix it locally.

---

## v2.98.0 — (2026-04-18)

### Added

- `gitmap mv LEFT RIGHT` (alias `move`) — moves LEFT's contents into RIGHT (excluding `.git/`), then deletes LEFT entirely. Both endpoints can be local folders or remote git URLs (with optional `:branch` suffix); URL endpoints are cloned (or fast-forward pulled if already on disk with matching origin), and after the move the RIGHT-side URL is committed (`gitmap mv from <LEFT-display>`) and pushed.
- `gitmap merge-both LEFT RIGHT` (alias `mb`) — bidirectional file-level merge: each side gains every file the other has but it doesn't; conflicting files (different content on both sides) trigger the `[L]eft / [R]ight / [S]kip / [A]ll-left / [B]all-right / [Q]uit` interactive prompt.
- `gitmap merge-left LEFT RIGHT` (alias `ml`) — one-way merge that writes only into LEFT (RIGHT is read-only). With `-y`, RIGHT wins by default.
- `gitmap merge-right LEFT RIGHT` (alias `mr`) — one-way merge that writes only into RIGHT (LEFT is read-only). With `-y`, LEFT wins by default.
- Bypass flags shared by all four merge commands: `-y` / `--yes` / `-a` / `--accept-all` skip the prompt; `--prefer-left`, `--prefer-right`, `--prefer-newer`, `--prefer-skip` override the per-command default policy. `merge-both -y` defaults to `--prefer-newer`.
- URL-side commit/push controls: `--no-push` (commit but skip push), `--no-commit` (copy files but skip both). `--force-folder` replaces a folder whose origin doesn't match the requested URL. `--pull` opt-in for `git pull --ff-only` on folder endpoints. `--dry-run` prints every action and writes nothing. `--include-vcs` and `--include-node-modules` override the default ignore list.
- New `gitmap/movemerge/` package with focused files (<200 lines each, <15 lines per function): `types.go`, `endpoint.go` + `endpoint_test.go` (URL classification + `:branch` suffix + scp-style `git@host:user/repo` preservation), `walk.go` (default ignore list `.git/` / `node_modules/` / `.gitmap/release-assets/`), `copy.go` (mode-preserving file copy with symlink replication), `conflict.go` + `conflict_test.go` (L/R/S/A/B/Q resolver with sticky All-Left/All-Right and `--prefer-newer` mtime tie-break), `diff.go` (SHA-256 classification into MissingLeft / MissingRight / Conflict / Identical), `git.go` (clone / pull --ff-only / add-commit-push), `resolve.go` (full endpoint resolver with origin-match check), `guard.go` (same-folder + nested-ancestor protection), `merge.go`, `move.go`, `finalize.go` (URL-side commit + push), `log.go` (structured `[mv]` / `[merge-*]` prefix lines).
- CLI wiring: `cmd/move.go`, `cmd/merge.go`, `cmd/movemergeflags.go` (shared flag binder), `cmd/dispatchmovemerge.go` hooked into `cmd/root.go`. New constants in `constants/constants_movemerge.go` (command IDs, aliases, flag names, log prefixes, commit message templates, error formats) plus `GitAddCmd`, `GitAddAllArg`, `GitCommitCmd`, `GitMessageArg` reused for the post-merge git plumbing.

### Notes

- `mv` does NOT prompt — its semantic is destructively "move-and-delete-LEFT". Use `merge-right` for the safer copy-with-prompt variant.
- Same-folder and nested-folder protection trips before any file write: LEFT and RIGHT may not resolve to the same absolute path, and neither may be a strict ancestor of the other on disk.
- `gitmap diff LEFT RIGHT` (added in v2.97.0) is the recommended dry-run preview before `gitmap merge-both` — every conflict it lists will trigger the interactive prompt.


### Added

- `gitmap diff LEFT RIGHT` (alias `df`) — read-only preview of what `gitmap merge-both / merge-left / merge-right` would change between two folders. Lists conflicts (different content on both sides), missing-on-LEFT, missing-on-RIGHT, and (optionally) identical files. Writes nothing, commits nothing, pushes nothing.
- Flags: `--json` (machine-readable output with `{summary, entries}` payload), `--only-conflicts`, `--only-missing`, `--include-identical`, `--include-vcs`, `--include-node-modules`. Honours the same default ignore list as `merge-*` (`.git/`, `node_modules/`, `.gitmap/release-assets/`).
- New `gitmap/diff/` package: `endpoint.go` (folder-only resolver — URL endpoints are intentionally rejected with a hint to clone first), `tree.go` (parallel walk + SHA-256 classification), `report.go` (text/JSON renderer + `Summary` tally). Unit tests cover all four diff kinds and the default ignore list.
- `gitmap/helptext/diff.md` and `gitmap/cmd/diff.go` + `gitmap/cmd/dispatchdiff.go` wire the command into the existing dispatcher chain in `root.go`.

### Notes

- `diff` is the recommended dry-run preview before `merge-both`: every conflict it lists will trigger the `[L]eft / [R]ight / [S]kip / [A]ll-left / [B]all-right / [Q]uit` prompt during merge-both.
- URL endpoints are rejected on purpose so `diff` remains strictly side-effect-free (no network, no clone, no temp folders). Clone first via `gitmap clone <url>`, then diff the resulting folder.


## v2.96.0 — (2026-04-18)

### Added

- Help text files for the move/merge command family: `gitmap/helptext/mv.md`, `merge-both.md`, `merge-left.md`, `merge-right.md`. Each follows the standard template (overview, alias, usage, flags, prerequisites, 3 examples with sample output, exit codes, see-also).
- `gitmap help <command>` now prints the embedded help file for any command (e.g. `gitmap help mv`, `gitmap help merge-both`). Previously `gitmap help` only showed the global usage banner. The lookup uses the existing `helptext.Print` function, so every command in `gitmap/helptext/*.md` is auto-discovered.

### Changed

- `dispatchUtility` in `gitmap/cmd/rootutility.go` now intercepts `gitmap help <name>` before falling through to the global usage printer. A small `isFlagToken` helper distinguishes `gitmap help --groups` (still goes to grouped usage) from `gitmap help mv` (prints `mv.md`).


## v2.95.0 — (2026-04-18)

### Added

- `gitmap setup print-path-snippet --shell <bash|zsh|fish|pwsh> --dir <path> --manager <label>` — emits the canonical marker-block PATH snippet to stdout. Used by `run.sh` and `gitmap/scripts/install.sh` so all three drivers produce byte-identical rc-file output. Single source of truth lives in `constants_pathsnippet.go`.
- `gitmap setup` now writes the marker-block snippet to the user's profile on every run (idempotent: rewrites the existing block in place, otherwise appends after a blank line). Different `--manager` values create coexisting blocks so `run.sh`, `installer`, and `gitmap setup` never overwrite each other.
- `setup.WritePathSnippet()` and `setup.RenderPathSnippet()` Go helpers with full unit-test coverage (`pathsnippet_test.go`, `pathsnippetwriter_test.go`).

### Changed

- `run.sh::register_on_path` and `gitmap/scripts/install.sh::add_path_to_profile` now ask the freshly-built/installed gitmap binary for snippet bytes via `gitmap setup print-path-snippet`. Inline heredocs remain as a first-run fallback only.

## v2.94.0 — (2026-04-18)

### Fixed

- `Get-LastRelease.ps1` reported the OLDEST version (e.g. `v2.82.0`) because `list-versions --limit 1` returns ascending order. Now sorts all versions descending and falls back to the binary's own `version` output if needed.
- Stale active PATH binary (e.g. `E:\bin-run\gitmap.exe`) is no longer kept alive by copying the new build into it. New `Migrate-StaleActiveBinary` helper deletes the stale binary, removes empty parent dirs, and strips the location from user PATH so future shells use the wrapped deploy target only.
- `powershell.json` `deployPath` is now rewritten after every successful deploy via `Sync-ConfigDeployPath` so the "Config binary:" readout reflects the actual install location and future runs default to the same target.

## v2.83.0 — (2026-04-16)

### Fixed

- `gitmap update-cleanup` now scans the active PATH directory, the PATH-derived deploy directory, the configured deploy directory, and the repo build output directory so stale `.old` backups are removed even when `powershell.json` points to an older location.
- `gitmap update-cleanup` now removes leftover `gitmap-update-*` artifacts from deploy/build locations in addition to `%TEMP%`, preventing handoff files from being left behind after update flows that switch between deploy targets.

## v2.82.0 — (2026-04-16)

### Fixed

- Regenerated `package-lock.json` to sync with `package.json` — resolves CI `npm ci` failure caused by missing entries for testing libs, axios, framer-motion, vitest, and other dependencies added without a lockfile refresh.

## v2.81.0 — (2026-04-16)

### Fixed

- `go-winres` CI icon size error — Windows `.ico` resources require images ≤256x256 but `icon.png` was 512x512. Created `icon-256.png` (LANCZOS resize) and updated `winres.json` to reference it.
- Documented root cause and prevention in `spec/08-generic-update/09-winres-icon-constraint.md`.

## v2.80.0 — (2026-04-16)

### Added

- Hidden `set-source-repo` command — persists source repo path to DB so `gitmap update` always uses the correct location after repo moves.
- Post-deploy repo path sync in `run.ps1` — automatically calls `set-source-repo` after every successful deploy to keep the DB current.
- Repo path sync spec (`spec/08-generic-update/08-repo-path-sync.md`) — documents the post-deploy sync pattern for AI implementers.
- Help file for `set-source-repo` command (`gitmap/helptext/set-source-repo.md`).

### Fixed

- `go-winres` CI failure — moved `winres.json` from `gitmap/` to `gitmap/winres/` where `go-winres make` expects it.

### Changed

- Cross-references updated in `02f-self-update-orchestration.md` and `03-self-update-mechanism.md` to include repo path sync spec.

## v2.78.0 — (2026-04-16)

### Added

- Console-safe handoff spec (`spec/08-generic-update/07-console-safe-handoff.md`) — documents the blocking `cmd.Run()` pattern that prevents terminal detachment during self-update on Windows.
- Installer banner now displays version number (`gitmap installer v1.0.0`).

### Changed

- `install.ps1`: `Resolve-Version` now prints full HTTP status code, URL, response body, and potential causes on GitHub API failure instead of a generic error.
- `gitmap-updater/cmd/github.go`: `fetchLatestTag` error output now includes URL, response body, and troubleshooting hints.
- Standardized lowercase "gitmap" branding across all installer output messages.

### Fixed

- `ShouldPrintInstallHint` now uses case-insensitive matching for GitHub repo URL detection.

## v2.76.0 — (2026-04-16)

### Added

- New `gitmap version-history` (`vh`) command displays all version transitions for the current repo with `--limit N` and `--json` flags.
- Full database ERD (Mermaid) added to `spec/01-app/gitmap-database-erd.mmd` covering all 22 tables including `RepoVersionHistory`.
- Updated `spec/01-app/59-clone-next.md` and `spec/01-app/87-clone-next-flatten.md` to reflect flatten-by-default behavior (no `--flatten` flag required).

---

## v2.75.0 — (2026-04-16)

### Added

- `gitmap clone-next` now flattens by default: clones into the base name folder (no version suffix) instead of the versioned folder name. For example, `gitmap cn v++` inside `macro-ahk-v15` clones `macro-ahk-v16` into `macro-ahk/`.
- `gitmap clone <url>` auto-flattens versioned URLs when no custom folder is given. `gitmap clone https://github.com/user/wp-onboarding-v13` clones into `wp-onboarding/`.
- New `RepoVersionHistory` SQLite table tracks every version transition (from/to version tags, numbers, and flattened path) with timestamps.
- `Repos` table gains `CurrentVersionTag` and `CurrentVersionNum` columns, updated on each clone-next operation.
- Version transitions are printed to terminal: `Recorded version transition v15 -> v16`.
- If the flattened target folder already exists during clone-next, it is automatically removed and re-cloned fresh.

---

## v2.74.0 — (2026-04-16)

### Added

- `gitmap doctor` now checks setup config resolution from the installed binary location and warns when `git-setup.json` cannot be found.
- `gitmap doctor` now verifies the shell wrapper is loaded by checking the `GITMAP_WRAPPER` environment variable, with fix instructions when missing.
- Post-setup verification step warns users if the shell wrapper is not active after `gitmap setup` completes, with reload instructions.
- Shell wrapper scripts (Bash, Zsh, PowerShell) now export `GITMAP_WRAPPER=1` so the binary can detect wrapper-vs-raw invocation.
- `gitmap cd` prints a stderr warning when called without the shell wrapper, guiding users to run `gitmap setup` or reload their profile.

### Fixed

- `gitmap setup` now resolves `git-setup.json` relative to the binary's installation path instead of the current working directory, fixing "file not found" errors when running from arbitrary directories.

---

## v2.72.0 — (2026-04-16)

### Fixed

- VS Code admin-mode bypass: `runVSCodeCommand` now captures `CombinedOutput` and waits for the process exit code instead of fire-and-forget, ensuring CLI errors are properly detected before falling through to the next strategy.
- `tryVSCodeDetached` launches `Code.exe` with an isolated `--user-data-dir` (`%TEMP%\gitmap-vscode-user-data`) so the new instance does not attempt to hand off to an elevated single-instance, fully bypassing the "Another instance of Code is already running as administrator" lock.
- Added `resolveVSCodeExecutable` with multi-path discovery (`LookPath`, CLI sibling, `LocalAppData`, `Program Files`, `Program Files (x86)`) to reliably find the desktop binary when the CLI wrapper is unavailable.
- Extracted all VS Code constants (binary names, flags, paths, messages) into `constants/constants_vscode.go`.

---

## v2.71.0 — (2026-04-16)

### Added

- VS Code admin mode bypass: `openInVSCode` now uses a 3-tier launch strategy (`--reuse-window` → `--new-window` → `cmd /C start` detached) to handle the "Another instance of Code is already running as administrator" error.
- Added `tryVSCodeReuse`, `tryVSCodeNewWindow`, and `tryVSCodeDetached` helper functions in `cmd/clonevscode.go`.
- Added `ErrVSCodeAdminLock` constant for admin-mode warning message.

### Fixed

- `gitmap update` PATH sync now includes full 3-step fallback: direct `Copy-Item`, rename-then-copy (`Move-Item` to `.old` + `Copy-Item` with rollback), and kill stale `gitmap.exe` processes via `Stop-Process` before final retry.
- Updated `UpdatePSSync` PowerShell block in `constants/constants_update.go` with rename and kill-process recovery strategies.
- Updated `spec/01-app/89-update-path-sync.md` to document all sync fallback steps and error scenarios.

---

## v2.70.0 — (2026-04-16)

### Added

- `gitmap clone <url>` now auto-registers cloned repositories with GitHub Desktop by default (no manual prompt).
- `gitmap clone <url>` automatically opens the cloned folder in VS Code (`code --reuse-window`), with `--new-window` fallback for admin-mode conflicts.
- Added `isVSCodeAvailable()` detection via `exec.LookPath` in `cmd/clonevscode.go`.

### Fixed

- `gitmap update` now auto-syncs the active PATH binary when it differs from the deployed binary, resolving the `[FAIL] Active PATH version does not match deployed version` error.
- Added `Copy-Item` sync step with rename and kill-process fallbacks in the update PowerShell script.

---

## v2.69.1 — (2026-04-11)

### Fixed

- Fixed `errorlint` violation in `cmd/helpdashboard.go`: replaced direct `!= io.EOF` comparison with `errors.Is` to handle wrapped errors correctly.

### Changed

- Linked "Riseup Asia LLC" in the author Role row to [riseup-asia.com](https://riseup-asia.com).
- Changed Riseup Asia subheading from centered to left-aligned and linked it to [riseup-asia.com](https://riseup-asia.com).

---

## v2.69.0 — (2026-04-09)

### Added

- Windows binaries now embed a custom emerald green terminal icon, application manifest, and version info via `go-winres`.
- Added `gitmap/winres.json` and `gitmap/assets/icon.png` for Windows resource generation.
- Release pipeline generates `.syso` resource files before compilation, injecting the release version into the binary metadata.
- Added `spec/pipeline/09-binary-icon-branding.md` documenting the full `go-winres` workflow for AI/engineer handoff.
- Added the gitmap icon to the README header.

### Fixed

- Fixed `run.ps1 -d` switch: replaced `[Alias("d")]` on `[string]$DeployPath` with a dedicated `[switch]$Deploy` parameter so `-d` works without requiring a path argument.

---

## v2.68.1 — (2026-04-09)

### Fixed

- Fixed gosec G305 (file traversal) and G110 (decompression bomb) in `helpdashboard.go` zip extraction — paths are now validated against the target directory and extraction is size-limited to 100 MB.
- Fixed `run.ps1 -d` failing with "Missing an argument for parameter 'DeployPath'" — added `[Alias("d")]` to `$DeployPath` so `-d` resolves unambiguously.

---

## v2.68.0 — (2026-04-09)

### Fixed

- Fixed `TempReleases` migration crash: `ALTER TABLE RENAME COLUMN "Commit"` failed with `no such column` when the column was already renamed or never existed. Migration now silently skips the rename when the column is absent.

### Added

- Release pipeline now builds the docs-site (React/Vite) and bundles `dist/` into `docs-site.zip` as a release asset.
- Install scripts (`install.ps1`, `install.sh`) automatically download and extract `docs-site.zip` alongside the binary.
- `gitmap hd` auto-extracts `docs-site.zip` on first run if the `docs-site/` directory is missing — no manual setup needed.
- Added 5 new pipeline specification files (`04`–`08`) covering installation flow, changelog integration, version/help system, environment variable setup, and terminal output standards.
- Added AI Handoff Checklist to `spec/pipeline/README.md` with recommended reading order for onboarding.

## v2.67.0 — Smart Deploy & Rename-First (2026-04-08)

### Improvements

- `run.ps1` and `run.sh` now auto-detect the globally installed `gitmap` binary location and deploy there instead of using a hardcoded path.
- Deploy target resolution follows a 3-tier priority: `--deploy-path` CLI flag → globally installed PATH location → `powershell.json` default.
- First-time installs use the config default; subsequent builds automatically deploy to the active binary's directory.
- Added `Resolve-DeployTarget` function to `run.ps1` and `resolve_deploy_target` function to `run.sh` for full cross-platform parity.
- Deploy step now uses **rename-first strategy**: renames the existing binary to `.old` before copying the new one, avoiding Windows file-lock failures when deploying to a running binary.
- Rollback restores the `.old` file via rename (not copy) for consistency.
- Added "Build once, package once" constraint to `spec/05-coding-guidelines/17-cicd-patterns.md` and `spec/04-generic-cli/11-build-deploy.md`.
- Updated `spec/01-app/09-build-deploy.md` with deploy target resolution and rename-first deploy documentation.
- Added smart deploy path resolution and rename-first deploy to cross-platform parity table in `spec/01-app/42-cross-platform.md`.
- Replaced hardcoded `E:\bin-run` path in `gitmap doctor` fix suggestion with dynamic guidance.

## v2.66.0 — CI Hardening & Pipeline Docs (2026-04-08)

### Improvements

- Pinned `govulncheck` to `v1.1.4` in CI and vulncheck workflows for reproducible builds.
- Updated GitHub Actions to Node.js 24 compatible versions (`actions/checkout@v6`, `actions/setup-go@v6`).
- Added `FORCE_JAVASCRIPT_ACTIONS_TO_NODE24: true` environment variable across all workflows.
- Created portable `spec/pipeline/` documentation folder (CI, release, vulnerability scanning) for cross-AI shareability.
- Added CI Tool Versions pinning table to dependency specs (13, 17, 27) for consistency.
- Aligned severity response times across all dependency management specs.
- Updated stale action version examples in specs 17 and 27 from `@v4`/`@v5` to `@v6`.
- Added cross-reference from `spec/03-general/08-ci-pipeline.md` to `spec/pipeline/`.

### Bug Fixes

- Fixed `ShouldPrintInstallHint` not matching SSH remote URLs (`git@github.com:org/repo.git`) due to colon separator not being normalized to a slash.
- Fixed vulncheck pipeline logic error where `-q` flag on initial `grep` suppressed stdout, breaking the vulnerability classification pipe.

## v2.65.0 — Install UX Overhaul (2026-04-07)

### Improvements

- Install flow now shows a structured **Install Plan** box before execution with tool, version, manager, and command.
- Added numbered step progress: `[1/4] Updating...`, `[2/4] Installing...`, `[3/4] Verifying...`, `[4/4] Recording...`.
- Chocolatey installs now use `--no-progress` flag to suppress GUI popups and prevent blocking on interactive apps like Notepad++.
- Winget installs now use `--silent` flag for unattended installs.
- NPP verification now checks the expected exe path (`C:\Program Files\Notepad++\notepad++.exe`) directly instead of relying on PATH lookup.
- NPP settings zip path now resolves relative to the binary directory (not CWD), fixing "file not found" errors when gitmap is installed globally.
- Detected version is printed during verification for better diagnostics.
- Install command completion is confirmed with a success message before proceeding to verification.

### Bug Fixes

- Fixed NPP install blocking the terminal when Notepad++ GUI launched during Chocolatey install (missing `--no-progress`).
- Fixed post-install verification always failing for NPP because `notepad++` binary is not on PATH.
- Fixed settings zip not found when running `gitmap install npp` from a directory other than the source repo root.

## v2.64.0 — Install Scripts Command (2026-04-07)

### New Commands

- Added `gitmap install scripts` — clones gitmap scripts (install.ps1, install.sh, run.ps1, run.sh, etc.) to a local folder for easy access.
  - **Windows**: resolves the deploy drive from `powershell.json`, defaults to `D:\gitmap-scripts`.
  - **Linux/macOS**: installs to `~/Desktop/gitmap-scripts`.

## v2.63.0 — Installed Directory & Linux Update Flow (2026-04-07)

### New Commands

- Added `gitmap installed-dir` (alias `id`) — prints the full binary path and directory of the active gitmap installation, resolving symlinks to the real location.

### Update Command

- Linux/macOS update now uses `run.sh --update` instead of PowerShell, enabling native shell-based self-update on Unix systems.
- After pulling latest source and rebuilding, the active PATH binary is automatically synced to the new version.
- Added install path resolution using `which gitmap` with `EvalSymlinks` fallback for accurate binary location.
- If `run.sh` is missing from the source repo, a clear error is shown instead of a PowerShell failure.

### Bug Fixes

- Fixed `gitmap update` on Linux: handoff binary no longer uses `.exe` extension and now gets `chmod +x` permission.
- Fixed tilde `~` not expanding in update repo path prompt (e.g. `~/repos/gitmap` was treated as literal `~/`).
- Fixed `gitmap install` on Ubuntu: `apt-get update` now runs before package installation to prevent exit code 100 errors.
- Added `-y`/`--yes` flag to `gitmap install` for non-interactive installs with confirmation prompt.
- Install failures now write detailed error logs to `.gitmap/logs/` with version, manager, command, and reason.
- Fixed `install.sh` installer: `TMP_DIR` unbound variable error on exit caused by subshell scoping.

## v2.62.0 — CI Release Branch Protection (2026-04-07)

### CI/CD

- Release branches (`release/**`) are no longer cancelled by `cancel-in-progress` — every release commit now runs the full CI and release pipeline to completion.
- CI workflow uses a conditional expression: `cancel-in-progress: ${{ !startsWith(github.ref, 'refs/heads/release/') }}` to protect release branches while still cancelling superseded runs on `main` and feature branches.
- Release workflow changed to `cancel-in-progress: false` unconditionally.
- Updated CI pipeline spec (`spec/03-general/08-ci-pipeline.md`) with release branch protection documentation.

## v2.61.0 — Install Hint Polish & Post-Mortem #17 (2026-04-07)

### Release Command

- Improved post-release install hint formatting with emoji labels (📦 🪟 🐧) and better spacing.
- Removed hash-style comments in favor of OS-specific emoji indicators for Windows and Linux/macOS install one-liners.
- Extracted `ShouldPrintInstallHint()` as an exported function for testability.
- Added unit tests for install hint repo detection (11 cases covering gitmap and non-gitmap repos).

### Documentation

- Added Post-Mortem #17: Go Flag Ordering — Silent Flag Drop, documenting the `flag` package behavior and `reorderFlagsBeforeArgs()` fix.

## v2.60.0 — Auto-Detect Pending Release Branch (2026-04-07)

### Release Command

- Running `gitmap release` or `gitmap r` while on a `release/*` branch with no tag now auto-detects and completes the pending release instead of erroring about a duplicate branch.
- Running `gitmap release v1.1.0` while on `release/v1.1.0` with no tag delegates to `ExecuteFromBranch` automatically.
- Added `tryDelegateFromCurrentBranch()` for no-version detection and `tryDelegateFromBranch()` for explicit-version detection.
- Added `MsgReleaseBranchPending` constant for the delegation message.

## v2.59.0 — Post-Release Install Hints (2026-04-07)

### Release Command

- After a successful release, if the repo's remote origin matches the gitmap source repository prefix (`github.com/alimtvnetwork/gitmap-v4`), the CLI now prints install one-liner commands for both Windows (PowerShell) and Linux/macOS (Bash).
- Added `GitmapRepoPrefix` constant for repo detection and `MsgInstallHintHeader`, `MsgInstallHintWindows`, `MsgInstallHintUnix` message constants.
- Install hints appear after `Release complete` in all release paths: standard, branch-based, and metadata-only.
- Non-gitmap repos are unaffected — no install hints are printed.

## v2.58.0 — Release Flag Ordering Fix (2026-04-07)

### Bug Fix

- Fixed `-y` / `--yes` flag being silently ignored when placed after the version argument (e.g., `gitmap release v2.55 -y`).
- Root cause: Go's `flag` package stops parsing at the first non-flag argument, so flags after the version were never processed.
- Added `reorderFlagsBeforeArgs()` helper in `releaseargs.go` — reorders CLI args so all flags precede positional arguments before `flag.Parse()`.
- Affects `release`, `release-self` (`r`, `rs`), and all commands sharing `parseReleaseFlags`.

## v2.57.0 — README & Memory Updates (2026-04-07)

### Documentation

- Split README Quick Start into focused code blocks: separate Install (Windows + Linux/macOS), Scan, and Navigate sections.
- Created `one-liner-installer` memory documenting both `install.ps1` and `install.sh` as CI-generated versioned release assets.

## v2.56.1 — Clone-on-Missing-Path for Update (2026-04-07)

### Update Command

- When the user provides a non-existent path during the `gitmap update` interactive prompt, the system now clones the gitmap source repository into that directory instead of rejecting it.
- After a successful clone, the path is validated, saved to the SQLite Settings DB, and used for the update — no re-prompting on future runs.
- Added `SourceRepoCloneURL`, `MsgUpdateCloning`, `MsgUpdateCloneOK`, and `ErrUpdateCloneFailed` constants.

## v2.56.0 — Release Pipeline install.sh & CI Fix (2026-04-07)

### Release Pipeline

- Added `install.sh` generation to `release.yml` — version-pinned Bash installer is now created and attached as a release asset alongside `install.ps1`.
- Release body now includes both PowerShell and Bash one-liner install instructions.

### CI Pipeline Fix

- Eliminated separate `mark-success` job — inlined cache write as the final step of `test-summary` to prevent `cancel-in-progress` from cancelling the SHA marker after all validation passed.
- `test-summary` now depends on `[sha-check, lint, vulncheck, test]` to ensure full validation before caching.

### Documentation

- Updated `spec/01-app/82-install-script.md` — documented `install.sh` with CLI flags (`--version`, `--dir`, `--arch`, `--no-path`), version-pinned examples, `.tar.gz`/`.zip` fallback, 4-priority binary detection, and shell-aware auto-PATH append (bash/zsh/fish).
- Updated `spec/01-app/12-release-command.md` — CI release pipeline section now mentions `install.sh` alongside `install.ps1` in both steps list and release body format.
- Added "Known Behavior: Concurrency Cancellation" section to `spec/02-app-issues/16-ci-passthrough-gate-pattern.md` — documented and resolved by inlining cache write.
- Updated post-release auto-commit memory to reflect the new `-y` flag behavior.

### Testing

- Added unit test for `-y` flag in autocommit — verifies `promptAndCommit` skips stdin when `yes=true`.

## v2.55.0 — Release Auto-Confirm, Docs & Installer Fix (2026-04-07)

### Post-Mortems Documentation

- Created `spec/02-app-issues/13-release-pipeline-dist-directory.md` — documents `cd: dist` CI failure root cause and 4 prevention rules.
- Created `spec/02-app-issues/14-security-hardening-gosec-fixes.md` — documents G305, G110, format verb, and Code Red fixes with prevention rules.
- Added Post-Mortems page (`/post-mortems`) to docs site with category filters, version tags, and color-coded icons for all 15 documented issues.

### Coding Guidelines Updates

- Added "Lessons Learned" section to `spec/05-coding-guidelines/17-cicd-patterns.md` — never `cd` in CI, validate directories, pin tool versions.
- Added Section 10 (Zip Extraction Security) to `spec/05-coding-guidelines/08-security-secrets.md` — mandatory G305/G110 checks.
- Added Sections 7–8 to `spec/05-coding-guidelines/04-error-handling.md` — Code Red Rule and Format Verb Compliance.

### Installer Fixes

- Fixed PowerShell installer crash caused by `Invoke-WebRequest` progress bar rendering during `irm | iex`.
- Added `$ProgressPreference = "SilentlyContinue"` to `install.ps1`.
- Fixed versioned binary detection — installer now matches `gitmap-v*-windows-(amd64|arm64).exe` patterns from CI archives.
- Wrapped installer `Main` function in `try/catch` with friendly error message and manual download fallback.

### CI Pipeline: Passthrough Gate Pattern

- Replaced job-level `if` skipping with step-level conditionals in `ci.yml` so all jobs always report ✅ Success.
- Previously, SHA-deduplicated runs showed grey "skipped" status which looked like failures; now cached SHAs print "Already validated" and exit green.
- Updated `spec/05-coding-guidelines/29-ci-sha-deduplication.md` with the passthrough pattern documentation.
- Pinned `golangci-lint` to `v1.64.8` in `ci.yml` to match `setup.sh`.

### Release Command: Auto-Confirm (`-y` / `--yes`)

- Added `-y` / `--yes` flag to `release`, `release-self`, `release-branch`, and `release-pending` commands.
- When set, all interactive prompts (e.g. "Auto-commit all changes?") are automatically confirmed without user input.
- Enables fully non-interactive release workflows: `gitmap release v2.55.0 -y`.
- Bumped version to `v2.55.0`.

### Unix Installer (`install.sh`)

- Created `gitmap/scripts/install.sh` — cross-platform Bash installer for Linux and macOS.
- Supports `--version`, `--dir`, `--arch`, `--no-path` flags matching the PowerShell installer feature set.
- Includes SHA256 checksum verification, versioned binary detection, `.tar.gz`/`.zip` fallback.
- Auto-detects shell (bash/zsh/fish) and appends PATH entry to the correct profile file.
- Rename-first strategy for safe upgrades of running binaries.

### Changelog Improvements

- Added release dates to all changelog entries with available metadata (sourced from `.gitmap/release/*.json`).
- Backfilled v2.54.1, v2.54.2, v2.54.3, and v2.53.0 entries in the docs site changelog data.
- Removed duplicate Code Red content from v2.54.0 (now properly in v2.54.1).

### Build Reproducibility

- Pinned `golangci-lint` to `v1.64.8` in `setup.sh` instead of `@latest`.

---

## v2.54.3 — Security Hardening & Lint Compliance (2026-04-07)

### Zip Extraction Security (installnpp.go)

- Fixed **G305** (path traversal): `extractZipEntry` now validates that resolved destination paths stay within the target directory using absolute path prefix checks.
- Fixed **G110** (decompression bomb): `io.Copy` replaced with `io.LimitReader` capped at 10 MB per extracted file.

### Lint Configuration Documentation

- Added inline comments to all 8 gosec exclusions in `.golangci.yml` documenting why each is necessary (G104, G204, G304, G306, G401, G404, G505, G101).

---

## v2.54.2 — Format Verb Audit (2026-04-07)

### fmt.Fprintf Argument Mismatch Fix

- Fixed `cmd/tasksync.go:138` where `fmt.Fprintf` format string expected 2 arguments but only 1 was passed, causing a `go vet` failure.
- Audited all `fmt.Fprintf`, `fmt.Printf`, and `fmt.Errorf` calls across `cmd/`, `release/`, and `store/` packages (~140 call sites, 38+ files) — confirmed 100% compliance.

---

## v2.54.1 — Code Red Error Audit (2026-04-07)

### Mandatory Error Path Logging

- Completed full Code Red audit: every file/path-related error log now includes the exact file path, the operation attempted, and the specific failure reason.
- Standardized format: `Error: [message] at [path]: [error] (operation: [op], reason: [reason])`.
- Updated 35+ constants and 36+ call sites across the entire codebase.
- Generic "file not found" messages without paths are now prohibited by convention.

---

## v2.54.0 — Update Path Recovery & CI Optimization (2026-04-07)

### Update Path Recovery

- `gitmap update` now validates the saved source repo path exists on disk before using it.
- Falls back to the SQLite DB (`source_repo_path` setting) in the binary's `data/` folder.
- Prompts the user interactively when both embedded and saved paths are missing or stale.
- Successfully resolved paths are persisted to the DB for future runs.
- New file `cmd/updaterepo.go` extracts path resolution helpers for the 200-line file limit.

### CI Build Removal

- Removed cross-platform binary builds from the main CI pipeline (`ci.yml`).
- Binaries are now produced exclusively by the release pipeline (`release.yml`) on `release/**` branches and `v*` tags.

### CI Concurrency Cancellation

- All workflows (`ci.yml`, `release.yml`, `vulncheck.yml`) now cancel in-progress runs when a new commit is pushed to the same branch.
- Concurrency groups use `github.ref` so different branches run independently.

### Release Pipeline Fix

- Fixed `cd dist` failure in `release.yml` — the compress/checksum step was running inside `gitmap-updater/` (no `dist/` folder) instead of `gitmap/dist/` where binaries are output.
- Extracted compress and checksum into a separate step with explicit `working-directory: gitmap/dist`.

### SHA-Based Build Deduplication

- CI pipeline now skips redundant runs when the same commit SHA has already passed all checks.
- A `sha-check` gate job probes the GitHub Actions cache for `ci-passed-<SHA>` before any work begins.
- On full pipeline success, a `mark-success` job caches a marker so future runs for the same SHA short-circuit.
- Failed pipelines never cache — re-running the same SHA executes the full pipeline.

---

## v2.53.0 — Help Dashboard & Install Docs

### Help Dashboard Command

- New `gitmap help-dashboard` (alias `hd`) command to serve the documentation site locally.
- Dual-mode resolution: serves pre-built `dist/` via Go's built-in HTTP server; falls back to `npm install && npm run dev` if static assets are missing.
- `--port` flag to configure the serving port (default: 5173).
- Automatically opens the docs site in the default browser on launch.
- Graceful shutdown on Ctrl+C for both static and dev modes.
- New constants file `constants_helpdashboard.go` with all messages, defaults, and error strings.

### Install & Help Dashboard Docs Pages

- Added `/help-dashboard` docs page with terminal demos for static mode, dev fallback, and custom port usage.
- Added `/install` docs page documenting `install` and `uninstall` commands, supported tools, databases, and package managers.
- Both pages include feature cards, flags tables, file layout references, and interactive terminal demos.

## v2.52.0 — Lock Detection & Install System Overhaul

### Lock Detection (clone-next)

- `clone-next` now detects processes locking the current folder when deletion fails.
- On Windows, uses Sysinternals `handle.exe` or PowerShell WMI to identify locking processes.
- On Unix/macOS, uses `lsof` for process detection.
- Prompts the user to terminate blocking processes, then retries folder removal automatically.
- New `lockcheck` package with platform-specific implementations (`lockcheck_windows.go`, `lockcheck_unix.go`).

### Install System Overhaul

- Added SQLite-based installation tracking (`InstalledTools` table) with granular version columns (Major, Minor, Patch, Build) and timestamps.
- Expanded tool support: 11 databases (MySQL, PostgreSQL, Redis, MongoDB, SQLite, MariaDB, CockroachDB, Cassandra, Neo4j, InfluxDB, DynamoDB Local).
- Package manager mappings for Chocolatey, Winget, Apt, Homebrew, and Snap.
- New `gitmap uninstall <tool>` command with `--dry-run`, `--force`, and `--purge` flags.
- README redesigned with centered headers, badges, and grouped command/tool tables.

- Reorganized `gitmap help` output into 17 categorized command groups (Scanning, Cloning, Git Operations, Navigation, Release, etc.).
- Added `--compact` flag to `gitmap help` for a minimal command-and-alias-only listing.
- `gitmap help --compact <group>` filters compact output by group name (case-insensitive, falls back to all groups on no match).
- Added color-coded group headers using ANSI escape codes (bold cyan) for improved terminal readability.
- Added Quick Start section with common command examples at the top of help output.
- Each group header includes a hint to run commands with `--help` or `-h` for detailed usage and examples.
- Modularized help implementation across `rootusage.go`, `rootusagecompact.go`, `rootusageflags.go`, and `constants_helpgroups.go`.
- Repository renamed from `git-repo-navigator` to `gitmap-v4`; all URLs, scripts, and references updated.

## v2.49.1 — Update UX & Versioned Binaries (2026-04-06)

- Added `--repo-path` flag to `update` command: override the source repo path for a one-time update.
- The `--repo-path` flag is automatically forwarded through the handoff binary to `update-runner`.
- Resolution priority: `--repo-path` flag → embedded constant → friendly error with recovery options.
- Improved "repo path not embedded" error with actionable recovery steps (one-liner install, clone & build, manual download, `--repo-path` override).
- CI release binaries now include version in filenames (e.g., `gitmap-v4.49.1-windows-amd64.zip`).
- Updated `install.ps1` (standalone and release-embedded) to handle versioned asset filenames.
- CI release workflow now explicitly marks stable releases as "latest" via `make_latest`.
- Updated `helptext/update.md` with `--repo-path` flag docs, troubleshooting section, and error recovery examples.
- Added `gitmap-updater` — standalone tool to update gitmap via GitHub releases (no source repo required).
- `gitmap update` auto-delegates to `gitmap-updater` when no repo path is available and the updater is on PATH.
- Updater uses handoff-copy pattern to avoid Windows file locks during self-replacement.
- CI release pipeline now builds and ships `gitmap-updater` binaries for all 6 platform targets.

## v2.49.0 — Opt-in Binary Builds & Gitignore Safety (2026-04-06)

- Go binary cross-compilation is now opt-in: use `--bin` or `-b` to build executables during release.
- Removed `--no-assets` flag (replaced by the inverse `--bin` flag).
- `gitmap setup` now ensures `release-assets` and `.gitmap/release-assets` are in `.gitignore`.
- Release workflow auto-appends missing release-related paths to `.gitignore` before each release.
- Added `release-assets` and `.gitmap/release-assets` to `.gitignore` to prevent tracking build artifacts.
- CI release workflow now triggers on `release/*` branch push (in addition to tags).
- Each GitHub release includes: changelog entry, SHA256 checksums, release metadata table, and asset matrix.
- Version-specific `install.ps1` script is auto-generated and attached to each release for one-liner install.
- Pre-release versions (containing `-`) are automatically marked as prerelease on GitHub.

## v2.48.1 — Clone-Next Auto-Navigate (2026-04-03)

- `clone-next` now automatically changes into the newly cloned directory after removing the old folder.
- Prints `→ Now in <target>` confirmation after navigating to the new clone.

## v2.48.0 — Tag Discovery & DB Caching

- `list-releases` now scans git tags via `git for-each-ref` and includes tag-only releases with `source=tag`.
- All discovered releases (repo metadata + tags) are automatically upserted into the SQLite `Releases` table on every `lr` invocation.
- Added `--source tag` filter to `list-releases` for viewing tag-discovered releases.
- Updated helptext and spec to document three-source resolution order and caching behavior.

## v2.47.0 — Release Self Hardening (2026-04-03)

- Changed `release-self` primary alias from `rself` to `rs` (rescan moved to `rsc`).
- Added SQLite DB fallback for source repo discovery (`source_repo_path` in Settings table).
- Skip directory switch if already in the gitmap source repo directory.
- Updated spec, helptext, React docs page, and commands catalog to reflect changes.

## v2.46.0 — Release Self

- Added `release-self` (`rself`) command: release gitmap itself from any directory.
- Auto-fallback: `gitmap release` outside a Git repo now triggers self-release automatically.
- Source repo discovery via `os.Executable()` + symlink resolution + `.git` root walk.
- Returns to original working directory after release with confirmation message.
- Full flag parity with `release` (--bump, --assets, --draft, --dry-run, etc.).
- Added React docs page for release-self with terminal demos and error scenarios.

## v2.45.0 — Docs Site Update (2026-04-03)

- Updated CloneNext docs page with `--create-remote` flag, usage, and terminal example.
- Added repo creation failure to error handling table on docs site.

## v2.44.0 — Clone-Next Spec Update

- Updated `clone-next` spec to document `--create-remote` as opt-in.
- Removed mandatory repo creation from default workflow and examples.
- Added Example 5 showing `--create-remote` usage in spec.
- Marked deferred implementation phases 1–3 as complete.

## v2.43.0 — Clone-Next Hardening

- Auto-cd to parent directory before folder removal to prevent Windows file lock errors.
- Added `--create-remote` flag: optionally create the target GitHub repo before clone (requires `GITHUB_TOKEN`).
- Repo creation is now opt-in instead of mandatory; default `gitmap cn v+1` clones directly.

## v2.42.0 — Clone-Next Simplification

- Removed forced GitHub repo existence check and automatic creation from `clone-next`.
- `gitmap cn v+1` now clones directly without requiring `GITHUB_TOKEN`.
- Repo creation is no longer a blocking prerequisite before clone.

## v2.41.0 — Clone-Next Phase 3 (2026-04-03)

- GitHub repo existence check and automatic creation before clone via GitHub API.
- Requires `GITHUB_TOKEN` for repo creation; creates under org with user fallback.
- Added `ParseOwnerRepo` utility for HTTPS and SSH remote URL parsing.

## v2.40.0 — Clone-Next Command

- Added `clone-next` (alias `cn`) command: clone the next versioned iteration of a repo into its parent directory.
- Supports `v++` and `v+1` (increment current version by 1) and `vN` (jump to explicit version).
- Remote-first repo name resolution: derives base name and version from `remote.origin.url`, not the local folder name.
- GitHub repo existence check before clone: queries `GET /repos/{owner}/{repo}` via GitHub API.
- Automatic GitHub repo creation when target does not exist: creates under org (fallback to user) via GitHub API.
- Requires `GITHUB_TOKEN` environment variable for repo creation.
- Added `ParseOwnerRepo` utility to extract owner/repo from HTTPS and SSH remote URLs.
- Added `--delete` flag: auto-remove current version folder after successful clone.
- Added `--keep` flag: keep current folder without prompting for removal.
- Added `--no-desktop` flag: skip GitHub Desktop registration.
- Added `--ssh-key` / `-K` flag: use a named SSH key for Git operations.
- Added `--verbose` flag: show detailed clone-next diagnostics.
- Clone-Next Flags section added to `gitmap help` output.
- Version argument validation: rejects `v0`, negative values, and malformed inputs with clear errors.
- Case-insensitive version parsing (`V++`, `V+1` accepted).
- No-suffix repos default to `-v2` on increment.
- Added constants for all clone-next messages, errors, and flag descriptions.
- Added unit tests for `ParseRepoName`, `ResolveTarget`, `TargetRepoName`, and `ReplaceRepoInURL`.
- Spec: `spec/01-app/59-clone-next.md` with full workflow, examples, and acceptance criteria.

## v2.37.0 — v2.39.0

- Internal improvements and minor fixes (see individual commits).

## v2.36.7 — Integration Tests

- Added SkipMeta integration test (`skipmeta_test.go`): 6 test cases verifying `SkipMeta: true` prevents metadata and `latest.json` creation.
- Added release rollback integration test (`rollback_test.go`): 5 test cases verifying branch/tag cleanup on simulated push failure.
- Added end-to-end release test (`e2e_test.go`): full cycle from version bump through metadata commit on a temp repo with bare remote.
- E2E edge-case coverage: dry-run (no side effects), no-commit (staged only), skip-meta (no JSON), and duplicate version blocking.
- Added edge-case test suite (`edgecase_test.go`): pre-release parsing/comparison, bump resolution (all levels, from-zero, from-prerelease), parse validation, version ordering, multi-release sequences, out-of-order metadata, and rc-to-stable promotion.
- Added TUI Temp Releases view (`tempreleases.go`, `trformat.go`): 9th tab with flat list, detail panel, and grouped-by-prefix aggregation.
- Added `--stop-on-fail` flag to `pull` and `exec` commands: halts batch after first failure.
- Enhanced `BatchProgress` with per-item failure tracking (`FailWithError`), detailed failure reports, and exit code 3 on partial failures.
- Added `batchreport.go` with `PrintFailureReport()` and `ExitCodeForBatch()` helpers.

## v2.36.6 — Wave 2 Refactoring (14 Files)
- Split `assets.go` → `assets.go` + `assetsbuild.go` (build helpers: `buildSingleTarget`, `buildEnv`).
- Split `zipgroupops.go` → `zipgroupops.go` + `zipgroupshow.go` (display: `runZipGroupList`, `expandFolder`).
- Split `tui.go` → `tui.go` + `tuiview.go` (rendering: `View`, `renderTabs`, `renderContent`).
- Split `aliasops.go` → `aliasops.go` + `aliassuggest.go` (interactive: `runAliasSuggest`, `promptAliasSuggestion`).
- Split `tempreleaseops.go` → `tempreleaseops.go` + `tempreleaselist.go` (listing: `runTempReleaseList`, `printTRList`).
- Split `listreleases.go` → `listreleases.go` + `listreleasesload.go` (data: `loadReleasesFromRepo`, `sortRecordsByDate`).
- Split `listversions.go` → `listversions.go` + `listversionsutil.go` (collection: `collectVersionTags`, `printVersionEntriesJSON`).
- Split `sshgen.go` → `sshgen.go` + `sshgenutil.go` (utils: `validateSSHKeygen`, `resolveGitEmail`).
- Split `scanprojects.go` → `scanprojects.go` + `scanprojectsmeta.go` (metadata: `upsertGoProjectMeta`, `cleanStaleProjects`).
- Split `amendexec.go` → `amendexec.go` + `amendexecprint.go` (output: `buildEnvFilter`, `printAmendProgress`).
- Split `status.go` → `status.go` + `statusprint.go` (formatting: `printStatusTable`, `buildSummaryParts`).
- Split `exec.go` → `exec.go` + `execprint.go` (formatting: `printExecResult`, `printExecBanner`).
- Split `logs.go` → `logs.go` + `logsview.go` (view: `viewList`, `viewDetail`).
- Split `compress.go` → `compress.go` + `compresstar.go` (tar logic: `createTarGz`, `addFileToTar`).
- Added refactoring specs 65–78 for all 14 file splits.
- All source files comply with the 200-line limit; no functional changes.

## v2.36.5 — Extended Refactoring
- Split `ziparchive.go` (362 lines) into three files under `release/`:
  - `ziparchive.go` (~171 lines): orchestration, DB group routing, ad-hoc path resolution.
  - `zipio.go` (~152 lines): ZIP I/O with max Deflate compression, SHA-1 hashing, archive summary.
  - `zipdryrun.go` (~60 lines): dry-run preview for zip groups and ad-hoc archives.
- Split `autocommit.go` (352 lines) into two files under `release/`:
  - `autocommit.go` (~179 lines): orchestration, file classification, user prompts.
  - `autocommitgit.go` (~185 lines): Git primitives, push/retry, rebase recovery.
- Split `seowriteloop.go` (340 lines) into two files under `cmd/`:
  - `seowriteloop.go` (~198 lines): commit loop, rotation orchestration, signal handling.
  - `seowritegit.go` (~153 lines): Git stage/commit/push, rotation file I/O, output formatting.
- Split `workflowbranch.go` (310 lines) into two files under `release/`:
  - `workflowbranch.go` (~179 lines): branch-based releases, pending branch discovery.
  - `workflowpending.go` (~138 lines): metadata-based pending discovery and release.
- Split `workflow.go` (291 lines) into two files under `release/`:
  - `workflow.go` (~183 lines): `Execute`, `Options`/`Result` types, step execution.
  - `workflowvalidate.go` (~115 lines): duplicate detection, orphaned metadata, version resolution.
- Added refactoring specs: `60-refactor-ziparchive.md`, `61-refactor-autocommit.md`, `62-refactor-seowriteloop.md`, `63-refactor-workflowbranch.md`, `64-refactor-workflow.md`.
- All `release/` and `cmd/` files comply with the 200-line limit; no functional changes.

## v2.36.4
- Split `workflowfinalize.go` (498 lines) into four domain-specific files under `release/`:
  - `workflowfinalize.go` (~190 lines): core pipeline orchestration and metadata persistence.
  - `workflowdryrun.go` (~123 lines): dry-run preview functions and `returnToBranch`.
  - `workflowzip.go` (~108 lines): zip group building, ad-hoc archives, and checksum collection.
  - `workflowgithub.go` (~104 lines): GitHub release uploads and Go cross-compilation.
- Split `root.go` (388 lines) into seven domain-specific dispatch files under `cmd/`:
  - `root.go` (72 lines): entry point and top-level router.
  - `rootcore.go` (44 lines): scan, clone, pull, status, exec commands.
  - `rootrelease.go` (48 lines): release workflow commands.
  - `rootutility.go` (56 lines): update, revert, version, help, docs.
  - `rootdata.go` (98 lines): data management, history, profiles, TUI.
  - `roottooling.go` (91 lines): dev tooling and maintenance commands.
  - `rootprojectrepos.go` (38 lines): project type query commands.
- Eliminated `dispatchMisc` (166 lines); replaced by `dispatchData` + `dispatchTooling`.
  - `workflowdryrun.go` (~123 lines): dry-run preview functions and `returnToBranch`.
  - `workflowzip.go` (~108 lines): zip group building, ad-hoc archives, and checksum collection.
  - `workflowgithub.go` (~104 lines): GitHub release uploads and Go cross-compilation.
- All files comply with the 200-line limit; no functional changes.
- Added refactoring specs: `spec/01-app/58-refactor-workflowfinalize.md`, `spec/01-app/59-refactor-root-dispatch.md`.

## v2.36.3 (2026-03-26)
- Bumped compiled version constant to v2.36.3.
- Refactored legacy directory migration into shared `localdirs` package for reuse across CLI startup and release workflow.
- Release workflow now re-runs migration after returning to the original branch, preventing `.release/` from persisting when older branches restore tracked legacy files.
- Auto-commit `classifyFiles` now treats legacy `.release/` paths as release files for silent commit handling.
- Simplified doctor legacy directory check to always pass (migration handles cleanup automatically).
- Removed unused legacy directory warning/fix constants from `constants_doctor.go`.

## v2.36.2 (2026-03-26)
- Bumped compiled version constant to v2.36.2.
- Fixed legacy directory migration to merge files when target already exists instead of skipping.
- Legacy directories (`.release/`, `gitmap-output/`, `.deployed/`) are now fully removed after merging into `.gitmap/`.
- Added `mergeAndRemoveLegacy()` with file-walk merge and `os.RemoveAll` cleanup.
- Replaced Unicode characters in migration messages with ASCII for Windows console compatibility.

## v2.36.1 (2026-03-26)
- Bumped compiled version constant to v2.36.1.
- Added automatic database migration from legacy UUID TEXT IDs to INTEGER AUTOINCREMENT IDs.
- Migration detects TEXT-typed `Id` column in `Repos` via `PRAGMA table_info`, rebuilds the table preserving data, and drops dependent FK tables (project detection, group-repo associations) for clean repopulation.
- Fixed FK constraint violation (`787`) during `scan` when legacy UUID IDs were present in the `Repos` table.

## v2.36.0
- Bumped compiled version constant to v2.36.0.
- Added automatic legacy directory migration: `gitmap-output/` → `.gitmap/output/`, `.release/` → `.gitmap/release/`, `.deployed/` → `.gitmap/deployed/`.
- Migration runs at CLI startup before any command dispatch; skips if target already exists.
- Added `DeployedDirName` subdirectory constant and legacy directory name constants.

## v2.35.1
- Bumped compiled version constant to v2.35.1.
- Added legacy UUID data detection to all remaining DB query paths: `group show`, `group list`, `stats`, `history`, `status`, and `export`.
- All DB query errors from legacy string-based IDs now show a recovery prompt (`rescan` or `db-reset`) instead of raw SQL errors.

## v2.35.0
- Bumped compiled version constant to v2.35.0.
- Consolidated `.release/` and `gitmap-output/` under unified `.gitmap/` directory (`release/`, `output/`).
- Centralized all path constants (`GitMapDir`, `DefaultReleaseDir`, `DefaultOutputDir`) for single-point configuration.
- Migrated all database primary keys from UUID strings to `INTEGER PRIMARY KEY AUTOINCREMENT` (`int64`).
- Removed `github.com/google/uuid` dependency.
- Added `doctor` check (12th) that warns if legacy `.release/` or `gitmap-output/` directories exist.
- Updated all helptext, spec documents, and docs site to reference `.gitmap/` paths.

## v2.34.0 (2026-03-26)
- Bumped compiled version constant to v2.34.0.
- Fixed `list-releases` to read `.release/v*.json` from the current repo first, falling back to the database only when no local files exist.
- Added `SourceRepo` constant to release model for repo-sourced release records.

## v2.33.0 (2026-03-26)
- Bumped compiled version constant to v2.33.0.
- Fixed auto-commit push rejection when remote branch advances during release: added `pull --rebase` recovery with single retry.
- Added 16-stage summary table with anchor links to verbose logging spec.

## v2.32.0
- Bumped compiled version constant to v2.32.0.
- Documented autocommit verbose logging as pipeline stage 16 in the verbose logging spec.

## v2.31.0 (2026-03-26)
- Bumped compiled version constant to v2.31.0.
- Added verbose logging to auto-commit step: logs version, file counts, staging, commit message, and push target.

## v2.30.0 (2026-03-26)
- Bumped compiled version constant to v2.30.0.
- Renamed TempReleases `Commit` column to `CommitSha` to avoid SQLite reserved keyword conflict.
- Added automatic database migration (`ALTER TABLE RENAME COLUMN`) for existing TempReleases tables.
- Added JSON struct tags to `model.TempRelease` for backward-compatible serialization.

## v2.29.0
- Bumped compiled version constant to v2.29.0.
- Fixed TempReleases SQL syntax error: quoted reserved keyword `Commit` in CREATE TABLE, INSERT, and SELECT statements.
- Documented metadata persistence and rollback log points in verbose logging spec (stages 14–15 of 15).

## v2.28.0
- Bumped compiled version constant to v2.28.0.
- Added verbose logging to release pipeline: version resolution, source resolution, git operations, asset collection, staging, cross-compilation, compression, checksums, zip groups, ad-hoc zips, GitHub upload, retry, metadata persistence, and rollback.
- Updated verbose logging spec with all 15 pipeline stages documented.
- Added pull conflict handling to run.ps1 and run.sh with stash/discard/clean/quit prompt.
- Added --force-pull flag to both build scripts for non-interactive CI usage.
- Fixed set -e early exit bug in run.sh git pull error handling.
- Fixed parseCommitLines redeclaration conflict between temprelease.go and changeloggen.go.
- Fixed hasListFlag redeclaration conflict between tempreleaseops.go and completion.go.

## v2.27.0 (2026-03-22)
- Bumped compiled version constant to v2.27.0.
- Added doctor validation checks for config.json, database migration, lock file, and network connectivity.
- Added TUI release trigger overlay with patch/minor/major/custom version bump selection.
- Integrated batch progress tracking into pull, exec, and status commands with success/fail/skip counters.
- Added BatchProgress tracker to cloner package with quiet mode for programmatic use.
- Added TUI interaction tests covering tab switching, browser navigation, fuzzy search, and release triggers.
- Added alias suggestion tests covering auto-suggestion, conflict detection, and idempotent re-runs.

## v2.24.0 (2026-03-20)
- Bumped compiled version constant to v2.24.0.
- Moved release metadata writing from the release branch to the original branch, letting auto-commit handle `.release/` files after returning.
- Removed `commitReleaseMeta` step from the release branch workflow; the release branch now only contains the branch, tag, and push.
- Simplified `pushAndFinalize` to always complete without metadata writes (metadata is now the caller's responsibility).

## v2.23.0 (2026-03-20)
- Bumped compiled version constant to v2.23.0.
- Added `--notes` / `-N` flag to `release-branch` and `release-pending` commands, matching the `release` command.
- Updated docs site Release page with metadata-first workflow diagram, release notes feature card, and `--notes` flag documentation.

## v2.22.0 (2026-03-19)
- Bumped compiled version constant to v2.22.0.
- Persisted zip group metadata in `.release/vX.Y.Z.json` via new `zipGroups` field on `ReleaseMeta`.
- Documented `-A`/`--alias` flag in help text for `pull`, `exec`, `status`, and `cd` commands.
- Added shell completion support for `alias` and `zip-group` subcommands across PowerShell, Bash, and Zsh.
- Added `--list-aliases` and `--list-zip-groups` completion list flags with dynamic DB lookups.
- Added unit tests for `collectZipGroupNames` covering persistent groups, ad-hoc bundles, and merged output.

## v2.21.0
- Bumped compiled version constant to v2.21.0.
- Refactored `assetsupload.go` into three focused files: `githubapi.go` (API types/helpers), `assetsupload.go` (upload logic), `remoteorigin.go` (git URL parsing).
- Rebuilt Project Detection docs page with detection pipeline, tabbed type cards, metadata extraction deep-dive, DB schema, JSON output, and package layout sections.
- Added "How detection works" link from Projects dashboard to Detection page.
- Added unit tests for `store/location.go` covering symlink resolution, fallback, double-nesting prevention, and profile DB filenames.
- Added unit tests for `remoteorigin.go` covering HTTPS, SSH, and invalid URL parsing.

## v2.20.0
- **Fixed**: `OpenDefault()` double-nesting bug where profile config resolved to `<binary>/data/data/profiles.json`.
- Added `DefaultDBPath()` diagnostic helper to `store/location.go`.
- `gitmap ls` now prints resolved DB path when `--verbose` is passed or when zero repos are found.
- Created `spec/01-app/44-list-db-diagnostic.md` for path resolution contract.

## v2.19.0
- Bumped compiled version constant to v2.19.0.

## v2.18.0
- Added batch status terminal demo to Batch Actions page showing dirty/clean state across repos.
- Fixed missing `os/exec` import in release asset upload.
- Resolved `deriveSlug` redeclaration conflict in project repos output.
- Removed unused `os` import from audit command.

## v2.17.0
- Added 30-second auto-refresh timer to TUI dashboard via `tea.Tick`.
- Dashboard refresh interval configurable via `dashboardRefresh` in `config.json`.
- Added `--refresh` flag to `interactive` command for CLI-level override.
- Refresh interval validates with fallback to default 30s when missing or invalid.

## v2.16.0
- Wired real `gitutil.Status()` into TUI dashboard for live dirty/clean indicators.
- Dashboard now shows ahead/behind counts and stash per repo.
- Async background refresh on TUI startup; manual refresh via `r` key.
- Summary bar with aggregate dirty/behind/stash counts and UTC timestamp.

## v2.15.1
- **Fixed**: Database now resolves to `<binary-location>/data/gitmap.db` instead of CWD-relative `gitmap-output/data/`.
- Added `store.OpenDefault()` and `store.OpenDefaultProfile()` for binary-relative database access.
- Added `store/location.go` with `BinaryDataDir()` using `os.Executable()` + `filepath.EvalSymlinks()`.
- Updated all 13 database callers across the codebase to use binary-relative paths.
- Removed unused `resolveAuditOutputDir()` and `resolveDefaultOutputDir()` helpers.

## v2.15.0
- Added cross-platform build support: `run.sh` (Linux/macOS) with full parity to `run.ps1`.
- Fixed Makefile flags to match `run.sh` argument format (`--no-pull`, `--no-deploy`, `--update`).
- Added GitHub Actions CI workflow: test on push, cross-compile 6 OS/arch targets.
- Added GitHub Actions Release workflow: auto-release on `v*` tags with compression and checksums.
- Added interactive TUI mode (`gitmap interactive` / `gitmap i`) built with Bubble Tea.
- TUI repo browser with fuzzy search, multi-select, and keyboard navigation.
- TUI batch actions: pull, exec, status across selected repos.
- TUI group management: browse, create, delete groups interactively.
- TUI status dashboard with live repo status view.
- Added Build System section to Architecture documentation page.
- Added spec documents: `42-cross-platform.md` and `43-interactive-tui.md`.

## v2.14.0
- Added Go release assets: automatic cross-compilation for 6 OS/arch targets (windows/linux/darwin × amd64/arm64).
- Added GitHub Releases API integration for asset upload — no `gh` CLI or external tools needed.
- Added `--compress` flag to wrap release assets in `.zip` (Windows) or `.tar.gz` (Linux/macOS).
- Added `--checksums` flag to generate SHA256 `checksums.txt` for all release assets.
- Added `--no-assets` flag to skip automatic Go binary compilation.
- Added `--targets` flag for custom cross-compile target selection (e.g. `windows/amd64,linux/arm64`).
- Improved `gitmap ls <type>` output with labeled fields (Repo, Path, Indicator) and inline `cd` examples.
- Added shell completion for `release`, `release-branch`, `group`, `multi-group`, and `list` commands.
- Fixed duplicate hints appearing after `gitmap ls <type>` output.

## v2.13.0
- Added group activation: `gitmap g <name>` sets a persistent active group for batch pull/status/exec.
- Added `multi-group` (mg) command for selecting and operating on multiple groups at once.
- Added `gitmap ls <type>` filtering: `gitmap ls go`, `gitmap ls node`, `gitmap ls groups`.
- Added contextual helper hints shown after command output to aid discoverability.
- Added Settings table for persistent key-value configuration in SQLite.

## v2.12.0 (2026-03-14)
- Added global ⌘K command palette searching across commands, flags, and pages.

## v2.11.0
- Added Changelog page with timeline view and expand/collapse controls.
- Added Flag Reference page with sortable, searchable table of all flags.
- Added Interactive Examples page with animated terminal demos.

## v2.10.0 (2026-03-13)
- Version bump for next development cycle.

## v2.9.0 (2026-03-13)
- Completed flags and examples for all 22 command entries on the documentation site.
- Added detailed flag tables and usage examples for `seo-write`, `doctor`, `update`, `pull`, `version`, `history-reset`, and `db-reset`.
- Filled in flags and examples for 15 commands missing both: `rescan`, `desktop-sync`, `status`, `latest-branch`, `release-branch`, `release-pending`, `changelog`, `group`, `list`, `diff-profiles`, `export`, `import`, `profile`, `bookmark`, and `stats`.

## v2.28.0
- Removed unused `detector` import from `cmd/scan.go` that caused build failure.
- Updated documentation site fonts: Ubuntu for headings, Poppins for body text, Ubuntu Mono for code blocks.

## v2.27.0 (2026-03-22)
- Added `gitmap cd` (`go`) command: jump to any tracked repo by slug or partial name.
- Subcommands: `cd repos`, `cd set-default`, `cd clear-default`; supports `--group` and `--pick` flags.
- Added `gitmap watch` (`w`) command: live terminal dashboard monitoring repo status.
- Supports `--interval`, `--group`, `--no-fetch`, and `--json` snapshot mode.
- Added `gitmap diff-profiles` (`dp`) command: compare two profiles side-by-side.
- Supports `--all` and `--json` output flags.
- Added clone progress bars with retry logic and Windows long-path warnings.
- Built documentation site with interactive terminal preview for the watch command.
- Added `gitmap/Makefile` as a thin wrapper around `run.sh` for standard `make` workflows.
  - Targets: `build`, `run` (with `ARGS=`), `test`, `update`, `no-pull`, `no-deploy`, `clean`, `help`.
- Added Makefile documentation page to the docs site with target reference, examples, and argument-passing guide.
- Added `run.sh` cross-platform build script: Bash equivalent of `run.ps1` for Linux and macOS.
  - Full pipeline: pull, tidy, build, deploy with `-ldflags` version embedding.
  - Reads config from `powershell.json` via `jq` or `python3` fallback.
  - Supports `-t` (test with report), `-n` (no-pull), `-d` (no-deploy), and `-u` (update) flags.
- Added `gitmap gomod` (`gm`) command: rename Go module path across an entire repo with branch safety.
  - Replaces module directive in `go.mod` and all matching paths across **all files** by default.
  - Use `--ext "*.go,*.md,*.txt"` to restrict replacement to specific file extensions.
  - Creates `backup/before-replace-<slug>` and `feature/replace-<slug>` branches automatically.
  - Commits changes on the feature branch and merges back to the original branch.
  - Supports `--dry-run`, `--no-merge`, `--no-tidy`, `--verbose`, and `--ext` flags.

## v2.26.0 (2026-03-22)
- Version bump to v2.26.0 following `gitmap profile` command addition.
- All profile subcommands (`create`, `list`, `switch`, `delete`, `show`) fully integrated and documented.

## v2.25.0 (2026-03-22)
- Added `gitmap profile` (`pf`) command: manage multiple database profiles (work, personal, etc.).
- Subcommands: `create`, `list`, `switch`, `delete`, `show`.
- Each profile has its own SQLite database file (`gitmap-{name}.db`).
- Default profile uses existing `gitmap.db` for full backward compatibility.
- Profile config stored in `gitmap-output/data/profiles.json`.
- All commands automatically use the active profile's database.

## v2.24.0 (2026-03-20)
- Added `gitmap import` (`im`) command: restore database from a `gitmap-export.json` backup file.
- Merge semantics: upserts repos/releases, INSERT OR IGNORE for history/bookmarks/groups.
- Group members re-linked by resolving `repoSlugs` against the Repos table.
- Requires `--confirm` flag to prevent accidental data changes.

## v2.23.0 (2026-03-20)
- Added `gitmap export` (`ex`) command: export the full database as a portable JSON file.
- Exports all tables: repos, groups (with member repo slugs), releases, command history, and bookmarks.
- Default output: `gitmap-export.json`; accepts optional custom file path.
- Summary line shows counts for each exported section.

## v2.22.0 (2026-03-19)
- Added `gitmap bookmark` (`bk`) command: save and replay frequently-used command+flag combinations.
- Subcommands: `save`, `list`, `run`, `delete` — full CRUD for saved bookmarks.
- `bookmark run <name>` replays the saved command through standard dispatch (appears in audit history).
- `bookmark list --json` outputs bookmarks as JSON.
- New `Bookmarks` SQLite table with unique name constraint.
- `db-reset --confirm` now also clears the Bookmarks table.

## v2.21.0
- Added `gitmap stats` (`ss`) command: aggregated usage statistics from command history.
- Shows most-used commands, success/fail counts, failure rates, and avg/min/max durations.
- Supports `--command <name>` filter and `--json` output.
- Summary row displays overall totals across all commands.

## v2.20.0
- Added `gitmap history` (`hi`) command: queryable audit trail of all CLI command executions.
- Three detail levels: `--detail basic` (command + timestamp), `--detail standard` (+ flags + duration), `--detail detailed` (+ args + repos + summary).
- Supports `--command <name>` filter, `--limit N`, and `--json` output.
- Added `gitmap history-reset` (`hr`) command: clears audit history (requires `--confirm`).
- New `CommandHistory` SQLite table auto-records every command with start/end timestamps, duration, exit code, and affected repo count.
- `db-reset --confirm` now also clears the CommandHistory table.

## v2.19.0
- Added `gitmap amend` (`am`) command: rewrite author name/email on existing commits with three modes (all, range, HEAD).
- Supports `--branch` flag to operate on a specific branch (auto-switches back to original branch after completion).
- SHA as first positional argument: `gitmap amend <sha> --name "Name"` rewrites from that commit to HEAD.
- `--dry-run` previews affected commits without modifying history or writing audit records.
- `--force-push` auto-runs `git push --force-with-lease` after amend.
- Audit trail: every amend operation writes a JSON log to `.gitmap/amendments/amend-<timestamp>.json` with full details.
- Database persistence: amendment records saved to `Amendments` SQLite table for queryable history.
- `db-reset --confirm` now also clears the `Amendments` table.
- Added `--author-name` and `--author-email` flags to `gitmap seo-write` (`sw`): set custom author on each commit.
- SEO-write dry-run now displays the author that would be used when author flags are set.

## v2.18.0
- Added `gitmap seo-write` (`sw`) command: automated SEO commit scheduler that stages, commits, and pushes files on a randomized interval.
- Supports CSV input mode (`--csv`) for user-provided title/description pairs.
- Supports template mode with placeholder substitution (`{service}`, `{area}`, `{url}`, `{company}`, `{phone}`, `{email}`, `{address}`).
- Pre-seeded `data/seo-templates.json` with 25 title and 20 description templates (500 unique combinations).
- Added `CommitTemplates` SQLite table for persistent template storage with auto-seeding on first run.
- Rotation mode: when pending files are exhausted, appends/reverts text in a target file to maintain commit activity.
- Configurable interval (`--interval min-max`), commit limit (`--max-commits`), file selection (`--files`), and dry-run preview.
- Added `--template <path>` flag to load templates from a custom JSON file at runtime.
- Added `--create-template` / `ct` shorthand to scaffold a sample `seo-templates.json` in the current directory.
- Graceful shutdown on Ctrl+C (finishes current commit before exiting).

## v2.17.0
- Added `Source` column to the `Releases` table: tracks whether each release was created via `gitmap release` (`release`) or imported from `.release/` files (`import`).
- Added `--source` flag to `gitmap list-releases` (`lr`): filter releases by origin (`--source release` or `--source import`).
- Added `--source` flag to `gitmap list-versions` (`lv`): cross-references git tags with the Releases DB to filter by source and display source metadata.
- Added `--source` flag to `gitmap changelog` (`cl`): filter changelog entries by release source.
- Terminal and JSON output for `list-releases` and `list-versions` now includes the Source field.

## v2.16.0
- Added `gitmap list-releases` (`lr`) command: queries the Releases DB table and displays stored releases with `--json` and `--limit N` support.
- Enhanced `gitmap scan` to import `.release/v*.json` metadata files into the Releases DB table automatically after each scan.

## v2.15.0
- Added `--limit N` flag to `gitmap list-versions` (`lv`): show only the top N versions (0 or omitted = all).

## v2.14.0
- Added `Releases` table to SQLite database: stores release metadata (version, tag, branch, commit, changelog, flags) persistently.
- Release workflow now auto-persists metadata to the database after successful releases.
- Converted all database table and column names from snake_case to PascalCase (`Repos`, `Groups`, `GroupRepos`, `Releases`).
- Added `store/release.go` with `UpsertRelease`, `ListReleases`, `FindReleaseByTag` methods.
- Added `model/release.go` with `ReleaseRecord` struct.
- Note: existing databases will need `gitmap db-reset --confirm` to adopt the new schema.

## v2.13.0
- Release metadata JSON (`.release/vX.Y.Z.json`) now includes a `changelog` field with notes from CHANGELOG.md (gracefully omitted if unreadable).
- `gitmap list-versions` (`lv`) now shows changelog notes as sub-points under each version in terminal output.
- `gitmap list-versions --json` includes changelog array per version in JSON output.

## v2.12.0 (2026-03-14)
- Added `gitmap list-versions` (`lv`) command: lists all release tags sorted highest-first, with `--json` output support.
- Added `gitmap revert <version>` command: checks out a release tag and rebuilds/deploys via handoff (same mechanism as `update`).

## v2.11.0
- Added constants inventory audit section to compliance spec, documenting ~280 constants across 9 files and 17 categories.

## v2.10.0 (2026-03-13)
- Full compliance audit (Wave 1 + Wave 2): all 75 source files pass code style rules.
  - Trimmed 4 oversized files: `workflow.go`, `terminal.go`, `safe_pull.go`, `setup.go` (all under 200 lines).
  - Fixed all negation and switch violations across `changelog.go`, `github.go`, `metadata.go`, `config.go`, `verbose.go`, `semver.go`.
  - Extracted missing constants to dedicated constants files.

## v2.9.0 (2026-03-13)
- Full code style refactor of `latest-branch` command:
  - Split `cmd/latestbranch.go` into 3 files: handler, resolve, output (all under 200 lines).
  - Split `gitutil/latestbranch.go` into 2 files: core operations, resolve helpers.
  - All functions comply with 8-15 line limit. Positive logic throughout.
  - Blank line before every return. No magic strings. Chained if+return replaces switch.
  - Extracted git constants and display message constants.

## v2.8.0 (2026-03-06)
- Added `--filter` flag to `latest-branch`: filter branches by glob pattern (e.g. `feature/*`) or substring match.

## v2.7.0
- Added `--sort` flag to `latest-branch`: supports `date` (default, descending) and `name` (alphabetical ascending).

## v2.6.0
- Centralized date display formatting: all dates now convert to local timezone and display as `DD-Mon-YYYY hh:mm AM/PM`.
- Added `gitutil/dateformat.go` with `FormatDisplayDate` and `FormatDisplayDateUTC` functions.
- Updated `latest-branch` terminal, JSON, and CSV output to use the new date format.

## v2.5.1
- Added `--no-fetch` flag to `latest-branch`: skips `git fetch --all --prune` when remote refs are already up to date.

## v2.5.0 (2026-03-06)
- Added `--format` flag to `latest-branch`: supports `terminal` (default), `json`, and `csv` output formats.
  - CSV outputs a header row + data rows to stdout, suitable for piping and spreadsheets.
  - `--json` remains as shorthand for `--format json`.
- Refactored `latest-branch` output into dedicated functions per format.

## v2.4.1
- Added positional integer shorthand for `latest-branch`: `gitmap lb 3` is equivalent to `gitmap lb --top 3`.

## v2.4.0 (2026-03-06)
- Added `gitmap latest-branch` (`lb`) command: finds the most recently updated remote branch by commit date and displays name, SHA, date, and subject.
  - Flags: `--remote`, `--all-remotes`, `--contains-fallback`, `--top N`, `--json`.
  - Positional integer shorthand: `gitmap lb 3` is equivalent to `gitmap lb --top 3`.

## v2.3.12 (2026-03-06)
- Spec, issue post-mortems, and memory aligned to codify synchronous update handoff and rename-first PATH sync as permanent rules.
- Rename-first PATH sync in `-Update` mode: renames active binary to `.old` before copying, eliminating lock-retry loops.
- Parent `update` handoff uses `cmd.Start()` + `os.Exit(0)` to release file lock before worker runs.
- Handoff diagnostic log prints active exe and copy paths at update start.
- Spec consistency pass: all four update-flow specs now enforce identical rules.

## v2.3.10 (2026-03-06)
- Fixed `Read-Host` error in non-interactive PowerShell sessions during update by removing trailing prompt.
- Parent `update` process now exits immediately (handoff copy runs synchronously via `update-runner`).
- Added diagnostic log at update start showing active exe path and handoff copy path.
- Update script now uses unique temp file names (`gitmap-update-*.ps1`) to avoid stale script collisions.

## v2.3.9
- Version bump for rebuild validation after update-runner handoff changes.

- Replaced `update --from-copy` with hidden `update-runner` command for cleaner handoff separation.
- Handoff copy now created in the same directory as the active binary (fallback to %TEMP% if locked).
- Added `-Update` flag to `run.ps1`: runs full update pipeline (pull, build, deploy, sync) with post-update validation and cleanup.
- Update script delegates entire pipeline to `run.ps1 -Update`.
- Before/after version output derived from actual executables, not static constants.
- Mandatory `update-cleanup` runs after successful update to remove handoff and `.old` artifacts.
- Cleanup now scans both `%TEMP%` and same-directory for leftover `gitmap-update-*.exe` files.

- Added `gitmap doctor --fix-path` flag: automatically syncs the active PATH binary from the deployed binary using retry (20×500ms), rename fallback, and stale-process termination, with clear confirmation output.
- Doctor diagnostics now suggest `--fix-path` when version mismatches are detected.

## v2.3.6
- Added stale-process fallback during PATH-binary sync (`update` + `run.ps1`): if copy+rename fail, it now stops stale `gitmap.exe` processes bound to the old path and retries once.
- Improved failure guidance to run the deployed binary directly when active PATH binary remains locked.

## v2.3.5
- Hardened `gitmap update` PATH sync with retry + rename fallback, and it now exits with failure if active PATH binary remains stale.
- Clarified update output labels to distinguish source version (`constants.go`) vs active executable version.
- Added same rename-fallback PATH sync behavior in `run.ps1`.

## v2.3.4
- Updated PATH-binary sync in `run.ps1` and `gitmap update` to use retry-on-lock behavior (20 attempts × 500ms), matching the self-update spec.
- Added explicit recovery guidance when active PATH binary is still locked, including an exact `Copy-Item` fix command.

## v2.3.3
- Added `gitmap doctor` command: reports PATH binary, deployed binary, version mismatches, git/go availability, and recommends exact fix commands.

## v2.3.2
- `gitmap update` now syncs the active PATH binary with the deployed binary, so commands like `release` are available immediately.
- `gitmap update` now prints changelog bullet points after update (or no-op update) for quick visibility.
- Added `gitmap changelog --open` and `gitmap changelog.md` to open `CHANGELOG.md` in the default app.

## v2.3.1
- Added `gitmap changelog` command for concise, CLI-friendly release notes.
- Improved `gitmap update` output to show deployed binary/version and warn if PATH points to another binary.
- `gitmap update` now prints latest changelog notes after a successful update.

## v2.3.0
- Added `gitmap release-pending` (`rp`) to release all `release/v*` branches missing tags.
- `gitmap release` and `gitmap release-branch` now switch back to the previous branch after completion.

## v2.2.3
- Fixed PowerShell parser-breaking characters in update/deploy output paths.
- Improved deployment rollback messaging in `run.ps1`.

## v2.2.2
- Added additional parser safety fixes for update script output.

## v2.2.1
- Patched PowerShell parsing edge cases affecting update flow.
