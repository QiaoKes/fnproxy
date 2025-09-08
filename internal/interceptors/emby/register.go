package emby

import (
	"fnproxy/pkg/proxy"
	"go.uber.org/zap"
)

// RegisterInterceptors 注册所有Emby相关的拦截器
func RegisterInterceptors(server *proxy.Server, logger *zap.Logger, targetHost string, targetPort int) {
	// 注册认证拦截器
	authInterceptor := NewAuthInterceptor(logger, targetHost, targetPort)

	// 注册到 /emby/Users/AuthenticateByName 路径
	server.Register("/emby/Users/AuthenticateByName", authInterceptor.Intercept)

	logger.Info("Registered Emby interceptors",
		zap.String("target_host", targetHost),
		zap.Int("target_port", targetPort))
}
