package handler

const (
	SimilarPath = "/emby/Items/:itemid/Similar" // 相似影视
)

// SimilarResp 相似影视响应体
type SimilarResp struct {
	Items            []any `json:"Items"`
	TotalRecordCount int   `json:"TotalRecordCount"`
	StartIndex       int   `json:"StartIndex"`
}
