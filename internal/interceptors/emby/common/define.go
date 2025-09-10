package common

const (
	EmbyAuthPath   = "/emby/Users/AuthenticateByName"
	SystemInfoPath = "/emby/System/Info"

	PageViewPath     = "/emby/Users/:userid/Items"
	ArtPicturePath   = "/emby/Items/:itemid/Images/Backdrop/:index" // + ?maxWidth=xxx&maxHeight=xxx&tag=xxx
	VideoSummaryPath = "/emby/Users/:userid/Items/:itemid"          // 影视简介

	ItemSimilarPath     = "/emby/Items/:itemid/Similar" // 相似影视
	AlbumsSimilarPath   = "/emby/Albums/:itemid/Similar"
	ArtistsSimilarPath  = "/emby/Artists/:itemid/Similar"
	MoviesSimilarPath   = "/emby/Movies/:itemid/Similar"
	ShowsSimilarPath    = "/emby/Shows/:itemid/Similar"
	TrailersSimilarPath = "/emby/Trailers/:itemid/Similar"

	UserItemsPath = "/emby/Users/:userid/Items/:itemid" // 用户影视列表
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

type EpisodesResp struct {
	Items []Episode `json:"Items"`
}

type Episode struct {
	Id                      string        `json:"Id"`
	Type                    string        `json:"Type"`
	Name                    string        `json:"Name"`
	SeriesId                string        `json:"SeriesId"`
	SeriesName              string        `json:"SeriesName"`
	SeasonId                string        `json:"SeasonId"`
	ParentIndexNumber       *int          `json:"ParentIndexNumber"` // season number
	IndexNumber             *int          `json:"IndexNumber"`
	PremiereDate            string        `json:"PremiereDate"`
	DateCreated             string        `json:"DateCreated"`
	ServerId                string        `json:"ServerId"`
	LocationType            string        `json:"LocationType"`
	ParentBackdropItemId    string        `json:"ParentBackdropItemId"`
	ParentBackdropImageTags []string      `json:"ParentBackdropImageTags"`
	ParentLogoItemId        string        `json:"ParentLogoItemId"`
	ParentLogoImageTag      string        `json:"ParentLogoImageTag"`
	SeriesPrimaryImageTag   string        `json:"SeriesPrimaryImageTag"`
	MediaSources            []MediaSource `json:"MediaSources"`
	UserData                *UserData     `json:"UserData"`
}

type MediaSource struct {
	Path string `json:"Path"`
}

type UserData struct {
	Played                bool   `json:"Played"`
	PlayCount             int    `json:"PlayCount"`
	PlaybackPositionTicks int64  `json:"PlaybackPositionTicks"`
	IsFavorite            bool   `json:"IsFavorite"`
	Key                   string `json:"Key"`
}

type SeasonResp struct {
	Name                     string            `json:"Name,omitempty"`
	ServerId                 string            `json:"ServerId,omitempty"`
	Id                       string            `json:"Id,omitempty"`
	Etag                     string            `json:"Etag,omitempty"`
	DateCreated              string            `json:"DateCreated,omitempty"`
	CanDelete                bool              `json:"CanDelete"`
	CanDownload              bool              `json:"CanDownload"`
	SortName                 string            `json:"SortName,omitempty"`
	PremiereDate             string            `json:"PremiereDate,omitempty"`
	ExternalUrls             []any             `json:"ExternalUrls"`
	Path                     string            `json:"Path,omitempty"`
	EnableMediaSourceDisplay bool              `json:"EnableMediaSourceDisplay"`
	ChannelId                *string           `json:"ChannelId"`
	Taglines                 []string          `json:"Taglines"`
	Genres                   []string          `json:"Genres"`
	PlayAccess               string            `json:"PlayAccess,omitempty"`
	ProductionYear           int               `json:"ProductionYear,omitempty"`
	IndexNumber              int               `json:"IndexNumber,omitempty"`
	RemoteTrailers           []any             `json:"RemoteTrailers"`
	ProviderIds              map[string]string `json:"ProviderIds,omitempty"`
	IsFolder                 bool              `json:"IsFolder"`
	ParentId                 string            `json:"ParentId,omitempty"`
	Type                     string            `json:"Type"`
	People                   []any             `json:"People"`
	Studios                  []any             `json:"Studios"`
	GenreItems               []any             `json:"GenreItems"`
	ParentLogoItemId         string            `json:"ParentLogoItemId,omitempty"`
	ParentBackdropItemId     string            `json:"ParentBackdropItemId,omitempty"`
	ParentBackdropImageTags  []string          `json:"ParentBackdropImageTags,omitempty"`
	LocalTrailerCount        int               `json:"LocalTrailerCount"`
	UserData                 *SeasonUserData   `json:"UserData,omitempty"`
	RecursiveItemCount       int               `json:"RecursiveItemCount"`
	ChildCount               int               `json:"ChildCount"`
	SeriesName               string            `json:"SeriesName,omitempty"`
	SeriesId                 string            `json:"SeriesId,omitempty"`
	SpecialFeatureCount      int               `json:"SpecialFeatureCount"`
	DisplayPreferencesId     string            `json:"DisplayPreferencesId,omitempty"`
	Tags                     []string          `json:"Tags"`
	PrimaryImageAspectRatio  float64           `json:"PrimaryImageAspectRatio,omitempty"`
	SeriesPrimaryImageTag    string            `json:"SeriesPrimaryImageTag,omitempty"`
	ImageTags                map[string]string `json:"ImageTags,omitempty"`
	BackdropImageTags        []string          `json:"BackdropImageTags"`
	ParentLogoImageTag       string            `json:"ParentLogoImageTag,omitempty"`
	ImageBlurHashes          map[string]any    `json:"ImageBlurHashes,omitempty"`
	SeriesStudio             string            `json:"SeriesStudio,omitempty"`
	LocationType             string            `json:"LocationType,omitempty"`
	LockedFields             []string          `json:"LockedFields"`
	LockData                 bool              `json:"LockData"`
	// 根据需要可继续补充
}

type SeasonUserData struct {
	PlayedPercentage      float64 `json:"PlayedPercentage"`
	UnplayedItemCount     int     `json:"UnplayedItemCount"`
	PlaybackPositionTicks int64   `json:"PlaybackPositionTicks"`
	PlayCount             int     `json:"PlayCount"`
	IsFavorite            bool    `json:"IsFavorite"`
	Played                bool    `json:"Played"`
	Key                   string  `json:"Key"`
}
