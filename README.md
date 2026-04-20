<div align="center">

<img src="gitmap/assets/icon.png" alt="GitMap icon" width="80" height="80">

# GitMap

**Git repository scanner, manager, and navigator CLI**

[![CI](https://github.com/alimtvnetwork/gitmap-v4/actions/workflows/ci.yml/badge.svg)](https://github.com/alimtvnetwork/gitmap-v4/actions/workflows/ci.yml)
[![golangci-lint](https://github.com/alimtvnetwork/gitmap-v4/actions/workflows/ci.yml/badge.svg?event=push)](https://github.com/alimtvnetwork/gitmap-v4/actions/workflows/ci.yml)
[![Vulncheck](https://github.com/alimtvnetwork/gitmap-v4/actions/workflows/vulncheck.yml/badge.svg)](https://github.com/alimtvnetwork/gitmap-v4/actions/workflows/vulncheck.yml)
[![GitHub Release](https://img.shields.io/github/v/release/alimtvnetwork/gitmap-v4?style=flat-square&label=version)](https://github.com/alimtvnetwork/gitmap-v4/releases)
[![Go](https://img.shields.io/badge/Go-1.23+-00ADD8?style=flat-square&logo=go&logoColor=white)](https://go.dev)
[![Platform](https://img.shields.io/badge/platform-Windows%20%7C%20Linux%20%7C%20macOS-lightgrey?style=flat-square)](https://github.com/alimtvnetwork/gitmap-v4)
[![License](https://img.shields.io/badge/license-MIT-green?style=flat-square)](./LICENSE)
[![Go Report Card](https://goreportcard.com/badge/github.com/alimtvnetwork/gitmap-v4?style=flat-square)](https://goreportcard.com/report/github.com/alimtvnetwork/gitmap-v4)

_Scan, catalog, clone, and manage all your Git repositories from a single CLI._

</div>

---

## Quick Start

### Install — Quick (pick your install folder)

Prompts for the install drive/folder (press Enter for the default), then runs the full installer.

```powershell
# Windows (PowerShell) — interactive, choose drive/folder
irm https://raw.githubusercontent.com/alimtvnetwork/gitmap-v4/main/install-quick.ps1 | iex
```

```bash
# Linux / macOS — interactive, choose folder
curl -fsSL https://raw.githubusercontent.com/alimtvnetwork/gitmap-v4/main/install-quick.sh | bash
```

### Install — Full (defaults, no prompt)

```powershell
# Windows (PowerShell)
irm https://raw.githubusercontent.com/alimtvnetwork/gitmap-v4/main/gitmap/scripts/install.ps1 | iex
```

```bash
# Linux / macOS
curl -fsSL https://raw.githubusercontent.com/alimtvnetwork/gitmap-v4/main/gitmap/scripts/install.sh | sh
```

### Uninstall — Quick (one-liner)

Removes the gitmap binary, deploy folder, PATH entries, and (optionally) the user data folder. First tries the canonical `gitmap self-uninstall`; falls back to a manual sweep if gitmap is no longer on PATH.

```powershell
# Windows (PowerShell)
irm https://raw.githubusercontent.com/alimtvnetwork/gitmap-v4/main/uninstall-quick.ps1 | iex
```

```bash
# Linux / macOS
curl -fsSL https://raw.githubusercontent.com/alimtvnetwork/gitmap-v4/main/uninstall-quick.sh | bash
```

Useful flags (both scripts):

| Flag | Effect |
|---|---|
| `-Yes` / `-y` `--yes` | Skip the "delete user data?" prompt and assume yes |
| `-KeepData` / `--keep-data` | Always keep `%APPDATA%\gitmap` (Windows) or `~/.config/gitmap` (Unix) |
| `-InstallDir` / `--dir` | Override the auto-detected deploy root |

### Scan repos and see results

```bash
gitmap scan ~/projects
gitmap ls
```

### Navigate and pull

```bash
gitmap cd my-api
gitmap pull --all
```

Every command supports `--help` or `-h` for detailed usage with examples.

---

## Installation

### One-Liner Install (recommended)

**Quick installers — prompt for an install folder, then delegate to the full installer:**

```powershell
# Windows (PowerShell)
irm https://raw.githubusercontent.com/alimtvnetwork/gitmap-v4/main/install-quick.ps1 | iex
```

```bash
# Linux / macOS
curl -fsSL https://raw.githubusercontent.com/alimtvnetwork/gitmap-v4/main/install-quick.sh | bash
```

**Windows (PowerShell — full bootstrap, works on any machine):**

```powershell
Set-ExecutionPolicy Bypass -Scope Process -Force; [System.Net.ServicePointManager]::SecurityProtocol = [System.Net.ServicePointManager]::SecurityProtocol -bor 3072; iex ((New-Object System.Net.WebClient).DownloadString('https://raw.githubusercontent.com/alimtvnetwork/gitmap-v4/main/gitmap/scripts/install.ps1'))
```

**Windows (short form, PowerShell 5+):**

```powershell
irm https://raw.githubusercontent.com/alimtvnetwork/gitmap-v4/main/gitmap/scripts/install.ps1 | iex
```

**Linux / macOS (Bash):**

```bash
curl -fsSL https://raw.githubusercontent.com/alimtvnetwork/gitmap-v4/main/gitmap/scripts/install.sh | sh
```

### Installer Options

**Windows (PowerShell):**

| Flag | Description | Example |
|------|-------------|---------|
| `-Version` | Pin a specific release | `-Version v2.51.0` |
| `-InstallDir` | Custom install directory | `-InstallDir C:\tools\gitmap` |
| `-Arch` | Force architecture (`amd64`, `arm64`) | `-Arch arm64` |
| `-NoPath` | Skip adding to user PATH | `-NoPath` |

**Linux / macOS (Bash):**

| Flag | Description | Example |
|------|-------------|---------|
| `--version` | Pin a specific release | `--version v2.55.0` |
| `--dir` | Custom install directory | `--dir /opt/gitmap` |
| `--arch` | Force architecture (`amd64`, `arm64`) | `--arch arm64` |
| `--no-path` | Skip adding to PATH | `--no-path` |

**Specific version install (one-liner):**

```powershell
irm https://raw.githubusercontent.com/alimtvnetwork/gitmap-v4/main/gitmap/scripts/install.ps1 | iex; Install-Gitmap -Version "v2.51.0"
```

**Specific version + custom directory (one-liner):**

```powershell
irm https://raw.githubusercontent.com/alimtvnetwork/gitmap-v4/main/gitmap/scripts/install.ps1 | iex; Install-Gitmap -Version "v2.51.0" -InstallDir "D:\DevTools\gitmap"
```

**Custom directory install (downloaded script):**

```powershell
.\install.ps1 -InstallDir "D:\DevTools\gitmap"
```

**Pinned version + custom directory (downloaded script):**

```powershell
.\install.ps1 -Version v2.51.0 -InstallDir "C:\tools\gitmap"
```

**Linux / macOS — specific version:**

```bash
curl -fsSL https://raw.githubusercontent.com/alimtvnetwork/gitmap-v4/main/gitmap/scripts/install.sh | sh -s -- --version v2.51.0
```

> **Tip:** Use `gitmap list-versions` to see all available release versions before pinning.

### Clone & Setup (Development)

```bash
git clone https://github.com/alimtvnetwork/gitmap-v4.git
cd gitmap-v4
./setup.sh
```

The setup script installs the pre-commit hook (golangci-lint), verifies your Go toolchain, and downloads dependencies. See [CONTRIBUTING.md](CONTRIBUTING.md) for the full development workflow.

---

## What It Does

A portable CLI that scans directory trees for Git repositories, extracts clone URLs and branch info, and outputs structured data. Every scan produces **all outputs** automatically:

- **Terminal** — formatted table to stdout
- **CSV** — `gitmap.csv`
- **JSON** — `gitmap.json`
- **Folder Structure** — `folder-structure.md` (tree view of discovered repos)

All files are written to `.gitmap/output/` at the root of the scanned directory.

---

## Command Reference

<div align="center">

### Scanning & Discovery

</div>

| Command | Alias | Description |
|---------|-------|-------------|
| `scan` | `s` | Scan directory for Git repos |
| `rescan` | `rsc` | Re-scan previously scanned directories |
| `list` | `ls` | Show all tracked repos with slugs |

```bash
gitmap scan ~/projects --output json --mode ssh
gitmap ls go                    # list Go projects
gitmap rescan                   # re-scan all known directories
```

→ [scan](gitmap/helptext/scan.md) · [rescan](gitmap/helptext/rescan.md) · [list](gitmap/helptext/list.md)

---

<div align="center">

### Cloning & Sync

</div>

| Command | Alias | Description |
|---------|-------|-------------|
| `clone` | `c` | Clone from a structured file OR a direct URL |
| `clone-next` | `cn` | Clone next versioned iteration of current repo |
| `desktop-sync` | `ds` | Sync tracked repos with GitHub Desktop |

```bash
# clone from a structured file
gitmap clone json --target-dir ./restored
gitmap clone csv                                # auto-resolves to ./gitmap-output/gitmap.csv
gitmap clone ./gitmap-output/gitmap.json --safe-pull
gitmap clone ./gitmap-output/gitmap.json --github-desktop

# clone a single repo by URL (auto-flattens versioned URLs)
gitmap clone https://github.com/alimtvnetwork/gitmap-v4
gitmap clone https://github.com/alimtvnetwork/gitmap-v4 my-folder
gitmap clone git@github.com:alimtvnetwork/gitmap-v4.git my-folder
gitmap clone https://github.com/alimtvnetwork/gitmap-v4 --replace   # see spec 96

# clone-next: jump to the next (or specific) versioned sibling
gitmap cn v++                                   # my-app-v3 -> my-app-v4
gitmap cn v15 --delete                          # jump to v15, delete current
gitmap cn v++ --create-remote                   # create GitHub repo if missing
gitmap cn v++ --no-flatten                      # keep nested folder layout
```

→ [clone](gitmap/helptext/clone.md) · [clone-next](gitmap/helptext/clone-next.md) · [desktop-sync](gitmap/helptext/desktop-sync.md)

---

<div align="center">

### Move & Merge

</div>

| Command | Alias | Description |
|---------|-------|-------------|
| `mv` | `move` | Move LEFT into RIGHT, then delete LEFT |
| `merge-both` | — | Fill missing files on BOTH sides; prompt on conflicts |
| `merge-left` | — | Copy from RIGHT into LEFT; prompt on conflicts |
| `merge-right` | — | Copy from LEFT into RIGHT; prompt on conflicts |

Each side (LEFT / RIGHT) can be a local folder OR a remote URL.
URL endpoints are auto-cloned (or pulled if already cloned), and
the result is committed + pushed back when the operation completes.

```bash
# move: classic file copy + delete source
gitmap mv ./gitmap-v4 ./gitmap-v4
gitmap mv ./gitmap-v4 https://github.com/alimtvnetwork/gitmap-v4
gitmap mv https://github.com/alimtvnetwork/gitmap-v4 ./another-folder
gitmap mv https://github.com/alimtvnetwork/gitmap-v4 \
         https://github.com/alimtvnetwork/gitmap-v4

# merge-both: bidirectional fill (each side gains what the other has)
gitmap merge-both ./gitmap-v4 ./gitmap-v4
gitmap merge-both ./gitmap-v4 https://github.com/alimtvnetwork/gitmap-v4
gitmap merge-both https://github.com/alimtvnetwork/gitmap-v4 \
                  https://github.com/alimtvnetwork/gitmap-v4

# merge-left: take RIGHT into LEFT
gitmap merge-left ./gitmap-v4 ./gitmap-v4
gitmap merge-left ./local https://github.com/alimtvnetwork/gitmap-v4

# merge-right: take LEFT into RIGHT
gitmap merge-right ./gitmap-v4 ./gitmap-v4
gitmap merge-right ./local https://github.com/alimtvnetwork/gitmap-v4

# bypass conflict prompts: source-side wins by default
gitmap merge-right ./gitmap-v4 ./gitmap-v4 -y
gitmap merge-both  ./gitmap-v4 ./gitmap-v4 -y --prefer-newer

# pin remote branch + preview
gitmap merge-right ./local https://github.com/owner/repo:develop
gitmap mv ./gitmap-v4 ./gitmap-v4 --dry-run
```

Conflict prompt keys: **L**eft / **R**ight / **S**kip /
**A**ll-left / **B**all-right / **Q**uit. Pass `-y` (or `-a`) to
bypass; combine with `--prefer-left` / `--prefer-right` /
`--prefer-newer` / `--prefer-skip` to override the default policy.

→ [spec/01-app/97-move-and-merge.md](spec/01-app/97-move-and-merge.md)

---

<div align="center">

### Git Operations

</div>

| Command | Alias | Description |
|---------|-------|-------------|
| `pull` | `p` | Pull a specific repo by name |
| `exec` | `x` | Run git command across all repos |
| `status` | `st` | Show repo status dashboard |
| `watch` | `w` | Live-refresh repo status dashboard |
| `has-any-updates` | `hau` | Check if remote has new commits |
| `latest-branch` | `lb` | Find most recently updated remote branch |

```bash
gitmap pull --group work --all
gitmap exec fetch --prune
gitmap watch --interval 10 --group work
gitmap lb 5 --format csv
```

→ [pull](gitmap/helptext/pull.md) · [exec](gitmap/helptext/exec.md) · [status](gitmap/helptext/status.md) · [watch](gitmap/helptext/watch.md) · [latest-branch](gitmap/helptext/latest-branch.md)

---

<div align="center">

### Navigation & Organization

</div>

| Command | Alias | Description |
|---------|-------|-------------|
| `cd` | `go` | Navigate to a tracked repo directory |
| `group` | `g` | Manage repo groups / activate for batch ops |
| `multi-group` | `mg` | Select multiple groups for batch operations |
| `alias` | `a` | Assign short names to repos |
| `as` | `s-alias` | Register the current Git repo + name in one shot (run from inside the repo) |
| `diff-profiles` | `dp` | Compare repos across two profiles |

```bash
gitmap cd my-api
gitmap g work && gitmap g pull
gitmap mg backend,frontend && gitmap mg status
gitmap alias set api github/user/api-gateway
gitmap as backend           # registers the current repo as 'backend' + adds it to the DB
gitmap as                   # uses the folder basename as the alias
gitmap alias suggest --apply
```

→ [cd](gitmap/helptext/cd.md) · [group](gitmap/helptext/group.md) · [multi-group](gitmap/helptext/multi-group.md) · [alias](gitmap/helptext/alias.md) · [as](gitmap/helptext/as.md) · [diff-profiles](gitmap/helptext/diff-profiles.md)

---

<div align="center">

### Release & Versioning

</div>

| Command | Alias | Description |
|---------|-------|-------------|
| `release` | `r` | Create release branch, tag, and push |
| `release-alias` | `ra` | Release a repo by its registered alias from anywhere |
| `release-alias-pull` | `rap` | `release-alias` with implicit `--pull` (pull-then-release) |
| `release-self` | `rs` | Release gitmap itself from any directory |
| `release-branch` | `rb` | Create release branch without tagging |
| `temp-release` | `tr` | Create lightweight temp release branches |

```bash
gitmap release --bump patch
gitmap release --bump minor --bin --compress --checksums
gitmap release v3.0.0 -N "Major redesign"

# Release any aliased repo from anywhere — no `cd` required
gitmap as my-api                                   # one-time, run from inside the repo
gitmap release-alias my-api v1.4.0
gitmap ra my-api v1.4.0 --pull                     # pull --ff-only, then release
gitmap release-alias-pull my-api v1.4.0            # equivalent thin verb
gitmap rap my-api v1.4.0 --dry-run

gitmap release-self --bump patch
gitmap tr 10 v1.$$ -s 5
```

> Dirty trees are auto-stashed before `release-alias` runs and restored on
> exit. Pass `--no-stash` to abort instead, or `--dry-run` to preview.

→ [release](gitmap/helptext/release.md) · [release-alias](gitmap/helptext/release-alias.md) · [release-alias-pull](gitmap/helptext/release-alias-pull.md) · [release-self](gitmap/helptext/release-self.md) · [release-branch](gitmap/helptext/release-branch.md) · [temp-release](gitmap/helptext/temp-release.md)

---

<div align="center">

### Release History & Info

</div>

| Command | Alias | Description |
|---------|-------|-------------|
| `changelog` | `cl` | Show release notes |
| `changelog-generate` | `cg` | Auto-generate changelog from commits |
| `list-versions` | `lv` | List all available Git release tags |
| `list-releases` | `lr` | List release metadata from database |
| `release-pending` | `rp` | Show unreleased commits since last tag |
| `revert` | — | Revert to a specific release version |
| `clear-release-json` | `crj` | Remove orphaned release metadata files |
| `prune` | `pr` | Delete stale release branches |

```bash
gitmap changelog v2.49.0
gitmap release-pending
gitmap list-versions --json --limit 5
gitmap cg --from v2.22.0 --to v2.24.0 --write
gitmap revert v2.48.0
```

→ [changelog](gitmap/helptext/changelog.md) · [list-versions](gitmap/helptext/list-versions.md) · [list-releases](gitmap/helptext/list-releases.md) · [release-pending](gitmap/helptext/release-pending.md) · [revert](gitmap/helptext/revert.md) · [clear-release-json](gitmap/helptext/clear-release-json.md) · [prune](gitmap/helptext/prune.md)

> **CI Pipeline:** Pushing a `release/*` branch or `v*` tag triggers GitHub Actions to cross-compile 6 targets, generate checksums, and create a GitHub release with changelog and install instructions.

---

<div align="center">

### Data, Profiles & Bookmarks

</div>

| Command | Alias | Description |
|---------|-------|-------------|
| `export` | `ex` | Export database to file |
| `import` | `im` | Import repos from file |
| `profile` | `pf` | Manage database profiles |
| `bookmark` | `bk` | Save and run bookmarked commands |
| `db-reset` | — | Reset the local SQLite database |

```bash
gitmap export && gitmap import gitmap-export.json
gitmap profile create work && gitmap profile switch work
gitmap bookmark save daily scan ~/projects
gitmap bookmark run daily
```

→ [export](gitmap/helptext/export.md) · [import](gitmap/helptext/import.md) · [profile](gitmap/helptext/profile.md) · [bookmark](gitmap/helptext/bookmark.md) · [db-reset](gitmap/helptext/db-reset.md)

---

<div align="center">

### History, Stats & Author Amendment

</div>

| Command | Alias | Description |
|---------|-------|-------------|
| `history` | `hi` | Show CLI command execution history |
| `history-reset` | `hr` | Clear command execution history |
| `stats` | `ss` | Show aggregated usage and performance metrics |
| `amend` | `am` | Rewrite commit author info |
| `amend-list` | `al` | List previous author amendments |

```bash
gitmap history --limit 10
gitmap stats --json
gitmap amend --name "John Doe" --email "john@example.com" --dry-run
```

→ [history](gitmap/helptext/history.md) · [stats](gitmap/helptext/stats.md) · [amend](gitmap/helptext/amend.md) · [amend-list](gitmap/helptext/amend-list.md)

---

<div align="center">

### Project Detection

</div>

| Command | Alias | Description |
|---------|-------|-------------|
| `go-repos` | `gr` | List detected Go projects |
| `node-repos` | `nr` | List detected Node.js projects |
| `react-repos` | `rr` | List detected React projects |
| `cpp-repos` | `cr` | List detected C++ projects |
| `csharp-repos` | `csr` | List detected C# projects |

```bash
gitmap go-repos
gitmap csharp-repos --json
```

→ [go-repos](gitmap/helptext/go-repos.md) · [node-repos](gitmap/helptext/node-repos.md) · [react-repos](gitmap/helptext/react-repos.md) · [cpp-repos](gitmap/helptext/cpp-repos.md) · [csharp-repos](gitmap/helptext/csharp-repos.md)

---

<div align="center">

### Tool Installation

</div>

Install developer tools and databases via platform package managers directly from the CLI.

#### Core Tools

| Tool | Keyword | Description |
|------|---------|-------------|
| Visual Studio Code | `vscode` | Code editor |
| Node.js | `node` | JavaScript runtime (includes Yarn, Bun) |
| pnpm | `pnpm` | Fast package manager |
| Python | `python` | Programming language |
| Go | `go` | Programming language |
| Git + LFS + gh | `git`, `git-lfs`, `gh` | Version control ecosystem |
| GitHub Desktop | `github-desktop` | Git GUI |
| C++ (MinGW) | `cpp` | C++ compiler |
| PHP | `php` | Programming language |
| PowerShell | `powershell` | Shell |

#### Databases

| Tool | Keyword | Description |
|------|---------|-------------|
| MySQL | `mysql` | Open-source relational database |
| MariaDB | `mariadb` | MySQL-compatible fork |
| PostgreSQL | `postgresql` | Advanced relational database |
| SQLite | `sqlite` | Embedded file-based database |
| MongoDB | `mongodb` | Document-oriented NoSQL |
| CouchDB | `couchdb` | Document database with REST API |
| Redis | `redis` | In-memory key-value store |
| Cassandra | `cassandra` | Wide-column distributed NoSQL |
| Neo4j | `neo4j` | Graph database |
| Elasticsearch | `elasticsearch` | Full-text search and analytics |
| DuckDB | `duckdb` | Analytical columnar database |

```bash
# Install a tool
gitmap install node
gitmap install postgresql

# Pin a specific version
gitmap install node --version 20.11.1

# Check if installed (no install)
gitmap install go --check

# Preview install command
gitmap install redis --dry-run

# Force a specific package manager
gitmap install vscode --manager winget

# List all supported tools
gitmap install --list

# Uninstall a tool
gitmap uninstall redis
```

**Default package managers by platform:**

| Platform | Default | Fallback |
|----------|---------|----------|
| Windows | Chocolatey | Winget |
| macOS | Homebrew | — |
| Linux | apt | snap |

Override in `config.json` → `install.defaultManager` or per-command with `--manager`.

→ [install](gitmap/helptext/install.md)

---

<div align="center">

### SSH Key Management

</div>

| Command | Alias | Description |
|---------|-------|-------------|
| `ssh` | — | Generate and manage SSH keys |

```bash
gitmap ssh --name work --path ~/.ssh/id_rsa_work
gitmap ssh cat --name work
gitmap ssh list
gitmap ssh config
```

→ [ssh](gitmap/helptext/ssh.md)

---

<div align="center">

### Zip Groups (Release Archives)

</div>

| Command | Alias | Description |
|---------|-------|-------------|
| `zip-group` | `z` | Manage named file collections for release archives |

```bash
gitmap z create docs-bundle
gitmap z add docs-bundle ./README.md ./CHANGELOG.md ./docs/
gitmap z show docs-bundle
gitmap release v3.0.0 --zip-group docs-bundle
```

→ [zip-group](gitmap/helptext/zip-group.md)

---

<div align="center">

### Environment & File-Sync

</div>

| Command | Alias | Description |
|---------|-------|-------------|
| `env` | `ev` | Manage persistent environment variables and PATH |
| `task` | `tk` | Manage file-sync watch tasks |

```bash
gitmap env set GOPATH "/home/user/go"
gitmap env path add /usr/local/go/bin
gitmap env list
gitmap task create my-sync --src ./src --dest ./backup
gitmap tk run my-sync --interval 10
```

→ [env](gitmap/helptext/env.md) · [task](gitmap/helptext/task.md)

---

<div align="center">

### Utilities

</div>

| Command | Alias | Description |
|---------|-------|-------------|
| `setup` | — | Interactive first-time configuration wizard |
| `doctor` | — | Diagnose PATH, deploy, and version issues |
| `update` | — | Self-update from source repo or gitmap-updater |
| `version` | `v` | Show version number |
| `completion` | `cmp` | Generate shell tab-completion scripts |
| `interactive` | `i` | Launch full-screen interactive TUI |
| `docs` | `d` | Open documentation website in browser |
| `seo-write` | `sw` | Auto-commit SEO messages |
| `gomod` | `gm` | Rename Go module path across repo |
| `dashboard` | `db` | Generate interactive HTML dashboard |

```bash
gitmap doctor --fix-path
gitmap update
gitmap completion powershell
gitmap interactive --refresh 10
gitmap dashboard --limit 100 --open
```

→ [setup](gitmap/helptext/setup.md) · [doctor](gitmap/helptext/doctor.md) · [update](gitmap/helptext/update.md) · [completion](gitmap/helptext/completion.md) · [interactive](gitmap/helptext/interactive.md) · [dashboard](gitmap/helptext/dashboard.md)

---

## Build & Deploy

### Makefile Targets

| Target | Description |
|--------|-------------|
| `make all` | Lint → Test → Build (default) |
| `make setup` | Install hooks and dev tools |
| `make lint` | Run golangci-lint |
| `make test` | Run all tests |
| `make build` | Compile for current platform |
| `make vulncheck` | Scan dependencies for CVEs |
| `make release BUMP=patch` | Lint, test, then release |
| `make release-dry` | Preview release without executing |
| `make clean` | Remove build artifacts |

### Build from Source

```bash
cd gitmap && go build -o ../gitmap .
```

### Build via run.ps1 (Windows)

```powershell
.\run.ps1                        # Full pipeline: pull, build, deploy
.\run.ps1 -R scan                # Build + scan parent folder
.\run.ps1 -R scan D:\repos --mode ssh
```

| Flag | Description |
|------|-------------|
| `-NoPull` | Skip `git pull` |
| `-NoDeploy` | Skip deploy step |
| `-Update` | Update mode with post-update validation |
| `-R` | Run gitmap after build (trailing args forwarded) |

---

## Project Structure

```
gitmap/                        # Go CLI source
  cmd/                         # Command handlers
  constants/                   # All string constants (no magic strings)
  completion/                  # Shell completion generators
  release/                     # Release workflow and semver
  store/                       # SQLite database layer
  formatter/                   # Output formatters
  helptext/                    # Embedded markdown help files
  scripts/                     # Install/uninstall scripts
gitmap-updater/                # Standalone update tool
spec/                          # Specifications per feature
src/                           # React documentation site
.github/workflows/             # CI/CD pipelines
```

---

## Web UI Dashboard

GitMap includes a React-based documentation and dashboard UI:

```bash
npm install && npm run dev     # opens at http://localhost:5173
```

**Tech Stack:** Vite · TypeScript · React · shadcn/ui · Tailwind CSS

---

## Author

<div align="center">

### [Md. Alim Ul Karim](https://www.google.com/search?q=alim+ul+karim)

**[Creator & Lead Architect](https://alimkarim.com)** | [Chief Software Engineer](https://www.google.com/search?q=alim+ul+karim), [Riseup Asia LLC](https://riseup-asia.com)

</div>

A system architect with **20+ years** of professional software engineering experience across enterprise, fintech, and distributed systems. His technology stack spans **.NET/C# (18+ years)**, **JavaScript (10+ years)**, **TypeScript (6+ years)**, and **Golang (4+ years)**.

Recognized as a **top 1% talent at Crossover** and one of the top software architects globally. He is also the **Chief Software Engineer of [Riseup Asia LLC](https://riseup-asia.com)** and maintains an active presence on **[Stack Overflow](https://stackoverflow.com/users/361646/alim-ul-karim)** (2,452+ reputation, member since 2010) and **LinkedIn** (12,500+ followers).

|  |  |
|---|---|
| **Website** | [alimkarim.com](https://alimkarim.com/) · [my.alimkarim.com](https://my.alimkarim.com/) |
| **LinkedIn** | [linkedin.com/in/alimkarim](https://linkedin.com/in/alimkarim) |
| **Stack Overflow** | [stackoverflow.com/users/361646/alim-ul-karim](https://stackoverflow.com/users/361646/alim-ul-karim) |
| **Google** | [Alim Ul Karim](https://www.google.com/search?q=Alim+Ul+Karim) |
| **Role** | Chief Software Engineer, [Riseup Asia LLC](https://riseup-asia.com) |

### Riseup Asia LLC

[Top Leading Software Company in WY (2026)](https://riseup-asia.com)

| | |
|---|---|
| **Website** | [riseup-asia.com](https://riseup-asia.com) |
| **Facebook** | [riseupasia.talent](https://www.facebook.com/riseupasia.talent/) |
| **LinkedIn** | [Riseup Asia](https://www.linkedin.com/company/105304484/) |
| **YouTube** | [@riseup-asia](https://www.youtube.com/@riseup-asia) |

## License

This project is licensed under the [MIT License](./LICENSE).
