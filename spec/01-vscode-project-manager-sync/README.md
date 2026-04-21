# VS Code Project Manager Sync

> Status: **Spec locked, implementation pending**
> Owner: gitmap CLI
> Version target: v3.38.0
> Sample fixture: [`sample-projects.json`](./sample-projects.json) (273 entries, captured from a real user environment)

## 1. Goal

Keep the `alefragnani.project-manager` VS Code extension's `projects.json` in
lock-step with the gitmap database so every scanned repo or path the user
explicitly registers via `gitmap code` shows up immediately in the VS Code
**Project Manager** sidebar.

The DB is the source of truth. `projects.json` is a synced **export**.

## 2. Schema (locked from sample)

Each entry in `projects.json` is an object with exactly these fields:

```json
{
  "name": "gitmap",
  "rootPath": "d:\\wp-work\\riseup-asia\\gitmap",
  "paths": [],
  "tags": [],
  "enabled": true,
  "profile": ""
}
```

| Field      | Type      | gitmap behavior                                                  |
|------------|-----------|------------------------------------------------------------------|
| `name`     | string    | DB alias. On first insert: folder basename. Updated by `gitmap as`. |
| `rootPath` | string    | **Match key.** Absolute path. Native separators per OS.          |
| `paths`    | string[]  | Always `[]` from gitmap. Preserved on upsert if user edited it.  |
| `tags`     | string[]  | Always `[]` from gitmap (v1). Preserved on upsert if user edited.|
| `enabled`  | boolean   | `true` on insert. Preserved on upsert.                           |
| `profile`  | string    | `""` on insert. Preserved on upsert.                             |

**Multi-root (`paths`) and auto-tags are deferred** to a future revision.

## 3. File location — derived from VS Code user-data root

Per user request, do **not** hardcode the full path. First resolve the
**VS Code user-data root**, then append the extension-relative tail.

### 3.1 User-data root discovery

| OS      | Resolution order                                                                          |
|---------|-------------------------------------------------------------------------------------------|
| Windows | `%APPDATA%\Code` → fallback `%USERPROFILE%\AppData\Roaming\Code`                          |
| macOS   | `$HOME/Library/Application Support/Code`                                                  |
| Linux   | `$XDG_CONFIG_HOME/Code` → fallback `$HOME/.config/Code`                                   |

If the root directory does not exist, gitmap reports a clear error
("VS Code user data directory not found at <path> — is VS Code installed?")
and exits non-zero.

### 3.2 Relative tail (constant across all OSes)

```
User/globalStorage/alefragnani.project-manager/projects.json
```

Final path = `filepath.Join(userDataRoot, "User", "globalStorage", "alefragnani.project-manager", "projects.json")`.

If the file does not exist, gitmap creates it with `[]`. If the parent
directory does not exist, gitmap returns an error rather than creating
extension folders silently (the extension must be installed).

## 4. Atomicity

All writes go through:

1. Read existing file (or `[]` if missing).
2. Decode → mutate in memory → encode with tab indent (matches sample fixture).
3. Write to `projects.json.tmp` in the same directory.
4. `os.Rename` over the original.

Failures at any step leave the original file untouched. A trailing newline
is appended for git-friendliness.

## 5. CLI surface

### 5.1 `gitmap scan` — auto-sync (default ON)

After the existing scan + DB upsert phase, gitmap reads every
`VSCodeProject` row and reconciles `projects.json`:

- New `rootPath` → append entry with gitmap defaults.
- Existing `rootPath` (case-insensitive on Windows) → update only `name`.
  Leave `paths`, `tags`, `enabled`, `profile` untouched.
- Foreign entries (rootPath not in DB) → **preserved**, never deleted.

Flags:

- `--no-vscode-sync` — skip the sync phase entirely.

Summary line printed:

```
✓ VS Code Project Manager: 12 added, 3 updated, 0 skipped (38 total in projects.json)
```

`scan` **never opens VS Code.**

### 5.2 `gitmap code [alias] [path]` — register + open

| Invocation                  | Behavior                                                                 |
|-----------------------------|--------------------------------------------------------------------------|
| `gitmap code`               | Use git repo root (if inside one) else CWD; alias = folder basename.     |
| `gitmap code myalias`       | Same path resolution; alias overridden to `myalias`.                     |
| `gitmap code myalias /path` | Use `/path` (any path, no git requirement); alias = `myalias`.           |

Steps:

1. Resolve absolute `rootPath` (`filepath.Abs` + `EvalSymlinks`).
2. Upsert into DB `VSCodeProject` table by `rootPath`.
3. Sync `projects.json` (atomic).
4. Launch `code "<rootPath>"`. If `code` is not on PATH, print:
   ```
   VS Code CLI not found on PATH.
   Open VS Code → Cmd/Ctrl+Shift+P → "Shell Command: Install 'code' command in PATH".
   ```

### 5.3 `gitmap as <newalias>` — alias rename mirror

Existing `gitmap as` flow gains a post-hook: after the DB rename succeeds,
it calls the same projects.json sync helper so the matching `rootPath` row
gets its `name` updated. No new flag.

## 6. Database

Extends the unified gitmap SQLite DB at the binary path
(see `mem://tech/database-location`). Per `mem://tech/database-architecture`
all identifiers are **PascalCase** with `INTEGER PRIMARY KEY AUTOINCREMENT`.

```sql
CREATE TABLE IF NOT EXISTS VSCodeProject (
    Id          INTEGER PRIMARY KEY AUTOINCREMENT,
    RootPath    TEXT NOT NULL,
    Name        TEXT NOT NULL,
    Enabled     INTEGER NOT NULL DEFAULT 1,
    Profile     TEXT NOT NULL DEFAULT '',
    LastSeenAt  TEXT NOT NULL,
    CreatedAt   TEXT NOT NULL,
    UpdatedAt   TEXT NOT NULL
);
CREATE UNIQUE INDEX IF NOT EXISTS UX_VSCodeProject_RootPath
    ON VSCodeProject (RootPath COLLATE NOCASE);
```

Migration is idempotent (existing migration runner pattern).
`tags` and `paths` are not stored in the DB — they live only in
`projects.json` and are preserved on each sync.

## 7. Constants

All new strings land in `gitmap/constants/constants_vscode.go`
(per `mem://tech/constants-structure`). Examples:

```go
VSCodePMExtensionDir   = "alefragnani.project-manager"
VSCodePMProjectsFile   = "projects.json"
VSCodeUserSubDir       = "User"
VSCodeGlobalStorageDir = "globalStorage"
VSCodeUserDataDirNameWin = "Code"
VSCodeUserDataMacRel     = "Library/Application Support/Code"
VSCodeUserDataLinuxRel   = ".config/Code"
```

No magic strings in resolver, sync, or command files.

## 8. Errors

Per `mem://tech/code-red-error-management` — every failure logs to
`os.Stderr` with the standardized format and surfaces a non-zero exit.
Specific cases:

| Condition                                  | Message                                                                  |
|--------------------------------------------|--------------------------------------------------------------------------|
| User-data root missing                     | `vscode: user data directory not found at <path>`                        |
| Extension dir missing                      | `vscode: project-manager extension dir not found at <path>`              |
| `projects.json` corrupt JSON               | `vscode: projects.json is not valid JSON: <err> (left untouched)`        |
| Atomic rename failure                      | `vscode: failed to commit projects.json: <err>`                          |
| `code` CLI missing in PATH (gitmap code)   | actionable install hint above                                            |

## 9. Acceptance criteria

1. `scan` populates `projects.json` (all DB rows reconciled), no VS Code launch.
2. Re-running `scan` is idempotent — zero duplicates by `rootPath`.
3. `gitmap code` inside a git repo → repo root added, alias = basename, VS Code opens.
4. `gitmap code myalias` → alias overridden to `myalias`, VS Code opens.
5. `gitmap code myalias D:\anywhere` → non-git path added and opened.
6. `gitmap as newalias` mirrors `name` change to `projects.json`.
7. Foreign entries in `projects.json` are preserved across all operations.
8. Atomic writes — kill -9 mid-write never produces corrupted JSON.
9. Cross-platform: identical behavior on Windows / macOS / Linux.
10. The string `git map` (with a space) appears nowhere in code, help, or logs.

## 10. Flow diagrams

### `gitmap code [alias] [path]`

```
user → gitmap code [alias] [path]
        │
        ▼
resolve rootPath  (arg | git root | cwd)
        │
        ▼
upsert VSCodeProject  (DB, key=RootPath)
        │
        ▼
sync projects.json   (atomic, preserve foreign + user fields)
        │
        ▼
exec  code "<rootPath>"  (or print install-hint if missing)
```

### `gitmap scan`

```
user → gitmap scan [dir]
        │
        ▼
walk → existing scan/upsert pipeline
        │
        ▼
for each new repo  →  VSCodeProject upsert
        │
        ▼
sync projects.json   (no VS Code launch)
        │
        ▼
print summary  (added / updated / skipped / total)
```

## 11. Out of scope (v1)

- Multi-root workspace support via `paths`.
- Auto-derived tags (`git`, `node`, `go`, language detection).
- Reverse sync (mutating DB from external `projects.json` edits).
- Profile assignment (always `""` on insert; preserved on upsert).

## 12. See also

- `mem://features/vscode-project-manager-sync`
- `mem://tech/database-architecture`
- `mem://tech/database-location`
- `mem://tech/constants-structure`
- `mem://tech/code-red-error-management`
- `gitmap/constants/constants_vscode.go` (existing executable discovery constants)
