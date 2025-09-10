package proxy

import (
	"fmt"
	"testing"
)

func TestHRMatcher_ArtPicturePath(t *testing.T) {
	matcher := NewHRMatcher()

	// 注册模式
	pattern := "/emby/Items/:itemid/Images/Backdrop/:index"
	matcher.Register("GET", pattern)

	// 测试目标URL路径（去掉查询参数）
	testPath := "/emby/Items/c7d22542a722404280d06d4da578fec6/Images/Backdrop/0"

	t.Logf("Testing pattern: %s", pattern)
	t.Logf("Testing path: %s", testPath)

	matchedPattern, params, ok := matcher.Lookup("GET", testPath)
	if !ok {
		t.Errorf("Expected pattern %s to match path %s, but it didn't match", pattern, testPath)
		return
	}

	if matchedPattern != pattern {
		t.Errorf("Expected matched pattern to be %s, but got %s", pattern, matchedPattern)
	}

	// 验证参数
	expectedParams := map[string]string{
		"itemid": "c7d22542a722404280d06d4da578fec6",
		"index":  "0",
	}

	for key, expectedValue := range expectedParams {
		if actualValue, exists := params[key]; !exists {
			t.Errorf("Expected parameter %s to exist, but it doesn't", key)
		} else if actualValue != expectedValue {
			t.Errorf("Expected parameter %s to be %s, but got %s", key, expectedValue, actualValue)
		}
	}

	t.Logf("Successfully matched pattern: %s with path: %s", matchedPattern, testPath)
	t.Logf("Extracted parameters: %v", params)
}

func TestHRMatcher_DebugLookup(t *testing.T) {
	matcher := NewHRMatcher()

	// 注册模式
	pattern := "/emby/Items/:itemid/Images/Backdrop/:index"
	matcher.Register("GET", pattern)

	// 测试路径
	testPath := "/emby/Items/c7d22542a722404280d06d4da578fec6/Images/Backdrop/0"

	// 直接测试 httprouter 的 Lookup
	h, _, _ := matcher.r.Lookup("GET", testPath)
	t.Logf("httprouter.Lookup result: handler=%v", h != nil)

	// 测试我们的 patterns 存储
	t.Logf("Registered patterns for GET: %v", matcher.patterns["GET"])

	// 测试我们的 Lookup
	matchedPattern, _, ok := matcher.Lookup("GET", testPath)
	t.Logf("Our Lookup result: pattern=%s, ok=%v", matchedPattern, ok)
}

func TestHRMatcher_EmbyImagePaths(t *testing.T) {
	matcher := NewHRMatcher()

	// 注册多个 Emby 相关的模式
	patterns := []string{
		"/emby/Items/:itemid/Images/Backdrop/:index",
		"/emby/Items/:itemid/Images/Primary",
		"/emby/Items/:itemid/Video/:type",
	}

	for _, pattern := range patterns {
		matcher.Register("GET", pattern)
	}

	// 测试用例
	testCases := []struct {
		path        string
		expected    string
		shouldMatch bool
	}{
		{
			path:        "/emby/Items/c7d22542a722404280d06d4da578fec6/Images/Backdrop/0",
			expected:    "/emby/Items/:itemid/Images/Backdrop/:index",
			shouldMatch: true,
		},
		{
			path:        "/emby/Items/abc123/Images/Primary",
			expected:    "/emby/Items/:itemid/Images/Primary",
			shouldMatch: true,
		},
		{
			path:        "/emby/Items/xyz789/Video/Logo",
			expected:    "/emby/Items/:itemid/Video/:type",
			shouldMatch: true,
		},
	}

	for i, tc := range testCases {
		t.Run(fmt.Sprintf("case_%d", i), func(t *testing.T) {
			matchedPattern, _, ok := matcher.Lookup("GET", tc.path)

			if tc.shouldMatch && !ok {
				t.Errorf("Expected path %s to match, but it didn't", tc.path)
				return
			}

			if !tc.shouldMatch && ok {
				t.Errorf("Expected path %s to not match, but it matched with %s", tc.path, matchedPattern)
				return
			}

			if tc.shouldMatch && matchedPattern != tc.expected {
				t.Errorf("Path %s: expected pattern %s, got %s", tc.path, tc.expected, matchedPattern)
			}

			t.Logf("Path %s matched pattern %s", tc.path, matchedPattern)
		})
	}
}

func TestHRMatcher_URLWithQueryParams(t *testing.T) {
	matcher := NewHRMatcher()

	pattern := "/emby/Items/:itemid/Images/Backdrop/:index"
	matcher.Register("GET", pattern)

	// 测试带查询参数的完整URL路径（注意：Lookup应该只接收路径部分，不包含查询参数）
	fullURL := "http://10.0.0.115:8005/emby/Items/c7d22542a722404280d06d4da578fec6/Images/Backdrop/0?maxWidth=720&tag=c7d22542a722404280d06d4da578fec6&quality=90"
	pathOnly := "/emby/Items/c7d22542a722404280d06d4da578fec6/Images/Backdrop/0"

	t.Logf("Full URL: %s", fullURL)
	t.Logf("Path only: %s", pathOnly)

	// 应该使用路径部分（不含查询参数）进行匹配
	matchedPattern, _, ok := matcher.Lookup("GET", pathOnly)
	if !ok {
		t.Errorf("Expected pattern %s to match path %s", pattern, pathOnly)
		return
	}

	if matchedPattern != pattern {
		t.Errorf("Expected matched pattern to be %s, got %s", pattern, matchedPattern)
	}

	t.Logf("Successfully matched: %s", matchedPattern)
}
