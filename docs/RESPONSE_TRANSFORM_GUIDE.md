# 响应转换功能使用指南

## 🔄 响应转换概述

我已经为您实现了完整的响应转换功能，可以在代理转发后拦截和修改后端服务器的响应。这个功能特别适用于：

- 统一响应格式
- 过滤敏感数据
- 添加缓存头
- 错误响应包装
- 数据格式转换

## 🛠️ 实现原理

### 1. 响应拦截机制
```go
// 启用响应拦截
ctx.EnableResponseIntercept()

// 在拦截器中检查是否有响应数据
if responseBody := ctx.ResponseHelper.GetResponseBody(); responseBody != nil {
    // 处理响应数据
}
```

### 2. 响应处理流程
1. 请求转发到后端服务器
2. 拦截后端响应
3. 在拦截器中处理响应数据
4. 修改响应内容
5. 返回给客户端

## 📋 响应转换示例

### 1. JSON格式统一包装
```go
server.Register("/api/transform/*", func(ctx *interceptor.Context) interceptor.InterceptorResult {
    if responseBody := ctx.ResponseHelper.GetResponseBody(); responseBody != nil {
        // 获取原始响应
        originalResponse := string(responseBody)
        
        // 包装为统一格式
        transformedResponse := map[string]interface{}{
            "success": true,
            "timestamp": time.Now().Unix(),
            "data": json.RawMessage(originalResponse),
            "meta": map[string]interface{}{
                "processed_by": "fnproxy",
                "version": "1.0",
            },
        }
        
        // 写回转换后的响应
        if newBody, err := json.Marshal(transformedResponse); err == nil {
            ctx.ResponseHelper.SetResponseBody(newBody)
            ctx.ResponseHelper.SetHeader("Content-Type", "application/json")
        }
    }
    return interceptor.Continue
})
```

**效果示例：**
```json
// 原始响应
{"users": [...]}

// 转换后响应
{
    "success": true,
    "timestamp": 1699123456,
    "data": {"users": [...]},
    "meta": {
        "processed_by": "fnproxy",
        "version": "1.0"
    }
}
```

### 2. 敏感数据过滤
```go
server.Register("/api/secure/*", func(ctx *interceptor.Context) interceptor.InterceptorResult {
    if responseBody := ctx.ResponseHelper.GetResponseBody(); responseBody != nil {
        var responseData map[string]interface{}
        if err := json.Unmarshal(responseBody, &responseData); err == nil {
            // 移除敏感字段
            delete(responseData, "password")
            delete(responseData, "secret")
            delete(responseData, "private_key")
            delete(responseData, "token")
            
            // 标记已过滤
            responseData["_filtered"] = true
            
            if filteredBody, err := json.Marshal(responseData); err == nil {
                ctx.ResponseHelper.SetResponseBody(filteredBody)
                ctx.ResponseHelper.SetHeader("X-Data-Filtered", "true")
            }
        }
    }
    return interceptor.Continue
})
```

### 3. 错误响应转换
```go
server.Register("/api/errors/*", func(ctx *interceptor.Context) interceptor.InterceptorResult {
    if responseBody := ctx.ResponseHelper.GetResponseBody(); responseBody != nil {
        status := ctx.ResponseHelper.GetResponseStatus()
        
        // 只处理错误响应
        if status >= 400 {
            errorResponse := map[string]interface{}{
                "error": true,
                "code": status,
                "message": "请求处理失败",
                "timestamp": time.Now().Unix(),
                "original_error": string(responseBody),
            }
            
            if errorBody, err := json.Marshal(errorResponse); err == nil {
                ctx.ResponseHelper.SetResponseBody(errorBody)
            }
        }
    }
    return interceptor.Continue
})
```

### 4. 缓存头和响应修改
```go
server.Register("/api/cache/*", func(ctx *interceptor.Context) interceptor.InterceptorResult {
    if responseBody := ctx.ResponseHelper.GetResponseBody(); responseBody != nil {
        // 添加缓存控制头
        ctx.ResponseHelper.SetHeader("Cache-Control", "public, max-age=300")
        ctx.ResponseHelper.SetHeader("ETag", fmt.Sprintf("\"%x\"", md5.Sum(responseBody)))
        
        // 在响应中添加缓存信息
        var responseData map[string]interface{}
        if err := json.Unmarshal(responseBody, &responseData); err == nil {
            responseData["_cache_info"] = map[string]interface{}{
                "cached_at": time.Now().Unix(),
                "ttl": 300,
            }
            
            if modifiedBody, err := json.Marshal(responseData); err == nil {
                ctx.ResponseHelper.SetResponseBody(modifiedBody)
            }
        }
    }
    return interceptor.Continue
})
```

## 🔧 响应处理API

### 获取响应数据
```go
// 获取响应体
responseBody := ctx.ResponseHelper.GetResponseBody()

// 获取响应状态码
status := ctx.ResponseHelper.GetResponseStatus()
```

### 修改响应数据
```go
// 设置新的响应体
ctx.ResponseHelper.SetResponseBody(newBody)

// 修改状态码
ctx.ResponseHelper.SetResponseStatus(200)

// 添加响应头
ctx.ResponseHelper.SetHeader("X-Custom-Header", "value")
```

## 🎯 使用场景

### 1. API网关功能
- 统一响应格式
- 添加API版本信息
- 注入元数据

### 2. 安全增强
- 过滤敏感字段
- 数据脱敏
- 访问日志

### 3. 性能优化
- 添加缓存头
- 响应压缩
- 内容优化

### 4. 错误处理
- 友好错误信息
- 错误码转换
- 调试信息注入

## 📊 测试响应转换

### 1. 测试JSON格式转换
```bash
# 访问转换端点
curl http://localhost:2345/api/transform/users

# 预期看到包装后的响应格式
```

### 2. 测试敏感数据过滤
```bash
# 访问安全端点
curl http://localhost:2345/api/secure/profile

# 预期看到敏感字段被移除
```

### 3. 测试错误响应转换
```bash
# 触发错误响应
curl http://localhost:2345/api/errors/nonexistent

# 预期看到统一的错误格式
```

## ⚠️ 注意事项

1. **性能考虑**: 响应转换会增加内存使用和处理时间
2. **内容类型**: 主要适用于JSON响应，二进制数据需特殊处理
3. **大文件处理**: 大响应体可能影响性能，建议设置大小限制
4. **错误处理**: 转换过程中的错误需要妥善处理

## 🚀 高级用法

### 条件响应转换
```go
server.Register("/api/conditional/*", func(ctx *interceptor.Context) interceptor.InterceptorResult {
    if responseBody := ctx.ResponseHelper.GetResponseBody(); responseBody != nil {
        // 只对特定用户代理进行转换
        userAgent := ctx.Headers.Get("User-Agent")
        if strings.Contains(userAgent, "mobile") {
            // 移动端特殊处理
            // ...转换逻辑
        }
    }
    return interceptor.Continue
})
```

### 响应流式处理
```go
// 对于大响应，可以考虑流式处理
// 避免将整个响应加载到内存中
```

现在您的代理服务器具备了完整的响应转换能力！可以灵活地处理和修改后端服务器的响应。
