---
name: Version Bump Procedure
description: When user says "bump minor/patch/version" or "release", update Version const + 2 release JSON files + CHANGELOG. Do NOT defer to gitmap r.
type: feature
---
# Version Bump Procedure

When the user says "bump the minor", "bump version", "release it", "cut a release", or similar, perform ALL of these steps directly. Do NOT defer to `gitmap r` — release info lives in the const + JSON files and the AI is expected to update them.

## Files to update (atomic — batch in parallel)

1. **`gitmap/constants/constants.go`** — `const Version = "X.Y.Z"`. Minor = X.Y+1.0, patch = X.Y.Z+1, major = X+1.0.0.
2. **`.gitmap/release/latest.json`** — overwrite with `{"version": "X.Y.Z", "tag": "vX.Y.Z", "branch": "release/vX.Y.Z"}`.
3. **`.gitmap/release/vX.Y.Z.json`** — create new file with: `version`, `branch` (`release/vX.Y.Z`), `sourceBranch` ("main"), `commit` (empty if unknown), `tag` (`vX.Y.Z`), `assets: []`, `changelog: [...bullet strings...]`, `isDraft: false`, `isPreRelease: false`, `createdAt` (ISO 8601 UTC), `isLatest: true`.
4. **`CHANGELOG.md`** — rename `## Unreleased — ...` to `## vX.Y.Z — (YYYY-MM-DD) — <summary>`.

## Rules

- The previous "never edit .gitmap/release/* by hand" rule was OVERRIDDEN by the user — the AI must update these files directly.
- **DO NOT touch ANY file under the `gitmap/` source folder during a bump EXCEPT the single `Version = "X.Y.Z"` line in `gitmap/constants/constants.go`.** No new `.go` files, no edits to other constants, no refactors. If a code change is needed, the user will ask for it separately.
- Use `code--line_replace` for const + latest.json + CHANGELOG; `code--write` for the new vX.Y.Z.json.
- `createdAt` = real ISO 8601 UTC timestamp.
- `changelog` array entries are plain bullet strings (no leading `- ` markers).
- Bump `constants.SchemaVersionCurrent` ONLY when Migrate() gains a new structural step (new CREATE/ALTER/phase/seed/ID rename). Cosmetic changes do NOT require a schema bump.