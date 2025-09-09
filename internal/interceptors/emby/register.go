package emby

import (
	"fnproxy/internal/interceptors/emby/common"
	"fnproxy/internal/interceptors/emby/handler"
	"fnproxy/pkg/proxy"
	"net/http"
)

// RegisterInterceptors 注册所有Emby相关的拦截器
func RegisterInterceptors(server *proxy.Server) {
	// 全局拦截器
	server.RegisterGlobalBoth(handler.NewGlobalInterceptor().ReqIntercept, handler.NewGlobalInterceptor().RespIntercept)

	// 认证相关
	server.RegisterPreRequest(http.MethodPost, common.EmbyAuthPath, handler.NewAuthInterceptor().Intercept)

	// 页面浏览相关
	server.RegisterPreRequest(http.MethodGet, common.PageViewPath, handler.NewPageViewInterceptor().ViewIntercept)
	server.RegisterPreRequest(http.MethodGet, common.ArtPicturePath, handler.NewPageViewInterceptor().ArtPictureIntercept)

	// 处理相似影视请求
	server.Register(http.MethodGet, common.ItemSimilarPath, handler.GetSimilar)
	server.Register(http.MethodGet, common.ArtistsSimilarPath, handler.GetSimilar)
	server.Register(http.MethodGet, common.MoviesSimilarPath, handler.GetSimilar)
	server.Register(http.MethodGet, common.AlbumsSimilarPath, handler.GetSimilar)
	server.Register(http.MethodGet, common.ShowsSimilarPath, handler.GetSimilar)
	server.Register(http.MethodGet, common.TrailersSimilarPath, handler.GetSimilar)
}
