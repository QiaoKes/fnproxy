package handler

const (
	ItemSimilarPath     = "/emby/Items/:itemid/Similar" // 相似影视
	AlbumsSimilarPath   = "/emby/Albums/:itemid/Similar"
	ArtistsSimilarPath  = "/emby/Artists/:itemid/Similar"
	MoviesSimilarPath   = "/emby/Movies/:itemid/Similar"
	ShowsSimilarPath    = "/emby/Shows/:itemid/Similar"
	TrailersSimilarPath = "/emby/Trailers/:itemid/Similar"
)

// SimilarResp 相似影视响应体
type SimilarResp struct {
	Items            []any `json:"Items"`
	TotalRecordCount int   `json:"TotalRecordCount"`
	StartIndex       int   `json:"StartIndex"`
}
