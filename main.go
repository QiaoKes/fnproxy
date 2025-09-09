package main

import (
	"fmt"
	"fnproxy/internal/api"
	"fnproxy/internal/interceptors/emby/cache"
	"fnproxy/internal/interceptors/emby/handler"
	"fnproxy/pkg/config"
	"fnproxy/pkg/logger"
	"fnproxy/pkg/proxy"

	"go.uber.org/zap"
)

func main() {
	// 加载配置
	cfg, err := config.Load()
	if err != nil {
		panic(fmt.Sprintf("Failed to load config: %v", err))
	}

	target := cfg.Target
	var url string
	if target.Https {
		url = fmt.Sprintf("https://%s:%d", target.Host, target.Port)
	} else {
		url = fmt.Sprintf("http://%s:%d", target.Host, target.Port)
	}

	err = cache.InitCacheManager(url, cfg.User.Username, cfg.User.Password)
	if err != nil {
		panic(fmt.Sprintf("Failed to init cache manager: %v", err))
	}

	logger.SetLevel(logger.LogLevel(cfg.Log.Level))

	// 创建代理服务器
	server := proxy.NewServer(cfg)

	// 注册自定义API（不转发）
	api.RegisterAPIs(server)

	// 注册Emby拦截器
	handler.RegisterInterceptors(server)

	// 注册示例拦截器（可选，用于演示）
	//examples.RegisterExampleInterceptors(server, logger)

	// 启动服务器
	logger.Info("Proxy server starting with response transformation...")
	if err := server.Start(); err != nil {
		logger.Fatal("Failed to start server", zap.Error(err))
	}
}
