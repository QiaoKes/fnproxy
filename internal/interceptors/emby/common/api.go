package common

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"fnproxy/pkg/logger"
	"net/http"
	"strings"
	"time"

	"go.uber.org/zap"
)

// FnApi 飞牛api - emby
type FnApi struct {
	url    string
	client *http.Client
}

func NewFnApi(url string) *FnApi {
	return &FnApi{
		url:    strings.TrimRight(url, "/"),
		client: &http.Client{Timeout: 10 * time.Second},
	}
}

// Login 会向指定 Emby 服务的 AuthenticateByName 接口发起请求
func (api *FnApi) Login(username, password string, device DeviceInfo) (*UserInfo, error) {
	url := fmt.Sprintf("%s%s", api.url, EmbyAuthPath)

	reqBody := LoginReq{
		Username: username,
		Pw:       password,
	}
	bodyBytes, _ := json.Marshal(reqBody)

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

	resp, err := api.client.Do(req) // 复用 client
	if err != nil {
		logger.Error("auth request failed", zap.Error(err))
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		logger.Warn("auth request got non-200", zap.Int("status", resp.StatusCode))
		return nil, errors.New("auth request got non-200")
	}

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

// GetEpisodes 请求某个剧集的所有分集信息
func (api *FnApi) GetEpisodes(seriesId, seasonId, userId, auth, token string) (*EpisodesResp, error) {
	url := fmt.Sprintf("%s/emby/Shows/%s/Episodes?UserId=%s&SeasonId=%s",
		api.url, seriesId, userId, seasonId)

	req, _ := http.NewRequest("GET", url, nil)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set(EmbyTokenHeader, token) // "X-Emby-Token"
	req.Header.Set(EmbyAuthHeader, auth)   // "X-Emby-Authorization"

	resp, err := api.client.Do(req) // 复用 client
	if err != nil {
		logger.Error("episodes request failed", zap.Error(err))
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		logger.Warn("episodes request got non-200", zap.Int("status", resp.StatusCode))
		return nil, fmt.Errorf("episodes request got non-200, code=%d", resp.StatusCode)
	}

	episodesResp := &EpisodesResp{}
	if err := json.NewDecoder(resp.Body).Decode(episodesResp); err != nil {
		logger.Errorf("decode episodes resp failed:%s", err)
		return nil, err
	}

	return episodesResp, nil
}
