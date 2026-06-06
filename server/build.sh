#!/usr/bin/env bash
# ==============================================
# CS2Match Go Plugin 编译脚本 (Linux / macOS / WSL2)
# ==============================================
# 用途: 使用 Nakama 官方 pluginbuilder 镜像编译 .so 插件
# 输出: server/build/backend.so
# 前置: Docker
#
# 使用: bash server/build.sh
# ==============================================

set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_DIR="$(cd "$SCRIPT_DIR/.." && pwd)"
BUILD_DIR="${SCRIPT_DIR}/build"
OUTPUT="${BUILD_DIR}/backend.so"

# 确保 build 目录存在
mkdir -p "${BUILD_DIR}"

echo "=== CS2Match Go Plugin Build ==="
echo "Using: Nakama pluginbuilder 3.30.0"
echo "Output: ${OUTPUT}"

# 检查配置表是否已生成，如果没有则自动导表
if [ ! -f "${SCRIPT_DIR}/config/Tables.go" ]; then
    echo ""
    echo "Config tables not found, running gen-config first..."
    bash "${PROJECT_DIR}/scripts/gen-config.sh"
fi

# 使用 Nakama 官方 pluginbuilder 镜像编译
# 该镜像包含与 Nakama 3.30.0 完全匹配的 Go 1.24.5 及依赖
docker run --rm \
  --entrypoint "" \
  -v "${SCRIPT_DIR}:/app" \
  -w /app \
  registry.heroiclabs.com/heroiclabs/nakama-pluginbuilder:3.30.0 \
  go build -v -mod=mod -buildmode=plugin -trimpath -o build/backend.so .

echo "=== Build complete: ${OUTPUT} ==="
echo "Restart Nakama to reload: docker compose restart nakama"
