#!/usr/bin/env bash
# check-placeholder-modulepath.sh
#
# Fails CI if the legacy placeholder Go module path is reintroduced anywhere
# in the repo. The real module path is github.com/alimtvnetwork/gitmap-v5/gitmap
# (set in v3.27.0). The placeholder github.com/user/gitmap was a bootstrap
# leftover that broke `go install`, the Go module proxy, and Go Report Card.
#
# A stale `go.mod`, vendored snippet, generated docs page, or copy-pasted
# import block could silently bring it back. This guard catches that on PR.
#
# Allowed historical references (these files document the fix and MUST keep
# the literal placeholder string for context):
#   - CHANGELOG.md
#   - .github/scripts/check-placeholder-modulepath.sh  (this file)
#   - .github/workflows/ci.yml                         (job description)
#   - spec/02-app-issues/**                            (post-mortem docs)
#   - .gitmap/release/**                               (frozen release notes)
#   - .lovable/**                                      (AI session memory)

set -uo pipefail

PLACEHOLDER='github.com/user/gitmap'
REAL='github.com/alimtvnetwork/gitmap-v5/gitmap'

echo "→ Scanning for placeholder module path: ${PLACEHOLDER}"
echo "  (real path is ${REAL})"
echo ""

# grep -r with explicit excludes. -F = fixed string, -I = skip binaries,
# -n = line numbers, -l-less so we can show offending lines for fast triage.
matches=$(grep -rnFI "${PLACEHOLDER}" . \
  --exclude-dir='.git' \
  --exclude-dir='node_modules' \
  --exclude-dir='.gitmap' \
  --exclude-dir='.lovable' \
  --exclude-dir='spec' \
  --exclude='CHANGELOG.md' \
  --exclude='check-placeholder-modulepath.sh' \
  --exclude='ci.yml' \
  2>/dev/null) || true

if [ -z "${matches}" ]; then
  echo "✅ No placeholder module path references found."
  exit 0
fi

echo "::error::Placeholder module path '${PLACEHOLDER}' detected."
echo ""
echo "Offending lines:"
echo "${matches}" | sed 's/^/  /'
echo ""
echo "::error::Replace every occurrence with '${REAL}'."
echo "::error::See CHANGELOG.md v3.27.0 for the original migration."
echo ""
echo "Quick fix (run from repo root):"
echo "  grep -rln --exclude-dir=.git '${PLACEHOLDER}' . | \\"
echo "    xargs sed -i 's|${PLACEHOLDER}|${REAL}|g'"
exit 1
