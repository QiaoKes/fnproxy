package handler

import (
	"fnproxy/internal/interceptors/emby/cache"
	"fnproxy/internal/interceptors/emby/common"
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

	if !ctx.Headers.Contains(common.EmbyAuthHeader) || !ctx.Headers.Contains(common.EmbyTokenHeader) {
		ctx.Headers.Set(common.EmbyAuthHeader, cache.GetCacheManager().GetAuthHeader())
		ctx.Headers.Set(common.EmbyTokenHeader, cache.GetCacheManager().GetToken())
	}

	//logger.Info("Successfully modified Emby global request")

	return proxy.Continue
}
