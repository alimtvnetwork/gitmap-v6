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
# defined in the package. Uses awk to track `const (` block state so it never
# picks up identifiers from multi-line string literals (SQL keywords like
# SET/WHERE/AND) or struct-field assignments (ProjectTypeGoID int64 = 1 in a
# var block, Manifest.AppSubdir = ... in init code).
#
# Recognized forms:
#   const Name = ...                       (single-line)
#   const Name Type = ...                  (single-line typed)
#   const ( ... \n  Name = ... \n ... )    (block, untyped or typed)
#
# Inside a `const (` block we accept lines whose first non-whitespace token is
# a PascalCase identifier followed by either `=`, `<type> =`, or just whitespace
# then end-of-line (iota grouping). Inside multi-line raw strings (delimited
# by backticks) we skip everything.
extract_constants() {
  # IMPORTANT: this awk script must be portable across BOTH gawk and mawk.
  # GitHub Actions Ubuntu runners ship mawk as /usr/bin/awk, and mawk does
  # NOT support gawk's 3-argument match(string, regex, array) form — using
  # it causes a silent parse failure that makes the whole CI step exit 1
  # with no `::error::` output. We therefore use only POSIX-portable awk:
  # 2-arg match() + RSTART / RLENGTH + substr().
  awk '
    BEGIN { in_const = 0; in_rawstr = 0 }
    {
      line = $0

      # Toggle raw-string state on every backtick on the line.
      n = gsub(/`/, "`", line)
      if (n % 2 == 1) { in_rawstr = 1 - in_rawstr; next }
      if (in_rawstr) { next }

      # Track const block entry/exit.
      if (match(line, /^[[:space:]]*const[[:space:]]*\(/)) { in_const = 1; next }
      if (in_const && match(line, /^[[:space:]]*\)/))      { in_const = 0; next }

      # Strip leading whitespace.
      sub(/^[[:space:]]+/, "", line)

      # Single-line const declaration outside a block: peel "const ".
      out_of_block_decl = 0
      if (match(line, /^const[[:space:]]+/)) {
        out_of_block_decl = 1
        line = substr(line, RSTART + RLENGTH)
      } else if (!in_const) {
        # Outside a block and no leading "const " keyword — skip.
        next
      }

      # Now line should start with a PascalCase identifier.
      if (!match(line, /^[A-Z][A-Za-z0-9]+/)) { next }
      name = substr(line, RSTART, RLENGTH)
      rest = substr(line, RSTART + RLENGTH)

      # For block entries we require the rest to be `=`, `<type> =`, end of
      # line (iota continuation), or a comment. For peeled "const Name ..."
      # lines we accept anything (single-line consts always have `=`).
      if (out_of_block_decl) { print name; next }

      sub(/^[[:space:]]+[A-Za-z0-9_.]+/, "", rest) # optional type
      sub(/^[[:space:]]*/, "", rest)
      if (rest == "" || substr(rest, 1, 1) == "=" || substr(rest, 1, 2) == "//") {
        print name
      }
    }
  ' "$CONST_DIR"/*.go | LC_ALL=C sort -u
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
