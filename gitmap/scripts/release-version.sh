#!/usr/bin/env bash
# ─────────────────────────────────────────────────────────────────────
# release-version.sh — version-pinned gitmap installer
#
# Installs EXACTLY the version requested via --version. Never resolves
# "latest", never auto-upgrades, never silently substitutes.
#
# Spec: spec/01-app/105-release-version-script.md
#
# Usage:
#   curl -fsSL https://gitmap.dev/scripts/release-version.sh \
#     | bash -s -- --version v3.36.0
#
# Options:
#   --version <tag>       REQUIRED. Release tag (e.g. v3.36.0).
#   --dir <path>          Install dir. Default: ~/.local/bin
#   --arch <arch>         Force amd64 or arm64. Default: auto-detect.
#   --no-path             Skip PATH modification.
#   --no-self-install     Skip the chained `gitmap self-install` step.
#   --allow-fallback      Use newest patch in same vMAJOR.MINOR if missing.
#   --quiet               Suppress prompts and progress output.
#   --json-errors         Emit fatal errors as a single-line JSON object on
#                         stderr (machine-readable contract for CI).
# ─────────────────────────────────────────────────────────────────────

# Re-exec under bash if invoked under sh/dash.
if [ -z "${BASH_VERSION:-}" ]; then
    if command -v bash >/dev/null 2>&1; then
        exec bash "$0" "$@"
    else
        printf 'release-version.sh requires bash.\n' >&2
        exit 1
    fi
fi

set -euo pipefail

REPO="alimtvnetwork/gitmap-v5"
BINARY_NAME="gitmap"

VERSION=""
INSTALL_DIR=""
ARCH_OVERRIDE=""
NO_PATH=0
NO_SELF_INSTALL=0
ALLOW_FALLBACK=0
QUIET=0
JSON_ERRORS=0
TMP_DIR=""

# Exit codes (spec 105)
EXIT_OK=0
EXIT_VERSION_MISSING=1
EXIT_NETWORK=2
EXIT_CHECKSUM=3
EXIT_UNSUPPORTED_ARCH=4
EXIT_PATH_FAIL=5
EXIT_SELF_INSTALL=6
EXIT_VERIFY=7

# Stable error code symbols (contract for JSON consumers).
ERR_INVALID_VERSION="INVALID_VERSION"
ERR_VERSION_NOT_FOUND="VERSION_NOT_FOUND"
ERR_NO_FALLBACK="NO_FALLBACK_AVAILABLE"
ERR_NON_INTERACTIVE="NON_INTERACTIVE_NO_SUBSTITUTE"
ERR_RECENT_LIST_FAILED="RECENT_LIST_FAILED"
ERR_USER_DECLINED="USER_DECLINED"
ERR_INVALID_CHOICE="INVALID_CHOICE"
ERR_NETWORK="NETWORK_ERROR"
ERR_CHECKSUM_MISMATCH="CHECKSUM_MISMATCH"
ERR_UNSUPPORTED_OS="UNSUPPORTED_OS"
ERR_UNSUPPORTED_ARCH="UNSUPPORTED_ARCH"
ERR_NO_ASSET="NO_MATCHING_ASSET"
ERR_EXTRACT_FAILED="EXTRACT_FAILED"
ERR_VERSION_MISMATCH="VERSION_MISMATCH"
ERR_SELF_INSTALL="SELF_INSTALL_FAILED"

cleanup() { [ -n "$TMP_DIR" ] && [ -d "$TMP_DIR" ] && rm -rf "$TMP_DIR"; }
trap cleanup EXIT

# ── Logging (ASCII only) ───────────────────────────────────────────
# Suppress all human-readable output when --json-errors is active so the
# JSON payload on stderr is the only thing consumers parse.
step() { [ "$QUIET" -eq 1 ] || [ "$JSON_ERRORS" -eq 1 ] || printf '  -> %s\n' "$*" >&2; }
ok()   { [ "$QUIET" -eq 1 ] || [ "$JSON_ERRORS" -eq 1 ] || printf '  OK %s\n' "$*" >&2; }
warn() { [ "$QUIET" -eq 1 ] || [ "$JSON_ERRORS" -eq 1 ] || printf '  !  %s\n' "$*" >&2; }
err()  { [ "$JSON_ERRORS" -eq 1 ] || printf '  X  %s\n' "$*" >&2; }

# json_escape escapes a value for safe inclusion in a JSON string literal.
# Handles backslash, double-quote, newline, tab — sufficient for our payloads.
json_escape() {
    printf '%s' "$1" \
        | sed -e 's/\\/\\\\/g' -e 's/"/\\"/g' \
        | awk 'BEGIN{ORS="\\n"} {print}' \
        | sed 's/\\n$//'
}

# fatal_error <code> <message> <exit_code> [details_json]
# Emits either a structured JSON error or a human-readable one, then exits.
fatal_error() {
    local code="$1" message="$2" exit_code="$3" details="${4:-{\}}"
    if [ "$JSON_ERRORS" -eq 1 ]; then
        local esc_msg esc_code esc_ver
        esc_msg="$(json_escape "$message")"
        esc_code="$(json_escape "$code")"
        esc_ver="$(json_escape "$VERSION")"
        printf '{"error":{"code":"%s","message":"%s","exitCode":%d,"requestedVersion":"%s","script":"release-version.sh","details":%s}}\n' \
            "$esc_code" "$esc_msg" "$exit_code" "$esc_ver" "$details" >&2
    else
        err "$message [code=$code]"
    fi
    exit "$exit_code"
}

# ── Arg parsing ─────────────────────────────────────────────────────
while [ $# -gt 0 ]; do
    case "$1" in
        --version)          VERSION="$2"; shift 2 ;;
        --dir)              INSTALL_DIR="$2"; shift 2 ;;
        --arch)             ARCH_OVERRIDE="$2"; shift 2 ;;
        --no-path)          NO_PATH=1; shift ;;
        --no-self-install)  NO_SELF_INSTALL=1; shift ;;
        --allow-fallback)   ALLOW_FALLBACK=1; shift ;;
        --quiet)            QUIET=1; shift ;;
        --json-errors)      JSON_ERRORS=1; shift ;;
        -h|--help)
            sed -n '2,24p' "$0"
            exit 0
            ;;
        *)
            fatal_error "INVALID_ARGUMENT" "Unknown argument: $1" $EXIT_VERSION_MISSING \
                "{\"argument\":\"$(json_escape "$1")\"}"
            ;;
    esac
done

# ── Version validation ─────────────────────────────────────────────
validate_version() {
    if [[ ! "$VERSION" =~ ^v[0-9]+\.[0-9]+\.[0-9]+(-[A-Za-z0-9.]+)?$ ]]; then
        fatal_error "$ERR_INVALID_VERSION" \
            "Invalid version tag: '$VERSION' (expected vMAJOR.MINOR.PATCH)" \
            $EXIT_VERSION_MISSING \
            "{\"provided\":\"$(json_escape "$VERSION")\",\"pattern\":\"^v[0-9]+\\\\.[0-9]+\\\\.[0-9]+\"}"
    fi
}

# ── OS / arch detection ────────────────────────────────────────────
detect_os() {
    local u
    u="$(uname -s)"
    case "$u" in
        Linux*)   echo "linux" ;;
        Darwin*)  echo "darwin" ;;
        MINGW*|MSYS*|CYGWIN*)
            err "release-version.sh does not run on Windows. Use release-version.ps1."
            exit $EXIT_UNSUPPORTED_ARCH
            ;;
        *)
            err "Unsupported OS: $u"
            exit $EXIT_UNSUPPORTED_ARCH
            ;;
    esac
}

detect_arch() {
    if [ -n "$ARCH_OVERRIDE" ]; then
        case "$ARCH_OVERRIDE" in
            amd64|arm64) echo "$ARCH_OVERRIDE"; return ;;
            *) err "Unsupported --arch: $ARCH_OVERRIDE"; exit $EXIT_UNSUPPORTED_ARCH ;;
        esac
    fi
    local m
    m="$(uname -m)"
    case "$m" in
        x86_64|amd64)        echo "amd64" ;;
        aarch64|arm64)       echo "arm64" ;;
        *) err "Unsupported architecture: $m"; exit $EXIT_UNSUPPORTED_ARCH ;;
    esac
}

# ── HTTP helpers ───────────────────────────────────────────────────
http_get() {
    local url="$1" out="${2:-}"
    local hdrs=(-H "User-Agent: gitmap-release-version-installer" -H "Accept: application/vnd.github+json")
    if [ -n "${GITHUB_TOKEN:-}" ]; then
        hdrs+=(-H "Authorization: Bearer $GITHUB_TOKEN")
    fi
    if command -v curl >/dev/null 2>&1; then
        if [ -n "$out" ]; then
            curl -fsSL "${hdrs[@]}" -o "$out" "$url"
        else
            curl -fsSL "${hdrs[@]}" "$url"
        fi
    elif command -v wget >/dev/null 2>&1; then
        if [ -n "$out" ]; then
            wget -qO "$out" "$url"
        else
            wget -qO- "$url"
        fi
    else
        err "Neither curl nor wget is available."
        exit $EXIT_NETWORK
    fi
}

github_api() {
    local path="$1"
    http_get "https://api.github.com/repos/$REPO$path" 2>/dev/null || return 1
}

# ── Resolve requested version (with optional fallback / prompt) ────
resolve_requested_version() {
    validate_version
    local body
    body="$(github_api "/releases/tags/$VERSION" || true)"
    if [ -n "$body" ] && echo "$body" | grep -q '"tag_name"'; then
        echo "$VERSION"
        return
    fi

    if [ "$ALLOW_FALLBACK" -eq 1 ]; then
        local fb
        fb="$(resolve_fallback_patch)"
        if [ -n "$fb" ]; then
            warn "Requested $VERSION missing; falling back to newest patch in series: $fb"
            echo "$fb"
            return
        fi
        fatal_error "$ERR_NO_FALLBACK" \
            "Requested version $VERSION is not published and no same-minor-series patch is available." \
            $EXIT_VERSION_MISSING \
            "{\"requested\":\"$(json_escape "$VERSION")\",\"fallbackAttempted\":true}"
    fi

    if ! is_interactive; then
        local recent_json
        recent_json="$(recent_releases_json)"
        fatal_error "$ERR_NON_INTERACTIVE" \
            "Requested version $VERSION is not published. Non-interactive session cannot prompt for substitution. Re-run with --allow-fallback to opt into same-minor patch substitution, or pin to one of the recent releases." \
            $EXIT_VERSION_MISSING \
            "{\"requested\":\"$(json_escape "$VERSION")\",\"interactive\":false,\"allowFallbackHint\":\"--allow-fallback\",\"recentReleases\":$recent_json}"
    fi

    interactive_pick
}

# is_interactive returns 0 only when we can safely prompt for input.
# We require: not --quiet, not --json-errors, not running under CI, AND a
# real /dev/tty.
is_interactive() {
    [ "$QUIET" -eq 1 ] && return 1
    [ "$JSON_ERRORS" -eq 1 ] && return 1
    [ "${CI:-}" = "true" ] || [ "${CI:-}" = "1" ] && return 1
    [ -r /dev/tty ] || return 1
    [ -w /dev/tty ] || return 1
    return 0
}

resolve_fallback_patch() {
    if [[ ! "$VERSION" =~ ^v([0-9]+)\.([0-9]+)\.[0-9]+ ]]; then return; fi
    local major="${BASH_REMATCH[1]}" minor="${BASH_REMATCH[2]}"
    github_api "/releases?per_page=100" \
        | grep -oE '"tag_name":[[:space:]]*"v[0-9]+\.[0-9]+\.[0-9]+"' \
        | sed -E 's/.*"(v[0-9.]+)".*/\1/' \
        | grep -E "^v$major\.$minor\." \
        | sort -t. -k3 -n -r \
        | head -n1
}

# recent_releases_json returns a JSON array of the 5 most recent release tags.
recent_releases_json() {
    local tags
    tags="$(github_api "/releases?per_page=5" 2>/dev/null \
        | grep -oE '"tag_name":[[:space:]]*"v[0-9]+\.[0-9]+\.[0-9]+"' \
        | sed -E 's/.*"(v[0-9.]+)".*/\1/' \
        | head -n5)"
    if [ -z "$tags" ]; then printf '[]'; return; fi
    local first=1 out="["
    while IFS= read -r tag; do
        [ -z "$tag" ] && continue
        if [ $first -eq 1 ]; then first=0; else out="$out,"; fi
        out="$out\"$tag\""
    done <<< "$tags"
    out="$out]"
    printf '%s' "$out"
}

interactive_pick() {
    local recent
    recent="$(github_api "/releases?per_page=5" \
        | grep -oE '"tag_name":[[:space:]]*"v[0-9]+\.[0-9]+\.[0-9]+"' \
        | sed -E 's/.*"(v[0-9.]+)".*/\1/' \
        | head -n5)"
    if [ -z "$recent" ]; then
        fatal_error "$ERR_RECENT_LIST_FAILED" \
            "Requested version $VERSION is not published and the recent-releases list could not be fetched." \
            $EXIT_VERSION_MISSING \
            "{\"requested\":\"$(json_escape "$VERSION")\"}"
    fi
    local i=1
    local -a choices=()
    echo "" >&2
    echo "  Requested: $VERSION (not found)" >&2
    echo "  Most recent published releases:" >&2
    while IFS= read -r tag; do
        choices+=("$tag")
        printf "    [%d] %s\n" "$i" "$tag" >&2
        i=$((i+1))
    done <<< "$recent"
    echo "    [N] Quit (default)" >&2
    printf "  Pick a number to install instead, or N to quit: " >&2

    local reply=""
    if ! read -r reply </dev/tty; then
        echo "" >&2
        fatal_error "$ERR_NON_INTERACTIVE" \
            "Could not read from /dev/tty; aborting." \
            $EXIT_VERSION_MISSING \
            "{\"requested\":\"$(json_escape "$VERSION")\"}"
    fi

    if [ -z "$reply" ] || [[ "$reply" =~ ^[Nn] ]]; then
        local recent_json
        recent_json="$(recent_releases_json)"
        fatal_error "$ERR_USER_DECLINED" \
            "User declined to substitute for missing version $VERSION." \
            $EXIT_VERSION_MISSING \
            "{\"requested\":\"$(json_escape "$VERSION")\",\"recentReleases\":$recent_json}"
    fi
    if ! [[ "$reply" =~ ^[0-9]+$ ]] || [ "$reply" -lt 1 ] || [ "$reply" -gt "${#choices[@]}" ]; then
        fatal_error "$ERR_INVALID_CHOICE" \
            "Invalid choice '$reply'; expected 1..${#choices[@]} or N." \
            $EXIT_VERSION_MISSING \
            "{\"reply\":\"$(json_escape "$reply")\",\"max\":${#choices[@]}}"
    fi
    local chosen="${choices[$((reply-1))]}"
    warn "User selected $chosen as substitute for $VERSION"
    echo "$chosen"
}

# ── Asset selection ────────────────────────────────────────────────
select_asset_url() {
    local resolved="$1" os="$2" arch="$3"
    local body expected loose
    body="$(github_api "/releases/tags/$resolved")"
    expected="${BINARY_NAME}-${resolved}-${os}-${arch}.tar.gz"

    # Try canonical .tar.gz first, then .zip.
    for cand in "$expected" "${BINARY_NAME}-${resolved}-${os}-${arch}.zip"; do
        local url
        url="$(echo "$body" | grep -oE "\"browser_download_url\":[[:space:]]*\"[^\"]*${cand}\"" \
              | head -n1 \
              | sed -E 's/.*"(https:[^"]+)"/\1/')"
        if [ -n "$url" ]; then
            echo "$url"
            return
        fi
    done

    # Loose match: anything ending with -<os>-<arch>.tar.gz|.zip
    loose="$(echo "$body" \
        | grep -oE "\"browser_download_url\":[[:space:]]*\"[^\"]+-${os}-${arch}\.(tar\.gz|zip)\"" \
        | head -n1 \
        | sed -E 's/.*"(https:[^"]+)"/\1/')"
    if [ -n "$loose" ]; then
        warn "Exact asset for $os/$arch missing; using closest match: $(basename "$loose")"
        echo "$loose"
        return
    fi

    err "No asset matching $os/$arch in release $resolved."
    err "Available assets:"
    echo "$body" | grep -oE '"name":[[:space:]]*"[^"]+"' | sed -E 's/.*"([^"]+)"/  - \1/' >&2
    exit $EXIT_UNSUPPORTED_ARCH
}

checksums_url() {
    local resolved="$1" body
    body="$(github_api "/releases/tags/$resolved")"
    echo "$body" | grep -oE '"browser_download_url":[[:space:]]*"[^"]+checksums\.txt"' \
        | head -n1 \
        | sed -E 's/.*"(https:[^"]+)"/\1/'
}

# ── Download + checksum ────────────────────────────────────────────
verify_checksum() {
    local archive="$1" name="$2" sums="$3"
    if [ ! -s "$sums" ]; then
        warn "No checksums.txt available; skipping verification."
        return
    fi
    local expected
    expected="$(grep -F "$name" "$sums" | awk '{print $1}' | head -n1)"
    if [ -z "$expected" ]; then
        warn "$name not listed in checksums.txt; skipping verification."
        return
    fi
    local actual=""
    if command -v sha256sum >/dev/null 2>&1; then
        actual="$(sha256sum "$archive" | awk '{print $1}')"
    elif command -v shasum >/dev/null 2>&1; then
        actual="$(shasum -a 256 "$archive" | awk '{print $1}')"
    else
        warn "No sha256sum/shasum tool; skipping verification."
        return
    fi
    if [ "$expected" != "$actual" ]; then
        err "Checksum mismatch for $name"
        err "  expected: $expected"
        err "  actual:   $actual"
        exit $EXIT_CHECKSUM
    fi
    ok "Checksum verified."
}

# ── Install + PATH + chain self-install ────────────────────────────
resolve_install_dir() {
    if [ -n "$INSTALL_DIR" ]; then echo "$INSTALL_DIR"; return; fi
    echo "${HOME}/.local/bin"
}

extract_archive() {
    local archive="$1" dest="$2"
    case "$archive" in
        *.tar.gz|*.tgz)
            tar -xzf "$archive" -C "$dest"
            ;;
        *.zip)
            if command -v unzip >/dev/null 2>&1; then
                unzip -qo "$archive" -d "$dest"
            else
                err "unzip required to extract $archive"
                exit $EXIT_VERIFY
            fi
            ;;
        *)
            err "Unknown archive format: $archive"
            exit $EXIT_VERIFY
            ;;
    esac
}

install_binary() {
    local archive="$1" install_dir="$2" os="$3" arch="$4"
    mkdir -p "$install_dir"
    local extract="$TMP_DIR/extract"
    mkdir -p "$extract"
    extract_archive "$archive" "$extract"

    local candidate
    candidate="$(find "$extract" -type f \( -name "$BINARY_NAME" \
        -o -name "${BINARY_NAME}-${os}-${arch}" \
        -o -regex ".*/${BINARY_NAME}-v[0-9][0-9.]*-${os}-${arch}" \) | head -n1)"

    if [ -z "$candidate" ]; then
        err "Archive did not contain a recognizable gitmap binary."
        exit $EXIT_VERIFY
    fi

    local dest="$install_dir/$BINARY_NAME"
    cp -f "$candidate" "$dest"
    chmod +x "$dest"
    ok "Installed: $dest"
    echo "$dest"
}

add_to_path() {
    local dir="$1"
    [ "$NO_PATH" -eq 1 ] && return
    case ":$PATH:" in
        *":$dir:"*) step "Already on PATH: $dir"; return ;;
    esac
    local profile=""
    case "${SHELL:-}" in
        */zsh)  profile="$HOME/.zshrc" ;;
        */bash) profile="$HOME/.bashrc" ;;
        *)      profile="$HOME/.profile" ;;
    esac
    {
        echo ""
        echo "# Added by gitmap release-version installer"
        echo "export PATH=\"\$PATH:$dir\""
    } >> "$profile" 2>/dev/null || {
        warn "Could not update $profile"
        return
    }
    ok "Appended PATH update to $profile (restart your shell to apply)"
}

verify_version() {
    local bin="$1" expected="$2"
    local reported
    if ! reported="$("$bin" --version 2>&1 | head -n1)"; then
        err "Could not run installed binary"
        exit $EXIT_VERIFY
    fi
    local stripped="${expected#v}"
    if [[ "$reported" != *"$stripped"* ]]; then
        err "Version mismatch: expected $expected, binary reported '$reported'"
        exit $EXIT_VERIFY
    fi
    ok "Verified: $reported"
}

chain_self_install() {
    local bin="$1"
    [ "$NO_SELF_INSTALL" -eq 1 ] && return
    step "Chaining gitmap self-install ..."
    if ! "$bin" self-install; then
        warn "self-install failed"
        exit $EXIT_SELF_INSTALL
    fi
}

# ── main ───────────────────────────────────────────────────────────
if [ -z "$VERSION" ]; then
    err "Required: --version vMAJOR.MINOR.PATCH"
    err "Example:  bash release-version.sh --version v3.36.0"
    exit $EXIT_VERSION_MISSING
fi

OS="$(detect_os)"
ARCH="$(detect_arch)"
step "Target: $OS/$ARCH"

RESOLVED="$(resolve_requested_version)"
step "Resolving release $RESOLVED ..."

ASSET_URL="$(select_asset_url "$RESOLVED" "$OS" "$ARCH")"
ASSET_NAME="$(basename "$ASSET_URL")"
SUMS_URL="$(checksums_url "$RESOLVED")"

TMP_DIR="$(mktemp -d -t gitmap-rv.XXXXXX)"
ARCHIVE="$TMP_DIR/$ASSET_NAME"
SUMS="$TMP_DIR/checksums.txt"

step "Downloading $ASSET_NAME ..."
http_get "$ASSET_URL" "$ARCHIVE" || { err "Download failed"; exit $EXIT_NETWORK; }
if [ -n "$SUMS_URL" ]; then
    http_get "$SUMS_URL" "$SUMS" 2>/dev/null || warn "Could not fetch checksums.txt"
fi
verify_checksum "$ARCHIVE" "$ASSET_NAME" "$SUMS"

INSTALL_DIR_RESOLVED="$(resolve_install_dir)"
BIN_PATH="$(install_binary "$ARCHIVE" "$INSTALL_DIR_RESOLVED" "$OS" "$ARCH")"
add_to_path "$INSTALL_DIR_RESOLVED"
verify_version "$BIN_PATH" "$RESOLVED"
chain_self_install "$BIN_PATH"

echo ""
ok "gitmap $RESOLVED installed to $INSTALL_DIR_RESOLVED"
exit $EXIT_OK
