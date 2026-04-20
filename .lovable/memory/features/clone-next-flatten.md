---
name: Clone-next flatten mode
description: Clone-next flattens by default — clones into base-name folder (no version suffix) and tracks version history in DB
type: feature
---

## Feature: Default Flatten for `clone-next`

### Status: ✅ Implemented (v2.75.0)

### Behavior (Default — No Flag Required)

When `gitmap cn v++` is used on a repo like `macro-ahk-v15`:

1. Target clone folder is `macro-ahk/` (base name only, version suffix stripped)
2. If `macro-ahk/` already exists, remove it entirely first (no prompt)
3. Clone the target repo (e.g., `macro-ahk-v16`) into `macro-ahk/`
4. The remote URL still points to `macro-ahk-v16` on GitHub — only the local folder name is flattened
5. Update the database with the new version info
6. Record version transition in `RepoVersionHistory`

### `gitmap clone <url>` Auto-Flatten

When cloning a versioned URL without a custom folder name:
- `gitmap clone https://github.com/user/wp-onboarding-v13` → clones into `wp-onboarding/`
- `gitmap clone https://github.com/user/wp-onboarding-v13 my-folder` → clones into `my-folder/` (no flatten)

### Database Schema

#### Repos table — version columns
- `CurrentVersionTag TEXT DEFAULT ''` — e.g., "v16"
- `CurrentVersionNum INTEGER DEFAULT 0` — e.g., 16

#### `RepoVersionHistory` table
Tracks every version transition with from/to tags, numbers, flattened path, and timestamp.

### Related Commands
- `gitmap version-history` (`vh`) — Display all version transitions for the current repo
