# self-install

Install (or re-install) the gitmap binary on this machine.

## Synopsis

```
gitmap self-install [--dir <path>] [--yes] [--version <tag>]
                    [--profile auto|both|zsh|bash|pwsh|fish]
                    [--show-path] [--force-lock]
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
     `raw.githubusercontent.com/alimtvnetwork/gitmap-v5/main/gitmap/scripts/`.
3. Writes the script to a temp file (UTF-8 BOM on PowerShell), runs it
   with `-InstallDir` / `--dir`, and forwards `--version` if pinned.

## --profile <mode>

Controls which shell profile files receive the PATH snippet on Unix.
Defaults to `auto`.

| Mode   | Writes PATH to                                             |
|--------|------------------------------------------------------------|
| `auto` | Detected shell profiles (current behavior, default)        |
| `both` | zsh + bash + .profile + fish (if installed) + pwsh         |
| `zsh`  | `~/.zshrc` and `~/.zprofile` only                          |
| `bash` | `~/.bashrc` and `~/.bash_profile` only                     |
| `pwsh` | `~/.config/powershell/Microsoft.PowerShell_profile.ps1`    |
| `fish` | `~/.config/fish/config.fish`                               |

`--profile both` is the recommended mode when running from pwsh on
macOS — it guarantees both the zsh login profile (used by
Terminal.app / iTerm) and the pwsh profile receive the PATH update so
gitmap works immediately whichever shell you open.

`--dual-shell` is kept as a hidden alias for `--profile both`.

## Examples

```
gitmap self-install
gitmap self-install --yes
gitmap self-install --dir D:\dev\gitmap
gitmap self-install --version v3.0.0
gitmap self-install --profile both        # macOS / pwsh user
gitmap self-install --profile pwsh        # only touch the pwsh profile
gitmap self-install --show-path           # audit which profiles got written
```

## See also

- `gitmap self-uninstall` — remove gitmap from this machine
- `gitmap update` — pull a newer build from the source repo
