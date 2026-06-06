@echo off
echo === CS2Match Go Plugin Build ===
docker run --rm --entrypoint "" -v "%~dp0:/app" -w /app registry.heroiclabs.com/heroiclabs/nakama-pluginbuilder:3.30.0 go build -v -mod=mod -buildmode=plugin -trimpath -o build/backend.so .
if %ERRORLEVEL% neq 0 (
    echo Build failed!
    exit /b 1
)
echo === Build complete: %~dp0build\backend.so ===
