<#
.SYNOPSIS
    Short interactive installer for gitmap on Windows.

.DESCRIPTION
    Prompts the user for an install drive/folder (with a sensible default),
    then delegates to the canonical gitmap/scripts/install.ps1 with that path.

    Versioned repo discovery: if the source repo URL ends with -v<N>, this
    script probes for higher-numbered sibling repos (-v<N+1>, -v<N+2>, ...)
    and delegates to the latest available one. See:
      spec/01-app/95-installer-script-find-latest-repo.md

    Run via one-liner:
      irm https://raw.githubusercontent.com/alimtvnetwork/gitmap-v4/main/install-quick.ps1 | iex

    Or locally:
      ./install-quick.ps1
      ./install-quick.ps1 -InstallDir "E:\Tools\gitmap"
      ./install-quick.ps1 -NoDiscovery
      ./install-quick.ps1 -ProbeCeiling 50
#>

param(
    [string]$InstallDir    = "",
    [string]$Version       = "",
    [switch]$NoDiscovery,
    [int]$ProbeCeiling     = 30
)

$ErrorActionPreference = "Stop"
$ProgressPreference    = "SilentlyContinue"

$Repo          = "alimtvnetwork/gitmap-v4"
$InstallerUrl  = "https://raw.githubusercontent.com/$Repo/main/gitmap/scripts/install.ps1"
$DefaultDir    = "D:\gitmap"

# ---------------------------------------------------------------------------
# Versioned repo discovery (spec/01-app/95-installer-script-find-latest-repo.md)
# ---------------------------------------------------------------------------

function Split-RepoSuffix([string]$repo) {
    # Returns @{ Owner=...; Stem=...; N=<int> } or $null if no -v<N> suffix.
    if ($repo -match '^([^/]+)/(.+)-v(\d+)$') {
        return @{
            Owner = $Matches[1]
            Stem  = $Matches[2]
            N     = [int]$Matches[3]
        }
    }
    return $null
}

function Test-RepoExists([string]$url) {
    try {
        $resp = Invoke-WebRequest -Uri $url -Method Head -TimeoutSec 5 `
            -UseBasicParsing -ErrorAction Stop
        return ($resp.StatusCode -eq 200)
    } catch {
        return $false
    }
}

function Resolve-EffectiveRepo([string]$repo, [int]$ceiling) {
    $parts = Split-RepoSuffix $repo
    if ($null -eq $parts) {
        Write-Host "  [discovery] no -v<N> suffix on '$repo'; installing baseline as-is"
        return $repo
    }

    $owner    = $parts.Owner
    $stem     = $parts.Stem
    $baseline = $parts.N
    $effective = $baseline

    Write-Host "  [discovery] baseline: $owner/$stem-v$baseline"
    Write-Host "  [discovery] probe ceiling: $ceiling"

    for ($m = $baseline + 1; $m -le $ceiling; $m++) {
        $url = "https://github.com/$owner/$stem-v$m"
        if (Test-RepoExists $url) {
            Write-Host "  [discovery] HEAD $url ... HIT"
            $effective = $m
        } else {
            Write-Host "  [discovery] HEAD $url ... MISS (fail-fast)"
            break
        }
    }

    if ($effective -eq $baseline) {
        Write-Host "  [discovery] no higher version found; using baseline -v$baseline"
        return $repo
    }

    Write-Host "  [discovery] effective: $owner/$stem-v$effective (was -v$baseline)"
    return "$owner/$stem-v$effective"
}

function Invoke-DelegatedInstaller([string]$effectiveRepo, [string]$installDir, [string]$version, [int]$ceiling) {
    $delegatedUrl = "https://raw.githubusercontent.com/$effectiveRepo/main/install-quick.ps1"
    Write-Host "  [discovery] delegating to $delegatedUrl"

    $env:INSTALLER_DELEGATED = "1"
    try {
        $script = (Invoke-WebRequest -Uri $delegatedUrl -UseBasicParsing -TimeoutSec 15).Content
    } catch {
        Write-Host "  [discovery] [WARN] could not fetch delegated installer: $_" -ForegroundColor Yellow
        Write-Host "  [discovery] falling back to baseline installer" -ForegroundColor Yellow
        Remove-Item Env:INSTALLER_DELEGATED -ErrorAction SilentlyContinue
        return $false
    }

    $block = [ScriptBlock]::Create($script)

    $passArgs = @{ ProbeCeiling = $ceiling }
    if (-not [string]::IsNullOrWhiteSpace($installDir)) { $passArgs.InstallDir = $installDir }
    if (-not [string]::IsNullOrWhiteSpace($version))    { $passArgs.Version    = $version }

    & $block @passArgs
    return $true
}

# ---------------------------------------------------------------------------
# Discovery: only run when not already delegated and not opted out.
# ---------------------------------------------------------------------------

$alreadyDelegated = ($env:INSTALLER_DELEGATED -eq "1")

if ($alreadyDelegated) {
    Write-Host "  [discovery] INSTALLER_DELEGATED=1; skipping discovery (loop guard)"
} elseif ($NoDiscovery) {
    Write-Host "  [discovery] -NoDiscovery set; skipping probe"
} else {
    $effective = Resolve-EffectiveRepo $Repo $ProbeCeiling
    if ($effective -ne $Repo) {
        $delegated = Invoke-DelegatedInstaller $effective $InstallDir $Version $ProbeCeiling
        if ($delegated) { return }
        # If delegation failed we fall through and install baseline.
    }
}

# ---------------------------------------------------------------------------
# Baseline install flow (unchanged behaviour).
# ---------------------------------------------------------------------------

function Read-InstallDir([string]$default) {
    Write-Host ""
    Write-Host "  gitmap quick installer" -ForegroundColor Cyan
    Write-Host "  ---------------------" -ForegroundColor DarkGray
    Write-Host "  Choose install folder. Press Enter to accept the default." -ForegroundColor Gray
    Write-Host "  Default: $default" -ForegroundColor DarkGray

    $answer = Read-Host "  Install path"
    if ([string]::IsNullOrWhiteSpace($answer)) { return $default }
    return $answer.Trim('"').Trim()
}

function Save-DeployPath([string]$dir) {
    # Persist the chosen install path so `gitmap install scripts` and
    # `run.ps1` pick the same drive/folder automatically.
    try {
        if (-not (Test-Path $dir)) {
            New-Item -ItemType Directory -Path $dir -Force | Out-Null
        }
        $cfgPath = Join-Path $dir "powershell.json"
        $cfg = [ordered]@{
            deployPath  = $dir
            buildOutput = "./bin"
            binaryName  = "gitmap.exe"
            goSource    = "./gitmap"
            copyData    = $true
        }
        ($cfg | ConvertTo-Json) | Set-Content -Path $cfgPath -Encoding UTF8
        Write-Host "  Saved deployPath -> $cfgPath" -ForegroundColor DarkGray
    } catch {
        Write-Host "  [WARN] Could not save powershell.json: $_" -ForegroundColor Yellow
    }
}

if ([string]::IsNullOrWhiteSpace($InstallDir)) {
    $InstallDir = Read-InstallDir $DefaultDir
}

Write-Host ""
Write-Host "  Installing gitmap to: $InstallDir" -ForegroundColor Green
Write-Host ""

Save-DeployPath $InstallDir

$script = (Invoke-WebRequest -Uri $InstallerUrl -UseBasicParsing).Content
$block  = [ScriptBlock]::Create($script)

if ($Version -ne "") {
    & $block -InstallDir $InstallDir -Version $Version
} else {
    & $block -InstallDir $InstallDir
}
