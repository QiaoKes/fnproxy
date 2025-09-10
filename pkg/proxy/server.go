package proxy

import (
	"errors"
	"fmt"
	"fnproxy/pkg/config"
	"fnproxy/pkg/logger"
	"net/http"
	"net/http/httputil"
	"net/url"
	"strings"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// Server 代理服务器
type Server struct {
	config   *config.Config
	registry *Registry
	proxy    *httputil.ReverseProxy
	engine   *gin.Engine
}

// NewServer 创建新的代理服务器
func NewServer(cfg *config.Config) *Server {
	// 创建目标URL
	targetURL := &url.URL{
		Scheme: "http",
		Host:   fmt.Sprintf("%s:%d", cfg.Target.Host, cfg.Target.Port),
	}

	// 创建反向代理
	proxy := httputil.NewSingleHostReverseProxy(targetURL)

	// 自定义Director函数
	originalDirector := proxy.Director
	proxy.Director = func(req *http.Request) {
		originalDirector(req)
		req.Host = targetURL.Host
		req.URL.Host = targetURL.Host
		req.URL.Scheme = targetURL.Scheme
	}

	// 设置Gin模式
	if logger.GetLevel() == logger.DEBUG {
		gin.SetMode(gin.DebugMode)
	} else {
		gin.SetMode(gin.ReleaseMode)
	}

	engine := gin.New()
	engine.Use(gin.CustomRecovery(func(c *gin.Context, rec any) {
		// 忽略 ReverseProxy 的控制流中止
		if err, ok := rec.(error); ok && errors.Is(err, http.ErrAbortHandler) {
			c.Abort()
			return
		}
		// 其它 panic 仍按 500 处理
		c.AbortWithStatus(http.StatusInternalServerError)
	}))

	server := &Server{
		config:   cfg,
		registry: NewRegistry(),
		proxy:    proxy,
		engine:   engine,
	}

	// 设置路由
	server.setupRoutes()

	return server
}

// GetConfig 获取当前配置
func (s *Server) GetConfig() *config.Config {
	return s.config
}

// setupRoutes 设置路由
func (s *Server) setupRoutes() {
	// 捕获所有请求的通用处理器
	s.engine.Any("/*path", s.handleRequest)
}

// handleRequest 处理请求
func (s *Server) handleRequest(c *gin.Context) {
	defer func() {
		if rec := recover(); rec != nil {
			if err, ok := rec.(error); ok && errors.Is(err, http.ErrAbortHandler) {
				// 局部静默，避免冒泡到上层中间件
				return
			}
			// 非预期 panic 继续抛，让 ZeroNoiseRecovery 处理
			panic(rec)
		}
	}()

	path := c.Param("path")
	if path == "" {
		path = "/"
	}

	// 创建拦截器上下文
	ctx := NewContext(c)

	method := c.Request.Method

	// 判断是否视频流
	isStreamLink := c.Request.Header.Get("Range") != ""

	// 获取全局拦截器
	globalInterceptors := s.registry.GetGlobalInterceptors()

	// 检查是否有匹配的拦截器并获取路径参数
	interceptor, pathParams := s.registry.GetInterceptorWithParams(method, ctx.Path)

	// 将路径参数设置到上下文中
	if pathParams != nil {
		ctx.PathParams = pathParams
	}

	var hasAfterResponse bool

	// 执行全局请求前拦截器
	for _, globalInterceptor := range globalInterceptors {
		if globalInterceptor.PreRequest != nil {
			// 全局处理器不支持cancel
			globalInterceptor.PreRequest(ctx)
		}

		if globalInterceptor.AfterResponse != nil {
			hasAfterResponse = true
		}
	}

	// 执行请求前拦截器
	if interceptor != nil && interceptor.PreRequest != nil {
		result := interceptor.PreRequest(ctx)

		if result == Cancel {
			logger.Info("Request intercepted and cancelled",
				zap.String("path", path),
				zap.String("method", c.Request.Method))
			return
		}
	}

	// 检查是否有响应后处理器
	if interceptor != nil && interceptor.AfterResponse != nil {
		hasAfterResponse = true
	}

	// 如果有响应后处理器且不是视频流，启用响应拦截
	if hasAfterResponse && !isStreamLink {
		ctx.EnableResponseIntercept()
	}

	// 没有拦截器或继续执行，进行透明转发
	logger.Debugf("Proxying request path:%s, args:%s method:%s", path, ctx.RequestHelper.GetQuery(), ctx.Method)

	s.proxy.ServeHTTP(c.Writer, ctx.Request)

	// 执行响应后处理
	if hasAfterResponse && !isStreamLink {
		var interceptors []*Interceptor
		interceptors = append(interceptors, globalInterceptors...)
		if interceptor != nil {
			interceptors = append(interceptors, interceptor)
		}
		s.processResponse(ctx, interceptors)
	}
}

// shouldIntercept 检查路径是否应该被拦截
func (s *Server) shouldIntercept(requestPath, registeredPath string) bool {
	// 精确匹配
	if requestPath == registeredPath {
		return true
	}

	// 前缀匹配（支持通配符）
	if strings.HasSuffix(registeredPath, "*") {
		prefix := strings.TrimSuffix(registeredPath, "*")
		return strings.HasPrefix(requestPath, prefix)
	}

	return false
}

// processResponse 处理响应
func (s *Server) processResponse(ctx *Context, interceptors []*Interceptor) {
	logger.Info("Processing response",
		zap.String("path", ctx.Path),
		zap.Int("response_size", len(ctx.ResponseHelper.GetResponseBody())))

	for _, interceptor := range interceptors {
		if interceptor.AfterResponse != nil {
			interceptor.AfterResponse(ctx)
		}
	}
	// 刷新响应到客户端
	ctx.FlushResponse()
}

// Register 注册路由
func (s *Server) Register(method, path string, callback PreRequestFunc) {
	s.registry.RegisterPreRequest(method, path, callback)
	logger.Infof("Registered router:%s", path)
}

// RegisterPreRequest 注册请求前处理器
func (s *Server) RegisterPreRequest(method, path string, preReq PreRequestFunc) {
	s.registry.RegisterPreRequest(method, path, preReq)
	logger.Info("Registered pre-request interceptor", zap.String("path", path))
}

// RegisterAfterResponse 注册响应后处理器
func (s *Server) RegisterAfterResponse(method, path string, afterResp AfterResponseFunc) {
	s.registry.RegisterAfterResponse(method, path, afterResp)
	logger.Info("Registered after-response interceptor", zap.String("path", path))
}

// RegisterBoth 注册请求前和响应后处理器
func (s *Server) RegisterBoth(method, path string, preReq PreRequestFunc, afterResp AfterResponseFunc) {
	s.registry.RegisterBoth(method, path, preReq, afterResp)
	logger.Info("Registered both pre-request and after-response interceptors", zap.String("path", path))
}

// RegisterGlobalBoth 注册全局请求处理器
func (s *Server) RegisterGlobalBoth(preReq PreRequestFunc, afterResp AfterResponseFunc) {
	interceptor := &Interceptor{
		PreRequest:    preReq,
		AfterResponse: afterResp,
	}
	s.registry.RegisterGlobalInterceptor(interceptor)
	logger.Info("Registered global interceptor")
}

// Start 启动服务器
func (s *Server) Start() error {
	logger.Info("Starting proxy server",
		zap.String("listen", s.config.Server.Listen),
		zap.String("target", fmt.Sprintf("%s:%d", s.config.Target.Host, s.config.Target.Port)))

	return s.engine.Run(s.config.Server.Listen)
}

// GetEngine 获取Gin引擎（用于测试）
func (s *Server) GetEngine() *gin.Engine {
	return s.engine
}
