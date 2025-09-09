package emby

import (
	"fmt"
	"fnproxy/pkg/logger"
	"fnproxy/pkg/utils"
)

// todo: 缓存要过期支持，否则无法长期使用
type UserCacheManager struct {
	username string
	password string
	user     *UserInfo
	device   *DeviceInfo
}

var userCache *UserCacheManager

// InitCacheManager 初始化用户缓存管理器
func InitCacheManager(url string, username, password string) error {
	device := DeviceInfo{
		Client:   "Yamby",
		Device:   "HUAWEI-HBP-AL00",
		DeviceId: utils.GenerateDeviceID(username),
		Version:  "1.0.0",
	}

	api := NewFnApi(url)
	resp, err := api.Login(username, password, device)
	if err != nil {
		logger.Errorf("Failed to login: %v", err)
		return err
	}

	userCache = &UserCacheManager{
		username: username,
		password: password,
		device:   &device,
		user:     resp,
	}

	return nil
}

// GetCacheManager 获取用户缓存管理器
func GetCacheManager() *UserCacheManager {
	return userCache
}

// GetAuthHeader 获取认证头
func (cm *UserCacheManager) GetAuthHeader() string {
	return fmt.Sprintf(`Emby UserId="%s", Client="%s", Device="%s", DeviceId="%s", Version="%s"`,
		cm.user.UserId, cm.device.Client, cm.device.Device, cm.device.DeviceId, cm.device.Version)
}

// GetToken 获取用户Token
func (cm *UserCacheManager) GetToken() string {
	return cm.user.Token
}
