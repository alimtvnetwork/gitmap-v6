# gitmap github-desktop

Register the current git repository (or a path you pass) with GitHub Desktop in one shot. No prior scan required.

## Alias

gd

## Usage

    gitmap github-desktop          # register cwd
    gitmap gd                      # short alias
    gitmap gd D:\path\to\repo      # register an explicit path

## Flags

None.

## Prerequisites

- GitHub Desktop must be installed and its `github` CLI must be on PATH.
- The target directory must contain a `.git` folder (or `.git` file for worktrees).

## How it differs from `desktop-sync`

| Command | Source of repos | Needs prior `gitmap scan`? |
|---------|-----------------|----------------------------|
| `github-desktop` (gd) | The single cwd (or explicit path arg) | No |
| `desktop-sync` (ds)   | Every repo in last scan output JSON    | Yes |

## Examples

### Example 1: Register the current repo

    cd D:\wp-work\riseup-asia\my-api
    gitmap gd

**Output:**

    Registering with GitHub Desktop: D:\wp-work\riseup-asia\my-api
    ✓ Registered with GitHub Desktop: D:\wp-work\riseup-asia\my-api

### Example 2: Register an explicit path

    gitmap github-desktop D:\projects\billing-svc

**Output:**

    Registering with GitHub Desktop: D:\projects\billing-svc
    ✓ Registered with GitHub Desktop: D:\projects\billing-svc

## See Also

- [desktop-sync](desktop-sync.md) — Bulk-sync every repo from the last scan
- [scan](scan.md) — Use `--github-desktop` to register during scan
- [clone](clone.md) — Use `--github-desktop` to register during clone
