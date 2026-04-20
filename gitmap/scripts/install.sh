#!/usr/bin/env bash
# Re-exec under bash if invoked via sh/dash (which lack pipefail, local, etc.)
if [ -z "${BASH_VERSION:-}" ]; then
    if command -v bash >/dev/null 2>&1; then
        case "${0##*/}" in
            sh|dash|ash|ksh|mksh)
                exec bash -s -- "$@"
                ;;
        esac

        exec bash "$0" "$@"
    else
        printf '\033[31m  Error: bash is required but not found. Install bash first.\033[0m\n' >&2
        exit 1
    fi
fi
# ─────────────────────────────────────────────────────────────────────
# gitmap installer for Linux and macOS
#
# Usage:
#   curl -fsSL https://raw.githubusercontent.com/alimtvnetwork/gitmap-v4/main/gitmap/scripts/install.sh | bash
#
# Options:
#   --version <tag>    Install a specific version (e.g. v2.55.0). Default: latest.
#   --dir <path>       Target directory. Default: ~/.local/bin
#   --arch <arch>      Force architecture (amd64, arm64). Default: auto-detect.
#   --no-path          Skip adding install directory to PATH.
#
# Examples:
#   curl -fsSL .../install.sh | bash
#   curl -fsSL .../install.sh | bash -s -- --version v2.55.0
#   ./install.sh --dir /opt/gitmap --arch arm64
# ─────────────────────────────────────────────────────────────────────

set -euo pipefail

REPO="alimtvnetwork/gitmap-v4"
BINARY_NAME="gitmap"
TMP_DIR=""
APP_DIR=""
PATH_SHELL=""
PATH_TARGET=""
PATH_LINE=""
PATH_STATUS=""
PATH_RELOAD=""

cleanup() {
    if [ -n "${TMP_DIR}" ] && [ -d "${TMP_DIR}" ]; then
        rm -rf "${TMP_DIR}"
    fi
}
trap cleanup EXIT

# ── Logging helpers ─────────────────────────────────────────────────

step()  { printf '  \033[36m%s\033[0m\n' "$*" >&2; }
ok()    { printf '  \033[32m%s\033[0m\n' "$*" >&2; }
err()   { printf '  \033[31m%s\033[0m\n' "$*" >&2; }

# ── Versioned repo discovery ────────────────────────────────────────
# spec/01-app/95-installer-script-find-latest-repo.md

# Parses "<owner>/<stem>-v<N>". Sets SUFFIX_OWNER, SUFFIX_STEM, SUFFIX_N.
parse_repo_suffix() {
    local repo="$1"
    if [[ "$repo" =~ ^([^/]+)/(.+)-v([0-9]+)$ ]]; then
        SUFFIX_OWNER="${BASH_REMATCH[1]}"
        SUFFIX_STEM="${BASH_REMATCH[2]}"
        SUFFIX_N="${BASH_REMATCH[3]}"
        return 0
    fi
    return 1
}

repo_exists() {
    curl -sfI --max-time 5 "$1" >/dev/null 2>&1
}

# Echoes the effective "<owner>/<stem>-v<M>" (or original repo when none higher).
resolve_effective_repo() {
    local repo="$1" ceiling="$2"
    if ! parse_repo_suffix "$repo"; then
        printf '  [discovery] no -v<N> suffix on '"'"'%s'"'"'; installing baseline as-is\n' "$repo" >&2
        echo "$repo"
        return 0
    fi

    local owner="$SUFFIX_OWNER" stem="$SUFFIX_STEM" baseline="$SUFFIX_N"
    local effective="$baseline" m url

    printf '  [discovery] baseline: %s/%s-v%s\n' "$owner" "$stem" "$baseline" >&2
    printf '  [discovery] probe ceiling: %s\n' "$ceiling" >&2

    for (( m = baseline + 1; m <= ceiling; m++ )); do
        url="https://github.com/${owner}/${stem}-v${m}"
        if repo_exists "$url"; then
            printf '  [discovery] HEAD %s ... HIT\n' "$url" >&2
            effective=$m
        else
            printf '  [discovery] HEAD %s ... MISS (fail-fast)\n' "$url" >&2
            break
        fi
    done

    if [ "$effective" = "$baseline" ]; then
        printf '  [discovery] no higher version found; using baseline -v%s\n' "$baseline" >&2
        echo "$repo"
    else
        printf '  [discovery] effective: %s/%s-v%s (was -v%s)\n' "$owner" "$stem" "$effective" "$baseline" >&2
        echo "${owner}/${stem}-v${effective}"
    fi
}

# Re-exec the full installer from the effective repo, passing through flags.
invoke_delegated_full_installer() {
    local effective_repo="$1"
    shift
    local delegated_url="https://raw.githubusercontent.com/${effective_repo}/main/gitmap/scripts/install.sh"
    printf '  [discovery] delegating to %s\n' "$delegated_url" >&2

    export INSTALLER_DELEGATED=1

    local script
    if ! script="$(curl -fsSL --max-time 15 "$delegated_url")"; then
        printf '  [discovery] [WARN] could not fetch delegated installer; falling back to baseline\n' >&2
        unset INSTALLER_DELEGATED
        return 1
    fi

    bash -c "$script" _ "$@"
    exit $?
}

# ── Detect OS ───────────────────────────────────────────────────────

detect_os() {
    local uname_out
    uname_out="$(uname -s)"
    case "${uname_out}" in
        Linux*)     echo "linux" ;;
        Darwin*)    echo "darwin" ;;
        MINGW*|MSYS*|CYGWIN*)
            err "Windows detected. Use the PowerShell installer instead:"
            err "  irm https://raw.githubusercontent.com/${REPO}/main/gitmap/scripts/install.ps1 | iex"
            exit 1
            ;;
        *)
            err "Unsupported OS: ${uname_out}"
            exit 1
            ;;
    esac
}

# ── Detect architecture ────────────────────────────────────────────

detect_arch() {
    local arch_flag="$1"
    if [ -n "${arch_flag}" ]; then
        echo "${arch_flag}"
        return
    fi

    local machine
    machine="$(uname -m)"
    case "${machine}" in
        x86_64|amd64)   echo "amd64" ;;
        aarch64|arm64)  echo "arm64" ;;
        *)
            err "Unsupported architecture: ${machine}"
            exit 1
            ;;
    esac
}

# ── Resolve version (latest or pinned) ─────────────────────────────

resolve_version() {
    local version="$1"
    if [ -n "${version}" ]; then
        echo "${version}"
        return
    fi

    step "Fetching latest release..."
    local url="https://api.github.com/repos/${REPO}/releases/latest"
    local tag

    if command -v curl >/dev/null 2>&1; then
        tag="$(curl -fsSL "${url}" | grep '"tag_name"' | head -1 | sed -E 's/.*"tag_name"[[:space:]]*:[[:space:]]*"([^"]+)".*/\1/')"
    elif command -v wget >/dev/null 2>&1; then
        tag="$(wget -qO- "${url}" | grep '"tag_name"' | head -1 | sed -E 's/.*"tag_name"[[:space:]]*:[[:space:]]*"([^"]+)".*/\1/')"
    else
        err "Neither curl nor wget found. Cannot fetch latest release."
        exit 1
    fi

    if [ -z "${tag}" ]; then
        err "Failed to determine latest version."
        exit 1
    fi

    echo "${tag}"
}

# ── Download helper ────────────────────────────────────────────────

download() {
    local url="$1" dest="$2"
    if command -v curl >/dev/null 2>&1; then
        curl -fsSL -o "${dest}" "${url}"
    elif command -v wget >/dev/null 2>&1; then
        wget -qO "${dest}" "${url}"
    else
        err "Neither curl nor wget found."
        exit 1
    fi
}

# ── Download and verify asset ──────────────────────────────────────

download_asset() {
    local version="$1" os="$2" arch="$3"
    local asset_name="${BINARY_NAME}-${version}-${os}-${arch}.tar.gz"
    local base_url="https://github.com/${REPO}/releases/download/${version}"
    local asset_url="${base_url}/${asset_name}"
    local checksum_url="${base_url}/checksums.txt"

    # TMP_DIR is set by the caller (main).

    local archive_path="${TMP_DIR}/${asset_name}"
    local checksum_path="${TMP_DIR}/checksums.txt"

    step "Downloading ${asset_name} (${version})..."
    download "${asset_url}" "${archive_path}"
    download "${checksum_url}" "${checksum_path}"

    # Verify checksum
    step "Verifying checksum..."
    local expected_line
    expected_line="$(grep "${asset_name}" "${checksum_path}" || true)"
    if [ -z "${expected_line}" ]; then
        # Try .zip variant (some releases may only have zip)
        asset_name="${BINARY_NAME}-${version}-${os}-${arch}.zip"
        asset_url="${base_url}/${asset_name}"
        archive_path="${TMP_DIR}/${asset_name}"

        step "Trying .zip variant..."
        download "${asset_url}" "${archive_path}"
        expected_line="$(grep "${asset_name}" "${checksum_path}" || true)"

        if [ -z "${expected_line}" ]; then
            err "Asset not found in checksums.txt"
            err "Tried: ${BINARY_NAME}-${version}-${os}-${arch}.tar.gz"
            err "Tried: ${asset_name}"
            exit 1
        fi
    fi

    local expected_hash
    expected_hash="$(echo "${expected_line}" | awk '{print $1}')"

    local actual_hash
    if command -v sha256sum >/dev/null 2>&1; then
        actual_hash="$(sha256sum "${archive_path}" | awk '{print $1}')"
    elif command -v shasum >/dev/null 2>&1; then
        actual_hash="$(shasum -a 256 "${archive_path}" | awk '{print $1}')"
    else
        err "No SHA256 tool found (sha256sum or shasum required)."
        exit 1
    fi

    if [ "${actual_hash}" != "${expected_hash}" ]; then
        err "Checksum mismatch!"
        err "  Expected: ${expected_hash}"
        err "  Got:      ${actual_hash}"
        exit 1
    fi

    ok "Checksum verified."
    echo "${archive_path}"
}

# ── Layout repair + pre-deploy cleanup (DFD-3, DFD-6) ──────────────
# Migrates legacy unwrapped install (<dir>/gitmap) into nested
# <dir>/gitmap/gitmap layout and removes prior-deploy artifacts.

repair_layout() {
    local target="$1"
    local app_dir="$target/${BINARY_NAME}"
    local legacy_binary="$target/${BINARY_NAME}"
    local wrapped_binary="$app_dir/${BINARY_NAME}"

    # Special case: when target ends with /<binary> the legacy and wrapped
    # paths collide. Skip — caller resolved into a parent dir.
    if [ -f "$legacy_binary" ] && [ ! -d "$app_dir" ]; then
        step "Layout: migrating legacy unwrapped install -> ${app_dir}"
        mkdir -p "$app_dir"
        local name src dst
        for name in "${BINARY_NAME}" data CHANGELOG.md docs docs-site; do
            src="$target/$name"
            dst="$app_dir/$name"
            [ ! -e "$src" ] && continue
            [ -e "$dst" ] && continue
            mv "$src" "$dst" 2>/dev/null && \
                step "  moved $name -> ${BINARY_NAME}/$name"
        done
    elif [ -f "$legacy_binary" ] && [ -f "$wrapped_binary" ]; then
        rm -f "$legacy_binary" 2>/dev/null && \
            step "Layout: removed leftover legacy binary $legacy_binary"
    else
        step "Layout: OK"
    fi
}

cleanup_prior_artifacts() {
    local target="$1" app_dir="$2"
    local stem="${BINARY_NAME}"
    local removed=0
    local dir pat f

    for dir in "$target" "$app_dir"; do
        [ ! -d "$dir" ] && continue
        for pat in "*.old" "${stem}-update-*" "updater-tmp-*"; do
            for f in "$dir"/$pat; do
                [ ! -e "$f" ] && continue
                rm -rf "$f" 2>/dev/null && {
                    step "[cleanup] removed $f"
                    removed=$((removed + 1))
                }
            done
        done
    done

    local tmp_root="${TMPDIR:-/tmp}"
    if [ -d "$tmp_root" ]; then
        for f in "$tmp_root/${stem}-update-"*; do
            [ ! -e "$f" ] && continue
            rm -rf "$f" 2>/dev/null && {
                step "[cleanup] removed temp $f"
                removed=$((removed + 1))
            }
        done
    fi

    if [ -d "$target" ]; then
        for f in "$target"/*.gitmap-tmp-*; do
            [ ! -d "$f" ] && continue
            rm -rf "$f" 2>/dev/null && {
                step "[cleanup] removed swap dir $f"
                removed=$((removed + 1))
            }
        done
    fi

    if [ "$removed" -gt 0 ]; then
        ok "[cleanup] removed $removed artifact(s)"
    else
        step "[cleanup] nothing to clean"
    fi
}

# ── Extract and install binary ─────────────────────────────────────

install_binary() {
    local archive_path="$1" install_dir="$2" os="$3" arch="$4" version="$5"

    # DFD-1/DFD-3: nested layout. install_dir is the deploy ROOT (e.g.
    # ~/.local/bin); the actual app folder is ${install_dir}/${BINARY_NAME}.
    repair_layout "${install_dir}"
    local app_dir="${install_dir}/${BINARY_NAME}"
    cleanup_prior_artifacts "${install_dir}" "${app_dir}"

    step "Installing to ${app_dir}..."
    mkdir -p "${app_dir}"

    local extract_dir="${TMP_DIR}/extract"
    mkdir -p "${extract_dir}"

    case "${archive_path}" in
        *.tar.gz|*.tgz)
            tar -xzf "${archive_path}" -C "${extract_dir}"
            ;;
        *.zip)
            if command -v unzip >/dev/null 2>&1; then
                unzip -qo "${archive_path}" -d "${extract_dir}"
            else
                err "unzip not found. Cannot extract .zip archive."
                exit 1
            fi
            ;;
        *)
            err "Unknown archive format: ${archive_path}"
            exit 1
            ;;
    esac

    local binary_path=""
    local candidate

    candidate="$(find "${extract_dir}" -type f -name "${BINARY_NAME}" | head -1)"
    [ -n "${candidate}" ] && binary_path="${candidate}"

    if [ -z "${binary_path}" ]; then
        candidate="$(find "${extract_dir}" -type f -name "${BINARY_NAME}-${os}-${arch}" | head -1)"
        [ -n "${candidate}" ] && binary_path="${candidate}"
    fi

    if [ -z "${binary_path}" ]; then
        candidate="$(find "${extract_dir}" -type f -regex ".*/${BINARY_NAME}-v[0-9][0-9.]*-${os}-${arch}" | head -1)"
        [ -n "${candidate}" ] && binary_path="${candidate}"
    fi

    if [ -z "${binary_path}" ]; then
        candidate="$(find "${extract_dir}" -type f -executable | head -1)"
        [ -n "${candidate}" ] && binary_path="${candidate}"
    fi

    if [ -z "${binary_path}" ]; then
        err "Archive did not contain a recognizable binary."
        find "${extract_dir}" -type f | while read -r f; do err "  ${f}"; done
        exit 1
    fi

    local target_path="${app_dir}/${BINARY_NAME}"

    if [ -f "${target_path}" ]; then
        mv -f "${target_path}" "${target_path}.old" 2>/dev/null || true
    fi

    mv -f "${binary_path}" "${target_path}"
    chmod +x "${target_path}"

    rm -f "${target_path}.old" 2>/dev/null || true

    if [ ! -f "${target_path}" ]; then
        err "Install failed: ${BINARY_NAME} was not written to ${app_dir}"
        exit 1
    fi

    ok "Installed ${BINARY_NAME} to ${app_dir}"

    # Echo the app dir so main() can use it for PATH + summary.
    APP_DIR="${app_dir}"
}

# ── Download and extract docs-site.zip release asset ───────────────
# Required for `gitmap help-dashboard` (hd). Best-effort: skip silently
# if the release does not bundle docs-site.zip (older versions).
install_docs_site() {
    local version="$1" install_dir="$2"
    local asset_name="docs-site.zip"
    local asset_url="https://github.com/${REPO}/releases/download/${version}/${asset_name}"
    local tmp_zip="${TMP_DIR}/${asset_name}"

    step "Downloading docs-site.zip (${version})..."

    if ! download "${asset_url}" "${tmp_zip}" 2>/dev/null; then
        step "  docs-site.zip not available for ${version} - skipping (gitmap hd may not work)"
        rm -f "${tmp_zip}" 2>/dev/null || true
        return 0
    fi

    # Remove any existing docs-site/ before extracting fresh.
    rm -rf "${install_dir}/docs-site" 2>/dev/null || true

    if ! command -v unzip >/dev/null 2>&1; then
        err "unzip not found - cannot extract docs-site.zip (install unzip and re-run)"
        rm -f "${tmp_zip}" 2>/dev/null || true
        return 0
    fi

    # The zip's internal layout is docs-site/dist/... so it extracts directly.
    if unzip -qo "${tmp_zip}" -d "${install_dir}"; then
        ok "Installed docs-site to ${install_dir}/docs-site"
    else
        err "Failed to extract docs-site.zip"
    fi

    rm -f "${tmp_zip}" 2>/dev/null || true
}

# ── Add to PATH ────────────────────────────────────────────────────

# add_path_to_profile writes an export line to a single profile file (idempotent).
# Returns 0 if written, 1 if already present.
# add_path_to_profile writes a marker-block snippet (per
# spec/04-generic-cli/21-post-install-shell-activation) to a single
# profile file. Idempotent: rewrites the existing block if present.
# Returns 0 if written, 1 if no-op.
add_path_to_profile() {
    local dir="$1" profile_file="$2" is_fish="$3"

    local marker_open="# gitmap shell wrapper v2 - managed by gitmap installer. Do not edit manually."
    local marker_close="# gitmap shell wrapper v2 end"

    # Single-source-of-truth: ask the freshly-installed gitmap binary
    # for the canonical snippet bytes. Falls back to an inline heredoc
    # if the binary isn't on PATH yet (called before install_binary
    # completed).
    local snippet=""
    local snippet_shell="bash"
    [ "${is_fish}" = true ] && snippet_shell="fish"
    local gitmap_bin=""
    if [ -x "${INSTALL_DIR:-}/gitmap" ]; then
        gitmap_bin="${INSTALL_DIR}/gitmap"
    elif command -v gitmap >/dev/null 2>&1; then
        gitmap_bin="$(command -v gitmap)"
    fi
    if [ -n "${gitmap_bin}" ]; then
        snippet="$("${gitmap_bin}" setup print-path-snippet \
            --shell "${snippet_shell}" --dir "${dir}" --manager "installer" 2>/dev/null || true)"
    fi
    if [ -z "${snippet}" ]; then
        if [ "${is_fish}" = true ]; then
            snippet="${marker_open}
set -gx GITMAP_WRAPPER 1
fish_add_path ${dir}
${marker_close}"
        else
            snippet="${marker_open}
export GITMAP_WRAPPER=1
case \":\${PATH}:\" in *\":${dir}:\"*) ;; *) export PATH=\"\$PATH:${dir}\" ;; esac
${marker_close}"
        fi
    fi

    mkdir -p "$(dirname "${profile_file}")"
    touch "${profile_file}"

    if grep -qF "${marker_open}" "${profile_file}" 2>/dev/null; then
        local tmp
        tmp="$(mktemp)"
        awk -v open="${marker_open}" -v close="${marker_close}" -v body="${snippet}" '
            $0 == open { skip = 1; print body; next }
            skip && $0 == close { skip = 0; next }
            !skip { print }
        ' "${profile_file}" > "${tmp}" && mv "${tmp}" "${profile_file}"
        return 1
    fi

    printf '\n%s\n' "${snippet}" >> "${profile_file}"
    return 0
}

add_to_path() {
    local dir="$1"
    local has_session_path=false

    case ":${PATH}:" in
        *":${dir}:"*)
            has_session_path=true
            ;;
    esac

    # Detect primary shell
    local shell_name
    shell_name="$(basename "${SHELL:-/bin/bash}")"
    PATH_SHELL="${shell_name}"

    local primary_profile=""
    local profiles_written=""
    local profiles_skipped=""

    # ── Write to all relevant POSIX/bash/zsh profiles ──────────────
    # This ensures gitmap is available regardless of which shell the user opens.

    # zsh profiles (both, to cover login + interactive shells)
    if [ "${shell_name}" = "zsh" ] || [ -f "${HOME}/.zshrc" ] || [ -f "${HOME}/.zprofile" ]; then
        # .zshrc — interactive shells (most terminal emulators)
        if add_path_to_profile "${dir}" "${HOME}/.zshrc" false; then
            profiles_written="${profiles_written} ~/.zshrc"
        else
            profiles_skipped="${profiles_skipped} ~/.zshrc"
        fi
        # .zprofile — login shells (macOS Terminal.app)
        if add_path_to_profile "${dir}" "${HOME}/.zprofile" false; then
            profiles_written="${profiles_written} ~/.zprofile"
        else
            profiles_skipped="${profiles_skipped} ~/.zprofile"
        fi
    fi

    # bash profiles
    if [ "${shell_name}" = "bash" ] || [ -f "${HOME}/.bashrc" ] || [ -f "${HOME}/.bash_profile" ]; then
        if add_path_to_profile "${dir}" "${HOME}/.bashrc" false; then
            profiles_written="${profiles_written} ~/.bashrc"
        else
            profiles_skipped="${profiles_skipped} ~/.bashrc"
        fi
        if [ -f "${HOME}/.bash_profile" ]; then
            if add_path_to_profile "${dir}" "${HOME}/.bash_profile" false; then
                profiles_written="${profiles_written} ~/.bash_profile"
            else
                profiles_skipped="${profiles_skipped} ~/.bash_profile"
            fi
        fi
    fi

    # POSIX ~/.profile — catch-all for sh and other POSIX shells
    if add_path_to_profile "${dir}" "${HOME}/.profile" false; then
        profiles_written="${profiles_written} ~/.profile"
    else
        profiles_skipped="${profiles_skipped} ~/.profile"
    fi

    # fish (only if fish is installed or is the default shell)
    if [ "${shell_name}" = "fish" ] || command -v fish >/dev/null 2>&1; then
        local fish_config="${HOME}/.config/fish/config.fish"
        if add_path_to_profile "${dir}" "${fish_config}" true; then
            profiles_written="${profiles_written} ~/.config/fish/config.fish"
        else
            profiles_skipped="${profiles_skipped} ~/.config/fish/config.fish"
        fi
    fi

    # Determine primary profile for reload instruction
    case "${shell_name}" in
        zsh)    primary_profile="${HOME}/.zshrc" ;;
        bash)   primary_profile="${HOME}/.bashrc" ;;
        fish)   primary_profile="${HOME}/.config/fish/config.fish" ;;
        *)      primary_profile="${HOME}/.profile" ;;
    esac

    PATH_TARGET="${primary_profile}"

    if [ "${shell_name}" = "fish" ]; then
        PATH_LINE="fish_add_path ${dir}"
        PATH_RELOAD="source ${primary_profile}"
    else
        PATH_LINE="export PATH=\"\$PATH:${dir}\""
        PATH_RELOAD=". ${primary_profile}"
    fi

    # Report what was written
    if [ -n "${profiles_written}" ]; then
        ok "Added to PATH in:${profiles_written}"
        PATH_STATUS="added"
    else
        step "PATH already configured in all profiles"
        PATH_STATUS="already present"
    fi

    if [ -n "${profiles_skipped}" ]; then
        step "Already present in:${profiles_skipped}"
    fi

    if [ "${has_session_path}" = true ]; then
        return
    fi

    # Update current session (only effective when script is sourced, not piped)
    export PATH="${PATH}:${dir}"
}

print_install_summary() {
    local installed_version="$1" bin_path="$2"

    echo ""
    step "Install summary"
    printf '    Version: %s\n' "${installed_version}" >&2
    printf '    Binary: %s\n' "${bin_path}" >&2
    printf '    Install dir: %s\n' "$(dirname "${bin_path}")" >&2
    if [ "${NO_PATH}" = true ]; then
        printf '    PATH target: skipped (--no-path)\n' >&2

        return
    fi
    printf '    Shell: %s\n' "${PATH_SHELL}" >&2
    printf '    PATH target: %s (%s)\n' "${PATH_TARGET}" "${PATH_STATUS}" >&2
    printf '    Reload: %s\n' "${PATH_RELOAD}" >&2
}

# ── Resolve install directory ──────────────────────────────────────

resolve_install_dir() {
    local dir="$1"
    if [ -n "${dir}" ]; then
        echo "${dir}"
        return
    fi

    # Use ~/.local/bin if it exists or is standard; fallback to /usr/local/bin
    if [ -d "${HOME}/.local/bin" ] || [ -w "${HOME}/.local" ]; then
        echo "${HOME}/.local/bin"
    elif [ -w "/usr/local/bin" ]; then
        echo "/usr/local/bin"
    else
        echo "${HOME}/.local/bin"
    fi
}

# ── Parse arguments ────────────────────────────────────────────────

parse_args() {
    VERSION=""
    INSTALL_DIR=""
    ARCH_FLAG=""
    NO_PATH=false
    NO_DISCOVERY=false
    PROBE_CEILING=30

    while [ $# -gt 0 ]; do
        case "$1" in
            --version)
                VERSION="$2"
                shift 2
                ;;
            --dir)
                INSTALL_DIR="$2"
                shift 2
                ;;
            --arch)
                ARCH_FLAG="$2"
                shift 2
                ;;
            --no-path)
                NO_PATH=true
                shift
                ;;
            --no-discovery)
                NO_DISCOVERY=true
                shift
                ;;
            --probe-ceiling)
                PROBE_CEILING="$2"
                shift 2
                ;;
            --help|-h)
                echo "Usage: install.sh [--version <tag>] [--dir <path>] [--arch <arch>] [--no-path] [--no-discovery] [--probe-ceiling <N>]"
                echo ""
                echo "Options:"
                echo "  --version <tag>        Install a specific version (e.g. v2.55.0)"
                echo "  --dir <path>           Target directory (default: ~/.local/bin)"
                echo "  --arch <arch>          Force architecture: amd64, arm64 (default: auto)"
                echo "  --no-path              Skip adding install directory to PATH"
                echo "  --no-discovery         Skip versioned-repo discovery (install baseline)"
                echo "  --probe-ceiling <N>    Highest -v<N> to probe (default: 30)"
                exit 0
                ;;
            *)
                err "Unknown option: $1"
                err "Run with --help for usage."
                exit 1
                ;;
        esac
    done
}

# ── Main ───────────────────────────────────────────────────────────

main() {
    echo ""
    echo "  gitmap installer"
    printf '  \033[90mgithub.com/%s\033[0m\n' "${REPO}"
    echo ""

    parse_args "$@"

    # Versioned repo discovery: re-exec from the latest -v<M> sibling repo.
    if [ "${INSTALLER_DELEGATED:-0}" = "1" ]; then
        printf '  [discovery] INSTALLER_DELEGATED=1; skipping discovery (loop guard)\n' >&2
    elif [ "${NO_DISCOVERY}" = "true" ]; then
        printf '  [discovery] --no-discovery set; skipping probe\n' >&2
    elif [ -n "${VERSION}" ]; then
        # Pinned-version contract (spec/07-generic-release/08-pinned-version-install-snippet.md):
        # When --version is supplied, install EXACTLY that version from the embedded REPO.
        # Skip versioned-repo discovery so a snippet copied from a v3.x release page
        # never silently jumps to the v4 repo's latest tag.
        printf '  [discovery] --version %s pinned; skipping repo probe (exact-version install)\n' "${VERSION}" >&2
    else
        local effective_repo
        effective_repo="$(resolve_effective_repo "${REPO}" "${PROBE_CEILING}")"
        if [ "${effective_repo}" != "${REPO}" ]; then
            invoke_delegated_full_installer "${effective_repo}" "$@" || true
        fi
    fi

    local os arch version install_dir archive_path

    os="$(detect_os)"
    arch="$(detect_arch "${ARCH_FLAG}")"
    version="$(resolve_version "${VERSION}")"
    install_dir="$(resolve_install_dir "${INSTALL_DIR}")"

    # Create TMP_DIR in parent scope so install_binary and cleanup can access it.
    TMP_DIR="$(mktemp -d)"
    archive_path="$(download_asset "${version}" "${os}" "${arch}")"

    APP_DIR=""
    install_binary "${archive_path}" "${install_dir}" "${os}" "${arch}" "${version}"

    # Bundle the docs site so `gitmap help-dashboard` works after install.
    install_docs_site "${version}" "${APP_DIR}"

    if [ "${NO_PATH}" = false ]; then
        add_to_path "${APP_DIR}"
    fi

    # Verify the binary works
    local bin_path="${APP_DIR}/${BINARY_NAME}"
    local installed_version="${version}"
    if [ -f "${bin_path}" ]; then
        echo ""
        local version_output
        if version_output="$("${bin_path}" version 2>&1)"; then
            installed_version="${version_output}"
            ok "gitmap ${version_output}"
        else
            err "Binary found but failed to run."
        fi
    else
        err "Binary not found at ${bin_path}"
    fi

    print_install_summary "${installed_version}" "${bin_path}"
    if [ "${NO_PATH}" = false ]; then
        echo ""
        printf '  \033[32mOK\033[0m  To start using gitmap \033[1mright now\033[0m, run:\n' >&2
        echo "" >&2
        printf '      \033[36m%s\033[0m\n' "${PATH_RELOAD}" >&2
        echo "" >&2
        printf '     Or open a new terminal window.\n' >&2
        echo "" >&2
        printf '  \033[90mInstalled to: %s\033[0m\n' "${bin_path}" >&2
        printf '  \033[90mApp folder on PATH: %s\033[0m\n' "${APP_DIR}" >&2
    fi

    echo ""
    ok "Done! Run 'gitmap --help' to get started."
    echo ""
}

main "$@"
