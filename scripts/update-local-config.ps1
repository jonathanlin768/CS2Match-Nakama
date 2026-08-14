# Refresh all generated configuration consumers used by the local Docker stack.
$ErrorActionPreference = "Stop"
$ProjectDir = Split-Path -Parent $PSScriptRoot
Set-Location $ProjectDir

function Invoke-CheckedStep {
    param(
        [Parameter(Mandatory = $true)][string]$Name,
        [Parameter(Mandatory = $true)][scriptblock]$Command
    )

    Write-Host ""
    Write-Host "=== $Name ===" -ForegroundColor Cyan
    & $Command
    if ($LASTEXITCODE -ne 0) {
        throw "$Name failed with exit code $LASTEXITCODE"
    }
}

Invoke-CheckedStep "Generate Luban config" {
    powershell.exe -NoProfile -ExecutionPolicy Bypass -File scripts/gen-config.ps1
}

Invoke-CheckedStep "Build Nakama Go plugin" {
    cmd.exe /d /c server\build.bat
}

Invoke-CheckedStep "Start backend containers" {
    docker compose up -d db nakama
}

Invoke-CheckedStep "Restart Nakama to load config" {
    docker compose restart nakama
}

Invoke-CheckedStep "Rebuild frontend image" {
    docker compose build --no-cache frontend
}

Invoke-CheckedStep "Recreate frontend container" {
    docker compose up -d --no-deps frontend
}

Write-Host ""
Write-Host "Local frontend and backend config updated successfully." -ForegroundColor Green
Write-Host "Frontend: http://localhost:3000"
