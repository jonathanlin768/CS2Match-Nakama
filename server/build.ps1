# ==============================================
# CS2Match Go Plugin 编译脚本 (Windows)
# ==============================================
# 用途: 编译 main.go 为 .so 插件，供 Nakama 容器加载
# 输出: server/build/backend.so
# 前置: WSL2 + Go 1.22+ (在 WSL2 中运行)
#
# 使用: wsl bash server/build.sh
#       或直接在 WSL2 终端中: bash server/build.sh
# ==============================================

Write-Host "=== CS2Match Go Plugin Build (Windows) ===" -ForegroundColor Cyan
Write-Host ""
Write-Host "This script must be run inside WSL2 (Windows Subsystem for Linux)"
Write-Host "because the Go plugin binary must be Linux ELF format."
Write-Host ""
Write-Host "Please run one of the following commands instead:" -ForegroundColor Yellow
Write-Host ""
Write-Host "  wsl bash server/build.sh" -ForegroundColor Green
Write-Host ""
Write-Host "Or open a WSL2 terminal and run:" -ForegroundColor Yellow
Write-Host ""
Write-Host "  cd /mnt/d/Project/CS2Match-Nakama" -ForegroundColor Green
Write-Host "  bash server/build.sh" -ForegroundColor Green
Write-Host ""

# 尝试在 WSL 中自动执行
$wslPath = (Get-Command wsl -ErrorAction SilentlyContinue)
if ($wslPath) {
    Write-Host "WSL detected! Attempting to compile in WSL..." -ForegroundColor Cyan
    $projectDir = $PWD.Path -replace '\\', '/' -replace '^([A-Z]):', '/mnt/$1'
    $projectDir = $projectDir.ToLower()
    wsl bash -c "cd '$projectDir' && bash server/build.sh"
} else {
    Write-Host "ERROR: WSL not found. Please install WSL2 and try again." -ForegroundColor Red
    Write-Host "https://learn.microsoft.com/en-us/windows/wsl/install" -ForegroundColor Yellow
    exit 1
}
