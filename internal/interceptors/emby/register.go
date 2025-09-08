package emby

import (
	"fnproxy/pkg/proxy"
	"net/http"
)

const (
	EmbyAuthPath = "/emby/Users/AuthenticateByName"
)

// RegisterInterceptors 注册所有Emby相关的拦截器
func RegisterInterceptors(server *proxy.Server) {
	server.RegisterPreRequest(http.MethodPost, EmbyAuthPath, NewAuthInterceptor().AuthIntercept)
}
