package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"time"

	"github.com/heroiclabs/nakama-common/runtime"
	cfg "windypath.com/cs2match/config"
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
	if err := cfg.Init(); err != nil {
		logger.Error("Failed to init config: %v", err)
		return err
	}
	logger.Info("Config tables loaded, count=%d", cfg.TableCount())

	// 打印示例道具
	if item := cfg.GetFirstItem(); item != nil {
		logger.Info("Sample item: id=%d name=%s desc=%s price=%d",
			item.Id, item.Name, item.Desc, item.Price)
	}

	if err := initializer.RegisterRpc("HealthCheck", healthCheckRPC); err != nil {
		logger.Error("Failed to register HealthCheck RPC: %v", err)
		return err
	}
	logger.Info("HealthCheck RPC registered")

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
