package proxy

import "strings"

// Registry 拦截器注册表
type Registry struct {
	interceptors       map[string]*Interceptor
	globalInterceptors []*Interceptor
	matcher            *HRMatcher
}

// makeKey 生成注册键，格式为 "METHOD:PATH"
func (r *Registry) makeKey(method, path string) string {
	return strings.ToUpper(strings.TrimSpace(method)) + ":" + path
}

// NewRegistry 创建新的注册表
func NewRegistry() *Registry {
	return &Registry{
		interceptors:       make(map[string]*Interceptor),
		globalInterceptors: make([]*Interceptor, 0),
		matcher:            NewHRMatcher(),
	}
}

// Register 注册完整的拦截器
func (r *Registry) Register(method, path string, interceptor *Interceptor) {
	key := r.makeKey(method, path)
	r.interceptors[key] = interceptor
}

// RegisterPreRequest 只注册请求前处理器
func (r *Registry) RegisterPreRequest(method, path string, preReq PreRequestFunc) {
	key := r.makeKey(method, path)
	if r.interceptors[key] == nil {
		r.interceptors[key] = &Interceptor{}
	}
	r.interceptors[key].PreRequest = preReq
	r.matcher.Register(method, path)
}

// RegisterAfterResponse 只注册响应后处理器
func (r *Registry) RegisterAfterResponse(method, path string, afterResp AfterResponseFunc) {
	key := r.makeKey(method, path)
	if r.interceptors[key] == nil {
		r.interceptors[key] = &Interceptor{}
	}
	r.interceptors[key].AfterResponse = afterResp
	r.matcher.Register(method, path)
}

// RegisterBoth 注册请求前和响应后处理器
func (r *Registry) RegisterBoth(method, path string, preReq PreRequestFunc, afterResp AfterResponseFunc) {
	key := r.makeKey(method, path)
	r.interceptors[key] = &Interceptor{
		PreRequest:    preReq,
		AfterResponse: afterResp,
	}
	r.matcher.Register(method, path)
}

// GetInterceptor 获取路径和方法对应的拦截器（支持 :param/*any 路径模式）
func (r *Registry) GetInterceptor(method, path string) *Interceptor {
	method = strings.ToUpper(strings.TrimSpace(method))

	// 1) 精确匹配：METHOD:/exact/path
	if ic := r.interceptors[r.makeKey(method, path)]; ic != nil {
		return ic
	}
	// 2) 通配方法 + 精确路径：*:/exact/path
	if ic := r.interceptors[r.makeKey("*", path)]; ic != nil {
		return ic
	}

	// 3) 用 httprouter 解析路径，拿到“命中的模式字符串”
	if r.matcher != nil {
		if pat, ok := r.matcher.Lookup(method, path); ok {
			// 3.1 同方法 + 模式
			if ic := r.interceptors[r.makeKey(method, pat)]; ic != nil {
				return ic
			}
			// 3.2 通配方法 + 模式
			if ic := r.interceptors[r.makeKey("*", pat)]; ic != nil {
				return ic
			}
		}
	}
	return nil
}

// GetAllInterceptors 获取所有拦截器
func (r *Registry) GetAllInterceptors() map[string]*Interceptor {
	return r.interceptors
}

// GetGlobalInterceptors 获取全局拦截器
func (r *Registry) GetGlobalInterceptors() []*Interceptor {
	return r.globalInterceptors
}

// HasPreRequest 检查是否有请求前处理器
func (r *Registry) HasPreRequest(method, path string) bool {
	interceptor := r.GetInterceptor(method, path)
	return interceptor != nil && interceptor.PreRequest != nil
}

// HasAfterResponse 检查是否有响应后处理器
func (r *Registry) HasAfterResponse(method, path string) bool {
	interceptor := r.GetInterceptor(method, path)
	return interceptor != nil && interceptor.AfterResponse != nil
}

// RegisterAnyMethod 为任何HTTP方法注册拦截器
func (r *Registry) RegisterAnyMethod(path string, interceptor *Interceptor) {
	r.Register("*", path, interceptor)
}

// RegisterAnyMethodPreRequest 为任何HTTP方法注册请求前处理器
func (r *Registry) RegisterAnyMethodPreRequest(path string, preReq PreRequestFunc) {
	r.RegisterPreRequest("*", path, preReq)
}

// RegisterAnyMethodAfterResponse 为任何HTTP方法注册响应后处理器
func (r *Registry) RegisterAnyMethodAfterResponse(path string, afterResp AfterResponseFunc) {
	r.RegisterAfterResponse("*", path, afterResp)
}

// RegisterAnyMethodBoth 为任何HTTP方法注册请求前和响应后处理器
func (r *Registry) RegisterAnyMethodBoth(path string, preReq PreRequestFunc, afterResp AfterResponseFunc) {
	r.RegisterBoth("*", path, preReq, afterResp)
}

// RegisterGlobalInterceptor 注册全局拦截器
func (r *Registry) RegisterGlobalInterceptor(interceptor *Interceptor) {
	r.globalInterceptors = append(r.globalInterceptors, interceptor)
}
