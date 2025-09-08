# fnproxy - API透传代理服务器

## 功能特性

1. **透明代理转发**: 监听 0.0.0.0:2345，将请求转发到 10.0.0.115:8005
2. **路由拦截系统**: 支持自定义拦截器，可选择取消或继续转发
3. **透明转发**: 未拦截的请求直接透明转发
4. **插件化注册**: 提供便捷的拦截器注册函数
5. **丰富的工具集**: 头部操作、请求/响应改写等

## 项目结构

```
fnproxy/
├── main.go              # 主程序入口
├── config.yaml          # 配置文件
├── config/
│   └── config.go        # 配置管理
├── interceptor/
│   └── interceptor.go   # 拦截器系统
└── proxy/
    └── server.go        # 代理服务器
```

## 配置文件 (config.yaml)

```yaml
server:
  listen: "0.0.0.0:2345"     # 监听地址和端口
  
target:
  host: "10.0.0.115"         # 目标服务器地址
  port: 8005                 # 目标服务器端口
  
log:
  level: "info"              # 日志级别: debug, info, warn, error
  
timeout:
  read: 30s                  # 读超时
  write: 30s                 # 写超时
  idle: 120s                 # 空闲超时
```

## 使用方法

### 1. 启动服务器

```bash
go run main.go
```

### 2. 注册拦截器

```go
// 注册拦截器的基本语法
server.Register("/path", func(ctx *interceptor.Context) interceptor.InterceptorResult {
    // 拦截逻辑
    return interceptor.Continue  // 或 common.Cancel
})
```

## 拦截器示例

### 1. 认证拦截器
```go
server.Register("/emby/auth", func(ctx *interceptor.Context) interceptor.InterceptorResult {
    token := ctx.Headers.Get("X-Emby-Token")
    if token == "" {
        ctx.ResponseHelper.SetJSON(map[string]interface{}{
            "error": "Missing authentication token",
            "code":  401,
        })
        ctx.ResponseHelper.SetStatus(http.StatusUnauthorized)
        return interceptor.Cancel
    }
    return interceptor.Continue
})
```

### 2. 版本转换拦截器
```go
server.Register("/api/v1/*", func(ctx *interceptor.Context) interceptor.InterceptorResult {
    // 将v1版本转换为v2
    newPath := "/api/v2" + ctx.Path[7:]
    ctx.Request.URL.Path = newPath
    ctx.Headers.Set("X-API-Version", "v2")
    return interceptor.Continue
})
```

### 3. CORS拦截器
```go
server.Register("/api/*", func(ctx *interceptor.Context) interceptor.InterceptorResult {
    ctx.ResponseHelper.SetHeader("Access-Control-Allow-Origin", "*")
    ctx.ResponseHelper.SetHeader("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
    
    if ctx.Method == "OPTIONS" {
        ctx.ResponseHelper.SetStatus(http.StatusOK)
        return interceptor.Cancel
    }
    return interceptor.Continue
})
```

### 4. 健康检查拦截器
```go
server.Register("/health", func(ctx *interceptor.Context) interceptor.InterceptorResult {
    ctx.ResponseHelper.SetJSON(map[string]interface{}{
        "status": "ok",
        "proxy": "running",
    })
    return interceptor.Cancel
})
```

## 拦截器工具集

### Header操作
```go
ctx.Headers.Set("key", "value")     // 设置header
ctx.Headers.Add("key", "value")     // 添加header
ctx.Headers.Del("key")              // 删除header
ctx.Headers.Get("key")              // 获取header
ctx.Headers.Filter("key1", "key2")  // 过滤header，只保留指定的
```

### 请求操作
```go
ctx.RequestHelper.SetQuery("key", "value")  // 设置查询参数
ctx.RequestHelper.SetBody([]byte("data"))   // 设置请求体
```

### 响应操作
```go
ctx.ResponseHelper.SetHeader("key", "value")  // 设置响应头
ctx.ResponseHelper.SetStatus(200)             // 设置状态码
ctx.ResponseHelper.SetJSON(data)              // 设置JSON响应
```

## 路径匹配规则

- **精确匹配**: `/exact/path`
- **前缀匹配**: `/api/*` (匹配所有以/api/开头的路径)
- **通配符**: `/*` (匹配所有路径)

## 编译和运行

```bash
# 安装依赖
go mod tidy

# 编译
go build -o fnproxy .

# 运行
./fnproxy
```

## 架构优势

1. **低耦合**: 模块化设计，各组件职责清晰
2. **高扩展性**: 插件化的拦截器系统
3. **成熟框架**: 基于Gin框架，性能优异
4. **代码简洁**: 清晰的API设计，易于使用和维护

## 技术栈

- **框架**: Gin (HTTP框架)
- **配置**: Viper (配置管理)
- **日志**: Zap (高性能日志)
- **代理**: httputil.ReverseProxy (标准库反向代理)
