package main

import (
	"fmt"
	"fnproxy/internal/common/config"
	"fnproxy/pkg/proxy"

	"go.uber.org/zap"
)

func main() {
	// 加载配置
	cfg, err := config.Load()
	if err != nil {
		panic(fmt.Sprintf("Failed to load config: %v", err))
	}

	// 初始化日志
	var logger *zap.Logger
	if cfg.Log.Level == "debug" {
		logger, _ = zap.NewDevelopment()
	} else {
		logger, _ = zap.NewProduction()
	}
	defer logger.Sync()

	// 创建代理服务器
	server := proxy.NewServer(cfg, logger)

	// 注册自定义API（不转发）
	//api.RegisterAPIs(server)

	// 注册Emby拦截器
	//emby.RegisterInterceptors(server, logger, cfg.Target.Host, cfg.Target.Port)

	// 注册示例拦截器（可选，用于演示）
	//examples.RegisterExampleInterceptors(server, logger)

	// 启动服务器
	logger.Info("Proxy server starting with response transformation...")
	if err := server.Start(); err != nil {
		logger.Fatal("Failed to start server", zap.Error(err))
	}
}
