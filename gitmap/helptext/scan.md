# gitmap scan

Scan a directory tree for Git repositories and record them in the local database.

## Alias

s

## Usage

    gitmap scan [dir] [flags]

## Flags

| Flag | Default | Description |
|------|---------|-------------|
| --config \<path\> | ./data/config.json | Config file path |
| --mode ssh\|https | https | Clone URL style |
| --output csv\|json\|terminal | terminal | Output format |
| --output-path \<dir\> | ./.gitmap/output | Output directory |
| --github-desktop | false | Add repos to GitHub Desktop |
| --open | false | Open output folder after scan |
| --quiet | false | Suppress clone help section |

## Prerequisites

- None (this is typically the first command you run)

## Examples

### Example 1: Scan a directory

    gitmap scan D:\wp-work

**Output:**

    Scanning D:\wp-work...
    [1/42] github/user/my-api
    [2/42] github/user/web-app
    [3/42] github/org/billing-svc
    ...
    Found 42 repositories
    ✓ Output written to ./.gitmap/output/
    ✓ Database updated (42 repos)

### Example 2: Scan with JSON output and SSH URLs

    gitmap scan ~/work --output json --mode ssh

**Output:**

    Scanning ~/work...
    Found 18 repositories
    ✓ .gitmap/output/gitmap.json written
    ✓ .gitmap/output/gitmap.csv written
    ✓ Clone URLs use SSH format (git@github.com:...)

### Example 3: Scan and register with GitHub Desktop

    gitmap scan D:\repos --github-desktop

**Output:**

    Scanning D:\repos...
    Found 12 repositories
    ✓ Output written to ./.gitmap/output/
    Registering with GitHub Desktop...
    [1/12] my-api... added
    [2/12] web-app... already registered
    ✓ 12 repos synced to GitHub Desktop (10 new, 2 existing)

### Example 4: Scan current directory quietly

    gitmap s . --quiet --output csv

**Output:**

    Scanning current directory...
    Found 7 repositories
    ✓ .gitmap/output/gitmap.csv written

## See Also

- [rescan](rescan.md) — Re-scan using cached parameters
- [clone](clone.md) — Clone repos from scan output
- [status](status.md) — View repo statuses after scanning
- [desktop-sync](desktop-sync.md) — Sync scanned repos to GitHub Desktop
- [export](export.md) — Export scanned data
