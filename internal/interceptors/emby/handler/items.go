package handler

import (
	"fnproxy/internal/interceptors/emby/cache"
	"fnproxy/internal/interceptors/emby/common"
	"fnproxy/pkg/proxy"
)

// ItemsInterceptor Items拦截器
type ItemsInterceptor struct {
}

// NewItemsInterceptor 创建新的Items拦截器
func NewItemsInterceptor() *ItemsInterceptor {
	return &ItemsInterceptor{}
}

// UserItemsIntercept 拦截User Items请求
func (ai *ItemsInterceptor) UserItemsIntercept(ctx *proxy.Context) proxy.InterceptorResult {
	if !ctx.Headers.Contains(common.EmbyAuthHeader) || !ctx.Headers.Contains(common.EmbyTokenHeader) {
		ctx.Headers.Set(common.EmbyAuthHeader, cache.GetCacheManager().GetAuthHeader())
		ctx.Headers.Set(common.EmbyTokenHeader, cache.GetCacheManager().GetToken())
	}

	return proxy.Continue
}
