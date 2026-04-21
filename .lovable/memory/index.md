# Project Memory

## Core
Strict code style: <200 lines/file, <15 lines/func, positive logic, pascal case constants, 'is/has' boolean prefixes.
Zero-swallow error policy. Explicitly log errors to os.Stderr using standardized format. Use `errors.Is`.
Version bump = update THREE artifact files directly: `.gitmap/release/latest.json`, new `.gitmap/release/vX.Y.Z.json`, and `CHANGELOG.md` (rename Unreleased heading) + the single Go const `Version` in `gitmap/constants/constants.go`. Do NOT modify ANY OTHER file under `gitmap/` source folder during a version bump. Do NOT defer to `gitmap r`. See [Version Bump Procedure](mem://project/version-bump-procedure).
NEVER touch `.gitmap/release-assets/` manually. NEVER touch the `gitmap/` Go source folder unless the user explicitly asks for a code change — bumps and memory updates must leave `gitmap/` untouched apart from the lone `Version` const line.
No magic strings. Centralize in constants. All CLI IDs must be exclusively in `constants_cli.go`.
Windows-first platform development strategy. Scripts must handle Windows encoding (UTF-8 BOM).
Go v1.24.13. golangci-lint pinned to v1.64.8, govulncheck pinned to v1.1.4.
SQLite connection pooling restricted to `SetMaxOpenConns(1)`.
v15 DB schema: PascalCase + singular table names + `{TableName}Id` PKs + FKs match referenced PK + `IsX` boolean prefix + abbreviations as words (`SshKey`, `Csharp*`). SQLite reserved word `Group` double-quoted in DDL/DML.
Unified `.gitmap/` directory structure at repository root for all artifacts.
Clone-next flattens by default (v2.75.0+): clones into base name folder, tracks versions in RepoVersionHistory.
Completion generator uses marker-comment opt-in (v3.0.0+): `// gitmap:cmd top-level` on const block, `// gitmap:cmd skip` per spec. CI `generate-check` enforces drift.
v15 legacy compat shims (JSON `draft`/`preRelease` overlay + SQLite `Draft`/`PreRelease` column rename) are KEPT through v3.x; removal scheduled for v4.0.0.
Current version: v3.21.0 (schema-version fast-path + db-migrate --force + post-update force-migrate + last-release detector fix + gitmap install clean-code).

## Memories
- [Version Bump Procedure](mem://project/version-bump-procedure) — Update Version const + latest.json + new vX.Y.Z.json + CHANGELOG.md heading directly. Do NOT defer to `gitmap r`.
- [v3.12.1 Session](mem://03-v3.12.1-session) — Legacy field migration, AST parity test, fresh ERD, v15 audit, version bump (v3.12.1)
- [v15 Legacy Compat Audit](mem://02-v15-legacy-compat-audit) — Keep JSON overlay + SQLite column rename through v3.x, remove in v4.0.0
- [v15 Rename Progress](mem://features/v15-rename-progress) — Phase 1 complete: all 22 tables singular + {Table}Id PKs + IsDraft/IsPreRelease + CSharp→Csharp (v3.5.0)
- [Deploy Layout & Binary Readout](mem://features/deploy-layout-and-binary-readout) — Deploy folder is `gitmap-cli` (not `gitmap`). Bare `gitmap` prints Active/Deployed/Config triplet. PATH-detection reuses existing install location.
- [Uninstall Quick Scripts](mem://features/uninstall-quick-scripts) — Root-level uninstall-quick.{ps1,sh} one-liners that delegate to `gitmap self-uninstall` with manual-sweep fallback. Cleans both `gitmap-cli/` and legacy `gitmap/`.
- [Scan Folder & Version Probe Schema](mem://features/scan-folder-and-version-probe) — ScanFolder + VersionProbe tables, `gitmap sf add/list/rm` (v3.7.0)
- [Version Probe](mem://features/version-probe) — Hybrid `git ls-remote` → shallow clone fallback. `gitmap probe [path|--all]`. Scan auto-tags repos with ScanFolderId. (v3.8.0)
- [Find-Next](mem://features/find-next) — `gitmap find-next` / `fn` reads latest VersionProbe rows where IsAvailable=1. `--scan-folder <id>` filter, `--json` output. Read-only. (v3.9.0)
- [Parallel Pull](mem://features/parallel-pull) — `gitmap pull --parallel <N>` worker pool with mutex-guarded BatchProgress. `--only-available` intersects targets via FindNext (fail-open). (v3.10.0)
- [Scan All](mem://features/scan-all) — `gitmap scan all` / `scan a` re-scans every ScanFolder root via N=4 worker pool, prompts to prune missing roots, exit codes 0/1/2/3. Spec 100. (planned v3.33.0)
- [Pull All](mem://features/pull-all) — `gitmap pull all` / `p all` / `pull a` updates every repo under CWD scan root via N=4 worker pool. run.ps1 (Win) / run.sh (Unix) replaces git pull. Strict ScanFolder match, no prefix fallback. Spec 101. (planned v3.34.0)
- [Scan GD](mem://features/scan-gd) — `gitmap scan gd` / `scan github-desktop` registers every repo under CWD scan root in GitHub Desktop. Sequential, idempotent. Coexists with existing `--github-desktop` flag. Spec 102. (planned v3.35.0)
- [Probe Depth](mem://features/probe-depth) — `gitmap probe --depth N` (default 1, max 10) walks up to N newer tags, shallow-clones each to verify, inserts one VersionProbe row per verified version. Adds optional IsPreRelease column. Backwards compatible. Spec 103. (planned v3.36.0)
- [Desktop Sync = GD Merge](mem://features/desktop-sync-merge) — `ds` is now an alias of `gd`. No scan dependency. GitHub Desktop install check is step 1. `.git` worktree files detected. Spec 11 + 10. (v3.37.0)
- [Clone Multi-URL](mem://features/clone-multi) — `gitmap clone` accepts space- and/or comma-separated URL lists; each positional arg is split on commas and flattened. `--github-desktop` registers each clone immediately. Spec 104. (planned v3.38.0)
- [Release-Version Script](mem://features/release-version-script) — Dedicated `release-version.ps1` / `.sh` for `/release/:version` pages. Pinned, never auto-upgrades. Ships as generic-parameterized + per-version baked snapshot. Spec 105. (planned v3.39.0)
- [Code Constraints](mem://style/code-constraints) — Strict rules for code style, structure, and pull requests
- [Code Quality Process](mem://style/code-quality-improvement-process) — Architectural principles and resilience patterns
- [README Branding](mem://style/readme-branding) — Strict layout and linking requirements for the project author section
- [Windows Environment](mem://constraints/windows-environment) — Long paths, short root recommendations for git
- [PowerShell Encoding](mem://constraints/powershell-encoding) — ASCII punctuation, Virtual Terminal Processing, stdout vs stderr
- [Navigation Helper](mem://features/navigation-helper) — Shell wrapper using GITMAP_SHELL_HANDOFF for cd/clone-next
- [Command Help System](mem://features/command-help-system) — 120-line limit per help file, 3-8 line realistic simulations
- [Clone-Next Flatten](mem://features/clone-next-flatten) — Default flatten: clone into base-name folder, version tracking in DB with RepoVersionHistory table (DONE v2.75.0)
- [Clone Direct URL](mem://features/clone-direct-url) — gitmap clone accepts direct HTTPS/SSH URLs with optional folder name, auto-flattens versioned URLs
- [Move & Merge Commands](mem://features/movemerge) — gitmap mv / merge-both / merge-left / merge-right with L/R/S/A/B/Q prompt + --prefer-* bypass + URL-side commit/push (v2.96.0)
- [Release Alias](mem://features/release-alias) — gitmap as / release-alias (ra) / release-alias-pull (rap) with auto-stash labeled by alias-version-unixts, label-match pop for concurrent safety (v3.0.0)
- [Self Install Uninstall](mem://features/self-install-uninstall) — gitmap self-install / self-uninstall manage the binary itself (separate from third-party install/uninstall). Embedded scripts via go:embed, Windows handoff, marker-block PATH cleanup
- [Marker Comments](mem://features/marker-comments) — Decentralized opt-in for completion generator: `// gitmap:cmd top-level` + `// gitmap:cmd skip`, CI drift check enforces sync (v3.0.0)
- [Database Architect](mem://tech/database-architecture) — Idempotent SQLite migrations, PascalCase schema helpers
- [Database Constraints](mem://tech/database-constraints) — Recursive reconciliation pattern, explicitly re-query database IDs
- [Database Location](mem://tech/database-location) — SQLite state anchored to binary execution path via filepath.EvalSymlinks
- [Process Sync](mem://tech/process-synchronization) — Advisory file-based locking via gitmap.lock
- [DB Migration Strategy](mem://tech/database-migration-strategy) — Graceful recovery for breaking schema changes, intercepting scan errors
- [Static Analysis](mem://tech/static-analysis-security) — Linter setup, vulnerability response times, @latest installations prohibited
- [Security Hardening](mem://tech/security-hardening) — Zip extraction path validation, io.LimitReader for decompression bombs
- [Changelog System](mem://project/changelog-system) — Dual-mode Markdown/React changelog synced with local release metadata
- [Flag Parsing Logic](mem://tech/flag-parsing-logic) — Reordering flags before args to bypass Go's default flag package limitations
- [Go Namespace Rules](mem://tech/go-namespace-constraints) — Preventing redeclaration across files in the same Go package
- [Vulnerability Mitigation](mem://tech/vulnerability-mitigation-strategy) — Bypassing GO-2026-4601 in Go 1.24 via custom http Request
- [Config Pattern](mem://tech/config-pattern) — Three-layer configuration merge (defaults < config.json < CLI flags)
- [Script Generation](mem://tech/script-generation) — PowerShell text/template encoding with UTF-8 BOM
- [Constants Structure](mem://tech/constants-structure) — Avoiding redeclaration errors with unique suffixes and domain-specific files
- [Code Red Error Mgmt](mem://tech/code-red-error-management) — Zero-swallow error policy and os.Stderr standardized format
- [Internal Memory Standard](mem://project/internal-memory-standard) — Folder structure and file naming conventions for project planning
