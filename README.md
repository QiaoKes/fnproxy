# fnproxy

fnproxy 是一个强大的反向代理框架，基于 Go 语言开发。它提供了灵活的请求和响应拦截机制，支持透明转发、请求/响应修改、API 拦截和转发等功能。通过内置的工具类，可以轻松实现复杂的代理逻辑。

使用该框架，我们可以将 Jellyfin 客户端的请求转换为兼容飞牛影视服务器的 类Jellyfin API 格式，实现无缝对接。目前，Jellyfin 客户端 Yamby 可以正常使用该服务。

## 功能特性

### 1. 透传与请求/响应修改

fnproxy 支持透明转发请求，同时允许在请求前和响应后进行拦截和修改：

- **请求修改**：可以任意改写请求的 header、body、查询参数和路径
- **响应修改**：可以改写响应的 header、body 和状态码
- **拦截与转发**：支持根据路径和方法拦截特定 API，并进行自定义处理

#### 示例：修改请求头

```go
func (i *MyInterceptor) PreRequest(ctx *proxy.Context) proxy.InterceptorResult {
    // 添加自定义头
    ctx.Headers.Set("X-Custom-Header", "value")
    // 修改请求体
    newBody := []byte(`{"modified": true}`)
    ctx.RequestHelper.SetBody(newBody)
    return proxy.Continue
}
```

#### 示例：修改响应

```go
func (i *MyInterceptor) AfterResponse(ctx *proxy.Context) proxy.InterceptorResult {
    // 修改响应头
    ctx.ResponseHelper.SetHeader("Content-Type", "application/json")
    // 设置自定义响应
    ctx.ResponseHelper.SetJSON(map[string]string{"status": "modified"})
    return proxy.Continue
}
```

### 2. 自定义 API 支持

框架支持注册自定义 API 端点，可以完全接管请求处理：

```go
server.Register(http.MethodGet, "/custom/api", func(ctx *proxy.Context) proxy.InterceptorResult {
    ctx.ResponseHelper.SetJSONWithStatus(200, map[string]interface{}{
        "message": "Custom API response",
        "data": []string{"item1", "item2"},
    })
    return proxy.Cancel // 取消转发，直接返回自定义响应
})
```

### 3. 拦截器注册流程

fnproxy 支持多种方式注册拦截器，实现灵活的请求处理：

#### 全局拦截器

全局拦截器对所有请求生效，可以在请求前或响应后进行全局处理：

```go
// 注册全局拦截器（请求前和响应后）
server.RegisterGlobalBoth(
    func(ctx *proxy.Context) proxy.InterceptorResult {
        // 全局请求前处理，例如添加公共头、日志记录等
        ctx.Headers.Set("X-Global-Header", "value")
        return proxy.Continue
    },
    func(ctx *proxy.Context) proxy.InterceptorResult {
        // 全局响应后处理，例如添加响应头、错误处理等
        ctx.ResponseHelper.SetHeader("X-Powered-By", "fnproxy")
        return proxy.Continue
    },
)

// 只注册全局请求前处理器
server.RegisterGlobalInterceptor(&proxy.Interceptor{
    PreRequest: func(ctx *proxy.Context) proxy.InterceptorResult {
        // 全局请求前处理
        return proxy.Continue
    },
})

// 只注册全局响应后处理器
server.RegisterGlobalInterceptor(&proxy.Interceptor{
    AfterResponse: func(ctx *proxy.Context) proxy.InterceptorResult {
        // 全局响应后处理
        return proxy.Continue
    },
})
```

#### 路径拦截器

路径拦截器针对特定路径和HTTP方法进行拦截，支持精确匹配和参数匹配：

```go
// 注册完整的路径拦截器
server.Register(http.MethodGet, "/api/items", &proxy.Interceptor{
    PreRequest: func(ctx *proxy.Context) proxy.InterceptorResult {
        // 请求前处理
        return proxy.Continue
    },
    AfterResponse: func(ctx *proxy.Context) proxy.InterceptorResult {
        // 响应后处理
        return proxy.Continue
    },
})

// 只注册请求前处理器
server.RegisterPreRequest(http.MethodPost, "/api/auth", func(ctx *proxy.Context) proxy.InterceptorResult {
    // 处理认证逻辑
    return proxy.Continue
})

// 只注册响应后处理器
server.RegisterAfterResponse(http.MethodGet, "/api/data", func(ctx *proxy.Context) proxy.InterceptorResult {
    // 处理响应数据
    return proxy.Continue
})

// 同时注册请求前和响应后处理器
server.RegisterBoth(http.MethodPut, "/api/update", 
    func(ctx *proxy.Context) proxy.InterceptorResult {
        // 请求前处理
        return proxy.Continue
    },
    func(ctx *proxy.Context) proxy.InterceptorResult {
        // 响应后处理
        return proxy.Continue
    },
)
```

#### 复杂路径匹配

fnproxy 支持多级路径参数匹配，使用 `:param` 语法匹配路径中的动态部分：

```go
// 单级参数匹配
server.RegisterPreRequest(http.MethodGet, "/api/users/:id", func(ctx *proxy.Context) proxy.InterceptorResult {
    // 处理 /api/users/123 等路径
    return proxy.Continue
})

// 多级参数匹配
server.RegisterPreRequest(http.MethodGet, "/emby/Items/:itemid/Images/Backdrop/:index", func(ctx *proxy.Context) proxy.InterceptorResult {
    // 处理 /emby/Items/123/Images/Backdrop/0 等路径
    return proxy.Continue
})

// 通配方法匹配（*）
server.RegisterAnyMethodPreRequest("/api/logs", func(ctx *proxy.Context) proxy.InterceptorResult {
    // 对所有 HTTP 方法的 /api/logs 路径生效
    return proxy.Continue
})
```

路径匹配支持以下特性：

- **精确匹配**：完全匹配路径字符串
- **参数匹配**：使用 `:param` 匹配单个路径段，支持多级路径
- **方法通配**：使用 `*` 匹配所有 HTTP 方法
- **智能选择**：当多个模式匹配时，选择最具体的模式（静态段优先于参数段）

#### HeaderHelper - 请求头操作

```go
// 设置请求头
ctx.Headers.Set("Authorization", "Bearer token")
// 添加请求头
ctx.Headers.Add("X-Forwarded-For", "127.0.0.1")
// 删除请求头
ctx.Headers.Del("X-Debug")
// 获取请求头
auth := ctx.Headers.Get("Authorization")
// 检查头是否存在
if ctx.Headers.Contains("User-Agent") {
    // 处理逻辑
}
// 过滤头，只保留指定头
ctx.Headers.Filter("Authorization", "Content-Type")
```

#### RequestHelper - 请求操作

```go
// 设置请求体
ctx.RequestHelper.SetBody([]byte(`{"key": "value"}`))
// 获取请求体
body, err := ctx.RequestHelper.GetBody()
// 设置查询参数
ctx.RequestHelper.SetQuery("page", "1")
// 添加查询参数
ctx.RequestHelper.AddQuery("sort", "asc")
// 删除查询参数
ctx.RequestHelper.DelQuery("debug")
// 设置请求路径
ctx.RequestHelper.SetPath("/modified/path")
```

#### ResponseHelper - 响应操作

```go
// 设置响应头
ctx.ResponseHelper.SetHeader("Cache-Control", "no-cache")
// 添加响应头
ctx.ResponseHelper.AddHeader("X-Powered-By", "fnproxy")
// 删除响应头
ctx.ResponseHelper.DelHeader("Server")
// 设置状态码
ctx.ResponseHelper.SetStatus(404)
// 设置 JSON 响应
ctx.ResponseHelper.SetJSON(map[string]string{"error": "Not found"})
// 设置带状态码的 JSON 响应
ctx.ResponseHelper.SetJSONWithStatus(500, map[string]string{"error": "Internal error"})
```

## 配置示例

```yaml
server:
  listen: "0.0.0.0:2345"
target:
  host: "your-target-server.com"
  port: 8005
  https: false
user:
  username: "your-username"
  password: "your-password"
log:
  level: 1  # 0=debug, 1=info, 2=warn, 3=error
```

### 配置说明

- **server.listen**: 代理服务器监听地址和端口
- **target.host**: 目标服务器主机地址
- **target.port**: 目标服务器端口
- **target.https**: 是否使用HTTPS连接目标服务器
- **user.username**: 飞牛影视用户名
- **user.password**: 飞牛影视密码
- **log.level**: 日志级别 (0=debug, 1=info, 2=warn, 3=error)

当Jellyfin客户端发起请求时，如果请求中缺少认证token，fnproxy会使用配置的用户名和密码自动向目标服务器进行认证，获取有效的token并添加到请求中，确保请求能够正常转发到目标服务器。

```bash
# 编译
go build -o fnproxy main.go

# 运行
./fnproxy
```

代理服务将在 `http://localhost:2345` 启动，所有来自 Jellyfin 客户端的请求将被转换为飞牛影视兼容的格式，并转发到配置的目标服务器。

## Docker Compose 部署

你可以使用 Docker Compose 来快速部署 fnproxy 服务。项目中已提供 `docker-compose.yml` 文件作为部署模板。

### 部署步骤

1. 复制项目中的 [docker-compose.yml](docker-compose.yml) 文件
2. 根据需要修改环境变量配置
3. 运行以下命令启动服务：

```bash
docker-compose up -d
```

4. 服务将在 `http://localhost:2345` 启动

### 环境变量说明

- **SERVER_LISTEN**: 代理服务器监听地址 (默认: 0.0.0.0:2345)
- **TARGET_HOST**: 目标服务器主机地址
- **TARGET_PORT**: 目标服务器端口 (默认: 8005)
- **TARGET_HTTPS**: 是否使用HTTPS连接 (默认: false)
- **USER_USERNAME**: Jellyfin用户名，用于自动获取token
- **USER_PASSWORD**: Jellyfin密码，与用户名配合使用
- **LOG_LEVEL**: 日志级别 (0=debug, 1=info, 2=warn, 3=error)

## 架构说明

- **核心组件**：基于 Gin 框架的 HTTP 服务器
- **拦截器模式**：支持请求前和响应后拦截
- **插件化设计**：通过注册拦截器实现功能扩展
- **配置驱动**：支持 YAML 配置和环境变量
- **日志系统**：集成 Zap 日志框架，支持多级别日志

## 许可证

本项目采用 [MIT 许可证](LICENSE)。
