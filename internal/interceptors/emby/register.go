package emby

import (
	"fnproxy/pkg/proxy"
	"net/http"
)

const (
	EmbyAuthPath   = "/emby/Users/AuthenticateByName"
	SystemInfoPath = "/emby/System/Info"
)

// RegisterInterceptors 注册所有Emby相关的拦截器
func RegisterInterceptors(server *proxy.Server) {
	server.RegisterGlobalBoth(NewGlobalInterceptor().AuthIntercept, nil)
	server.RegisterPreRequest(http.MethodPost, EmbyAuthPath, NewAuthInterceptor().AuthIntercept)
	server.RegisterAfterResponse(http.MethodGet, SystemInfoPath, NewSystemInfoInterceptor().SystemInfoIntercept)
}
