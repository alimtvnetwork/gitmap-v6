# self-uninstall

Remove gitmap from this machine.

## Synopsis

```
gitmap self-uninstall [--confirm] [--keep-data] [--keep-snippet]
```

## What it removes

| Target              | Path                                                |
|---------------------|-----------------------------------------------------|
| Binary + deploy dir | Folder containing the running `gitmap` / `gitmap.exe` |
| Data dir            | `<deploy>/.gitmap/` (skip with `--keep-data`)       |
| PATH snippet        | `# gitmap shell wrapper v*` block in shell profile  |
| Completion files    | `gitmap-completion.{bash,zsh,fish}` in deploy dir   |

The shell profile cleared is `Microsoft.PowerShell_profile.ps1` on
Windows and `~/.bashrc` on Unix.

## Confirmation

Without `--confirm`, the command lists the targets and waits for the
user to type `yes`. Pass `--confirm` to skip the prompt.

## Windows handoff

The running `gitmap.exe` cannot delete itself while loaded. When the
binary lives inside the deploy dir, self-uninstall copies itself to
`%TEMP%\gitmap-handoff-<pid>.exe` and re-execs the hidden
`self-uninstall-runner` from there, which performs the deletion and
then schedules its own removal via `cmd.exe /C ... del`.

## Re-installing afterwards

```
# Linux / macOS
curl -fsSL https://raw.githubusercontent.com/alimtvnetwork/gitmap-v4/main/gitmap/scripts/install.sh | bash

# Windows (PowerShell)
irm https://raw.githubusercontent.com/alimtvnetwork/gitmap-v4/main/gitmap/scripts/install.ps1 | iex
```

…or, if you still have a copy of the binary somewhere, just run
`gitmap self-install`.

## Examples

```
gitmap self-uninstall
gitmap self-uninstall --confirm
gitmap self-uninstall --confirm --keep-data
gitmap self-uninstall --confirm --keep-snippet --keep-data
```

## See also

- `gitmap self-install` — install or re-install the gitmap binary
- `spec/04-generic-cli/21-post-install-shell-activation/04-idempotency.md` —
  marker block conventions used to locate and strip the PATH snippet
