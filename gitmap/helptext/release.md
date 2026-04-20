# gitmap release

Create a release: tag, push, and optionally publish a GitHub release.

## Alias

r

## Usage

    gitmap release [version] [flags]

## Flags

| Flag | Default | Description |
|------|---------|-------------|
| --assets \<path\> | — | Attach files to release |
| --commit \<sha\> | HEAD | Release from specific commit |
| --branch \<name\> | current | Release from branch |
| --bump major\|minor\|patch | — | Auto-increment version |
| --notes \<text\> / -N \<text\> | — | Release notes or title for the release |
| --draft | false | Create unpublished draft |
| --dry-run | false | Preview without executing |
| --compress | false | Wrap assets in .zip (Windows) or .tar.gz archives |
| --checksums | false | Generate SHA256 checksums.txt for assets |
| --bin / -b | false | Opt-in: cross-compile Go binaries locally |
| --no-assets | false | Skip Go binary cross-compilation |
| --targets \<list\> | all 6 | Cross-compile targets: windows/amd64,linux/arm64 |
| --list-targets | false | Print resolved target matrix and exit |
| --zip-group \<name\> | — | Include a persistent zip group as a release asset |
| -Z \<path\> | — | Add ad-hoc file or folder to zip as a release asset |
| --bundle \<name.zip\> | — | Bundle all -Z items into a single named archive |
| --no-commit | false | Skip post-release auto-commit and push |
| -y / --yes | false | Auto-confirm all prompts (e.g. commit) |
| --verbose | false | Write detailed debug log to a timestamped file |

## Prerequisites

- Must be inside a Git repository with at least one commit
- GitHub CLI (`gh`) recommended for publishing

## Orphaned Metadata Recovery

If a `.gitmap/release/vX.Y.Z.json` file exists but neither the Git tag nor
the release branch is found, the command warns and prompts:

    ⚠ Release metadata exists for v2.3.10 but no tag or branch was found.
    → Do you want to remove the release JSON and proceed? (y/N):

Answering `y` removes the stale JSON file and proceeds with the release.
Answering `n` (or pressing Enter) aborts the release.

## Examples

### Example 1: Release with auto-bump (patch)

    gitmap release --bump patch

**Output:**

    v2.21.0 → v2.21.1
    Creating branch release/v2.21.1... done
    Creating tag v2.21.1... done
    Pushing branch and tag... done
    Cross-compiling Go binaries...
      ✓ gitmap_v2.21.1_windows_amd64.exe
      ✓ gitmap_v2.21.1_windows_arm64.exe
      ✓ gitmap_v2.21.1_linux_amd64
      ✓ gitmap_v2.21.1_linux_arm64
      ✓ gitmap_v2.21.1_darwin_amd64
      ✓ gitmap_v2.21.1_darwin_arm64
    Uploading to GitHub... done
    ✓ Metadata written to .gitmap/release/v2.21.1.json
    ✓ Released v2.21.1

### Example 2: Dry-run preview with minor bump

    gitmap r --bump minor --dry-run

**Output:**

    [DRY RUN] v2.21.0 → v2.22.0
    [DRY RUN] Would create branch release/v2.22.0
    [DRY RUN] Would create tag v2.22.0
    [DRY RUN] Would push branch and tag
    [DRY RUN] Would cross-compile 6 targets
    [DRY RUN] Would upload assets to GitHub
    No changes made.

### Example 3: Release with assets and compression

    gitmap release v3.0.0 --assets ./dist/ --compress --checksums

**Output:**

    Creating branch release/v3.0.0... done
    Creating tag v3.0.0... done
    Pushing branch and tag... done
    Attaching assets from ./dist/...
      ✓ Compressed gitmap_v3.0.0_windows_amd64.exe → gitmap_v3.0.0_windows_amd64.zip
      ✓ Compressed gitmap_v3.0.0_linux_amd64 → gitmap_v3.0.0_linux_amd64.tar.gz
      ✓ Generated checksums.txt (SHA256)
    Uploading to GitHub... done
    ✓ Released v3.0.0

### Example 4: Release with a persistent zip group

    gitmap release v3.0.0 --zip-group docs-bundle

**Output:**

    Creating branch release/v3.0.0... done
    Creating tag v3.0.0... done
    Pushing branch and tag... done
    ✓ Compressed docs-bundle → docs-bundle_v3.0.0.zip
    Uploading to GitHub... done
    ✓ Released v3.0.0

### Example 5: Release with notes

    gitmap release --bump patch -N 'Hotfix for auth timeout'

**Output:**

    v2.21.0 → v2.21.1
    → Release notes: Hotfix for auth timeout
    Creating branch release/v2.21.1... done
    Creating tag v2.21.1... done
    Pushing branch and tag... done
    ✓ Metadata written to .gitmap/release/v2.21.1.json
    ✓ Released v2.21.1

### Example 6: Release as a draft from a specific branch

    gitmap release v3.0.0-rc1 --branch develop --draft

**Output:**

    Creating branch release/v3.0.0-rc1 from develop... done
    Creating tag v3.0.0-rc1... done
    Pushing branch and tag... done
    ✓ Draft release created (not published)
    ✓ Metadata written to .gitmap/release/v3.0.0-rc1.json
    ✓ Released v3.0.0-rc1 (draft)

### Example 7: List resolved cross-compile targets

    gitmap release --list-targets

**Output:**

    Resolved 6 target(s):
    Source: built-in defaults
      windows/amd64
      windows/arm64
      linux/amd64
      linux/arm64
      darwin/amd64
      darwin/arm64

### Example 8: Release with auto-confirm and install hints (gitmap repo)

    gitmap release v2.61.0 -y

**Output:**

    Creating branch release/v2.61.0... done
    Creating tag v2.61.0... done
    Pushing branch and tag... done
    ✓ Metadata written to .gitmap/release/v2.61.0.json

    Release v2.61.0 complete.


      📦 Install gitmap v2.61.0

      🪟 Windows (PowerShell)
         irm https://raw.githubusercontent.com/alimtvnetwork/gitmap-v4/main/gitmap/scripts/install.ps1 | iex

      🐧 Linux / macOS
         curl -fsSL https://raw.githubusercontent.com/alimtvnetwork/gitmap-v4/main/gitmap/scripts/install.sh | sh

Install hints only appear when the repo's remote origin matches the
gitmap source repository. Non-gitmap repos are unaffected.

## CI Release Pipeline

Pushing a `release/*` branch or `v*` tag triggers a GitHub Actions
workflow that automatically:

1. Cross-compiles 6 Go binaries (windows/linux/darwin × amd64/arm64)
2. Compresses assets (.zip for Windows, .tar.gz for Unix)
3. Generates SHA256 checksums
4. Creates version-pinned `install.ps1` and `install.sh` installers
5. Publishes a GitHub Release with changelog, metadata, and all assets

Local `--bin` builds are opt-in for development; the CI pipeline is
the recommended path for production releases.

## See Also

- [release-branch](release-branch.md) — Create a release branch without tagging
- [release-pending](release-pending.md) — Show unreleased commits
- [changelog](changelog.md) — View release notes
- [list-versions](list-versions.md) — List release tags
- [list-releases](list-releases.md) — List stored release metadata
- [revert](revert.md) — Revert to a previous release
- [zip-group](zip-group.md) — Manage zip group definitions
