package common

const (
	EmbyAuthPath        = "/emby/Users/AuthenticateByName"
	SystemInfoPath      = "/emby/System/Info"
	PageViewPath        = "/emby/Users/:userid/Items"
	ArtPicturePath      = "/emby/Items/:itemid/Images/Backdrop/:index" // + ?maxWidth=xxx&maxHeight=xxx&tag=xxx
	ItemSimilarPath     = "/emby/Items/:itemid/Similar"                // 相似影视
	AlbumsSimilarPath   = "/emby/Albums/:itemid/Similar"
	ArtistsSimilarPath  = "/emby/Artists/:itemid/Similar"
	MoviesSimilarPath   = "/emby/Movies/:itemid/Similar"
	ShowsSimilarPath    = "/emby/Shows/:itemid/Similar"
	TrailersSimilarPath = "/emby/Trailers/:itemid/Similar"
)

const (
	EmbyAuthHeader  = "X-Emby-Authorization"
	EmbyTokenHeader = "X-Emby-Token"
)

// SimilarResp 相似影视响应体
type SimilarResp struct {
	Items            []any `json:"Items"`
	TotalRecordCount int   `json:"TotalRecordCount"`
	StartIndex       int   `json:"StartIndex"`
}

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
