# gitmap changelog

Display release notes for a specific version or the latest release.

## Alias

cl

## Usage

    gitmap changelog [version] [--open] [--source]

## Flags

| Flag | Default | Description |
|------|---------|-------------|
| --latest | false | Show latest release notes |
| --open | false | Open changelog in browser |
| --source | false | Show source file path |

## Prerequisites

- Must be inside a Git repository with release metadata

## Examples

### Example 1: Show latest changelog

    gitmap changelog --latest

**Output:**

    ═══════════════════════════════════════════
    v2.22.0 — 2025-03-10
    ═══════════════════════════════════════════
    - Add interactive TUI with dashboard view
    - Add zip-group support for release assets
    - Add alias suggest --apply flag
    - Fix watch interval minimum validation
    - Fix cd picker numbering off-by-one
    ═══════════════════════════════════════════
    5 changes | Branch: release/v2.22.0

### Example 2: Show changelog for a specific version

    gitmap cl v2.20.0

**Output:**

    ═══════════════════════════════════════════
    v2.20.0 — 2025-02-28
    ═══════════════════════════════════════════
    - Add clear-release-json command with --dry-run
    - Add release-pending metadata source recovery
    - Fix orphaned metadata detection prompt
    ═══════════════════════════════════════════
    3 changes | Branch: release/v2.20.0

### Example 3: Show changelog with source path

    gitmap changelog v2.20.0 --source

**Output:**

    Source: .gitmap/release/v2.20.0.json
    v2.20.0 — 2025-02-28
    - Add clear-release-json command with --dry-run
    - Add release-pending metadata source recovery

### Example 4: Open changelog in browser

    gitmap changelog --latest --open

**Output:**

    v2.22.0 — 2025-03-10
    ...
    Opening in browser... done

## See Also

- [release](release.md) — Create a release
- [list-versions](list-versions.md) — List release tags
- [list-releases](list-releases.md) — List stored release metadata
- [release-pending](release-pending.md) — Show unreleased commits
