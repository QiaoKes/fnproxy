package main

import (
	"crypto/md5"
	"encoding/json"
	"fmt"
	"fnproxy/internal/api"
	"fnproxy/internal/common/config"
	proxy2 "fnproxy/pkg/proxy"
	"strings"
	"time"

	"go.uber.org/zap"
)

func main() {
	// 加载配置
	cfg, err := config.Load()
	if err != nil {
		panic(fmt.Sprintf("Failed to load config: %v", err))
	}

	// 初始化日志
	var logger *zap.Logger
	if cfg.Log.Level == "debug" {
		logger, _ = zap.NewDevelopment()
	} else {
		logger, _ = zap.NewProduction()
	}
	defer logger.Sync()

	// 创建代理服务器
	server := proxy2.NewServer(cfg, logger)

	// 注册自定义API（不转发）
	api.RegisterAPIs(server)

	// 注册响应转换拦截器
	registerResponseTransformers(server, logger)

	// 启动服务器
	logger.Info("Proxy server starting with response transformation...")
	if err := server.Start(); err != nil {
		logger.Fatal("Failed to start server", zap.Error(err))
	}
}

// registerResponseTransformers 注册响应转换拦截器
func registerResponseTransformers(server *proxy2.Server, logger *zap.Logger) {

	// 1. 统一响应格式转换器 - 包装所有响应为统一格式
	server.Register("/api/wrap/*", func(ctx *proxy2.Context) proxy2.InterceptorResult {
		// 请求前处理：添加请求标识
		ctx.Headers.Set("X-Request-ID", fmt.Sprintf("req_%d", time.Now().UnixNano()))

		// 响应后处理：检查是否有响应数据
		if responseBody := ctx.ResponseHelper.GetResponseBody(); responseBody != nil {
			logger.Info("Wrapping response in unified format", zap.String("path", ctx.Path))

			// 解析原始响应
			var originalData interface{}
			err := json.Unmarshal(responseBody, &originalData)
			if err != nil {
				originalData = string(responseBody) // 如果解析失败，保留原始字符串

			}

			// 创建统一响应格式
			wrappedResponse := map[string]interface{}{
				"success":    true,
				"timestamp":  time.Now().Unix(),
				"request_id": ctx.Headers.Get("X-Request-ID"),
				"data":       originalData,
				"meta": map[string]interface{}{
					"processed_by": "fnproxy",
					"path":         ctx.Path,
					"version":      "2.0",
				},
			}

			// 设置新的响应体
			if newBody, err := json.Marshal(wrappedResponse); err == nil {
				ctx.ResponseHelper.SetResponseBody(newBody)
				ctx.ResponseHelper.SetHeader("Content-Type", "application/json; charset=utf-8")
				ctx.ResponseHelper.SetHeader("X-Response-Wrapped", "true")
			}
		}

		return proxy2.Continue
	})

	// 2. 敏感数据脱敏转换器
	server.Register("/api/user/*", func(ctx *proxy2.Context) proxy2.InterceptorResult {
		if responseBody := ctx.ResponseHelper.GetResponseBody(); responseBody != nil {
			logger.Info("Sanitizing sensitive user data", zap.String("path", ctx.Path))

			var responseData map[string]interface{}
			if err := json.Unmarshal(responseBody, &responseData); err == nil {
				// 递归处理数据，脱敏敏感字段
				sanitizeData(responseData)

				// 添加脱敏标记
				responseData["_sanitized"] = true
				responseData["_sanitized_at"] = time.Now().Unix()

				if sanitizedBody, err := json.Marshal(responseData); err == nil {
					ctx.ResponseHelper.SetResponseBody(sanitizedBody)
					ctx.ResponseHelper.SetHeader("X-Data-Sanitized", "true")
					ctx.ResponseHelper.SetHeader("X-Privacy-Level", "high")
				}
			}
		}

		return proxy2.Continue
	})

	// 3. 错误响应美化转换器
	server.Register("/api/errors/*", func(ctx *proxy2.Context) proxy2.InterceptorResult {
		if responseBody := ctx.ResponseHelper.GetResponseBody(); responseBody != nil {
			status := ctx.ResponseHelper.GetResponseStatus()

			// 只处理错误响应 (4xx, 5xx)
			if status >= 400 {
				logger.Info("Beautifying error response",
					zap.String("path", ctx.Path),
					zap.Int("status", status))

				var originalError map[string]interface{}
				json.Unmarshal(responseBody, &originalError)

				// 创建友好的错误响应
				friendlyError := map[string]interface{}{
					"error":     true,
					"code":      status,
					"message":   getErrorMessage(status),
					"details":   originalError,
					"timestamp": time.Now().Unix(),
					"path":      ctx.Path,
					"help":      "请联系技术支持或查看API文档",
				}

				if errorBody, err := json.Marshal(friendlyError); err == nil {
					ctx.ResponseHelper.SetResponseBody(errorBody)
					ctx.ResponseHelper.SetHeader("Content-Type", "application/json")
					ctx.ResponseHelper.SetHeader("X-Error-Handled", "true")
				}
			}
		}

		return proxy2.Continue
	})

	// 4. 响应缓存和性能优化
	server.Register("/api/cache/*", func(ctx *proxy2.Context) proxy2.InterceptorResult {
		if responseBody := ctx.ResponseHelper.GetResponseBody(); responseBody != nil {
			logger.Info("Adding cache headers and optimizing response", zap.String("path", ctx.Path))

			// 计算响应的ETag
			etag := fmt.Sprintf("\"%x\"", md5.Sum(responseBody))

			// 添加缓存相关头部
			ctx.ResponseHelper.SetHeader("Cache-Control", "public, max-age=300, s-maxage=600")
			ctx.ResponseHelper.SetHeader("ETag", etag)
			ctx.ResponseHelper.SetHeader("Vary", "Accept-Encoding, Authorization")
			ctx.ResponseHelper.SetHeader("X-Cache-Status", "MISS")

			// 在响应中注入缓存信息
			var responseData map[string]interface{}
			if err := json.Unmarshal(responseBody, &responseData); err == nil {
				responseData["_cache"] = map[string]interface{}{
					"etag":      etag,
					"cached_at": time.Now().Unix(),
					"ttl":       300,
					"cache_key": fmt.Sprintf("cache_%x", md5.Sum([]byte(ctx.Path))),
				}

				if cachedBody, err := json.Marshal(responseData); err == nil {
					ctx.ResponseHelper.SetResponseBody(cachedBody)
				}
			}
		}

		return proxy2.Continue
	})

	// 5. API版本转换和向后兼容
	server.Register("/api/v1/*", func(ctx *proxy2.Context) proxy2.InterceptorResult {
		// 请求转换：将v1请求转换为v2
		newPath := "/api/v2" + ctx.Path[7:]
		ctx.Request.URL.Path = newPath
		ctx.Headers.Set("X-API-Version", "v2")

		// 响应转换：将v2响应转换回v1格式
		if responseBody := ctx.ResponseHelper.GetResponseBody(); responseBody != nil {
			logger.Info("Converting v2 response to v1 format", zap.String("original_path", ctx.Path))

			var v2Response map[string]interface{}
			if err := json.Unmarshal(responseBody, &v2Response); err == nil {
				// 转换为v1格式（示例：移除某些v2特有字段）
				v1Response := convertV2ToV1(v2Response)

				if v1Body, err := json.Marshal(v1Response); err == nil {
					ctx.ResponseHelper.SetResponseBody(v1Body)
					ctx.ResponseHelper.SetHeader("X-API-Version", "v1")
					ctx.ResponseHelper.SetHeader("X-Converted-From", "v2")
				}
			}
		}

		return proxy2.Continue
	})

	// 6. 数据格式转换器 (JSON to XML, 等)
	server.Register("/api/xml/*", func(ctx *proxy2.Context) proxy2.InterceptorResult {
		// 检查客户端是否请求XML格式
		acceptHeader := ctx.Headers.Get("Accept")
		if strings.Contains(acceptHeader, "application/xml") {
			if responseBody := ctx.ResponseHelper.GetResponseBody(); responseBody != nil {
				logger.Info("Converting JSON response to XML", zap.String("path", ctx.Path))

				// 这里可以实现JSON到XML的转换
				// 为简化示例，我们添加XML wrapper
				xmlResponse := fmt.Sprintf(`<?xml version="1.0" encoding="UTF-8"?>
<response>
    <data><![CDATA[%s]]></data>
    <timestamp>%d</timestamp>
    <format>xml</format>
</response>`, string(responseBody), time.Now().Unix())

				ctx.ResponseHelper.SetResponseBody([]byte(xmlResponse))
				ctx.ResponseHelper.SetHeader("Content-Type", "application/xml; charset=utf-8")
				ctx.ResponseHelper.SetHeader("X-Format-Converted", "json-to-xml")
			}
		}

		return proxy2.Continue
	})

	// 7. 通用响应头增强器
	server.Register("/*", func(ctx *proxy2.Context) proxy2.InterceptorResult {
		// 为所有响应添加安全和性能相关头部
		if responseBody := ctx.ResponseHelper.GetResponseBody(); responseBody != nil {
			// 安全头部
			ctx.ResponseHelper.SetHeader("X-Content-Type-Options", "nosniff")
			ctx.ResponseHelper.SetHeader("X-Frame-Options", "DENY")
			ctx.ResponseHelper.SetHeader("X-XSS-Protection", "1; mode=block")

			// 性能头部
			ctx.ResponseHelper.SetHeader("X-Response-Time", fmt.Sprintf("%dms", time.Now().UnixMilli()%1000))
			ctx.ResponseHelper.SetHeader("X-Proxy-Server", "fnproxy/1.0")

			// 调试头部
			if ctx.Headers.Get("X-Debug") != "" {
				ctx.ResponseHelper.SetHeader("X-Response-Size", fmt.Sprintf("%d", len(responseBody)))
				ctx.ResponseHelper.SetHeader("X-Processing-Time", "< 1ms")
			}
		}

		return proxy2.Continue
	})

	// 8. 健康检查端点
	server.Register("/health", func(ctx *proxy2.Context) proxy2.InterceptorResult {
		ctx.ResponseHelper.SetJSON(map[string]interface{}{
			"status":   "ok",
			"proxy":    "running",
			"target":   "10.0.0.115:8005",
			"features": []string{"response_transform", "header_injection", "data_sanitization"},
		})
		return proxy2.Cancel
	})

	logger.Info("Response transformation interceptors registered successfully")
}

// sanitizeData 递归脱敏敏感数据
func sanitizeData(data map[string]interface{}) {
	sensitiveFields := []string{"password", "secret", "token", "private_key", "credit_card", "ssn"}

	for key, value := range data {
		keyLower := strings.ToLower(key)

		// 检查是否是敏感字段
		for _, sensitive := range sensitiveFields {
			if strings.Contains(keyLower, sensitive) {
				data[key] = "***SANITIZED***"
				break
			}
		}

		// 递归处理嵌套对象
		if nestedMap, ok := value.(map[string]interface{}); ok {
			sanitizeData(nestedMap)
		}
	}
}

// getErrorMessage 获取友好的错误消息
func getErrorMessage(status int) string {
	switch status {
	case 400:
		return "请求参数错误，请检查您的输入"
	case 401:
		return "未授权访问，请先登录"
	case 403:
		return "无权限访问此资源"
	case 404:
		return "请求的资源不存在"
	case 500:
		return "服务器内部错误，请稍后重试"
	default:
		return "请求处理失败"
	}
}

// convertV2ToV1 将v2响应格式转换为v1格式
func convertV2ToV1(v2Response map[string]interface{}) map[string]interface{} {
	// 简化示例：移除v2特有的字段，保持向后兼容
	v1Response := make(map[string]interface{})

	for key, value := range v2Response {
		// 过滤掉v2特有字段
		if key != "meta" && key != "_links" && key != "pagination" {
			v1Response[key] = value
		}
	}

	// 添加v1格式的状态字段
	v1Response["status"] = "success"
	v1Response["api_version"] = "1.0"

	return v1Response
}
