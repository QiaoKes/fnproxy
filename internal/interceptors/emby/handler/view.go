package handler

import (
	"bytes"
	"encoding/json"
	"fnproxy/internal/interceptors/emby/cache"
	"fnproxy/internal/interceptors/emby/common"
	"fnproxy/pkg/config"
	"fnproxy/pkg/logger"
	"fnproxy/pkg/proxy"
	"fnproxy/pkg/utils"
	"go.uber.org/zap"
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
	index := ctx.GetPathParamWithDefault("index", "0")
	path = strings.Replace(path, "Backdrop/"+index, "Primary", 1)

	ctx.RequestHelper.SetPath(path)
	logger.Info("Successfully modified Emby art picture request")
	return proxy.Continue
}

// VideoSummaryAfterIntercept  拦截Emby视频详情返回值
func (pi *PageViewInterceptor) VideoSummaryAfterIntercept(ctx *proxy.Context) proxy.InterceptorResult {
	if len(ctx.GetPathParam("itemid")) != 32 {
		return proxy.Continue
	}

	body := ctx.ResponseHelper.GetResponseBody()
	if !bytes.Equal(bytes.TrimSpace(body), []byte("null")) {
		return proxy.Continue
	}

	userId := ctx.Headers.Get("userid")
	seasonId := ctx.GetPathParam("itemid")

	api := common.NewFnApi(config.GetConfig().Target.Url)
	c := cache.GetCacheManager()
	resp, err := api.GetEpisodes(seasonId, seasonId, userId, c.GetAuthHeader(), c.GetToken())
	if err != nil {
		logger.Errorf("Failed to get episodes for season %s: %v", seasonId, err)
		return proxy.Continue
	}

	transformed := common.TransformEpisodesToSeason(resp)
	// 序列化转换后的数据
	transformedBody, err := json.Marshal(transformed)
	if err != nil {
		logger.Error("Failed to marshal transformed system info", zap.Error(err))
		return proxy.Continue
	}

	logger.Infof("Successfully modified Emby video summary")
	enc := ctx.Response.Header().Get("Content-Encoding")
	_ = ctx.ResponseHelper.SetJSONBodyWithEncoding(transformedBody, enc)
	return proxy.Cancel
}
