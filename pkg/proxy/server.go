package proxy

import (
	"fmt"
	"fnproxy/internal/common/config"
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
	logger   *zap.Logger
	registry *Registry
	proxy    *httputil.ReverseProxy
	engine   *gin.Engine
}

// NewServer 创建新的代理服务器
func NewServer(cfg *config.Config, logger *zap.Logger) *Server {
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
	if cfg.Log.Level == "debug" {
		gin.SetMode(gin.DebugMode)
	} else {
		gin.SetMode(gin.ReleaseMode)
	}

	engine := gin.New()
	engine.Use(gin.Recovery())

	server := &Server{
		config:   cfg,
		logger:   logger,
		registry: NewRegistry(),
		proxy:    proxy,
		engine:   engine,
	}

	// 设置路由
	server.setupRoutes()

	return server
}

// setupRoutes 设置路由
func (s *Server) setupRoutes() {
	// 捕获所有请求的通用处理器
	s.engine.Any("/*path", s.handleRequest)
}

// handleRequest 处理请求
func (s *Server) handleRequest(c *gin.Context) {
	path := c.Param("path")
	if path == "" {
		path = "/"
	}

	// 创建拦截器上下文
	ctx := NewContext(c)

	// 检查是否有匹配的拦截器
	interceptor := s.registry.GetInterceptor(path)
	var hasAfterResponse bool

	// 执行请求前拦截器
	if interceptor != nil && interceptor.PreRequest != nil {
		result := interceptor.PreRequest(ctx)
		if result == Cancel {
			s.logger.Info("Request intercepted and cancelled",
				zap.String("path", path),
				zap.String("method", c.Request.Method))
			return
		}
	}

	// 检查是否有响应后处理器
	if interceptor != nil && interceptor.AfterResponse != nil {
		hasAfterResponse = true
		ctx.EnableResponseIntercept()
	}

	// 没有拦截器或继续执行，进行透明转发
	s.logger.Info("Proxying request",
		zap.String("path", path),
		zap.String("method", c.Request.Method),
		zap.String("target", fmt.Sprintf("%s:%d", s.config.Target.Host, s.config.Target.Port)))

	s.proxy.ServeHTTP(c.Writer, c.Request)

	// 执行响应后处理
	if hasAfterResponse {
		s.processResponse(ctx, interceptor)
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
func (s *Server) processResponse(ctx *Context, interceptor *Interceptor) {
	s.logger.Info("Processing response",
		zap.String("path", ctx.Path),
		zap.Int("response_size", len(ctx.ResponseHelper.GetResponseBody())))

	// 执行响应后处理器
	if interceptor.AfterResponse != nil {
		interceptor.AfterResponse(ctx)
	}

	// 刷新响应到客户端
	ctx.FlushResponse()
}

// Register 注册完整拦截器
func (s *Server) Register(path string, interceptor *Interceptor) {
	s.registry.Register(path, interceptor)
	s.logger.Info("Registered interceptor", zap.String("path", path))
}

// RegisterPreRequest 注册请求前处理器
func (s *Server) RegisterPreRequest(path string, preReq PreRequestFunc) {
	s.registry.RegisterPreRequest(path, preReq)
	s.logger.Info("Registered pre-request interceptor", zap.String("path", path))
}

// RegisterAfterResponse 注册响应后处理器
func (s *Server) RegisterAfterResponse(path string, afterResp AfterResponseFunc) {
	s.registry.RegisterAfterResponse(path, afterResp)
	s.logger.Info("Registered after-response interceptor", zap.String("path", path))
}

// RegisterBoth 注册请求前和响应后处理器
func (s *Server) RegisterBoth(path string, preReq PreRequestFunc, afterResp AfterResponseFunc) {
	s.registry.RegisterBoth(path, preReq, afterResp)
	s.logger.Info("Registered both pre-request and after-response interceptors", zap.String("path", path))
}

// Start 启动服务器
func (s *Server) Start() error {
	s.logger.Info("Starting proxy server",
		zap.String("listen", s.config.Server.Listen),
		zap.String("target", fmt.Sprintf("%s:%d", s.config.Target.Host, s.config.Target.Port)))

	return s.engine.Run(s.config.Server.Listen)
}

// GetEngine 获取Gin引擎（用于测试）
func (s *Server) GetEngine() *gin.Engine {
	return s.engine
}
