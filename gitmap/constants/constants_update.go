package constants

// Update handoff file patterns.
const (
	UpdateCopyFmtExe  = "gitmap-update-%d.exe"
	UpdateCopyFmtUnix = "gitmap-update-%d"
	UpdateCopyGlob    = "gitmap-update-*"
	UpdateScriptGlob  = "gitmap-update-*.ps1"
)

// Update flags.
const FlagVerbose = "--verbose"
const FlagRepoPath = "--repo-path"

// Updater fallback.
const (
	UpdaterBin         = "gitmap-updater"
	MsgUpdaterFallback = "  → No source repo found. Delegating to %s...\n\n"
)

// Update UI messages.
const (
	MsgUpdateActive      = "  → Active: %s\n  → Handoff: %s\n"
	MsgUpdateCleanStart  = "\n  Cleaning up update artifacts..."
	MsgUpdateCleanDone   = "  ✓ Removed %d file(s)\n\n"
	MsgUpdateCleanNone   = "  ✓ Nothing to clean up"
	MsgUpdateTempRemoved = "  → Removed temp copy: %s\n"
	MsgUpdateOldRemoved  = "  → Removed backup: %s\n"
	UpdateRunnerLogStart = "update-runner starting, repo=%s"
	UpdateScriptLogExec  = "executing update script: %s"
	UpdateScriptLogExit  = "update script exited: err=%v"
)

// Update error messages.
const (
	ErrUpdateExecFind          = "Error finding executable: %v\n"
	ErrUpdateCopyFail          = "Error creating update copy: %v\n"
	ErrUpdateNoRunSH           = "  ✗ run.sh not found at %s — cannot update on this platform without it.\n"
	ErrUpdateCleanupExecPath   = "Error: could not resolve executable path at active-binary: %v (operation: resolve executable, reason: os.Executable failed)\n"
	ErrUpdateCleanupConfigRead = "Error: could not read cleanup config at %s: %v (operation: read config, reason: cleanup path resolution unavailable)\n"
	ErrUpdateCleanupGlob       = "Error: could not enumerate cleanup matches at %s: %v (operation: glob, reason: invalid cleanup pattern)\n"
	ErrUpdateCleanupRemove     = "Error: could not remove cleanup artifact at %s: %v (operation: remove, reason: file may be locked or missing)\n"
)

// Unix update messages.
const (
	MsgUpdateInstallDir = "  → Installed directory: %s\n"
)

// Update path resolution messages.
const (
	MsgUpdatePathMissing = "\n  ⚠ The saved source repository path no longer exists on disk.\n"
	MsgUpdatePathPrompt  = "  Enter the new path to the gitmap source repo: "
	ErrUpdatePathInvalid = "  ✗ Directory not found at %s (operation: resolve, reason: file does not exist)\n"
)

// Clone-on-missing-path constants.
const (
	SourceRepoCloneURL   = "https://github.com/alimtvnetwork/gitmap-v4.git"
	MsgUpdateCloning     = "\n  ■ Path does not exist. Cloning gitmap source into %s...\n"
	MsgUpdateCloneOK     = "  ✓ Cloned successfully.\n"
	ErrUpdateCloneFailed = "  ✗ Clone failed: %v\n"
)

// Update PowerShell script template sections.
const (
	UpdatePSHeader = `# gitmap self-update script (auto-generated)
Set-Location "%s"

# Refresh run.ps1 from origin BEFORE invoking it, so a stale/buggy local
# copy can't break the update flow (e.g. positional-binding errors from
# old code paths). Best-effort: silently skip if git is unavailable or
# the repo has uncommitted run.ps1 changes.
try {
    $gitCmd = Get-Command git -ErrorAction SilentlyContinue
    if ($gitCmd) {
        $statusOut = & git status --porcelain -- run.ps1 2>$null
        if ([string]::IsNullOrWhiteSpace($statusOut)) {
            & git fetch --quiet origin 2>$null | Out-Null
            $headBranch = (& git symbolic-ref --quiet --short HEAD 2>$null)
            if ($headBranch) {
                & git checkout --quiet "origin/$headBranch" -- run.ps1 2>$null | Out-Null
                Write-Host "  [INFO] Refreshed run.ps1 from origin/$headBranch" -ForegroundColor DarkGray
            }
        } else {
            Write-Host "  [INFO] Local run.ps1 has uncommitted changes; skipping refresh" -ForegroundColor DarkGray
        }
    }
} catch {
    Write-Host "  [WARN] Could not refresh run.ps1: $_" -ForegroundColor Yellow
}
`
	UpdatePSDeployDetect = `
$configPath = Join-Path "%s" "gitmap\powershell.json"
$deployedBinary = $null
$configDeployedBinary = $null
if (Test-Path $configPath) {
    $cfg = Get-Content $configPath | ConvertFrom-Json
    if ($cfg.deployPath) {
	    $configDeployedBinary = Join-Path $cfg.deployPath "gitmap-cli\gitmap.exe"
    }
}

$activeCmdForDeploy = Get-Command gitmap -ErrorAction SilentlyContinue
if ($activeCmdForDeploy -and (Test-Path $activeCmdForDeploy.Source)) {
    $resolvedActiveBinary = (Resolve-Path $activeCmdForDeploy.Source).Path
    $resolvedActiveDir = Split-Path $resolvedActiveBinary -Parent
    if ((Split-Path $resolvedActiveDir -Leaf) -in @("gitmap-cli","gitmap")) {
        $effectiveDeployTarget = Split-Path $resolvedActiveDir -Parent
    } else {
        $effectiveDeployTarget = Split-Path $resolvedActiveDir -Parent
    }
    if ($effectiveDeployTarget) {
        $deployedBinary = Join-Path $effectiveDeployTarget "gitmap-cli\gitmap.exe"
    }
}

if ((-not $deployedBinary) -and $configDeployedBinary) {
    $deployedBinary = $configDeployedBinary
}
`
	UpdatePSVersionBefore = `
$activeBinary = $null
$activeBefore = "unknown"
$cmdBefore = Get-Command gitmap -ErrorAction SilentlyContinue
if ($cmdBefore -and (Test-Path $cmdBefore.Source)) {
    $activeBinary = $cmdBefore.Source
    $activeBefore = & $activeBinary version 2>&1
}
`
	UpdatePSRunUpdate = `
Write-Host ""
Write-Host "  Starting update via run.ps1 -Update" -ForegroundColor Cyan
& "%s" -Update
$runExit = $LASTEXITCODE
if (($runExit -ne 0) -and ($runExit -ne $null)) {
    exit $runExit
}
`
	UpdatePSSync = `
# Auto-sync deployed binary to active PATH binary if they differ.
if ($activeBinary -and $deployedBinary -and (Test-Path $deployedBinary)) {
    $resolvedActive = (Resolve-Path $activeBinary -ErrorAction SilentlyContinue).Path
    $resolvedDeployed = (Resolve-Path $deployedBinary -ErrorAction SilentlyContinue).Path
    if ($resolvedActive -and $resolvedDeployed -and ($resolvedActive -ne $resolvedDeployed)) {
        $deployedVer = & $deployedBinary version 2>&1
        $activeVer = & $activeBinary version 2>&1
        if ($deployedVer -ne $activeVer) {
            Write-Host ""
            Write-Host "  Syncing deployed binary to active PATH location..." -ForegroundColor Cyan
            Write-Host "    From: $resolvedDeployed" -ForegroundColor DarkGray
            Write-Host "    To:   $resolvedActive" -ForegroundColor DarkGray
            $syncOK = $false

            # Step 1: Try direct Copy-Item.
            try {
                Copy-Item -Path $resolvedDeployed -Destination $resolvedActive -Force
                $syncOK = $true
                Write-Host "  [OK] Synced successfully." -ForegroundColor Green
            } catch {
                Write-Host "  [WARN] Copy-Item failed: $_" -ForegroundColor Yellow
            }

            # Step 2: Rename-then-copy fallback.
            if (-not $syncOK) {
                Write-Host "  Trying rename-then-copy fallback..." -ForegroundColor Cyan
                $backupPath = "$resolvedActive.old"
                try {
                    if (Test-Path $backupPath) { Remove-Item $backupPath -Force -ErrorAction SilentlyContinue }
                    Move-Item -Path $resolvedActive -Destination $backupPath -Force
                    Copy-Item -Path $resolvedDeployed -Destination $resolvedActive -Force
                    $syncOK = $true
                    Write-Host "  [OK] Synced via rename fallback." -ForegroundColor Green
                } catch {
                    Write-Host "  [WARN] Rename fallback failed: $_" -ForegroundColor Yellow
                    # Restore backup if rename succeeded but copy failed.
                    if ((Test-Path $backupPath) -and (-not (Test-Path $resolvedActive))) {
                        Move-Item -Path $backupPath -Destination $resolvedActive -Force -ErrorAction SilentlyContinue
                    }
                }
            }

            # Step 3: Kill stale processes and retry.
            if (-not $syncOK) {
                Write-Host "  Killing stale gitmap processes..." -ForegroundColor Cyan
                $currentPID = $PID
                $stale = Get-CimInstance Win32_Process -Filter "Name='gitmap.exe'" -ErrorAction SilentlyContinue |
                    Where-Object { $_.ProcessId -ne $currentPID }
                foreach ($proc in $stale) {
                    try {
                        Stop-Process -Id $proc.ProcessId -Force -ErrorAction SilentlyContinue
                        Write-Host "    Stopped PID $($proc.ProcessId)" -ForegroundColor DarkGray
                    } catch {}
                }
                if ($stale) { Start-Sleep -Milliseconds 500 }
                try {
                    Copy-Item -Path $resolvedDeployed -Destination $resolvedActive -Force
                    $syncOK = $true
                    Write-Host "  [OK] Synced after killing stale processes." -ForegroundColor Green
                } catch {
                    Write-Host "  [WARN] Still could not sync: $_" -ForegroundColor Yellow
                }
            }

            if (-not $syncOK) {
                Write-Host "  [HINT] Run 'gitmap doctor --fix-path' manually." -ForegroundColor Yellow
            }
        }
    }
}
`
	UpdatePSVersionAfter = `
$activeAfter = "unknown"
$deployedAfter = "unknown"
$cmdAfter = Get-Command gitmap -ErrorAction SilentlyContinue
if ($cmdAfter -and (Test-Path $cmdAfter.Source)) {
    $activeBinary = $cmdAfter.Source
    $activeAfter = & $activeBinary version 2>&1
} else {
    Write-Host "  [TRACE] Get-Command gitmap: not found in PATH" -ForegroundColor DarkGray
}
if ($deployedBinary -and (Test-Path $deployedBinary)) {
    $deployedAfter = & $deployedBinary version 2>&1
} else {
    if (-not $deployedBinary) {
        Write-Host "  [TRACE] deployedBinary: not resolved (check powershell.json deployPath)" -ForegroundColor DarkGray
    } else {
        Write-Host "  [TRACE] deployedBinary: path not found: $deployedBinary" -ForegroundColor DarkGray
    }
}
`
	UpdatePSVerify = `
Write-Host ""
Write-Host "  Version before:   $activeBefore" -ForegroundColor DarkGray
Write-Host "  Version active:   $activeAfter" -ForegroundColor DarkGray
Write-Host "  Version deployed: $deployedAfter" -ForegroundColor DarkGray
Write-Host "  Active binary:    $activeBinary" -ForegroundColor DarkGray
Write-Host "  Deployed binary:  $(if ($deployedBinary) { $deployedBinary } else { '(not resolved)' })" -ForegroundColor DarkGray
if ($configDeployedBinary -and $deployedBinary -and ($configDeployedBinary -ne $deployedBinary)) {
    Write-Host "  Config binary:    $configDeployedBinary" -ForegroundColor DarkGray
}

$lastReleaseScript = Join-Path (Join-Path (Join-Path "%s" "gitmap") "scripts") "Get-LastRelease.ps1"
if (Test-Path $lastReleaseScript) {
    & $lastReleaseScript -BinaryPath $activeBinary -RepoRoot "%s"
}

if ($activeAfter -ne "unknown" -and $deployedAfter -eq "unknown") {
    Write-Host ""
    Write-Host "  [WARN] Deployed binary could not be verified (not resolved or missing)." -ForegroundColor Yellow
    Write-Host "  [TRACE] activeAfter=$activeAfter  deployedAfter=$deployedAfter" -ForegroundColor DarkGray
    if ($configDeployedBinary -and $deployedBinary -and ($configDeployedBinary -ne $deployedBinary)) {
        Write-Host "  [HINT] powershell.json points to an older deploy location; using PATH-derived target for verification." -ForegroundColor Yellow
    } else {
        Write-Host "  [HINT] Check that powershell.json 'deployPath' points to the correct directory" -ForegroundColor Yellow
        Write-Host "         and that the binary exists at: $deployedBinary" -ForegroundColor Yellow
    }
    Write-Host "  [OK] Active PATH binary updated successfully: $activeAfter" -ForegroundColor Green
} elseif (($activeAfter -eq "unknown") -or ($activeAfter -ne $deployedAfter)) {
    Write-Host ""
    Write-Host "  [FAIL] Active PATH version does not match deployed version." -ForegroundColor Red
    Write-Host "  [TRACE] activeAfter=$activeAfter  deployedAfter=$deployedAfter" -ForegroundColor DarkGray
    if ($activeAfter -eq "unknown") {
        Write-Host "  [HINT] Active binary not found in PATH." -ForegroundColor Yellow
    } elseif ($configDeployedBinary -and $deployedBinary -and ($configDeployedBinary -ne $deployedBinary)) {
        Write-Host "  [HINT] powershell.json still references a different deploy location than the active PATH binary." -ForegroundColor Yellow
    }
    exit 1
} else {
    Write-Host "  [OK] Active PATH binary matches deployed version." -ForegroundColor Green
}
`
	UpdatePSPostActions = `
if ($activeBinary -and (Test-Path $activeBinary)) {
    Write-Host ""
    Write-Host "  Latest changelog:" -ForegroundColor Cyan
    & $activeBinary changelog --latest

    Write-Host ""
    Write-Host "  Cleaning update artifacts..." -ForegroundColor DarkGray
    & $activeBinary update-cleanup
}

Write-Host ""
exit 0
`
)

// Revert PowerShell script template sections.
const (
	RevertPSHeader = `# gitmap revert script (auto-generated)
Set-Location "%s"
`
	RevertPSBuild = `
Write-Host ""
Write-Host "  Building from checked-out version..." -ForegroundColor Cyan
& "%s"
$runExit = $LASTEXITCODE
if (($runExit -ne 0) -and ($runExit -ne $null)) {
    exit $runExit
}
`
	RevertPSPostActions = `
$cmdAfter = Get-Command gitmap -ErrorAction SilentlyContinue
if ($cmdAfter -and (Test-Path $cmdAfter.Source)) {
    $activeAfter = & $cmdAfter.Source version 2>&1
    Write-Host "  Active version: $activeAfter" -ForegroundColor DarkGray

    Write-Host ""
    Write-Host "  Cleaning artifacts..." -ForegroundColor DarkGray
    & $cmdAfter.Source update-cleanup
}

Write-Host ""
exit 0
`
)

// Set-source-repo messages.
const (
	ErrSetSourceRepoNoPath  = "  ✗ set-source-repo requires a path argument\n"
	ErrSetSourceRepoInvalid = "  ✗ Invalid source repo path: %s\n"
	MsgSetSourceRepoDone    = "  ✓ Source repo path saved: %s\n"
)

// Backup file extension glob.
const OldBackupGlob = "*.old"

// PowerShell execution arguments.
const (
	PSBin            = "powershell"
	PSExecPolicy     = "-ExecutionPolicy"
	PSBypass         = "Bypass"
	PSNoProfile      = "-NoProfile"
	PSNoLogo         = "-NoLogo"
	PSFile           = "-File"
	PSNonInteractive = "-NonInteractive"
	PSCommand        = "-Command"
)
