package cache

import (
	"fmt"
	"fnproxy/internal/interceptors/emby/common"
	"fnproxy/pkg/logger"
	"fnproxy/pkg/utils"
	"sync"
)

type UserCacheManager struct {
	url      string
	username string
	password string
	user     common.UserInfo
	device   common.DeviceInfo
	lock     sync.RWMutex
}

var userCache *UserCacheManager

// InitCacheManager 初始化用户缓存管理器
func InitCacheManager(url string, username, password string) error {
	device := common.DeviceInfo{
		Client:   "Yamby",
		Device:   "HUAWEI-HBP-AL00",
		DeviceId: utils.GenerateDeviceID(username),
		Version:  "1.0.0",
	}

	userCache = &UserCacheManager{
		url:      url,
		username: username,
		password: password,
		device:   device,
	}

	return userCache.RefreshToken()
}

// GetCacheManager 获取用户缓存管理器
func GetCacheManager() *UserCacheManager {
	return userCache
}

// GetAuthHeader 获取认证头
func (cm *UserCacheManager) GetAuthHeader() string {
	cm.lock.RLock()
	defer cm.lock.RUnlock()
	return fmt.Sprintf(`Emby UserId="%s", Client="%s", Device="%s", DeviceId="%s", Version="%s"`,
		cm.user.UserId, cm.device.Client, cm.device.Device, cm.device.DeviceId, cm.device.Version)
}

// GetToken 获取用户Token
func (cm *UserCacheManager) GetToken() string {
	cm.lock.RLock()
	defer cm.lock.RUnlock()
	return cm.user.Token
}

// RefreshToken 刷新用户Token
func (cm *UserCacheManager) RefreshToken() error {
	cm.lock.Lock()
	defer cm.lock.Unlock()

	api := common.NewFnApi(cm.url)
	resp, err := api.Login(cm.username, cm.password, cm.device)
	if err != nil {
		logger.Errorf("Failed to login: %v", err)
		return err
	}

	cm.user = *resp

	logger.Infof("Successfully refreshed token for user %s", cm.username)
	return nil
}
