# Emby 认证拦截器使用指南

## 🎯 功能概述

专门为 Emby 认证请求改写创建的模块化拦截器，可以将原始的 Emby 客户端请求转换为目标格式。

## 📋 请求转换说明

### 原始请求格式：
```http
POST /emby/Users/AuthenticateByName
X-Emby-Authorization: Emby UserId=45d87baa-95f8-4909-9f62-6492346e862a,Client=Yamby,Device=HUAWEI-HBP-AL00,DeviceId=65c6fd36-48f9-4442-abb5-457ee0bebfe0,Version=1.6.1.3
User-Agent: Yamby/1.6.1.3(Android
Content-Type: application/json; charset=UTF-8

{"Username":"dqc1009","Password":"1037344122+","Pw":"1037344122+"}
```

### 目标请求格式：
```http
POST /emby/Users/AuthenticateByName
X-Emby-Authorization: Emby Client="%E7%BD%91%E6%98%93%E7%88%86%E7%B1%B3%E8%8A%B1 Android", Device="HBP-AL00", DeviceId="2c03a65388c4e4e1", Version="2.6.2"
User-Agent: okhttp/5.1.0
Content-Type: application/json

{"Username":"test","Pw":"1037344122+"}
```

## 🔧 模块化设计

### 1. **独立的拦截器类**
- `EmbyAuthInterceptor`: 专门处理 Emby 认证请求
- 可配置的转换参数
- 独立的注册方法

### 2. **配置结构体**
```go
type EmbyAuthConfig struct {
    TargetClient    string // 目标客户端名称
    TargetDevice    string // 目标设备型号
    TargetDeviceId  string // 目标设备ID
    TargetVersion   string // 目标版本号
    TargetUserAgent string // 目标User-Agent
    DefaultUsername string // 默认用户名（可选）
}
```

## 🚀 使用方法

### 1. **基本使用**
```go
// 创建默认配置
embyConfig := GetDefaultEmbyAuthConfig()

// 创建拦截器实例
embyInterceptor := NewEmbyAuthInterceptor(logger, embyConfig)

// 注册到服务器
embyInterceptor.Register(server)
```

### 2. **自定义配置**
```go
// 自定义配置
customConfig := EmbyAuthConfig{
    TargetClient:    "自定义客户端",
    TargetDevice:    "Custom-Device",
    TargetDeviceId:  "custom-device-id",
    TargetVersion:   "3.0.0",
    TargetUserAgent: "CustomApp/1.0",
    DefaultUsername: "admin", // 强制使用指定用户名
}

embyInterceptor := NewEmbyAuthInterceptor(logger, customConfig)
embyInterceptor.Register(server)
```

### 3. **运行时更新配置**
```go
// 运行时修改配置
newConfig := EmbyAuthConfig{
    TargetClient: "新的客户端名称",
    // ... 其他配置
}
embyInterceptor.SetConfig(newConfig)
```

## 📊 转换处理流程

1. **拦截匹配请求**: 只处理 `POST /emby/Users/AuthenticateByName`
2. **改写请求头**:
   - `X-Emby-Authorization`: 使用新的客户端信息
   - `User-Agent`: 替换为目标 User-Agent
   - `Content-Type`: 移除 charset 参数
3. **改写请求体**:
   - 保持或替换用户名
   - 将 `Password` 字段转换为 `Pw` 字段
   - 移除多余字段
4. **更新 Content-Length**: 自动计算新的内容长度

## 🎛️ 配置选项说明

| 配置项 | 默认值 | 说明 |
|--------|---------|------|
| `TargetClient` | `"%E7%BD%91%E6%98%93%E7%88%86%E7%B1%B3%E8%8A%B1 Android"` | 网易爆米花 Android |
| `TargetDevice` | `"HBP-AL00"` | 华为设备型号 |
| `TargetDeviceId` | `"2c03a65388c4e4e1"` | 设备唯一标识 |
| `TargetVersion` | `"2.6.2"` | 客户端版本 |
| `TargetUserAgent` | `"okhttp/5.1.0"` | HTTP客户端标识 |
| `DefaultUsername` | `"test"` | 默认用户名 |

## 📝 日志输出

拦截器会输出详细的日志信息：
- 请求拦截通知
- 原始请求信息
- 转换后的信息
- 错误信息（如有）

## 🔄 扩展性

### 添加新的拦截器模块：

1. **创建新的拦截器文件** (例如: `examples/other_interceptor.go`)
```go
package examples

type OtherInterceptor struct {
    // 配置和状态
}

func NewOtherInterceptor() *OtherInterceptor {
    return &OtherInterceptor{}
}

func (o *OtherInterceptor) Register(server *proxy.Server) {
    server.Register("/your/path", o.handleRequest)
}
```

2. **在 `interceptors.go` 中注册**
```go
func RegisterExampleInterceptors(server *proxy.Server, logger *zap.Logger) {
    // 注册 Emby 认证拦截器
    embyInterceptor := NewEmbyAuthInterceptor(logger, GetDefaultEmbyAuthConfig())
    embyInterceptor.Register(server)
    
    // 注册其他拦截器
    otherInterceptor := NewOtherInterceptor()
    otherInterceptor.Register(server)
    
    // ...
}
```

## 🧪 测试方法

### 1. **启动服务器**
```bash
go run main.go
```

### 2. **发送测试请求**
```bash
curl -X POST http://localhost:2345/emby/Users/AuthenticateByName \
  -H "X-Emby-Authorization: Emby UserId=test,Client=TestClient,Device=TestDevice,DeviceId=test123,Version=1.0" \
  -H "User-Agent: TestAgent/1.0" \
  -H "Content-Type: application/json; charset=UTF-8" \
  -d '{"Username":"testuser","Password":"testpass"}'
```

### 3. **查看日志**
观察控制台输出，确认请求被正确拦截和转换。

## ⚠️ 注意事项

1. **路径匹配**: 只拦截精确路径 `/emby/Users/AuthenticateByName`
2. **方法限制**: 只处理 POST 请求
3. **错误处理**: 如果请求体解析失败，会记录错误但继续转发
4. **配置持久性**: 运行时配置更改不会持久化，重启后恢复默认值

现在您的 Emby 认证拦截器已经完全模块化，可以独立管理和扩展！
