---
name: VS Code Project Manager Sync
description: gitmap scan auto-syncs and `gitmap code` registers + opens repos in alefragnani.project-manager projects.json
type: feature
---

# VS Code Project Manager Sync (v3.38.0)

Spec: `/spec/01-vscode-project-manager-sync/README.md`
Sample fixture: `/spec/01-vscode-project-manager-sync/sample-projects.json`

## Source of truth

- gitmap SQLite DB (table `VSCodeProject`) is the source of truth.
- `projects.json` (the alefragnani.project-manager extension file) is a synced export.

## projects.json schema (locked)

`name` (string), `rootPath` (string, absolute, native separators),
`paths` ([]), `tags` ([]), `enabled` (bool), `profile` (string).

`paths` and `tags` are always emitted as `[]` by gitmap on insert and
**preserved** if the user edited them. `enabled` and `profile` likewise
preserved on upsert.

## File path resolution (DO NOT hardcode full path)

Resolve VS Code user-data root first, then append:
`User/globalStorage/alefragnani.project-manager/projects.json`.

| OS      | Root                                                              |
|---------|-------------------------------------------------------------------|
| Windows | `%APPDATA%\Code` → fallback `%USERPROFILE%\AppData\Roaming\Code`  |
| macOS   | `$HOME/Library/Application Support/Code`                          |
| Linux   | `$XDG_CONFIG_HOME/Code` → fallback `$HOME/.config/Code`           |

## Match key

`rootPath`. Case-insensitive on Windows, case-sensitive on Unix.
SQLite unique index uses `COLLATE NOCASE`.

## Behavior rules

- `gitmap scan` syncs by default; `--no-vscode-sync` disables.
- `gitmap scan` **never** opens VS Code.
- `gitmap code [alias] [path]` upserts + opens VS Code via `code "<rootPath>"`.
  - No args → git repo root if inside one, else CWD; alias = folder basename.
  - `[alias]` → alias override.
  - `[alias] [path]` → arbitrary path, no git requirement.
- `gitmap as <newalias>` mirrors the DB name change to `projects.json` for the matching `rootPath`.
- Foreign entries in `projects.json` are always preserved.
- Atomic writes only: write to `.tmp` then `os.Rename`.

## Command name

`gitmap` (single word). The string `git map` (with a space) must
not appear anywhere in code, help, or logs.

## Out of scope (v1)

Multi-root `paths`, auto-derived tags, reverse sync, profile assignment.
