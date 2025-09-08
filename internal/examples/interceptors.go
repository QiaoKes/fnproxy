package examples

import (
	"crypto/md5"
	"encoding/json"
	"fmt"
	proxy2 "fnproxy/pkg/proxy"
	"net/http"
	"time"

	"go.uber.org/zap"
)

// RegisterExampleInterceptors 注册示例拦截器
func RegisterExampleInterceptors(server *proxy2.Server, logger *zap.Logger) {
	// 示例1: Emby认证拦截器
	server.Register("/emby/auth", func(ctx *proxy2.Context) proxy2.InterceptorResult {
		logger.Info("Intercepting Emby auth request")

		// 检查是否有有效的token
		token := ctx.Headers.Get("X-Emby-Token")
		if token == "" {
			// 返回自定义认证响应
			ctx.ResponseHelper.SetJSON(map[string]interface{}{
				"error": "Missing authentication token",
				"code":  401,
			})
			ctx.ResponseHelper.SetStatus(http.StatusUnauthorized)
			return proxy2.Cancel
		}

		// 添加自定义header
		ctx.Headers.Set("X-Proxy-Processed", "true")

		// 继续原有逻辑
		return proxy2.Continue
	})

	// 示例2: API版本转换拦截器
	server.Register("/api/v1/*", func(ctx *proxy2.Context) proxy2.InterceptorResult {
		logger.Info("Intercepting API v1 request", zap.String("path", ctx.Path))

		// 将v1版本转换为v2
		newPath := "/api/v2" + ctx.Path[7:] // 去掉"/api/v1"
		ctx.Request.URL.Path = newPath
		ctx.Headers.Set("X-API-Version", "v2")

		return proxy2.Continue
	})

	// 示例3: 响应转换拦截器 - JSON格式转换
	server.Register("/api/transform/*", func(ctx *proxy2.Context) proxy2.InterceptorResult {
		// 检查是否是响应处理阶段
		if responseBody := ctx.ResponseHelper.GetResponseBody(); responseBody != nil {
			logger.Info("Processing response transformation", zap.String("path", ctx.Path))

			// 获取原始响应
			originalResponse := string(responseBody)

			// 包装响应为统一格式
			transformedResponse := map[string]interface{}{
				"success":   true,
				"timestamp": time.Now().Unix(),
				"data":      json.RawMessage(originalResponse),
				"meta": map[string]interface{}{
					"processed_by":  "fnproxy",
					"original_path": ctx.Path,
					"version":       "1.0",
				},
			}

			// 将转换后的响应写回
			if newBody, err := json.Marshal(transformedResponse); err == nil {
				ctx.ResponseHelper.SetResponseBody(newBody)
				ctx.ResponseHelper.SetHeader("Content-Type", "application/json")
				ctx.ResponseHelper.SetHeader("X-Response-Transformed", "true")
			}
		}

		return proxy2.Continue
	})

	// 示例4: 响应过滤拦截器 - 移除敏感字段
	server.Register("/api/secure/*", func(ctx *proxy2.Context) proxy2.InterceptorResult {
		if responseBody := ctx.ResponseHelper.GetResponseBody(); responseBody != nil {
			logger.Info("Filtering sensitive data from response", zap.String("path", ctx.Path))

			var responseData map[string]interface{}
			if err := json.Unmarshal(responseBody, &responseData); err == nil {
				// 移除敏感字段
				delete(responseData, "password")
				delete(responseData, "secret")
				delete(responseData, "private_key")
				delete(responseData, "token")

				// 添加过滤标记
				responseData["_filtered"] = true
				responseData["_filter_time"] = time.Now().Unix()

				if filteredBody, err := json.Marshal(responseData); err == nil {
					ctx.ResponseHelper.SetResponseBody(filteredBody)
					ctx.ResponseHelper.SetHeader("X-Data-Filtered", "true")
				}
			}
		}

		return proxy2.Continue
	})

	// 示例5: 错误响应转换
	server.Register("/api/errors/*", func(ctx *proxy2.Context) proxy2.InterceptorResult {
		if responseBody := ctx.ResponseHelper.GetResponseBody(); responseBody != nil {
			status := ctx.ResponseHelper.GetResponseStatus()

			// 只处理错误响应
			if status >= 400 {
				logger.Info("Transforming error response",
					zap.String("path", ctx.Path),
					zap.Int("status", status))

				errorResponse := map[string]interface{}{
					"error":          true,
					"code":           status,
					"message":        "请求处理失败",
					"timestamp":      time.Now().Unix(),
					"path":           ctx.Path,
					"original_error": string(responseBody),
				}

				if errorBody, err := json.Marshal(errorResponse); err == nil {
					ctx.ResponseHelper.SetResponseBody(errorBody)
					ctx.ResponseHelper.SetHeader("Content-Type", "application/json")
				}
			}
		}

		return proxy2.Continue
	})

	// 示例6: 响应缓存和修改
	server.Register("/api/cache/*", func(ctx *proxy2.Context) proxy2.InterceptorResult {
		if responseBody := ctx.ResponseHelper.GetResponseBody(); responseBody != nil {
			logger.Info("Adding cache headers and modifying response", zap.String("path", ctx.Path))

			// 添加缓存控制头
			ctx.ResponseHelper.SetHeader("Cache-Control", "public, max-age=300")
			ctx.ResponseHelper.SetHeader("ETag", fmt.Sprintf("\"%x\"", md5.Sum(responseBody)))
			ctx.ResponseHelper.SetHeader("X-Cache-Status", "MISS")

			// 在响应中添加缓存信息
			var responseData map[string]interface{}
			if err := json.Unmarshal(responseBody, &responseData); err == nil {
				responseData["_cache_info"] = map[string]interface{}{
					"cached_at": time.Now().Unix(),
					"ttl":       300,
					"cache_key": fmt.Sprintf("cache_%x", md5.Sum([]byte(ctx.Path))),
				}

				if modifiedBody, err := json.Marshal(responseData); err == nil {
					ctx.ResponseHelper.SetResponseBody(modifiedBody)
				}
			}
		}

		return proxy2.Continue
	})

	// 示例7: 请求日志拦截器
	server.Register("/*", func(ctx *proxy2.Context) proxy2.InterceptorResult {
		logger.Info("Request intercepted",
			zap.String("method", ctx.Method),
			zap.String("path", ctx.Path),
			zap.String("user_agent", ctx.Headers.Get("User-Agent")),
			zap.String("remote_addr", ctx.Request.RemoteAddr))

		// 添加响应头
		ctx.ResponseHelper.SetHeader("X-Proxy-Timestamp", fmt.Sprintf("%d", time.Now().Unix()))

		return proxy2.Continue
	})

	// 示例8: 健康检查拦截器
	server.Register("/health", func(ctx *proxy2.Context) proxy2.InterceptorResult {
		ctx.ResponseHelper.SetJSON(map[string]interface{}{
			"status": "ok",
			"proxy":  "running",
			"target": fmt.Sprintf("%s:%d", "10.0.0.115", 8005),
		})
		return proxy2.Cancel
	})

	// 示例9: CORS处理拦截器
	server.Register("/api/*", func(ctx *proxy2.Context) proxy2.InterceptorResult {
		// 添加CORS头
		ctx.ResponseHelper.SetHeader("Access-Control-Allow-Origin", "*")
		ctx.ResponseHelper.SetHeader("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		ctx.ResponseHelper.SetHeader("Access-Control-Allow-Headers", "Content-Type, Authorization, X-Emby-Token")

		// 处理OPTIONS预检请求
		if ctx.Method == "OPTIONS" {
			ctx.ResponseHelper.SetStatus(http.StatusOK)
			return proxy2.Cancel
		}

		return proxy2.Continue
	})

	logger.Info("Example interceptors registered successfully")
}
