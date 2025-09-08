package emby

import (
	"fmt"
	"fnproxy/pkg/logger"
	"fnproxy/pkg/proxy"
	"strings"

	"go.uber.org/zap"
)

// AuthInterceptor Emby认证拦截器
type AuthInterceptor struct {
}

// NewAuthInterceptor 创建新的Emby认证拦截器
func NewAuthInterceptor() *AuthInterceptor {
	return &AuthInterceptor{}
}

// AuthRequest 认证请求体结构
type AuthRequest struct {
	Username string `json:"Username"`
	Password string `json:"Password,omitempty"`
	Pw       string `json:"Pw"`
}

// AuthorizationHeader Emby认证头解析结果
type AuthorizationHeader struct {
	Client   string
	Device   string
	DeviceId string
	Version  string
	UserId   string
}

// AuthIntercept 拦截Emby认证请求
func (ai *AuthInterceptor) AuthIntercept(ctx *proxy.Context) proxy.InterceptorResult {
	// 记录请求信息
	logger.Info("Intercepting Emby auth request",
		zap.String("path", ctx.Path),
		zap.String("method", ctx.Method),
		zap.String("original_host", ctx.Request.Host))

	// 解析并修改认证头
	if err := ai.modifyAuthHeader(ctx); err != nil {
		logger.Error("Failed to modify auth header", zap.Error(err))
		return proxy.Continue
	}

	logger.Info("Successfully modified Emby auth request")

	return proxy.Continue
}

// parseEmbyAuthHeader 解析Emby认证头
func (ai *AuthInterceptor) parseEmbyAuthHeader(authHeader string) (*AuthorizationHeader, error) {
	result := &AuthorizationHeader{}

	// 移除 "Emby " 前缀
	authHeader = strings.TrimPrefix(authHeader, "Emby ")

	// 解析各个字段
	pairs := strings.Split(authHeader, ",")
	for _, pair := range pairs {
		pair = strings.TrimSpace(pair)

		if strings.Contains(pair, "=") {
			parts := strings.SplitN(pair, "=", 2)
			key := strings.TrimSpace(parts[0])
			value := strings.TrimSpace(parts[1])

			switch key {
			case "UserId":
				result.UserId = value
			case "Client":
				result.Client = value
			case "Device":
				result.Device = value
			case "DeviceId":
				result.DeviceId = value
			case "Version":
				result.Version = value
			}
		}
	}

	return result, nil
}

// buildEmbyAuthHeader 构建新的Emby认证头
func (ai *AuthInterceptor) buildEmbyAuthHeader(auth *AuthorizationHeader) string {
	var parts []string

	// 只保留指定的字段，并确保都用引号包围
	if auth.Client != "" {
		client := ai.ensureQuoted(auth.Client)
		parts = append(parts, fmt.Sprintf("Client=%s", client))
	}

	if auth.Device != "" {
		device := ai.ensureQuoted(auth.Device)
		parts = append(parts, fmt.Sprintf("Device=%s", device))
	}

	if auth.DeviceId != "" {
		deviceId := ai.ensureQuoted(auth.DeviceId)
		parts = append(parts, fmt.Sprintf("DeviceId=%s", deviceId))
	}

	if auth.Version != "" {
		version := ai.ensureQuoted(auth.Version)
		parts = append(parts, fmt.Sprintf("Version=%s", version))
	}

	return "Emby " + strings.Join(parts, ", ")
}

// ensureQuoted 确保值被引号包围
func (ai *AuthInterceptor) ensureQuoted(value string) string {
	// 移除现有的引号
	value = strings.Trim(value, "\"")
	// 添加引号
	return fmt.Sprintf("\"%s\"", value)
}

// modifyAuthHeader 修改认证头
func (ai *AuthInterceptor) modifyAuthHeader(ctx *proxy.Context) error {
	authHeader := ctx.Headers.Get("X-Emby-Authorization")
	if authHeader == "" {
		return fmt.Errorf("missing X-Emby-Authorization header")
	}

	// 解析现有的认证头
	auth, err := ai.parseEmbyAuthHeader(authHeader)
	if err != nil {
		return fmt.Errorf("failed to parse auth header: %w", err)
	}

	// 构建新的认证头（只保留指定字段）
	newAuthHeader := ai.buildEmbyAuthHeader(auth)

	// 设置新的认证头
	ctx.Headers.Set("X-Emby-Authorization", newAuthHeader)

	logger.Info("Modified Emby authorization header",
		zap.String("original", authHeader),
		zap.String("modified", newAuthHeader))

	return nil
}
