package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"time"

	"github.com/heroiclabs/nakama-common/runtime"
	cfg "windypath.com/cs2match/config"
	"windypath.com/cs2match/server/internal/framework/matchengine"
	"windypath.com/cs2match/server/internal/match"
)

// InitModule 是 Nakama Go Plugin 的入口函数。
// Nakama 加载 .so 插件时自动调用此函数，在此注册所有 RPC、Match Handler、Hooks 等。
func InitModule(
	ctx context.Context,
	logger runtime.Logger,
	db *sql.DB,
	nk runtime.NakamaModule,
	initializer runtime.Initializer,
) error {
	logger.Info("CS2Match Go plugin loaded successfully")

	// 初始化配置表
	err := cfgInit(logger)
	if err != nil {
		return err
	}

	// 打印示例道具
	testItemCfg(logger)

	//注册RPC事件
	err = registerRpcFunc(initializer, logger)
	if err != nil {
		return err
	}

	return nil
}

func registerRpcFunc(initializer runtime.Initializer, logger runtime.Logger) error {
	if err := initializer.RegisterRpc("HealthCheck", healthCheckRPC); err != nil {
		logger.Error("Failed to register HealthCheck RPC: %v", err)
		return err
	}
	logger.Info("HealthCheck RPC registered")

	// 初始化 match 子系统与战斗引擎
	engineService := matchengine.NewService(logger)
	matchService := match.NewService(engineService, logger)

	if err := initializer.RegisterRpc("DebugSimuMatch", match.RPCDebugSimuMatch(matchService)); err != nil {
		logger.Error("Failed to register DebugSimuMatch RPC: %v", err)
		return err
	}
	logger.Info("DebugSimuMatch RPC registered")
	return nil
}

func testItemCfg(logger runtime.Logger) {
	if item := cfg.GetFirstItem(); item != nil {
		logger.Info("Sample item: id=%d name=%s desc=%s price=%d",
			item.Id, item.Name, item.Desc, item.Price)
	}
}

func cfgInit(logger runtime.Logger) error {
	if err := cfg.Init(); err != nil {
		logger.Error("Failed to init config: %v", err)
		return err
	}
	logger.Info("Config tables loaded, count=%d", cfg.TableCount())
	return nil
}

// healthCheckRPC 是一个简单的健康检查端点。
// 接受空 payload，返回服务器状态 JSON。
func healthCheckRPC(ctx context.Context, logger runtime.Logger, db *sql.DB, nk runtime.NakamaModule, payload string) (string, error) {
	response := map[string]interface{}{
		"status":    "ok",
		"timestamp": time.Now().UTC().Format(time.RFC3339),
		"version":   "0.1.0",
	}

	jsonBytes, err := json.Marshal(response)
	if err != nil {
		logger.Error("HealthCheck RPC failed to marshal response: %v", err)
		return "", err
	}

	return string(jsonBytes), nil
}
