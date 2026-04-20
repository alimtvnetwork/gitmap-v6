# Self-Update Orchestration (Windows-Safe)

Part of [PowerShell Build & Deploy Patterns](02-powershell-build-deploy.md).

When a CLI updates itself from a PATH-managed executable, use a two-phase handoff so the active binary lock is released before deploy.

## Phase 1: Handoff from active binary

1. `tool update` creates a handoff copy in the same active binary directory (for example `toolname-update-<pid>.exe`, fallback to `%TEMP%` if locked).
2. It launches the handoff copy with a hidden worker command (e.g. `update-runner`).
3. The parent **waits for the worker** using foreground/blocking execution (`cmd.Run()`). This keeps the terminal session stable. The handoff copy is a different file so there is no lock conflict with deploy.

## Phase 2: Execute update from handoff copy

1. Resolve repo root from embedded/configured repo path.
2. Run `run.ps1 -Update` (full pipeline: pull, build, deploy).
3. Sync active PATH binary using **rename-first** strategy in update mode:
   - Rename active binary to `.old` (Windows allows renaming a running exe).
   - Copy deployed binary to the active path.
   - Fall back to copy-retry loop (20 x 500ms) only if rename fails.
4. Read and print versions from the binaries (before update and after update) using `tool version`.
5. Show latest notes using the updated binary (`tool changelog --latest`).
6. Run `tool update-cleanup` to remove handoff and `.old` artifacts.

## Critical Rules

- The parent MUST use `cmd.Run()` (foreground/blocking). Using `cmd.Start()` + `os.Exit(0)` (async) breaks the terminal session.
- PATH sync MUST use rename-first in update mode. Copy-overwrite fails on Windows when any process holds the binary.
- Generated update scripts MUST NOT contain `Read-Host` or any interactive prompts — they run in non-interactive PowerShell sessions.

## Required Validation

- Fail the update if active version still does not match deployed version after sync.
- Version/changelog output must come from the updated executable, not static constants.
- Cleanup must run after successful update so rollback artifacts exist during deploy.

## Cross-References

- Generic spec: [05-handoff-mechanism.md](../08-generic-update/05-handoff-mechanism.md) §Solution: Copy-and-Handoff
- Generic spec: [06-cleanup.md](../08-generic-update/06-cleanup.md) §Cleanup Command
- Generic spec: [08-repo-path-sync.md](../08-generic-update/08-repo-path-sync.md) §Post-Deploy Repo Path Sync
- Self-update mechanism: [03-self-update-mechanism.md](03-self-update-mechanism.md)
