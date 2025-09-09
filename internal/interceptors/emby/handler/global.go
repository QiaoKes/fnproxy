package handler

import (
	"fnproxy/internal/interceptors/emby"
	"fnproxy/pkg/proxy"
)

// GlobalInterceptor 全局拦截器
type GlobalInterceptor struct {
}

// NewGlobalInterceptor 创建新的全局拦截器
func NewGlobalInterceptor() *GlobalInterceptor {
	return &GlobalInterceptor{}
}

// Intercept 拦截全局请求
func (ai *GlobalInterceptor) Intercept(ctx *proxy.Context) proxy.InterceptorResult {
	// 记录请求信息
	//logger.Info("Intercepting Emby global request",
	//	zap.String("path", ctx.Path),
	//	zap.String("method", ctx.Method),
	//	zap.String("original_host", ctx.Request.Host))

	if !ctx.Headers.Contains(emby.EmbyAuthHeader) || !ctx.Headers.Contains(emby.EmbyTokenHeader) {
		ctx.Headers.Set(emby.EmbyAuthHeader, emby.GetCacheManager().GetAuthHeader())
		ctx.Headers.Set(emby.EmbyTokenHeader, emby.GetCacheManager().GetToken())
	}

	//logger.Info("Successfully modified Emby global request")

	return proxy.Continue
}
