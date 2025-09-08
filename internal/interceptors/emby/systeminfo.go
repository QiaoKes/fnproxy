package emby

import (
	"encoding/json"
	"fnproxy/pkg/compress"
	"fnproxy/pkg/logger"
	"fnproxy/pkg/proxy"
	"strings"

	"go.uber.org/zap"
)

// SystemInfoInterceptor Emby系统信息拦截器
type SystemInfoInterceptor struct {
}

// NewSystemInfoInterceptor 创建新的Emby系统信息拦截器
func NewSystemInfoInterceptor() *SystemInfoInterceptor {
	return &SystemInfoInterceptor{}
}

// OriginalSystemInfo 原始系统信息结构
type OriginalSystemInfo struct {
	LocalAddress           string `json:"LocalAddress"`
	ServerName             string `json:"ServerName"`
	Version                string `json:"Version"`
	ProductName            string `json:"ProductName"`
	OperatingSystem        string `json:"OperatingSystem"`
	Id                     string `json:"Id"`
	StartupWizardCompleted bool   `json:"StartupWizardCompleted"`
}

// TransformedSystemInfo 转换后的系统信息结构
type TransformedSystemInfo struct {
	OperatingSystemDisplayName string        `json:"OperatingSystemDisplayName"`
	HasPendingRestart          bool          `json:"HasPendingRestart"`
	IsShuttingDown             bool          `json:"IsShuttingDown"`
	SupportsLibraryMonitor     bool          `json:"SupportsLibraryMonitor"`
	WebSocketPortNumber        int           `json:"WebSocketPortNumber"`
	CompletedInstallations     []interface{} `json:"CompletedInstallations"`
	CanSelfRestart             bool          `json:"CanSelfRestart"`
	CanLaunchWebBrowser        bool          `json:"CanLaunchWebBrowser"`
	ProgramDataPath            string        `json:"ProgramDataPath"`
	WebPath                    string        `json:"WebPath"`
	ItemsByNamePath            string        `json:"ItemsByNamePath"`
	CachePath                  string        `json:"CachePath"`
	LogPath                    string        `json:"LogPath"`
	InternalMetadataPath       string        `json:"InternalMetadataPath"`
	TranscodingTempPath        string        `json:"TranscodingTempPath"`
	HasUpdateAvailable         bool          `json:"HasUpdateAvailable"`
	EncoderLocation            string        `json:"EncoderLocation"`
	SystemArchitecture         string        `json:"SystemArchitecture"`
	LocalAddress               string        `json:"LocalAddress"`
	ServerName                 string        `json:"ServerName"`
	Version                    string        `json:"Version"`
	OperatingSystem            string        `json:"OperatingSystem"`
	Id                         string        `json:"Id"`
}

// SystemInfoIntercept 拦截Emby系统信息返回值
func (si *SystemInfoInterceptor) SystemInfoIntercept(ctx *proxy.Context) proxy.InterceptorResult {
	// 记录请求信息
	logger.Info("Intercepting Emby system info response",
		zap.String("path", ctx.Path),
		zap.String("method", ctx.Method))

	// 读取原始响应体
	body := ctx.ResponseHelper.GetResponseBody()
	if body == nil || len(body) == 0 {
		logger.Error("Failed to read response body")
		return proxy.Continue
	}

	plain, enc, _ := compress.Decode(body, ctx.Response.Header().Get("Content-Encoding"))

	// 用 rawBody 做 json.Unmarshal
	var originalInfo OriginalSystemInfo
	if err := json.Unmarshal(plain, &originalInfo); err != nil {
		logger.Error("Failed to parse original system info", zap.Error(err))
		return proxy.Continue
	}

	// 转换为新的系统信息结构
	transformedInfo := si.transformSystemInfo(&originalInfo)

	// 序列化转换后的数据
	transformedBody, err := json.Marshal(transformedInfo)
	if err != nil {
		logger.Error("Failed to marshal transformed system info", zap.Error(err))
		return proxy.Continue
	}

	// 修改响应体
	_ = ctx.ResponseHelper.SetJSONBodyWithEncoding(transformedBody, enc)

	logger.Info("Successfully transformed Emby system info response")
	return proxy.Continue
}

// transformSystemInfo 转换系统信息
func (si *SystemInfoInterceptor) transformSystemInfo(original *OriginalSystemInfo) *TransformedSystemInfo {
	return &TransformedSystemInfo{
		OperatingSystemDisplayName: original.OperatingSystem,
		HasPendingRestart:          false,
		IsShuttingDown:             false,
		SupportsLibraryMonitor:     true,
		CompletedInstallations:     []interface{}{},
		CanSelfRestart:             true,
		CanLaunchWebBrowser:        false,
		ProgramDataPath:            "/vol1/@appcenter/Jellyfin/data",
		WebPath:                    "/vol1/@appcenter/Jellyfin/service/web",
		ItemsByNamePath:            "/vol1/@appcenter/Jellyfin/data/metadata",
		CachePath:                  "/vol1/@appcenter/Jellyfin/cache",
		LogPath:                    "/vol1/@appcenter/Jellyfin/logs",
		InternalMetadataPath:       "/vol1/@appcenter/Jellyfin/data/metadata",
		TranscodingTempPath:        "/vol1/@appcenter/Jellyfin/data/transcodes",
		HasUpdateAvailable:         false,
		EncoderLocation:            "NotFound",
		SystemArchitecture:         "X64",
		ServerName:                 "nas",
		Version:                    "10.8.12",
		OperatingSystem:            original.OperatingSystem,
		Id:                         "2ae8b1b9064e40c288b6ca238e71c046",
	}
}

// extractIP 从LocalAddress中提取IP地址
func extractIP(localAddress string) string {
	if localAddress == "" {
		return "10.0.0.115"
	}

	// 如果包含端口，去除端口部分
	if strings.Contains(localAddress, ":") {
		parts := strings.Split(localAddress, ":")
		return parts[0]
	}

	return localAddress
}
