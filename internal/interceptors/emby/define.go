package emby

const (
	EmbyAuthPath   = "/emby/Users/AuthenticateByName"
	SystemInfoPath = "/emby/System/Info"
)

const (
	EmbyAuthHeader  = "X-Emby-Authorization"
	EmbyTokenHeader = "X-Emby-Token"
)

// UserInfo Emby用户信息
type UserInfo struct {
	UserId string // Emby用户ID
	Token  string // Emby用户Token
}

// DeviceInfo Emby设备信息
type DeviceInfo struct {
	Client   string // Emby客户端名称
	Device   string // Emby设备名称
	DeviceId string // Emby设备ID
	Version  string // Emby客户端版本
}

// LoginReq 登录请求体
type LoginReq struct {
	Username string `json:"Username"`
	Pw       string `json:"Pw"`
}

type User struct {
	Id   string `json:"Id"`   // Emby用户ID
	Name string `json:"Name"` // Emby用户名
}

// LoginResp 登录响应体
type LoginResp struct {
	User        User   `json:"User"`        // Emby用户信息
	AccessToken string `json:"AccessToken"` // Emby用户Token
}
