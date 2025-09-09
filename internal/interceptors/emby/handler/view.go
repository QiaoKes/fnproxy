package handler

import (
	"fnproxy/pkg/logger"
	"fnproxy/pkg/proxy"
	"fnproxy/pkg/utils"
	"strings"
)

// PageViewInterceptor Emby页面概览拦截器
type PageViewInterceptor struct {
}

// NewPageViewInterceptor 创建新的Emby页面概览拦截器
func NewPageViewInterceptor() *PageViewInterceptor {
	return &PageViewInterceptor{}
}

// ViewIntercept 拦截Emby页面概览请求
func (pi *PageViewInterceptor) ViewIntercept(ctx *proxy.Context) proxy.InterceptorResult {
	// 记录请求信息
	logger.Infof("Intercepting Emby view picture request before, path:%s, args:%s method:%s", utils.JsonPrint(ctx.Path), ctx.RequestHelper.GetQuery(), ctx.Method)

	ctx.RequestHelper.DelQuery("IncludeItemTypes")

	logger.Info("Successfully modified Emby view request")

	return proxy.Continue
}

// ArtPictureIntercept 拦截Emby封面图片请求
func (pi *PageViewInterceptor) ArtPictureIntercept(ctx *proxy.Context) proxy.InterceptorResult {
	logger.Infof("Intercepting Emby art picture request before, path:%s, args:%s method:%s", utils.JsonPrint(ctx.Path), ctx.RequestHelper.GetQuery(), ctx.Method)

	path := ctx.Path
	path = strings.Replace(path, "Backdrop/0", "Primary", 1)

	ctx.RequestHelper.SetPath(path)
	logger.Info("Successfully modified Emby art picture request")
	return proxy.Continue
}
