# 自定义API实现指南

## 概述

在这个代理框架中，您可以非常轻松地实现自定义API，而不进行转发。只需要注册拦截器并返回 `interceptor.Cancel` 即可。

## 实现方式

### 1. 基本API实现模式

```go
server.Register("/api/your-endpoint", func(ctx *interceptor.Context) interceptor.InterceptorResult {
    // 处理您的业务逻辑
    ctx.ResponseHelper.SetJSON(map[string]interface{}{
        "message": "Hello from custom API",
        "data": "your data here",
    })
    
    return interceptor.Cancel // 重要：返回Cancel表示不转发
})
```

### 2. 支持不同HTTP方法

```go
server.Register("/api/users", func(ctx *interceptor.Context) interceptor.InterceptorResult {
    switch ctx.Method {
    case "GET":
        // 获取用户列表
        ctx.ResponseHelper.SetJSON(userList)
    case "POST":
        // 创建新用户
        ctx.ResponseHelper.SetJSONWithStatus(201, newUser)
    case "PUT":
        // 更新用户
        ctx.ResponseHelper.SetJSON(updatedUser)
    case "DELETE":
        // 删除用户
        ctx.ResponseHelper.SetStatus(204)
    default:
        ctx.ResponseHelper.SetJSONWithStatus(405, map[string]string{
            "error": "Method not allowed",
        })
    }
    return interceptor.Cancel
})
```

## 已实现的API示例

### 用户管理API
- `GET /api/users` - 获取用户列表
- `POST /api/users` - 创建新用户
- `GET /api/users/{id}` - 获取用户详情
- `PUT /api/users/{id}` - 更新用户信息
- `DELETE /api/users/{id}` - 删除用户

### 系统信息API
- `GET /api/system/info` - 获取系统信息
- `GET /api/system/status` - 获取系统状态

### 代理管理API
- `GET /api/proxy/stats` - 获取代理统计信息
- `GET /api/proxy/config` - 获取代理配置
- `PUT /api/proxy/config` - 更新代理配置

### 认证API
- `POST /api/auth/login` - 用户登录
- `POST /api/auth/logout` - 用户登出

### 文件上传API
- `POST /api/upload` - 文件上传

### 数据搜索API
- `GET /api/data/search?q=keyword&page=1&limit=10` - 数据搜索

## API测试示例

启动服务器后，您可以使用以下命令测试API：

### 1. 获取用户列表
```bash
curl http://localhost:2345/api/users
```

### 2. 创建新用户
```bash
curl -X POST http://localhost:2345/api/users \
  -H "Content-Type: application/json" \
  -d '{"username":"newuser","email":"newuser@example.com"}'
```

### 3. 获取系统信息
```bash
curl http://localhost:2345/api/system/info
```

### 4. 获取代理统计
```bash
curl http://localhost:2345/api/proxy/stats
```

### 5. 用户登录
```bash
curl -X POST http://localhost:2345/api/auth/login \
  -H "Content-Type: application/json" \
  -d '{"username":"admin","password":"password"}'
```

### 6. 数据搜索
```bash
curl "http://localhost:2345/api/data/search?q=test&page=1&limit=5"
```

### 7. 健康检查
```bash
curl http://localhost:2345/health
```

## 可用的工具方法

### 请求操作
```go
// 获取查询参数
query := ctx.Query("q")
page := ctx.DefaultQuery("page", "1")

// 获取请求头
token := ctx.Headers.Get("Authorization")

// 获取请求体
body, err := ctx.RequestHelper.GetBody()

// 获取路径参数
userID := ctx.Param("id")
```

### 响应操作
```go
// 返回JSON响应
ctx.ResponseHelper.SetJSON(data)
ctx.ResponseHelper.SetJSONWithStatus(201, data)

// 返回字符串响应
ctx.ResponseHelper.SetString("Hello %s", name)
ctx.ResponseHelper.SetStringWithStatus(404, "Not found")

// 设置响应头
ctx.ResponseHelper.SetHeader("X-Custom-Header", "value")

// 设置状态码
ctx.ResponseHelper.SetStatus(204)
```

## 添加新API的步骤

### 1. 在api/handlers.go中添加处理函数
```go
func handleYourAPI(ctx *interceptor.Context) interceptor.InterceptorResult {
    // 您的业务逻辑
    ctx.ResponseHelper.SetJSON(map[string]interface{}{
        "message": "Your API response",
    })
    return interceptor.Cancel
}
```

### 2. 在RegisterAPIs函数中注册
```go
func RegisterAPIs(server interface{ Register(string, interceptor.InterceptorFunc) }) {
    // ...existing registrations...
    server.Register("/api/your-endpoint", handleYourAPI)
}
```

### 3. 重新编译和运行
```bash
go build -o fnproxy.exe .
./fnproxy.exe
```

## 路径匹配规则

- **精确匹配**: `/api/users` - 只匹配确切路径
- **前缀匹配**: `/api/users/*` - 匹配所有以/api/users/开头的路径
- **通配符**: `/*` - 匹配所有路径

## 响应格式建议

建议使用统一的响应格式：

```go
type APIResponse struct {
    Code    int         `json:"code"`
    Message string      `json:"message"`
    Data    interface{} `json:"data,omitempty"`
}
```

示例响应：
```json
{
    "code": 200,
    "message": "success",
    "data": {
        "users": [...]
    }
}
```

## 注意事项

1. **返回值重要**: 必须返回 `interceptor.Cancel` 才不会转发请求
2. **路径顺序**: 更具体的路径应该先注册，通配符路径后注册
3. **错误处理**: 建议统一处理错误并返回合适的HTTP状态码
4. **CORS**: 如果需要跨域访问，记得设置CORS头部
5. **认证**: 可以在拦截器中实现认证逻辑

通过这种方式，您可以在代理框架中实现任何自定义API，同时保持代理转发的功能。API和代理转发可以完美共存！
