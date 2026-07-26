# ============================================================
# gen-config.ps1 - Luban Export Script (Windows PowerShell)
# Run Luban via Docker, generate Server (Go) and Client (TS) configs
# ============================================================

param(
    [switch]$RebuildImage
)

$ErrorActionPreference = "Stop"
$ScriptDir = Split-Path -Parent $MyInvocation.MyCommand.Path
$ProjectDir = Split-Path -Parent $ScriptDir
$Image = "luban-runner"
$Conf = "configs/luban.conf"
$DefinesDir = Join-Path $ProjectDir "configs/Defines"

Write-Host "=============================================="
Write-Host "  Luban Config Export (Docker)"
Write-Host "=============================================="

# 1. Check Docker
Write-Host ""
Write-Host "[1/4] Checking Docker..."
$dockerAvailable = Get-Command docker -ErrorAction SilentlyContinue
if (-not $dockerAvailable) {
    Write-Host "[ERROR] Docker not found. Install Docker Desktop first." -ForegroundColor Red
    Write-Host "  https://www.docker.com/products/docker-desktop"
    exit 1
}
Write-Host "[OK] Docker ready" -ForegroundColor Green

# 2. Build Luban Docker image
Write-Host ""
Write-Host "[2/4] Preparing Luban image..."
Set-Location $ProjectDir
if (-not (Test-Path $DefinesDir)) {
    New-Item -ItemType Directory -Path $DefinesDir | Out-Null
    Write-Host "[OK] Created empty configs/Defines compatibility directory" -ForegroundColor Green
}
$imageExists = $false
docker image inspect $Image *> $null
if ($LASTEXITCODE -eq 0) {
    $imageExists = $true
}

if ($RebuildImage -or -not $imageExists) {
    docker build -t $Image tools/luban/ | Out-Null
    if ($LASTEXITCODE -ne 0) {
        Write-Host "[ERROR] Luban image build failed" -ForegroundColor Red
        exit 1
    }
    Write-Host "[OK] Luban image built" -ForegroundColor Green
} else {
    Write-Host "[OK] Luban image exists; skip build" -ForegroundColor Green
}

# 3. Clean old output (keep hand-written files)
Write-Host ""
Write-Host "[3/4] Cleaning old output..."

# server/config/ - keep hand-written module files and tests; delete generated code/data
Get-ChildItem server/config -ErrorAction SilentlyContinue |
    Where-Object { $_.Name -ne 'go.mod' -and $_.Name -ne 'loader.go' -and $_.Name -ne 'config_test.go' } |
    Remove-Item -Recurse -Force

# client/src/config/ - keep index.ts; delete the rest
Get-ChildItem client/src/config -ErrorAction SilentlyContinue |
    Where-Object { $_.Name -ne 'index.ts' } |
    Remove-Item -Recurse -Force

# client/public/data/config/ - all generated JSON, delete entirely
Remove-Item -Recurse -Force client/public/data/config -ErrorAction SilentlyContinue
Write-Host "[OK] Cleanup done" -ForegroundColor Green

# 4. Run export
Write-Host ""
Write-Host "[4/4] Running Luban export..."

# 4a. Server (Go) - two steps: code first, then data (avoid cleanup conflict)
Write-Host ""
Write-Host "  --- Server (go-json) ---"
docker run --rm -v "${PWD}:/workspace" $Image `
    -t all -c go-json `
    --conf $Conf `
    -x outputCodeDir=server/config `
    -x go-json.lubanGoModule=windypath.com/cs2match/config
if ($LASTEXITCODE -ne 0) {
    Write-Host "[ERROR] Server code generation failed" -ForegroundColor Red
    exit 1
}
docker run --rm -v "${PWD}:/workspace" $Image `
    -t all -d json `
    --conf $Conf `
    -x outputDataDir=server/config/data
if ($LASTEXITCODE -ne 0) {
    Write-Host "[ERROR] Server data generation failed" -ForegroundColor Red
    exit 1
}
Write-Host "[OK] Server config done" -ForegroundColor Green

# 4b. Client (TypeScript)
Write-Host ""
Write-Host "  --- Client (typescript-json) ---"
docker run --rm -v "${PWD}:/workspace" $Image `
    -t all -c typescript-json -d json `
    --conf $Conf `
    -x outputCodeDir=client/src/config `
    -x outputDataDir=client/public/data/config
if ($LASTEXITCODE -ne 0) {
    Write-Host "[ERROR] Client config generation failed" -ForegroundColor Red
    exit 1
}
Write-Host "[OK] Client config done" -ForegroundColor Green

# Post-process: add @ts-nocheck to generated TypeScript (suppresses unused-param warnings)
$schemaPath = "client/src/config/schema.ts"
if (Test-Path $schemaPath) {
    $schemaContent = Get-Content $schemaPath -Raw
    if ($schemaContent -notmatch '// @ts-nocheck') {
        Set-Content $schemaPath -Value ("// @ts-nocheck`n" + $schemaContent)
        Write-Host "[POST] Added @ts-nocheck to schema.ts" -ForegroundColor Yellow
    }
}

# Ensure hand-written files exist (Luban may have cleared outputCodeDir)
if (-not (Test-Path server/config/go.mod)) {
    Set-Content -Path server/config/go.mod -Value @"
module windypath.com/cs2match/config

go 1.24.5
"@
    Write-Host "[POST] Recreated server/config/go.mod" -ForegroundColor Yellow
}

if (-not (Test-Path server/config/loader.go)) {
    Set-Content -Path server/config/loader.go -Value @'
package cfg

import (
	"embed"
	"encoding/json"
	"fmt"
	"reflect"
	"strings"
)

//go:embed data/*.json
var configData embed.FS

// Global is the singleton config tables instance, available after Init()
var Global *Tables

// Init loads all config table data. Call from InitModule.
func Init() error {
	entries, err := configData.ReadDir("data")
	if err != nil {
		return fmt.Errorf("config: failed to read embedded config data: %w", err)
	}

	dataMap := make(map[string][]map[string]interface{})
	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".json") {
			continue
		}

		raw, err := configData.ReadFile("data/" + entry.Name())
		if err != nil {
			return fmt.Errorf("config: failed to read %s: %w", entry.Name(), err)
		}

		var rows []map[string]interface{}
		if err := json.Unmarshal(raw, &rows); err != nil {
			return fmt.Errorf("config: failed to parse %s: %w", entry.Name(), err)
		}

		tableName := strings.TrimSuffix(entry.Name(), ".json")
		dataMap[tableName] = rows
	}

	loader := func(tableName string) ([]map[string]interface{}, error) {
		data, ok := dataMap[tableName]
		if !ok {
			return nil, fmt.Errorf("config: table %s not found", tableName)
		}
		return data, nil
	}

	tables, err := NewTables(loader)
	if err != nil {
		return fmt.Errorf("config: failed to init tables: %w", err)
	}

	Global = tables
	return nil
}

// TableCount returns the number of loaded tables
func TableCount() int {
	if Global == nil {
		return 0
	}
	value := reflect.ValueOf(Global).Elem()
	count := 0
	for index := 0; index < value.NumField(); index++ {
		if !value.Field(index).IsNil() {
			count++
		}
	}
	return count
}

// GetPlayer returns a player config by id.
func GetPlayer(id string) *Player {
	if Global == nil || Global.TbPlayer == nil {
		return nil
	}
	return Global.TbPlayer.Get(id)
}

// GetFirstItem returns the first item (for debug logging)
func GetFirstItem() *item {
	if Global == nil || Global.Tbitem == nil {
		return nil
	}
	list := Global.Tbitem.GetDataList()
	if len(list) == 0 {
		return nil
	}
	return list[0]
}
'@
    Write-Host "[POST] Recreated server/config/loader.go" -ForegroundColor Yellow
}

if (-not (Test-Path server/config/config_test.go)) {
    Set-Content -Path server/config/config_test.go -Value @'
package cfg

import (
	"testing"
)

func TestConfigLoad(t *testing.T) {
	if err := Init(); err != nil {
		t.Fatalf("Init failed: %v", err)
	}
	if Global == nil {
		t.Fatal("Global is nil")
	}
	if Global.Tbitem == nil {
		t.Fatal("Tbitem not loaded")
	}
	if Global.TbPlayer == nil {
		t.Fatal("TbPlayer not loaded")
	}
	if Global.TbMapNode == nil {
		t.Fatal("TbMapNode not loaded")
	}
	if TableCount() < 11 {
		t.Fatalf("expected at least 11 tables, got %d", TableCount())
	}
	players := Global.TbPlayer.GetDataList()
	if len(players) != 10 {
		t.Fatalf("expected 10 players, got %d", len(players))
	}
	p := GetPlayer("player_niko")
	if p == nil {
		t.Fatal("player_niko not found")
	}
	if p.Name != "NiKo" {
		t.Fatalf("expected NiKo, got %s", p.Name)
	}
	t.Logf("Loaded %d tables and %d players, first: %s (entry=%d)", TableCount(), len(players), p.Name, p.Entry)
}
'@
    Write-Host "[POST] Recreated server/config/config_test.go" -ForegroundColor Yellow
}

if (-not (Test-Path client/src/config/index.ts)) {
    Set-Content -Path client/src/config/index.ts -Value @'
/**
 * Config table loader module
 * Auto-generated by Luban; do not modify schema.ts manually.
 */

export { Tables, Tbitem, item } from './schema';

const TABLE_NAMES = ['tbitem'];

export async function loadConfig(): Promise<InstanceType<typeof Tables>> {
  const { Tables } = await import('./schema');

  const dataCache: Record<string, unknown> = {};

  await Promise.all(
    TABLE_NAMES.map(async (name) => {
      const resp = await fetch(`/data/config/${name}.json`);
      if (!resp.ok) {
        throw new Error(`Failed to load config table "${name}": ${resp.status}`);
      }
      dataCache[name] = await resp.json();
    })
  );

  const loader = (tableName: string) => {
    const data = dataCache[tableName];
    if (!data) throw new Error(`Config table "${tableName}" not found`);
    return data;
  };

  return new Tables(loader);
}
'@
    Write-Host "[POST] Recreated client/src/config/index.ts" -ForegroundColor Yellow
}

Write-Host ""
Write-Host "=============================================="
Write-Host "  Export Complete!" -ForegroundColor Green
Write-Host "=============================================="
Write-Host ""
Write-Host "Output:"
Write-Host "  Server: server/config/"
Write-Host "  Client: client/src/config/ + client/public/data/config/"
