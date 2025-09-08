package proxy

// Registry 拦截器注册表
type Registry struct {
	interceptors       map[string]*Interceptor
	globalInterceptors []*Interceptor
}

// makeKey 生成注册键，格式为 "METHOD:PATH"
func (r *Registry) makeKey(method, path string) string {
	return method + ":" + path
}

// NewRegistry 创建新的注册表
func NewRegistry() *Registry {
	return &Registry{
		interceptors: make(map[string]*Interceptor),
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
}

// RegisterAfterResponse 只注册响应后处理器
func (r *Registry) RegisterAfterResponse(method, path string, afterResp AfterResponseFunc) {
	key := r.makeKey(method, path)
	if r.interceptors[key] == nil {
		r.interceptors[key] = &Interceptor{}
	}
	r.interceptors[key].AfterResponse = afterResp
}

// RegisterBoth 注册请求前和响应后处理器
func (r *Registry) RegisterBoth(method, path string, preReq PreRequestFunc, afterResp AfterResponseFunc) {
	key := r.makeKey(method, path)
	r.interceptors[key] = &Interceptor{
		PreRequest:    preReq,
		AfterResponse: afterResp,
	}
}

// GetInterceptor 获取路径和方法对应的拦截器
func (r *Registry) GetInterceptor(method, path string) *Interceptor {
	// 首先尝试精确匹配
	key := r.makeKey(method, path)
	if interceptor := r.interceptors[key]; interceptor != nil {
		return interceptor
	}

	// 尝试通配符方法匹配（*:PATH）
	wildcardKey := r.makeKey("*", path)
	if interceptor := r.interceptors[wildcardKey]; interceptor != nil {
		return interceptor
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
