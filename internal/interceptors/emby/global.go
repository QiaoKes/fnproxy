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

	if !ctx.Headers.Contains(EmbyAuthHeader) || !ctx.Headers.Contains(EmbyTokenHeader) {
		ctx.Headers.Set(EmbyAuthHeader, GetCacheManager().GetAuthHeader())
		ctx.Headers.Set(EmbyTokenHeader, GetCacheManager().GetToken())
	}

	logger.Info("Successfully modified Emby auth request")

	return proxy.Continue
}
