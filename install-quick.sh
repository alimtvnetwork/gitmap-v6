#!/usr/bin/env bash
# Short interactive installer for gitmap on Linux / macOS.
#
# Prompts for an install folder (with a sensible default), then delegates
# to the canonical gitmap/scripts/install.sh with that path.
#
# Versioned repo discovery: if the source repo URL ends with -v<N>, this
# script probes for higher-numbered sibling repos (-v<N+1>, -v<N+2>, ...)
# and delegates to the latest available one. See:
#   spec/01-app/95-installer-script-find-latest-repo.md
#
# Usage:
#   curl -fsSL https://raw.githubusercontent.com/alimtvnetwork/gitmap-v4/main/install-quick.sh | bash
#   ./install-quick.sh
#   ./install-quick.sh --dir /opt/gitmap
#   ./install-quick.sh --no-discovery
#   ./install-quick.sh --probe-ceiling 50

set -euo pipefail

REPO="alimtvnetwork/gitmap-v4"
INSTALLER_URL="https://raw.githubusercontent.com/${REPO}/main/gitmap/scripts/install.sh"
DEFAULT_DIR="${HOME}/.local/bin"

INSTALL_DIR=""
VERSION=""
NO_DISCOVERY=0
PROBE_CEILING=30

while [ $# -gt 0 ]; do
    case "$1" in
        --dir)            INSTALL_DIR="$2"; shift 2 ;;
        --version)        VERSION="$2";     shift 2 ;;
        --no-discovery)   NO_DISCOVERY=1;   shift ;;
        --probe-ceiling)  PROBE_CEILING="$2"; shift 2 ;;
        -h|--help)
            sed -n '2,18p' "$0"
            exit 0
            ;;
        *)
            printf '  Unknown argument: %s\n' "$1" >&2
            exit 1
            ;;
    esac
done

# ---------------------------------------------------------------------------
# Versioned repo discovery (spec/01-app/95-installer-script-find-latest-repo.md)
# ---------------------------------------------------------------------------

# Parses "<owner>/<stem>-v<N>". Sets globals: SUFFIX_OWNER, SUFFIX_STEM, SUFFIX_N.
# Returns 0 on match, 1 otherwise.
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
    local url="$1"
    curl -sfI --max-time 5 "$url" >/dev/null 2>&1
}

# Resolves the highest existing -v<M> repo. Echoes "<owner>/<stem>-v<M>"
# (or the original repo if no suffix or no higher version).
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

invoke_delegated_installer() {
    local effective_repo="$1"
    local delegated_url="https://raw.githubusercontent.com/${effective_repo}/main/install-quick.sh"

    printf '  [discovery] delegating to %s\n' "$delegated_url" >&2

    local pass_args=()
    [ -n "$INSTALL_DIR" ]   && pass_args+=(--dir "$INSTALL_DIR")
    [ -n "$VERSION" ]       && pass_args+=(--version "$VERSION")
    pass_args+=(--probe-ceiling "$PROBE_CEILING")

    # Loop guard for the delegated script.
    export INSTALLER_DELEGATED=1

    local script
    if ! script="$(curl -fsSL --max-time 15 "$delegated_url")"; then
        printf '  [discovery] [WARN] could not fetch delegated installer; falling back to baseline\n' >&2
        unset INSTALLER_DELEGATED
        return 1
    fi

    bash -c "$script" _ "${pass_args[@]}"
    exit $?
}

if [ "${INSTALLER_DELEGATED:-0}" = "1" ]; then
    printf '  [discovery] INSTALLER_DELEGATED=1; skipping discovery (loop guard)\n' >&2
elif [ "$NO_DISCOVERY" = "1" ]; then
    printf '  [discovery] --no-discovery set; skipping probe\n' >&2
else
    EFFECTIVE_REPO="$(resolve_effective_repo "$REPO" "$PROBE_CEILING")"
    if [ "$EFFECTIVE_REPO" != "$REPO" ]; then
        invoke_delegated_installer "$EFFECTIVE_REPO" || true
        # If delegation failed we fall through and install baseline.
    fi
fi

# ---------------------------------------------------------------------------
# Baseline install flow (unchanged behaviour).
# ---------------------------------------------------------------------------

prompt_dir() {
    printf '\n'
    printf '  \033[36mgitmap quick installer\033[0m\n'
    printf '  \033[90m---------------------\033[0m\n'
    printf '  Choose install folder. Press Enter to accept the default.\n'
    printf '  \033[90mDefault: %s\033[0m\n' "${DEFAULT_DIR}"
    printf '  Install path: '

    # Read from the controlling terminal so it works under `curl | bash`.
    if [ -r /dev/tty ]; then
        IFS= read -r answer < /dev/tty || answer=""
    else
        IFS= read -r answer || answer=""
    fi

    if [ -z "${answer}" ]; then
        echo "${DEFAULT_DIR}"
    else
        echo "${answer}"
    fi
}

if [ -z "${INSTALL_DIR}" ]; then
    INSTALL_DIR="$(prompt_dir)"
fi

printf '\n  \033[32mInstalling gitmap to: %s\033[0m\n\n' "${INSTALL_DIR}"

# Persist the chosen install path so `gitmap install scripts` and
# `run.sh` pick the same folder automatically.
save_deploy_path() {
    local dir="$1"
    mkdir -p "${dir}" 2>/dev/null || true
    local cfg="${dir}/powershell.json"
    cat > "${cfg}" <<EOF
{
  "deployPath": "${dir}",
  "buildOutput": "./bin",
  "binaryName": "gitmap",
  "goSource": "./gitmap",
  "copyData": true
}
EOF
    printf '  \033[90mSaved deployPath -> %s\033[0m\n' "${cfg}"
}

save_deploy_path "${INSTALL_DIR}" || printf '  \033[33m[WARN] Could not save powershell.json\033[0m\n'

ARGS=(--dir "${INSTALL_DIR}")
if [ -n "${VERSION}" ]; then
    ARGS+=(--version "${VERSION}")
fi

curl -fsSL "${INSTALLER_URL}" | bash -s -- "${ARGS[@]}"
