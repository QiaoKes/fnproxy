package emby

import (
	"fnproxy/pkg/logger"
	"fnproxy/pkg/proxy"
	"go.uber.org/zap"
)

// GlobalInterceptor 全局拦截器
type GlobalInterceptor struct {
}

// NewGlobalInterceptor 创建新的全局拦截器
func NewGlobalInterceptor() *GlobalInterceptor {
	return &GlobalInterceptor{}
}

// AuthIntercept 拦截全局请求
func (ai *GlobalInterceptor) AuthIntercept(ctx *proxy.Context) proxy.InterceptorResult {
	// 记录请求信息
	logger.Info("Intercepting Emby auth request",
		zap.String("path", ctx.Path),
		zap.String("method", ctx.Method),
		zap.String("original_host", ctx.Request.Host))

	ctx.Headers.Set("X-Emby-Authorization", "Emby UserId=\"4917013c4fd44e71904d24a169b6ddd5\", Client=\"MyApp\", Device=\"MyDevice\", DeviceId=\"abcdef123456\", Version=\"1.0.0\"")
	ctx.Headers.Set("X-Emby-Token", "80261cc7f7cc4d9ebc5d9c2deddbe35c")

	logger.Info("Successfully modified Emby auth request")

	return proxy.Continue
}
