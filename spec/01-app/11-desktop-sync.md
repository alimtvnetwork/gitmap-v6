# Desktop Sync Command

## Overview

The `desktop-sync` command reads the previously generated scan output
and registers all discovered repositories with GitHub Desktop — without
requiring a full re-scan.

## Usage

```bash
gitmap desktop-sync
```

No flags. The command always reads from `./.gitmap/output/gitmap.json`
in the current working directory.

## Prerequisites

1. **Scan output exists**: `.gitmap/output/` folder with `gitmap.json`
   must exist in the current directory. Run `gitmap scan` first.
2. **GitHub Desktop installed**: The `github` CLI must be on `PATH`.

## Behavior

### Step 1 — Validate Output Directory

Check if `.gitmap/output/` exists in the current directory.
If missing, print an error and exit with code 1.

### Step 2 — Validate JSON File

Check if `.gitmap/output/gitmap.json` exists.
If missing, print an error and exit with code 1.

### Step 3 — Check GitHub Desktop

Verify `github` CLI is available via `exec.LookPath`.
If not found, print an error and exit with code 1.

### Step 4 — Load Records

Parse `gitmap.json` into `[]ScanRecord`. If JSON parsing fails,
print an error and exit with code 1.

### Step 5 — Sync Each Repo

For each record:

| Condition | Action |
|-----------|--------|
| `absolutePath` is empty | Print failure, count as failed |
| Path doesn't exist on disk | Print skip message, count as skipped |
| Path exists | Call `github <absolutePath>`, count result |

### Step 6 — Print Summary

```
GitHub Desktop sync: 8 added, 2 skipped, 1 failed
```

## Error Handling

| Scenario | Behavior |
|----------|----------|
| No `.gitmap/output/` directory | Print error, exit 1 |
| No `gitmap.json` file | Print error, exit 1 |
| GitHub Desktop not installed | Print error, exit 1 |
| Invalid JSON | Print parse error, exit 1 |
| Repo path missing on disk | Skip with message, continue |
| GitHub Desktop CLI fails for a repo | Log failure, continue |

## Defensive Coding

- Every external call is wrapped with error handling.
- Per-repo failures do not stop the batch — all repos are attempted.
- Missing paths are detected before calling the CLI (avoid panics).
- Summary always prints, even if all repos fail.
- No panics — all errors are caught and logged.

## Implementation

| File | Responsibility |
|------|---------------|
| `cmd/desktopsync.go` | Command handler, validation, sync loop |
| `constants/constants.go` | All messages and the command name |
| `cmd/root.go` | Dispatch routing for `desktop-sync` |

## Flow Diagram

```
User runs: gitmap desktop-sync
   │
   ├─ Check ./.gitmap/output/ exists → error if missing
   ├─ Check ./.gitmap/output/gitmap.json exists → error if missing
   ├─ Check GitHub Desktop CLI available → error if missing
   ├─ Parse JSON → []ScanRecord
   │
   └─ For each record:
      ├─ No absolutePath → failed
      ├─ Path doesn't exist → skipped
      └─ Path exists → github <path>
         ├─ Success → added
         └─ Error → failed
   │
   └─ Print summary: added / skipped / failed
```
