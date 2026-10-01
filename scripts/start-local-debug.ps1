# Start the full local stack with a debug-built Nakama Go plugin.
Set-StrictMode -Version Latest
$ErrorActionPreference = 'Stop'

$projectDir = (Resolve-Path -LiteralPath (Join-Path $PSScriptRoot '..')).Path
Push-Location $projectDir
try {
    if (-not (Test-Path -LiteralPath '.env')) {
        Copy-Item -LiteralPath '.env.example' -Destination '.env'
        Write-Host 'Created .env from .env.example.'
    } else {
        Write-Host 'Using existing .env without changing it.'
    }

    $configJson = & docker compose config --format json
    if ($LASTEXITCODE -ne 0) {
        throw 'Unable to read Docker Compose configuration.'
    }
    $config = $configJson | ConvertFrom-Json
    $frontendPort = [int]$config.services.frontend.ports[0].published
    $serverKey = [string]$config.services.nakama.environment.NAKAMA_SERVER_KEY
    if ($config.services.frontend.build.args.VITE_NAKAMA_SERVER_KEY -cne $serverKey) {
        throw 'VITE_NAKAMA_SERVER_KEY must match NAKAMA_SERVER_KEY for browser authentication.'
    }

    & (Join-Path $projectDir 'server/build.ps1')
    if ($LASTEXITCODE -ne 0) {
        throw 'Plugin build failed; Docker services were not restarted.'
    }

    $composeStarted = $false
    for ($buildAttempt = 1; $buildAttempt -le 3; $buildAttempt++) {
        & docker compose up -d --build
        if ($LASTEXITCODE -eq 0) {
            $composeStarted = $true
            break
        }
        if ($buildAttempt -lt 3) {
            Write-Warning "Docker image build/start failed (attempt $buildAttempt of 3); retrying."
            Start-Sleep -Seconds 3
        }
    }
    if (-not $composeStarted) {
        throw 'docker compose up -d --build failed after three attempts.'
    }

    $rpcUri = 'http://127.0.0.1:7350/v2/rpc/HealthCheck?http_key=' + [Uri]::EscapeDataString($serverKey)

    $ready = $false
    for ($attempt = 0; $attempt -lt 60; $attempt++) {
        $health = @(
            (& docker inspect --format '{{.State.Health.Status}}' cs2match-db 2>$null),
            (& docker inspect --format '{{.State.Health.Status}}' cs2match-nakama 2>$null),
            (& docker inspect --format '{{.State.Health.Status}}' cs2match-frontend 2>$null)
        )
        if (@($health | Where-Object { $_ -ne 'healthy' }).Count -eq 0) {
            try {
                $socket = [Net.Sockets.TcpClient]::new()
                try {
                    $connected = $socket.ConnectAsync('127.0.0.1', 2345).Wait(2000)
                } finally {
                    $socket.Dispose()
                }
                if ($connected) {
                    $page = Invoke-WebRequest -UseBasicParsing -Uri "http://127.0.0.1:$frontendPort/" -TimeoutSec 5
                    $rpc = Invoke-RestMethod -Method Post -Uri $rpcUri -ContentType 'application/json' -Body '""' -TimeoutSec 5
                    $payload = $rpc.payload | ConvertFrom-Json
                    if ($page.StatusCode -eq 200 -and $payload.status -eq 'ok') {
                        $ready = $true
                        break
                    }
                }
            } catch {
                # Compose health may turn green before the host port and RPC are ready.
            }
        }
        Start-Sleep -Seconds 2
    }

    if (-not $ready) {
        & docker compose ps
        & docker compose logs --tail 80 nakama
        throw 'Local debug stack did not become ready: check Nakama logs, frontend, and port 2345.'
    }

    Write-Host "Local debug stack is ready. Frontend: http://127.0.0.1:$frontendPort/; Go Remote: 127.0.0.1:2345."
    & docker compose ps
} finally {
    Pop-Location
}
