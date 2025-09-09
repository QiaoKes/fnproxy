package api

import (
	"fnproxy/internal/api/handler"
	"fnproxy/pkg/proxy"
	"net/http"
)

// RegisterAPIs 注册所有自定义API
func RegisterAPIs(server *proxy.Server) {
	// 用户相关API
	server.Register(http.MethodGet, handler.SimilarPath, handler.GetSimilar)
}

//// handleUsers 处理用户列表API
//func handleUsers(ctx *common2.Context) common2.InterceptorResult {
//	switch ctx.Method {
//	case "GET":
//		// 获取用户列表
//		users := []UserInfo{
//			{ID: 1, Username: "admin", Email: "admin@example.com", Role: "admin"},
//			{ID: 2, Username: "user1", Email: "user1@example.com", Role: "user"},
//			{ID: 3, Username: "user2", Email: "user2@example.com", Role: "user"},
//		}
//
//		ctx.ResponseHelper.SetJSONWithStatus(http.StatusOK, APIResponse{
//			Code:    200,
//			Message: "success",
//			Data:    users,
//		})
//
//	case "POST":
//		// 创建新用户
//		ctx.ResponseHelper.SetJSONWithStatus(http.StatusCreated, APIResponse{
//			Code:    201,
//			Message: "User created successfully",
//			Data:    map[string]interface{}{"id": 4, "username": "newuser"},
//		})
//
//	default:
//		ctx.ResponseHelper.SetJSONWithStatus(http.StatusMethodNotAllowed, APIResponse{
//			Code:    405,
//			Message: "Method not allowed",
//		})
//	}
//
//	return common2.Cancel // 不转发，直接处理
//}
//
//// handleUserDetail 处理用户详情API
//func handleUserDetail(ctx *common2.Context) common2.InterceptorResult {
//	// 从路径中提取用户ID
//	path := ctx.Path
//	userID := path[len("/api/users/"):]
//
//	switch ctx.Method {
//	case "GET":
//		// 获取用户详情
//		user := UserInfo{
//			ID:       1,
//			Username: "admin",
//			Email:    "admin@example.com",
//			Role:     "admin",
//		}
//
//		ctx.ResponseHelper.SetJSONWithStatus(http.StatusOK, APIResponse{
//			Code:    200,
//			Message: "success",
//			Data:    user,
//		})
//
//	case "PUT":
//		// 更新用户信息
//		ctx.ResponseHelper.SetJSONWithStatus(http.StatusOK, APIResponse{
//			Code:    200,
//			Message: "User updated successfully",
//			Data:    map[string]interface{}{"id": userID},
//		})
//
//	case "DELETE":
//		// 删除用户
//		ctx.ResponseHelper.SetJSONWithStatus(http.StatusOK, APIResponse{
//			Code:    200,
//			Message: "User deleted successfully",
//			Data:    map[string]interface{}{"id": userID},
//		})
//
//	default:
//		ctx.ResponseHelper.SetJSONWithStatus(http.StatusMethodNotAllowed, APIResponse{
//			Code:    405,
//			Message: "Method not allowed",
//		})
//	}
//
//	return common2.Cancel
//}
//
//// handleSystemInfo 处理系统信息API
//func handleSystemInfo(ctx *common2.Context) common2.InterceptorResult {
//	systemInfo := map[string]interface{}{
//		"version":    "1.0.0",
//		"build_time": "2024-01-01 12:00:00",
//		"go_version": "1.23",
//		"os":         "linux",
//		"arch":       "amd64",
//		"uptime":     "24h30m15s",
//		"memory_usage": map[string]interface{}{
//			"used":  "128MB",
//			"total": "512MB",
//		},
//	}
//
//	ctx.ResponseHelper.SetJSONWithStatus(http.StatusOK, APIResponse{
//		Code:    200,
//		Message: "success",
//		Data:    systemInfo,
//	})
//
//	return common2.Cancel
//}
//
//// handleSystemStatus 处理系统状态API
//func handleSystemStatus(ctx *common2.Context) common2.InterceptorResult {
//	status := map[string]interface{}{
//		"status":    "healthy",
//		"timestamp": time.Now().Unix(),
//		"services": map[string]string{
//			"proxy":    "running",
//			"database": "connected",
//			"cache":    "available",
//		},
//		"metrics": map[string]interface{}{
//			"requests_per_second": 150,
//			"error_rate":          0.1,
//			"response_time_avg":   "25ms",
//		},
//	}
//
//	ctx.ResponseHelper.SetJSONWithStatus(http.StatusOK, APIResponse{
//		Code:    200,
//		Message: "success",
//		Data:    status,
//	})
//
//	return common2.Cancel
//}
//
//// handleProxyStats 处理代理统计API
//func handleProxyStats(ctx *common2.Context) common2.InterceptorResult {
//	stats := map[string]interface{}{
//		"total_requests":     10520,
//		"intercepted":        1250,
//		"forwarded":          9270,
//		"error_count":        15,
//		"avg_response_time":  "45ms",
//		"active_connections": 23,
//		"uptime":             "72h15m30s",
//	}
//
//	ctx.ResponseHelper.SetJSONWithStatus(http.StatusOK, APIResponse{
//		Code:    200,
//		Message: "success",
//		Data:    stats,
//	})
//
//	return common2.Cancel
//}
//
//// handleProxyConfig 处理代理配置API
//func handleProxyConfig(ctx *common2.Context) common2.InterceptorResult {
//	switch ctx.Method {
//	case "GET":
//		// 获取当前配置
//		config := map[string]interface{}{
//			"listen_port":  2345,
//			"target_host":  "10.0.0.115",
//			"target_port":  8005,
//			"log_level":    "info",
//			"timeout":      "30s",
//			"interceptors": []string{"/api/*", "/emby/auth", "/health"},
//		}
//
//		ctx.ResponseHelper.SetJSONWithStatus(http.StatusOK, APIResponse{
//			Code:    200,
//			Message: "success",
//			Data:    config,
//		})
//
//	case "PUT":
//		// 更新配置
//		ctx.ResponseHelper.SetJSONWithStatus(http.StatusOK, APIResponse{
//			Code:    200,
//			Message: "Configuration updated successfully",
//		})
//
//	default:
//		ctx.ResponseHelper.SetJSONWithStatus(http.StatusMethodNotAllowed, APIResponse{
//			Code:    405,
//			Message: "Method not allowed",
//		})
//	}
//
//	return common2.Cancel
//}
//
//// handleFileUpload 处理文件上传API
//func handleFileUpload(ctx *common2.Context) common2.InterceptorResult {
//	if ctx.Method != "POST" {
//		ctx.ResponseHelper.SetJSONWithStatus(http.StatusMethodNotAllowed, APIResponse{
//			Code:    405,
//			Message: "Only POST method allowed",
//		})
//		return common2.Cancel
//	}
//
//	// 模拟文件上传处理
//	ctx.ResponseHelper.SetJSONWithStatus(http.StatusOK, APIResponse{
//		Code:    200,
//		Message: "File uploaded successfully",
//		Data: map[string]interface{}{
//			"file_id":     "f_12345",
//			"file_name":   "document.pdf",
//			"file_size":   "2.5MB",
//			"upload_time": time.Now().Format("2006-01-02 15:04:05"),
//		},
//	})
//
//	return common2.Cancel
//}
//
//// handleLogin 处理登录API
//func handleLogin(ctx *common2.Context) common2.InterceptorResult {
//	if ctx.Method != "POST" {
//		ctx.ResponseHelper.SetJSONWithStatus(http.StatusMethodNotAllowed, APIResponse{
//			Code:    405,
//			Message: "Only POST method allowed",
//		})
//		return common2.Cancel
//	}
//
//	// 模拟登录处理
//	token := "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.example_token"
//
//	ctx.ResponseHelper.SetJSONWithStatus(http.StatusOK, APIResponse{
//		Code:    200,
//		Message: "Login successful",
//		Data: map[string]interface{}{
//			"token":      token,
//			"expires_in": 3600,
//			"user": map[string]interface{}{
//				"id":       1,
//				"username": "admin",
//				"role":     "admin",
//			},
//		},
//	})
//
//	return common2.Cancel
//}
//
//// handleLogout 处理登出API
//func handleLogout(ctx *common2.Context) common2.InterceptorResult {
//	if ctx.Method != "POST" {
//		ctx.ResponseHelper.SetJSONWithStatus(http.StatusMethodNotAllowed, APIResponse{
//			Code:    405,
//			Message: "Only POST method allowed",
//		})
//		return common2.Cancel
//	}
//
//	ctx.ResponseHelper.SetJSONWithStatus(http.StatusOK, APIResponse{
//		Code:    200,
//		Message: "Logout successful",
//	})
//
//	return common2.Cancel
//}
//
//// handleDataSearch 处理数据搜索API
//func handleDataSearch(ctx *common2.Context) common2.InterceptorResult {
//	// 获取查询参数
//	query := ctx.Query("q")
//	page := ctx.DefaultQuery("page", "1")
//	limit := ctx.DefaultQuery("limit", "10")
//
//	searchResults := map[string]interface{}{
//		"query":       query,
//		"page":        page,
//		"limit":       limit,
//		"total":       156,
//		"total_pages": 16,
//		"results": []map[string]interface{}{
//			{"id": 1, "title": "Result 1", "description": "Description for result 1"},
//			{"id": 2, "title": "Result 2", "description": "Description for result 2"},
//			{"id": 3, "title": "Result 3", "description": "Description for result 3"},
//		},
//	}
//
//	ctx.ResponseHelper.SetJSONWithStatus(http.StatusOK, APIResponse{
//		Code:    200,
//		Message: "success",
//		Data:    searchResults,
//	})
//
//	return common2.Cancel
//}
//
//// handleWebSocket 处理WebSocket连接API
//func handleWebSocket(ctx *common2.Context) common2.InterceptorResult {
//	// 在实际应用中，这里会升级为WebSocket连接
//	ctx.ResponseHelper.SetJSONWithStatus(http.StatusOK, APIResponse{
//		Code:    200,
//		Message: "WebSocket endpoint ready",
//		Data: map[string]interface{}{
//			"protocol":         "ws",
//			"endpoint":         "/api/ws/connect",
//			"supported_events": []string{"message", "notification", "status_update"},
//		},
//	})
//
//	return common2.Cancel
//}
