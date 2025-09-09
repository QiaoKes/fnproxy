package handler

import (
	"fnproxy/internal/interceptors/emby/common"
	"fnproxy/pkg/logger"
	"fnproxy/pkg/proxy"
	"net/http"
)

// GetSimilar 获取相似影视
func GetSimilar(ctx *proxy.Context) proxy.InterceptorResult {
	logger.Infof("GetSimilar similar request intercepted, path:%s, args:%s method:%s", ctx.Path, ctx.RequestHelper.GetQuery(), ctx.Method)
	ctx.ResponseHelper.SetJSONWithStatus(http.StatusOK, common.SimilarResp{
		Items:            make([]any, 0),
		TotalRecordCount: 0,
		StartIndex:       0,
	})
	return proxy.Cancel
}
