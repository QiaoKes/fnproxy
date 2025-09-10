package proxy

import (
	"github.com/julienschmidt/httprouter"
	"net/http"
	"strings"
)

type HRMatcher struct {
	r        *httprouter.Router
	patterns map[string][]string // key=METHOD，保存该方法下已注册的所有 pattern（含 :param/*any）
}

func NewHRMatcher() *HRMatcher {
	hr := httprouter.New()
	hr.RedirectTrailingSlash = false
	hr.RedirectFixedPath = false
	return &HRMatcher{
		r:        hr,
		patterns: make(map[string][]string),
	}
}

var allHTTPMethods = []string{
	http.MethodGet, http.MethodPost, http.MethodPut,
	http.MethodPatch, http.MethodDelete, http.MethodHead, http.MethodOptions,
}

// Register 把 (method, pattern) 挂到内部路由器。
func (m *HRMatcher) Register(method, pattern string) {
	method = strings.ToUpper(strings.TrimSpace(method))

	add := func(mm string) {
		// 实际 handler 可以是空的，占位即可
		m.r.Handle(mm, pattern, func(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
			// 在header中写入pattern
			w.Header().Set("X-Matched-Pattern", pattern)
		})
		m.patterns[mm] = append(m.patterns[mm], pattern)
	}
	if method == "*" {
		for _, mm := range allHTTPMethods {
			add(mm)
		}
	} else {
		add(method)
	}
}

// Lookup 喂一个本地请求给路由器，看是否命中；命中则返回匹配到的模式字符串和路径参数
func (m *HRMatcher) Lookup(method, path string) (matchedPattern string, params map[string]string, ok bool) {
	method = strings.ToUpper(strings.TrimSpace(method))
	tryOne := func(mm string) (string, map[string]string, bool) {
		// 严格命中：httprouter 的 Lookup 只有"真正匹配"才 ok==true
		h, ps, _ := m.r.Lookup(mm, path)
		if h == nil {
			return "", nil, false
		}

		// 将 httprouter.Params 转换为 map[string]string
		paramMap := make(map[string]string)
		for _, p := range ps {
			paramMap[p.Key] = p.Value
		}

		// 尝试获取命中的 pattern，通过模拟调用 handler 来获取 X-Matched-Pattern
		// 创建一个模拟的 ResponseWriter 来捕获 header
		mockWriter := &mockResponseWriter{header: make(http.Header)}
		mockRequest, _ := http.NewRequest(mm, path, nil)

		// 调用 handler 来设置 X-Matched-Pattern
		h(mockWriter, mockRequest, ps)

		// 从 header 中获取匹配的 pattern
		pattern := mockWriter.header.Get("X-Matched-Pattern")

		return pattern, paramMap, true
	}

	if method == "*" {
		for _, mm := range allHTTPMethods {
			if pat, params, ok := tryOne(mm); ok {
				return pat, params, true
			}
		}
		return "", nil, false
	}
	return tryOne(method)
}

type mockResponseWriter struct {
	header http.Header
}

func (m *mockResponseWriter) Header() http.Header {
	return m.header
}

func (m *mockResponseWriter) Write([]byte) (int, error) {
	return 0, nil
}

func (m *mockResponseWriter) WriteHeader(int) {}
