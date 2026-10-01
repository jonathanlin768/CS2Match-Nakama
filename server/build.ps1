# Build the Linux Nakama plugin with Nakama's pinned Go toolchain.
Set-StrictMode -Version Latest
$ErrorActionPreference = 'Stop'

$serverDir = (Resolve-Path -LiteralPath $PSScriptRoot).Path
$buildDir = Join-Path $serverDir 'build'
$temporaryOutput = Join-Path $buildDir 'backend.so.tmp'
$output = Join-Path $buildDir 'backend.so'
$moduleProxy = if ($env:CS2MATCH_GOPROXY) { $env:CS2MATCH_GOPROXY } else { 'https://goproxy.cn,direct' }

if (-not (Test-Path -LiteralPath (Join-Path $serverDir 'config/Tables.go'))) {
    throw 'Generated Go config is missing. Run scripts/gen-config.ps1 first.'
}

New-Item -ItemType Directory -Path $buildDir -Force | Out-Null
Remove-Item -LiteralPath $temporaryOutput -Force -ErrorAction SilentlyContinue

try {
    & docker run --rm --entrypoint go `
        --mount "type=bind,source=$serverDir,target=/app" `
        --mount 'type=volume,source=cs2match-go-mod-cache,target=/go/pkg/mod' `
        --mount 'type=volume,source=cs2match-go-build-cache,target=/root/.cache/go-build' `
        --env "GOPROXY=$moduleProxy" `
        --workdir /app `
        heroiclabs/nakama-pluginbuilder:3.30.0 `
        build -mod=mod -buildmode=plugin -trimpath '-gcflags=all=-N -l' `
        -o build/backend.so.tmp .
    if ($LASTEXITCODE -ne 0) {
        throw "Nakama plugin build failed with exit code $LASTEXITCODE."
    }
    if (-not (Test-Path -LiteralPath $temporaryOutput) -or (Get-Item -LiteralPath $temporaryOutput).Length -eq 0) {
        throw 'Nakama plugin build produced no output.'
    }
    Move-Item -LiteralPath $temporaryOutput -Destination $output -Force
    Write-Host "Debug plugin built: $output"
} finally {
    Remove-Item -LiteralPath $temporaryOutput -Force -ErrorAction SilentlyContinue
}
