#!/usr/bin/env bash
# check-constants-naming.sh
#
# Guards the `gitmap/constants/` package against new constants that lack a
# domain-prefix convention. Per spec/12-consolidated-guidelines/02-go-code-style.md
# the four canonical prefixes are:
#
#   CmdXxx       -> command names and aliases (e.g. CmdScan, CmdReleaseAlias)
#   MsgXxx       -> user-facing strings, error messages, format templates
#                   (also covers ErrXxx — treated as a Msg sub-family)
#   FlagXxx      -> flag names and descriptions (e.g. FlagDryRun, FlagDescBump)
#   DefaultXxx   -> default values used when no override is provided
#
# Reality check: the package today contains ~2,700 constants spread across
# ~50 historical prefixes (Doctor*, Git*, Hint*, Tool*, Term*, Choco*, ...).
# Renaming all of them in one PR is not viable, so this guard uses a
# grandfather baseline:
#
#   .github/scripts/constants-baseline.txt
#
# Every constant present at the time the guard was introduced is listed in
# the baseline and exempt. Only constants added AFTER the baseline must use
# one of the four canonical prefixes. The intent is to bend the curve so the
# package converges on Cmd/Msg/Flag/Default over time without breaking the
# build today.
#
# To regenerate the baseline after an approved rename pass:
#   bash .github/scripts/check-constants-naming.sh --regenerate-baseline
#
# Exit 1 on any new violation; 0 otherwise.

set -euo pipefail

CONST_DIR="${CONST_DIR:-gitmap/constants}"
BASELINE_FILE="${BASELINE_FILE:-.github/scripts/constants-baseline.txt}"
ALLOWED_PREFIX_REGEX='^(Cmd|Msg|Err|Flag|Default)'

if [ ! -d "$CONST_DIR" ]; then
  echo "::error::constants directory not found: $CONST_DIR"
  exit 1
fi

# extract_constants prints the unique sorted list of top-level constant names
# defined in the package. Matches both `Name = value` and `Name Type = value`
# forms inside `const (...)` blocks (and bare `const Name = ...`).
extract_constants() {
  grep -rhE '^\s*[A-Z][a-zA-Z0-9]+\s*(=|[A-Za-z0-9.]+\s*=)' "$CONST_DIR"/*.go 2>/dev/null \
    | grep -oE '^[[:space:]]*[A-Z][a-zA-Z0-9]+' \
    | sed 's/^[[:space:]]*//' \
    | sort -u
}

# regenerate mode: rewrite the baseline file from current source.
if [ "${1:-}" = "--regenerate-baseline" ]; then
  extract_constants > "$BASELINE_FILE"
  echo "Regenerated $BASELINE_FILE ($(wc -l < "$BASELINE_FILE") constants)."
  exit 0
fi

if [ ! -f "$BASELINE_FILE" ]; then
  echo "::error::baseline file missing: $BASELINE_FILE"
  echo "::error::run: bash .github/scripts/check-constants-naming.sh --regenerate-baseline"
  exit 1
fi

current=$(mktemp)
trap 'rm -f "$current"' EXIT
extract_constants > "$current"

# New constants = in current, NOT in baseline.
# `comm -23` requires both inputs sorted (they are: -u above + sorted baseline).
new_consts=$(comm -23 "$current" <(sort -u "$BASELINE_FILE"))

if [ -z "$new_consts" ]; then
  echo "No new constants added since baseline."
  exit 0
fi

violations=""
while IFS= read -r name; do
  [ -z "$name" ] && continue
  if ! [[ "$name" =~ $ALLOWED_PREFIX_REGEX ]]; then
    # Locate the file:line for a helpful error pointer.
    location=$(grep -nE "^\s*${name}\s*(=|[A-Za-z0-9.]+\s*=)" "$CONST_DIR"/*.go 2>/dev/null | head -1)
    violations+="${name}|${location}"$'\n'
  fi
done <<< "$new_consts"

if [ -n "$violations" ]; then
  echo "::error::New constants must use one of the canonical prefixes: Cmd*, Msg*, Err*, Flag*, Default*"
  echo "::error::See spec/12-consolidated-guidelines/02-go-code-style.md"
  echo ""
  while IFS='|' read -r name location; do
    [ -z "$name" ] && continue
    echo "::error::  $name"
    if [ -n "$location" ]; then
      echo "::error::    at $location"
    fi
  done <<< "$violations"
  echo ""
  echo "::error::Rename to use a canonical prefix, e.g.:"
  echo "::error::  ScanTimeout      -> DefaultScanTimeout"
  echo "::error::  HelpReleaseFlags -> MsgHelpReleaseFlags"
  echo "::error::  CloneVerbose     -> FlagCloneVerbose"
  echo "::error::  GitMainBranch    -> DefaultGitMainBranch"
  echo ""
  echo "::error::If the constant pre-existed and the diff is a rename, also run:"
  echo "::error::  bash .github/scripts/check-constants-naming.sh --regenerate-baseline"
  exit 1
fi

echo "All new constants use a canonical Cmd/Msg/Err/Flag/Default prefix."
