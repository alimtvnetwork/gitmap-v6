# gitmap update

Self-update gitmap from the source repository. Pulls latest, rebuilds, and deploys.

## Alias

None

## Usage

    gitmap update [--repo-path <path>] [--verbose]

## Flags

| Flag | Description |
|------|-------------|
| `--repo-path <path>` | Override the source repository path for this run |
| `--verbose` | Enable verbose logging to file |

## Prerequisites

- Git must be installed
- Source repository must be accessible

## Examples

### Example 1: Update to a newer version

    gitmap update

**Output:**

    ■ Checking for updates...
    Current version: v2.19.0
    Latest version:  v2.22.0
    v2.19.0 → v2.22.0
    ■ Pulling latest source...
    ■ Building gitmap.exe...
    ■ Deploying to E:\bin-run\gitmap.exe...
    ✓ Updated to v2.22.0
    → Run 'gitmap changelog --latest' to see what's new

### Example 2: Already up to date

    gitmap update

**Output:**

    ■ Checking for updates...
    Current version: v2.22.0
    Latest version:  v2.22.0
    ✓ Already up to date (v2.22.0)

### Example 3: Update with custom repo path

    gitmap update --repo-path C:\Projects\gitmap-v4

**Output:**

    → Repo path: C:\Projects\gitmap-v4
    ■ Pulling latest source...
    ■ Building gitmap.exe...
    ✓ Updated to v2.49.1

### Example 4: Update with network error

    gitmap update

**Output:**

    ■ Checking for updates...
    ✗ Failed to pull latest: network timeout
    → Check your internet connection and try again

### Example 5: No source repo linked — clone into new path

    gitmap update

**Output:**

    ⚠ The saved source repository path no longer exists on disk.

    Enter the new path to the gitmap source repo: D:\gitmap

    ■ Path does not exist. Cloning gitmap source into D:\gitmap...
    Cloning into 'D:\gitmap'...
    ✓ Cloned successfully.
    → Repo path: D:\gitmap
    ■ Pulling latest source...
    ■ Building gitmap.exe...
    ✓ Updated to v2.56.1

### Example 6: No source repo linked and no path provided

    gitmap update

**Output:**

    ✗ Source repository path not found.

    This binary was installed without a linked source repo, so 'update'
    cannot locate the code to pull and rebuild.

    How to fix:

      Option 1 — Re-install via the one-liner (recommended):
        irm https://raw.githubusercontent.com/.../install.ps1 | iex

      Option 2 — Clone the repo and build from source:
        git clone https://github.com/.../gitmap-v4.git C:\gitmap-src
        cd C:\gitmap-src
        .\run.ps1

      Option 3 — Download the latest release manually:
        https://github.com/.../gitmap-v4/releases/latest

      Option 4 — Use --repo-path to specify it manually:
        gitmap update --repo-path C:\gitmap-src

      After building from source, 'gitmap update' will work automatically.

## Updater Fallback

If no source repo is available and `gitmap-updater` is installed, `gitmap update`
automatically delegates to it. The updater checks GitHub releases and downloads
the latest version without needing a local source checkout.

    gitmap update

**Output (with updater installed):**

    → No source repo found. Delegating to gitmap-updater...

    ■ Checking for updates...
    Current version: v2.49.0
    Latest version:  v2.49.1
    v2.49.0 → v2.49.1
    ■ Downloading installer for v2.49.1...
    ■ Running installer...
    ✓ Update complete.

## Troubleshooting

If you installed gitmap from a GitHub release (e.g. via the one-liner installer),
the binary does not have a source repo path embedded. You have three choices:

1. **Install `gitmap-updater`** — it handles updates via GitHub releases automatically.
2. **Use `--repo-path`** to point at a local clone for a one-time update.
3. **Clone and rebuild** from source so future updates work automatically.

## See Also

- [version](version.md) — Check current version
- [doctor](doctor.md) — Diagnose installation issues
- [changelog](changelog.md) — View release notes for new version
