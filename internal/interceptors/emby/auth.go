package emby

import (
	"bytes"
	"encoding/json"
	"fmt"
	"fnproxy/pkg/proxy"
	"io"
	"net/url"
	"strings"

	"go.uber.org/zap"
)

// AuthInterceptor Emby认证拦截器
type AuthInterceptor struct {
	logger     *zap.Logger
	targetHost string
	targetPort int
}

// NewAuthInterceptor 创建新的Emby认证拦截器
func NewAuthInterceptor(logger *zap.Logger, targetHost string, targetPort int) *AuthInterceptor {
	return &AuthInterceptor{
		logger:     logger,
		targetHost: targetHost,
		targetPort: targetPort,
	}
}

// AuthRequest 认证请求体结构
type AuthRequest struct {
	Username string `json:"Username"`
	Password string `json:"Password,omitempty"`
	Pw       string `json:"Pw"`
}

// EmbyAuthorizationHeader Emby认证头解析结果
type EmbyAuthorizationHeader struct {
	Client   string
	Device   string
	DeviceId string
	Version  string
	UserId   string
}

// Intercept 拦截Emby认证请求
func (ai *AuthInterceptor) Intercept(ctx *proxy.Context) proxy.InterceptorResult {
	// 检查是否是Emby认证请求
	if !ai.isEmbyAuthRequest(ctx) {
		return proxy.Continue
	}

	ai.logger.Info("Intercepting Emby auth request",
		zap.String("path", ctx.Path),
		zap.String("method", ctx.Method),
		zap.String("original_host", ctx.Request.Host))

	// 修改目标主机和端口
	ai.modifyTargetHost(ctx)

	// 解析并修改认证头
	if err := ai.modifyAuthHeader(ctx); err != nil {
		ai.logger.Error("Failed to modify auth header", zap.Error(err))
		return proxy.Continue
	}

	// 修改请求体
	if err := ai.modifyRequestBody(ctx); err != nil {
		ai.logger.Error("Failed to modify request body", zap.Error(err))
		return proxy.Continue
	}

	// 添加其他必要的头
	ai.setAdditionalHeaders(ctx)

	ai.logger.Info("Successfully modified Emby auth request",
		zap.String("new_host", fmt.Sprintf("%s:%d", ai.targetHost, ai.targetPort)))

	return proxy.Continue
}

// isEmbyAuthRequest 检查是否是Emby认证请求
func (ai *AuthInterceptor) isEmbyAuthRequest(ctx *proxy.Context) bool {
	return ctx.Method == "POST" &&
		strings.Contains(ctx.Path, "/emby/Users/AuthenticateByName")
}

// modifyTargetHost 修改目标主机
func (ai *AuthInterceptor) modifyTargetHost(ctx *proxy.Context) {
	// 修改URL的主机部分
	ctx.Request.URL.Host = fmt.Sprintf("%s:%d", ai.targetHost, ai.targetPort)
	ctx.Request.Host = ctx.Request.URL.Host

	// 更新Host头
	ctx.Headers.Set("Host", ctx.Request.URL.Host)
}

// parseEmbyAuthHeader 解析Emby认证头
func (ai *AuthInterceptor) parseEmbyAuthHeader(authHeader string) (*EmbyAuthorizationHeader, error) {
	result := &EmbyAuthorizationHeader{}

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
func (ai *AuthInterceptor) buildEmbyAuthHeader(auth *EmbyAuthorizationHeader) string {
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
	// URL解码（如果需要）
	if decoded, err := url.QueryUnescape(value); err == nil {
		value = decoded
	}
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

	ai.logger.Info("Modified Emby authorization header",
		zap.String("original", authHeader),
		zap.String("modified", newAuthHeader))

	return nil
}

// modifyRequestBody 修改请求体
func (ai *AuthInterceptor) modifyRequestBody(ctx *proxy.Context) error {
	// 读取原始请求体
	body, err := io.ReadAll(ctx.Request.Body)
	if err != nil {
		return fmt.Errorf("failed to read request body: %w", err)
	}

	// 解析JSON
	var authReq AuthRequest
	if err := json.Unmarshal(body, &authReq); err != nil {
		return fmt.Errorf("failed to parse request body: %w", err)
	}

	ai.logger.Info("Original request body",
		zap.String("username", authReq.Username),
		zap.Bool("has_password", authReq.Password != ""),
		zap.Bool("has_pw", authReq.Pw != ""))

	// 如果有Password字段，移除它，只保留Pw字段
	newAuthReq := AuthRequest{
		Username: authReq.Username,
		Pw:       authReq.Pw,
	}

	// 如果Pw为空但Password不为空，使用Password的值
	if newAuthReq.Pw == "" && authReq.Password != "" {
		newAuthReq.Pw = authReq.Password
	}

	// 重新编码为JSON
	newBody, err := json.Marshal(newAuthReq)
	if err != nil {
		return fmt.Errorf("failed to marshal new request body: %w", err)
	}

	// 设置新的请求体
	ctx.Request.Body = io.NopCloser(bytes.NewReader(newBody))
	ctx.Request.ContentLength = int64(len(newBody))

	// 更新Content-Length头
	ctx.Headers.Set("Content-Length", fmt.Sprintf("%d", len(newBody)))

	ai.logger.Info("Modified request body",
		zap.String("new_body", string(newBody)))

	return nil
}

// setAdditionalHeaders 设置其他必要的头
func (ai *AuthInterceptor) setAdditionalHeaders(ctx *proxy.Context) {
	// 设置User-Agent为目标格式
	ctx.Headers.Set("User-Agent", "okhttp/5.1.0")

	// 确保Content-Type正确
	ctx.Headers.Set("Content-Type", "application/json")

	// 保持其他头不变
	ctx.Headers.Set("Connection", "Keep-Alive")
	ctx.Headers.Set("Accept-Encoding", "gzip")
}
