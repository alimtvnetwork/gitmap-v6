<#
.SYNOPSIS
    One-liner uninstaller for gitmap on Windows.

.DESCRIPTION
    Removes the gitmap binary, deploy folder, user PATH entry, and (optionally)
    the per-user data folder. Works whether gitmap was installed via:

      - install-quick.ps1 (one-liner)
      - gitmap/scripts/install.ps1 (canonical installer)
      - manual `run.ps1` build-and-deploy

    Strategy:
      1. If `gitmap` is on PATH and reports its install dir, prefer that.
         Delegate to `gitmap self-uninstall -y` (best path — uses the binary's
         own knowledge of marker-block PATH cleanup, scheduled-task removal,
         etc.).
      2. If `gitmap` is NOT on PATH (already partially removed, broken install),
         fall back to a manual sweep:
           - delete <root>/gitmap-cli/ AND legacy <root>/gitmap/
           - strip the deploy root from User PATH
           - prompt before deleting %APPDATA%/gitmap

    Run via one-liner:
      irm https://raw.githubusercontent.com/alimtvnetwork/gitmap-v4/main/uninstall-quick.ps1 | iex

    Or locally:
      ./uninstall-quick.ps1
      ./uninstall-quick.ps1 -KeepData
      ./uninstall-quick.ps1 -InstallDir "E:\bin-run"
      ./uninstall-quick.ps1 -Yes
#>

param(
    [string]$InstallDir = "",
    [switch]$KeepData,
    [switch]$Yes
)

$ErrorActionPreference = "Continue"
$ProgressPreference    = "SilentlyContinue"

function Write-Step($msg) { Write-Host "  $msg" -ForegroundColor Cyan }
function Write-Info($msg) { Write-Host "    $msg" -ForegroundColor DarkGray }
function Write-Ok($msg)   { Write-Host "    $msg" -ForegroundColor Green }
function Write-Warn($msg) { Write-Host "    $msg" -ForegroundColor Yellow }
function Write-Err($msg)  { Write-Host "    $msg" -ForegroundColor Red }

function Confirm-Or-Exit([string]$prompt) {
    if ($Yes) { return $true }
    Write-Host ""
    Write-Host "  $prompt [y/N]: " -ForegroundColor Yellow -NoNewline
    $answer = Read-Host
    return ($answer -match '^(y|yes)$')
}

# ---------------------------------------------------------------------------
# Step 1 — try the canonical `gitmap self-uninstall` (best path).
# ---------------------------------------------------------------------------

function Try-SelfUninstall {
    $cmd = Get-Command gitmap -ErrorAction SilentlyContinue
    if (-not $cmd) {
        Write-Info "gitmap not found on PATH, skipping self-uninstall (will sweep manually)"
        return $false
    }

    Write-Info "Active binary: $($cmd.Source)"
    Write-Info "Delegating to: gitmap self-uninstall -y"
    Write-Host ""
    try {
        & gitmap self-uninstall -y
        if ($LASTEXITCODE -eq 0) {
            Write-Ok "self-uninstall completed cleanly"
            return $true
        }
        Write-Warn "self-uninstall exited with code $LASTEXITCODE; falling back to manual sweep"
        return $false
    } catch {
        Write-Warn "self-uninstall threw: $_; falling back to manual sweep"
        return $false
    }
}

# ---------------------------------------------------------------------------
# Step 2 — manual sweep fallback.
# ---------------------------------------------------------------------------

function Resolve-DeployRoot {
    if ($InstallDir.Length -gt 0) { return $InstallDir }

    # Check the active binary's grandparent (deployRoot/gitmap-cli/gitmap.exe).
    $cmd = Get-Command gitmap -ErrorAction SilentlyContinue
    if ($cmd -and (Test-Path $cmd.Source)) {
        $parent = Split-Path (Resolve-Path $cmd.Source).Path -Parent
        $grand  = Split-Path $parent -Parent
        return $grand
    }

    # Common defaults to probe.
    foreach ($candidate in @("E:\bin-run", "D:\gitmap", "$env:LOCALAPPDATA\gitmap", "$env:USERPROFILE\gitmap")) {
        if (Test-Path (Join-Path $candidate "gitmap-cli\gitmap.exe")) { return $candidate }
        if (Test-Path (Join-Path $candidate "gitmap\gitmap.exe"))     { return $candidate }
        if (Test-Path (Join-Path $candidate "gitmap.exe"))            { return $candidate }
    }
    return ""
}

function Remove-DeployFolders([string]$root) {
    if (-not $root) {
        Write-Warn "could not locate a gitmap deploy root; skipping folder removal"
        return
    }

    foreach ($sub in @("gitmap-cli", "gitmap")) {
        $dir = Join-Path $root $sub
        if (Test-Path $dir) {
            try {
                Remove-Item $dir -Recurse -Force -ErrorAction Stop
                Write-Ok "removed $dir"
            } catch {
                Write-Err "could not remove $dir : $_"
                Write-Warn "close any terminal/process using gitmap.exe and retry"
            }
        }
    }

    # Also remove a flat deploy (no subfolder).
    $flatBin = Join-Path $root "gitmap.exe"
    if (Test-Path $flatBin) {
        try {
            Remove-Item $flatBin -Force -ErrorAction Stop
            Write-Ok "removed $flatBin"
        } catch {
            Write-Err "could not remove $flatBin : $_"
        }
    }
}

function Remove-FromUserPath([string]$root) {
    if (-not $root) { return }

    $userPath = [Environment]::GetEnvironmentVariable("Path", "User")
    if (-not $userPath) { return }

    $entries = $userPath -split ';' | Where-Object { $_ -ne "" }
    $toRemove = @(
        $root,
        (Join-Path $root "gitmap-cli"),
        (Join-Path $root "gitmap")
    )

    $kept = $entries | Where-Object {
        $entry = $_.TrimEnd('\')
        $match = $false
        foreach ($r in $toRemove) {
            if ($entry -ieq $r.TrimEnd('\')) { $match = $true; break }
        }
        -not $match
    }

    $newPath = ($kept -join ';')
    if ($newPath -ne $userPath) {
        [Environment]::SetEnvironmentVariable("Path", $newPath, "User")
        Write-Ok "stripped gitmap entries from User PATH (restart shells to take effect)"
    } else {
        Write-Info "no gitmap entries found in User PATH"
    }
}

function Remove-DataFolder {
    $appdata = Join-Path $env:APPDATA "gitmap"
    if (-not (Test-Path $appdata)) { return }

    if ($KeepData) {
        Write-Info "keeping data folder: $appdata"
        return
    }

    if (-not (Confirm-Or-Exit "Also delete user data at $appdata?")) {
        Write-Info "kept: $appdata"
        return
    }

    try {
        Remove-Item $appdata -Recurse -Force -ErrorAction Stop
        Write-Ok "removed $appdata"
    } catch {
        Write-Err "could not remove $appdata : $_"
    }
}

# ---------------------------------------------------------------------------
# Main
# ---------------------------------------------------------------------------

Write-Host ""
Write-Host "  gitmap quick uninstaller" -ForegroundColor Cyan
Write-Host "  ------------------------" -ForegroundColor DarkGray
Write-Host ""

Write-Step "[1/4] Trying canonical self-uninstall"
$ok = Try-SelfUninstall

if (-not $ok) {
    Write-Host ""
    Write-Step "[2/4] Manual sweep — locating deploy root"
    $root = Resolve-DeployRoot
    if ($root) { Write-Info "Deploy root: $root" } else { Write-Warn "no deploy root found" }

    Write-Host ""
    Write-Step "[3/4] Removing deploy folders"
    Remove-DeployFolders $root

    Write-Host ""
    Write-Step "[4/4] Cleaning User PATH"
    Remove-FromUserPath $root
}

Write-Host ""
Write-Step "User data"
Remove-DataFolder

Write-Host ""
Write-Host "  Done. Open a new terminal to refresh PATH." -ForegroundColor Green
Write-Host ""
