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
		m.r.Handle(mm, pattern, func(http.ResponseWriter, *http.Request, httprouter.Params) {})
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
		if h, ps, _ := m.r.Lookup(mm, path); h == nil {
			return "", nil, false
		} else {
			// 将 httprouter.Params 转换为 map[string]string
			paramMap := make(map[string]string)
			for _, p := range ps {
				paramMap[p.Key] = p.Value
			}

			// 在该方法已注册的 patterns 中选最具体的一条（静态段>参数段>通配*）
			best := ""
			bestScore := int(^uint(0)>>1) * -1 // very small
			for _, pat := range m.patterns[mm] {
				if matched, statics, params, stars, segs := strictMatch(pat, path); matched {
					// 评分：静态段多更好，参数少更好，*any 少更好，段数更"贴合"更好
					score := statics*10 - params*2 - stars*5 + segs // segs 作为轻微加权
					if score > bestScore || (score == bestScore && len(pat) > len(best)) {
						bestScore, best = score, pat
					}
				}
			}
			if best == "" {
				// 理论上不应该发生：Router 命中了，但我们没找到等价 pattern
				return "", nil, false
			}
			return best, paramMap, true
		}
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

// ExtractParams 从给定的模式和路径中提取参数（独立方法，可单独使用）
func (m *HRMatcher) ExtractParams(pattern, path string) map[string]string {
	params := make(map[string]string)
	ps := split(pattern)
	ss := split(path)

	i, j := 0, 0
	for i < len(ps) && j < len(ss) {
		pseg, sseg := ps[i], ss[j]
		switch {
		case strings.HasPrefix(pseg, "*"):
			// *any：吞掉余下所有段
			key := strings.TrimPrefix(pseg, "*")
			if key != "" {
				// 将剩余路径拼接
				remaining := strings.Join(ss[j:], "/")
				params[key] = remaining
			}
			return params
		case strings.HasPrefix(pseg, ":"):
			// :param：单段参数
			key := strings.TrimPrefix(pseg, ":")
			if key != "" {
				params[key] = sseg
			}
			i++
			j++
		default:
			// 静态段，跳过
			i++
			j++
		}
	}
	return params
}

// strictMatch 支持 :param（单段）、*any（吞剩余），绝不放宽为“相似路径”。
// 返回 matched, 静态段数, 参数段数, 星号段数, 段总数（用于评分）
func strictMatch(pattern, path string) (bool, int, int, int, int) {
	ps := split(pattern)
	ss := split(path)

	statics, params, stars := 0, 0, 0
	i, j := 0, 0
	for i < len(ps) && j < len(ss) {
		pseg, sseg := ps[i], ss[j]
		switch {
		case strings.HasPrefix(pseg, "*"):
			// *any：吞掉余下所有段（包括 0 段）
			stars++
			// * 必须是最后一个段（与 httprouter 一致）
			return i == len(ps)-1, statics, params, stars, len(ss)
		case strings.HasPrefix(pseg, ":"):
			if sseg == "" {
				return false, 0, 0, 0, 0
			}
			params++
			i++
			j++
		default:
			if pseg != sseg {
				return false, 0, 0, 0, 0
			}
			statics++
			i++
			j++
		}
	}
	// 正常结束：两边都应耗尽；若 pattern 剩最后一个是 *any，也算匹配
	if i == len(ps) && j == len(ss) {
		return true, statics, params, stars, len(ss)
	}
	if i == len(ps)-1 && strings.HasPrefix(ps[i], "*") && j == len(ss) {
		stars++
		return true, statics, params, stars, len(ss)
	}
	return false, 0, 0, 0, 0
}

func split(p string) []string {
	p = strings.Trim(p, "/")
	if p == "" {
		return nil
	}
	return strings.Split(p, "/")
}
