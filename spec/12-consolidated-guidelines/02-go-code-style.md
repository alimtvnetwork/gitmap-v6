# 02 — Go Code Style

Go-specific coding conventions prioritizing readability and consistency.

## Positive Conditionals Only

Always write positive conditions. No `!`, no `!=`.

```go
// Correct
if len(args) > 0 { process(args) }

// Wrong
if len(args) != 0 { process(args) }
```

## Function Length — 8 to 15 Lines

Extract a helper when exceeding 15 lines. Each function does one thing.

## File Length — Max 200 Lines

| Signal | Action |
|--------|--------|
| 2+ unrelated function groups | Split into separate files |
| Large switch statement | Each case becomes a file |
| Mixed types and logic | Separate `model.go` from logic |

## Package Granularity — One Responsibility

| Package | Responsibility |
|---------|----------------|
| `cmd` | CLI routing and flag parsing |
| `config` | Config loading and merging |
| `constants` | All shared string literals |
| `model` | Shared data structures |
| `store` | Database access |

Rules: No circular imports. `cmd` orchestrates; others never import `cmd`. `model` and `constants` are leaf packages.

## No Magic Strings

Every literal used for comparison, defaults, or messages goes in `constants`.

## Naming Conventions

| Element | Convention | Example |
|---------|-----------|---------|
| Package | Lowercase, single word | `scanner` |
| Exported func | PascalCase, verb-led | `BuildRecords` |
| Unexported func | camelCase, verb-led | `parseFlags` |
| Constants | PascalCase | `DefaultBranch` |
| Files | Lowercase, single word | `terminal.go` |

## `cmd/` Helper Naming — Collision Guard

Every file in `gitmap/cmd/` shares one Go namespace, so generic helper names collide across files (e.g. two `func invokeRelease(...)` declarations break the build). Enforced by `.github/scripts/check-cmd-naming.sh` in CI.

| Verb prefix | Rule | Example |
|-------------|------|---------|
| `executeXxx` | Canonical handler, **one per top-level command**. Noun = the command itself. | `executeRelease`, `executeClone`, `executeScan` |
| `handleXxx` | Canonical sub-handler for a command branch. | `handleChangelogOpen`, `handleCompletionList` |
| `invokeXxx` | Helper. **Must** carry a domain-narrowing suffix after the noun. | `invokeAliasRelease` (not `invokeRelease`) |
| `persistXxx` | Helper. **Must** name the persisted entity + sink. | `persistReleaseToDB` (not `persistAll`) |
| `runOneXxx` | Per-item loop helper. **Must** name the item type. | `runOnePullJob`, `runOneScanRelease` (not `runOne`) |

**Forbidden** (CI fails): bare `invoke()`, `persist()`, `runOne()`, and `(invoke|persist|runOne)+(Release|Task|Job|Item|All|One|Cmd)` combinations. The trailing noun must be project-specific so the function's scope is unambiguous from its name alone.

## Error Handling

Check errors immediately. Return errors up the stack. In `cmd` handlers: print and `os.Exit(1)`. Never `panic` for expected conditions.

---

Source: `spec/05-coding-guidelines/02-go-code-style.md`, `spec/03-general/06-code-style-rules.md`
