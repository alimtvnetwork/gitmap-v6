<#
.SYNOPSIS
    Detect and display the last released version.
.DESCRIPTION
    Resolves the latest release version using one of three strategies:
    1. gitmap list-versions --limit 1 (if gitmap binary is available)
    2. .release/latest.json (if present in repo root)
    3. git tag (fallback: highest vX.Y.Z tag)
.PARAMETER BinaryPath
    Optional path to the gitmap binary. If omitted, attempts Get-Command.
.PARAMETER RepoRoot
    Optional repo root for reading .release/latest.json. Defaults to CWD.
.PARAMETER Label
    Display label prefix. Defaults to "Last release".
#>

param(
    [string]$BinaryPath = "",
    [string]$RepoRoot = "",
    [string]$Label = "Last release"
)

function Get-ReleaseFromBinary {
    param([string]$Binary)

    if ($Binary.Length -eq 0) {
        $cmd = Get-Command gitmap -ErrorAction SilentlyContinue
        if ($cmd -and (Test-Path $cmd.Source)) {
            $Binary = $cmd.Source
        }
    }

    if ($Binary.Length -gt 0 -and (Test-Path $Binary)) {
        # Strategy A: parse `list-versions` output, take the HIGHEST semver
        # (the command lists ascending, so --limit 1 gave the oldest — bug
        # surfaced as "Last release: v2.82.0" while the binary was 2.93.0).
        try {
            $output = & $Binary list-versions 2>&1
            if ($LASTEXITCODE -eq 0 -and $output) {
                $allVersions = @()
                foreach ($l in ($output | Out-String).Trim() -split "`n") {
                    if ($l -match '(v\d+\.\d+\.\d+)') {
                        $allVersions += $Matches[1]
                    }
                }
                if ($allVersions.Count -gt 0) {
                    $sorted = $allVersions | Sort-Object {
                        $parts = $_.TrimStart('v') -split '\.'
                        [int]$parts[0] * 1000000 + [int]$parts[1] * 1000 + [int]$parts[2]
                    } -Descending
                    return $sorted[0]
                }
            }
        } catch {
        }

        # Strategy B: ask the binary its own version directly
        try {
            $vOut = & $Binary version 2>&1
            $vText = ($vOut | Out-String).Trim()
            if ($vText -match '(v?\d+\.\d+\.\d+)') {
                $v = $Matches[1]
                if (-not $v.StartsWith('v')) { $v = "v$v" }
                return $v
            }
        } catch {
        }
    }

    return $null
}

function Get-ReleaseFromJSON {
    param([string]$Root)

    if ($Root.Length -eq 0) {
        $Root = (Get-Location).Path
    }

    $latestFile = Join-Path (Join-Path $Root ".release") "latest.json"
    if (Test-Path $latestFile) {
        try {
            $data = Get-Content $latestFile -Raw | ConvertFrom-Json
            if ($data.tag) {
                return $data.tag
            }
            if ($data.version) {
                return "v$($data.version)"
            }
        } catch {
        }
    }

    return $null
}

function Get-ReleaseFromGitTag {
    try {
        $tags = git tag --list "v*" --sort=-version:refname 2>&1
        if ($LASTEXITCODE -eq 0 -and $tags) {
            $lines = ($tags | Out-String).Trim() -split "`n"
            foreach ($t in $lines) {
                $t = $t.Trim()
                if ($t -match '^v\d+\.\d+\.\d+$') {
                    return $t
                }
            }
        }
    } catch {
    }

    return $null
}

# -- Resolution order ------------------------------------------
$release = Get-ReleaseFromBinary -Binary $BinaryPath
$source = "binary"

if (-not $release) {
    $release = Get-ReleaseFromJSON -Root $RepoRoot
    $source = "latest.json"
}

if (-not $release) {
    $release = Get-ReleaseFromGitTag
    $source = "git tag"
}

if ($release) {
    Write-Host "  ${Label}:    $release ($source)" -ForegroundColor DarkGray
} else {
    Write-Host "  ${Label}:    unknown" -ForegroundColor DarkGray
}
