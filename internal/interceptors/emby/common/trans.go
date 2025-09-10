package common

import (
	"fmt"
	"path/filepath"
	"sort"
	"strings"
	"time"
)

// TransformEpisodesToSeason 将 EpisodesResp 转换为 SeasonResp
func TransformEpisodesToSeason(resp *EpisodesResp) *SeasonResp {
	if resp == nil || len(resp.Items) == 0 {
		return &SeasonResp{}
	}

	// 取第一集作为季级元信息来源
	first := resp.Items[0]

	// 计算季号、名称
	seasonNumber := 0
	if first.ParentIndexNumber != nil {
		seasonNumber = *first.ParentIndexNumber
	}
	seasonName := "Season"
	if seasonNumber > 0 {
		seasonName = fmt.Sprintf("第 %d 季", seasonNumber)
	}

	// Path：从第一集的文件路径上溯到季目录（上一级）
	seasonPath := ""
	if len(first.MediaSources) > 0 && first.MediaSources[0].Path != "" {
		// 例：.../Season 1/xxx.mkv -> .../Season 1
		seasonPath = filepath.Dir(first.MediaSources[0].Path)
	}

	// PremiereDate：取最早的
	premiere := pickEarliestTime(resp.Items, func(e Episode) string { return e.PremiereDate })
	premiereStr := formatTimeISO8601(premiere)

	// ProductionYear：从 PremiereDate 取年份
	productionYear := 0
	if !premiere.IsZero() {
		productionYear = premiere.Year()
	}

	// DateCreated：取最早的
	dateCreated := pickEarliestTime(resp.Items, func(e Episode) string { return e.DateCreated }).UTC()
	dateCreatedStr := formatTimeISO8601(dateCreated)

	// 统计 UserData
	total := len(resp.Items)
	unplayed := 0
	totalPlayCount := 0
	var anyKey string
	anyFavorite := false
	for _, it := range resp.Items {
		if it.UserData == nil || !it.UserData.Played {
			unplayed++
		}
		if it.UserData != nil {
			totalPlayCount += it.UserData.PlayCount
			if it.UserData.Key != "" && anyKey == "" {
				anyKey = it.UserData.Key
			}
			if it.UserData.IsFavorite {
				anyFavorite = true
			}
		}
	}
	playedPercentage := 0.0
	if total > 0 {
		playedPercentage = float64(total-unplayed) * 100.0 / float64(total)
	}

	// SortName：像样例一样左侧补零（4 位）
	sortName := "0000"
	if seasonNumber > 0 {
		sortName = fmt.Sprintf("%04d", seasonNumber)
	}

	// Series/Images
	seriesId := first.SeriesId
	seriesName := first.SeriesName

	// Backdrop/Logo/SeriesPrimaryImageTag：从第一集里取父级字段
	parentBackdropItemId := first.ParentBackdropItemId
	parentBackdropImageTags := dedupStrings(flatCollect(resp.Items, func(e Episode) []string { return e.ParentBackdropImageTags }))
	parentLogoItemId := first.ParentLogoItemId
	parentLogoImageTag := first.ParentLogoImageTag
	seriesPrimaryImageTag := first.SeriesPrimaryImageTag

	// LocationType & ServerId（如果每集不同，这里取第一条）
	locationType := first.LocationType
	serverId := first.ServerId

	// Season Id：用 SeasonId（若缺失则回退到 ParentId/SeriesId）
	seasonId := first.SeasonId
	parentId := "" // 接口1中是该季的上级（剧集/Show）Id；接口2没有明确提供 ParentId，通常可用 SeriesId 作为 ParentId。
	if seriesId != "" {
		parentId = seriesId
	}

	season := SeasonResp{
		Name:                     seasonName,
		ServerId:                 serverId,
		Id:                       seasonId,
		DateCreated:              dateCreatedStr,
		CanDelete:                true,
		CanDownload:              false,
		SortName:                 sortName,
		PremiereDate:             premiereStr,
		ExternalUrls:             []any{},
		Path:                     seasonPath,
		EnableMediaSourceDisplay: true,
		ChannelId:                nil,
		Taglines:                 []string{},
		Genres:                   []string{},
		PlayAccess:               "Full",
		ProductionYear:           productionYear,
		IndexNumber:              seasonNumber,
		RemoteTrailers:           []any{},
		ProviderIds:              map[string]string{}, // 无法从接口2推导，留空
		IsFolder:                 true,
		ParentId:                 parentId,
		Type:                     "Season",
		People:                   []any{},
		Studios:                  []any{},
		GenreItems:               []any{},
		ParentLogoItemId:         parentLogoItemId,
		ParentBackdropItemId:     parentBackdropItemId,
		ParentBackdropImageTags:  parentBackdropImageTags,
		LocalTrailerCount:        0,
		UserData: &SeasonUserData{
			PlayedPercentage:      playedPercentage,
			UnplayedItemCount:     unplayed,
			PlaybackPositionTicks: 0,
			PlayCount:             totalPlayCount,
			IsFavorite:            anyFavorite,
			Played:                total > 0 && unplayed == 0,
			Key:                   anyKey,
		},
		RecursiveItemCount:      total,
		ChildCount:              total,
		SeriesName:              seriesName,
		SeriesId:                seriesId,
		SpecialFeatureCount:     0,
		DisplayPreferencesId:    "",
		Tags:                    []string{},
		PrimaryImageAspectRatio: 0,
		SeriesPrimaryImageTag:   seriesPrimaryImageTag,
		ImageTags:               map[string]string{}, // Season 专属海报在接口2不可得，留空
		BackdropImageTags:       []string{},
		ParentLogoImageTag:      parentLogoImageTag,
		ImageBlurHashes:         map[string]any{}, // 不可推导，留空
		SeriesStudio:            "",
		LocationType:            coalesce(locationType, "FileSystem"),
		LockedFields:            []string{},
		LockData:                false,
	}

	return &season
}

// ---------- 小工具 ----------

func coalesce(s, def string) string {
	if strings.TrimSpace(s) == "" {
		return def
	}
	return s
}

func parseTime(s string) (time.Time, bool) {
	if strings.TrimSpace(s) == "" {
		return time.Time{}, false
	}
	// Jellyfin 常见时间格式：2025-01-11T00:00:00.0000000Z / 2025-04-02T13:54:54.5448833Z
	layouts := []string{
		time.RFC3339Nano,
		"2006-01-02T15:04:05Z07:00",           // RFC3339
		"2006-01-02T15:04:05.0000000Z07:00",   // 7 位小数
		"2006-01-02T15:04:05.000000000Z07:00", // 9 位小数
		"2006-01-02T15:04:05.0000000Z",        // 带 Z
		"2006-01-02T15:04:05.000000000Z",
	}
	for _, l := range layouts {
		if t, err := time.Parse(l, s); err == nil {
			return t, true
		}
	}
	return time.Time{}, false
}

func formatTimeISO8601(t time.Time) string {
	if t.IsZero() {
		return ""
	}
	// 统一使用 RFC3339Nano（与样例兼容）
	return t.UTC().Format(time.RFC3339Nano)
}

func pickEarliestTime(items []Episode, getter func(Episode) string) time.Time {
	var times []time.Time
	for _, it := range items {
		if tt, ok := parseTime(getter(it)); ok {
			times = append(times, tt)
		}
	}
	if len(times) == 0 {
		return time.Time{}
	}
	sort.Slice(times, func(i, j int) bool { return times[i].Before(times[j]) })
	return times[0]
}

func flatCollect[T any](items []Episode, getter func(Episode) []T) []T {
	var out []T
	for _, it := range items {
		out = append(out, getter(it)...)
	}
	return out
}

func dedupStrings(in []string) []string {
	seen := map[string]struct{}{}
	var out []string
	for _, s := range in {
		if _, ok := seen[s]; ok {
			continue
		}
		seen[s] = struct{}{}
		out = append(out, s)
	}
	return out
}
