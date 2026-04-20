# self-install

Install (or re-install) the gitmap binary on this machine.

## Synopsis

```
gitmap self-install [--dir <path>] [--yes] [--version <tag>]
```

## What it does

1. Resolves the install directory:
   - `--dir <path>` if supplied.
   - Default with prompt otherwise:
     - **Windows**: `D:\gitmap`
     - **Unix**: `~/.local/bin/gitmap`
   - `--yes` accepts the default without prompting.
2. Loads the platform installer from one of two sources:
   - **Embedded**: `install.ps1` / `install.sh` shipped inside the binary
     via `go:embed`. No network needed.
   - **Remote** (fallback): downloaded from
     `raw.githubusercontent.com/alimtvnetwork/gitmap-v4/main/gitmap/scripts/`.
3. Writes the script to a temp file (UTF-8 BOM on PowerShell), runs it
   with `-InstallDir` / `--dir`, and forwards `--version` if pinned.

## Examples

```
gitmap self-install
gitmap self-install --yes
gitmap self-install --dir D:\dev\gitmap
gitmap self-install --version v3.0.0
```

## See also

- `gitmap self-uninstall` — remove gitmap from this machine
- `gitmap update` — pull a newer build from the source repo
