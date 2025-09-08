package emby

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"fnproxy/pkg/logger"
	"go.uber.org/zap"
	"net/http"
	"strings"
	"time"
)

// FnApi 飞牛api - emby
type FnApi struct {
	url string
}

func NewFnApi(url string) *FnApi {
	return &FnApi{
		url: strings.TrimRight(url, "/"),
	}
}

// Login 会向指定 Emby 服务的 AuthenticateByName 接口发起请求
func (api *FnApi) Login(username, password string, device DeviceInfo) (*UserInfo, error) {
	// 拼接 URL
	url := fmt.Sprintf("%s%s", api.url, EmbyAuthPath)

	// 请求体
	reqBody := LoginReq{
		Username: username,
		Pw:       password,
	}
	bodyBytes, _ := json.Marshal(reqBody)

	// 构造请求
	req, _ := http.NewRequest("POST", url, bytes.NewReader(bodyBytes))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Emby-Authorization",
		fmt.Sprintf(`Emby Client="%s", Device="%s", DeviceId="%s", Version="%s"`,
			device.Client,
			device.Device,
			device.DeviceId,
			device.Version,
		),
	)

	// 发送
	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		logger.Error("auth request failed", zap.Error(err))
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		logger.Warn("auth request got non-200", zap.Int("status", resp.StatusCode))
		return nil, errors.New("auth request got non-200")
	}

	// 解析返回
	authResp := &LoginResp{}
	if err := json.NewDecoder(resp.Body).Decode(authResp); err != nil {
		logger.Errorf("decode auth resp failed:%s", err)
		return nil, err
	}

	if authResp.AccessToken == "" || authResp.User.Id == "" {
		logger.Error("auth response missing token or user id")
		return nil, errors.New("auth response missing token or user id")
	}

	return &UserInfo{
		UserId: authResp.User.Id,
		Token:  authResp.AccessToken,
	}, nil
}
