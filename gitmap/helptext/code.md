# gitmap code

Register a path with the VS Code Project Manager extension and open it in
VS Code in one step.

## Aliases

None.

## Usage

    gitmap code [alias] [path]

| Form                              | Behavior                                                                   |
|-----------------------------------|----------------------------------------------------------------------------|
| `gitmap code`                     | Use the git repo root if inside one, else the current directory. Alias = folder basename. |
| `gitmap code myalias`             | Same path resolution as above; alias overridden to `myalias`.              |
| `gitmap code myalias D:\anywhere` | Use any path (no git requirement). Alias = `myalias`.                       |

## What it does

1. Resolves the absolute `rootPath`.
2. Upserts a row in the gitmap `VSCodeProject` table keyed by `rootPath`.
3. Atomically syncs the entry into the alefragnani.project-manager
   `projects.json` file (preserving foreign entries and any user-set
   `tags` / `paths` / `enabled` / `profile` fields).
4. Launches VS Code on the resolved path via the `code` CLI.

`projects.json` location is derived per OS by first discovering the VS Code
**user-data root** (`%APPDATA%\Code` / `~/Library/Application Support/Code` /
`$XDG_CONFIG_HOME/Code`), then appending
`User/globalStorage/alefragnani.project-manager/projects.json`. The full path
is never hardcoded.

## Examples

    cd ~/code/my-app
    gitmap code                # rootPath = repo root, name = "my-app"

    gitmap code backend        # name overridden to "backend"

    gitmap code docs ~/Documents/spec   # any path, name = "docs"

## Errors

| Condition                                  | Exit | Action                                                  |
|--------------------------------------------|------|---------------------------------------------------------|
| Provided path does not exist               | 1    | Error printed; nothing written.                         |
| VS Code user-data dir missing              | 1    | Suggests installing VS Code.                            |
| project-manager extension dir missing      | 1    | Suggests installing the extension first.                |
| `code` CLI not on PATH                     | 0    | DB and `projects.json` still updated; install hint.     |

## See also

- `gitmap as` — register a short alias for a repo.
- `gitmap scan` — bulk-syncs every discovered repo into `projects.json`.
- Spec: `spec/01-vscode-project-manager-sync/README.md`
