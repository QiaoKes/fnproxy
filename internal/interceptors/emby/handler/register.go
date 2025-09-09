package handler

import (
	"fnproxy/internal/interceptors/emby"
	"fnproxy/pkg/proxy"
	"net/http"
)

// RegisterInterceptors 注册所有Emby相关的拦截器
func RegisterInterceptors(server *proxy.Server) {
	server.RegisterGlobalBoth(NewGlobalInterceptor().Intercept, nil)
	server.RegisterPreRequest(http.MethodPost, emby.EmbyAuthPath, NewAuthInterceptor().Intercept)
	//server.RegisterAfterResponse(http.MethodGet, SystemInfoPath, NewSystemInfoInterceptor().SystemInfoIntercept)
	server.RegisterPreRequest(http.MethodGet, emby.PageViewPath, NewPageViewInterceptor().ViewIntercept)
	server.RegisterPreRequest(http.MethodGet, emby.ArtPicturePath, NewPageViewInterceptor().ArtPictureIntercept)
}
