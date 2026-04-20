# gitmap desktop-sync

Sync all tracked repositories with GitHub Desktop.

## Alias

ds

## Usage

    gitmap desktop-sync

## Flags

None.

## Prerequisites

- Run `gitmap scan` first to populate the database (see scan.md)
- GitHub Desktop must be installed

## Examples

### Example 1: Sync all repos to GitHub Desktop

    gitmap desktop-sync

**Output:**

    Syncing 42 repos to GitHub Desktop...
    [1/42] my-api... added
    [2/42] web-app... added
    [3/42] billing-svc... already registered
    [4/42] auth-gateway... added
    ...
    ✓ 42 repos synced to GitHub Desktop
      15 newly added
      27 already registered

### Example 2: Sync after a rescan (picks up new repos)

    gitmap rescan
    gitmap ds

**Output:**

    Re-scanning D:\wp-work...
    Found 44 repositories (+2 new)
    ✓ Database updated

    Syncing 44 repos to GitHub Desktop...
    [1/44] new-service... added
    [2/44] analytics-api... added
    ...
    ✓ 44 repos synced (2 new, 42 existing)

## See Also

- [scan](scan.md) — Scan directories to populate the database
- [clone](clone.md) — Clone repos from output files
- [list](list.md) — View tracked repos
