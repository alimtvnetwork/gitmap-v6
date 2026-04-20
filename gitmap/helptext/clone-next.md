# gitmap clone-next

Clone the next or a specific versioned iteration of the current repository into the parent directory, using the base name (no version suffix) as the local folder.

## Alias

cn

## Usage

    gitmap clone-next <v++|vN> [flags]

## Flags

| Flag | Default | Description |
|------|---------|-------------|
| --delete | false | Auto-remove current versioned folder after clone |
| --keep | false | Keep current folder without prompting |
| --no-desktop | false | Skip GitHub Desktop registration |
| --ssh-key \<name\> | (none) | Use a named SSH key for the clone |
| --verbose | false | Write detailed debug log |
| --create-remote | false | Create target GitHub repo if missing (requires GITHUB_TOKEN) |

## Prerequisites

- Must be run inside a Git repository with a remote origin configured

## Flatten Behavior

By default, clone-next clones into the base name folder (without version suffix).
For example, running `gitmap cn v++` inside `macro-ahk-v11` will:
1. Clone `macro-ahk-v12` into `macro-ahk/` (not `macro-ahk-v12/`)
2. If `macro-ahk/` already exists, remove it first
3. The remote URL still points to `macro-ahk-v12` on GitHub
4. Record the version transition (v11 -> v12) in the database

## Examples

### Example 1: Increment version by one

    gitmap cn v++

**Output:**

    Removing existing macro-ahk for fresh clone...
    Cloning macro-ahk-v12 into macro-ahk (flattened)...
    ✓ Cloned macro-ahk-v12 into macro-ahk
    ✓ Recorded version transition v11 -> v12
    ✓ Registered macro-ahk-v12 with GitHub Desktop

### Example 2: Jump to a specific version with auto-delete

    gitmap cn v15 --delete

**Output:**

    Cloning macro-ahk-v15 into macro-ahk (flattened)...
    ✓ Cloned macro-ahk-v15 into macro-ahk
    ✓ Recorded version transition v12 -> v15
    ✓ Registered macro-ahk-v15 with GitHub Desktop
    ✓ Removed macro-ahk-v12

### Example 3: Lock detection when folder is in use

    gitmap cn v++ --delete

**Output:**

    Removing existing macro-ahk for fresh clone...
    Cloning macro-ahk-v12 into macro-ahk (flattened)...
    ✓ Cloned macro-ahk-v12 into macro-ahk
    ✓ Recorded version transition v11 -> v12
    ✓ Registered macro-ahk-v12 with GitHub Desktop
    Warning: could not remove macro-ahk-v11: unlinkat: access denied
    Checking for processes locking macro-ahk-v11...
    The following processes are using this folder:
      • Code.exe (PID 14320)
      • explorer.exe (PID 5928)
    Terminate these processes to allow deletion? [y/N] y
    Terminating Code.exe (PID 14320)...
    ✓ Terminated Code.exe
    Terminating explorer.exe (PID 5928)...
    ✓ Terminated explorer.exe
    Retrying folder removal...
    ✓ Removed macro-ahk-v11

## See Also

- [clone](clone.md) — Clone repos from output files
- [desktop-sync](desktop-sync.md) — Sync repos to GitHub Desktop
- [ssh](ssh.md) — Manage named SSH keys
