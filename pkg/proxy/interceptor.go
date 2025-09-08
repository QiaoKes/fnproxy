package proxy

import (
	"bytes"
	"io"
	"net/http"

	"github.com/gin-gonic/gin"
)

// PreRequestFunc 请求前处理函数类型
type PreRequestFunc func(*Context) InterceptorResult

// AfterResponseFunc 响应后处理函数类型
type AfterResponseFunc func(*Context) InterceptorResult

// InterceptorResult 拦截器返回结果
type InterceptorResult int

const (
	// Continue 继续原有逻辑，透明转发
	Continue InterceptorResult = iota
	// Cancel 取消原有逻辑，不转发
	Cancel
)

// Interceptor 拦截器结构，包含请求前和响应后处理
type Interceptor struct {
	PreRequest    PreRequestFunc    // 请求前处理
	AfterResponse AfterResponseFunc // 响应后处理
}

// Context 拦截器上下文，提供各种便捷工具
type Context struct {
	*gin.Context
	Request  *http.Request
	Response gin.ResponseWriter
	Path     string
	Method   string
	Headers  HeaderHelper
	RequestHelper
	ResponseHelper

	// 响应处理相关
	responseBody    *bytes.Buffer
	originalWriter  gin.ResponseWriter
	interceptWriter *ResponseInterceptor
}

// ResponseInterceptor 响应拦截器
type ResponseInterceptor struct {
	gin.ResponseWriter
	body       *bytes.Buffer
	statusCode int
	written    bool
}

// Write 拦截写入操作
func (w *ResponseInterceptor) Write(data []byte) (int, error) {
	if !w.written {
		w.written = true
	}
	return w.body.Write(data)
}

// WriteHeader 拦截状态码设置
func (w *ResponseInterceptor) WriteHeader(statusCode int) {
	w.statusCode = statusCode
	w.written = true
}

// Status 获取状态码
func (w *ResponseInterceptor) Status() int {
	return w.statusCode
}

// Size 获取响应大小
func (w *ResponseInterceptor) Size() int {
	return w.body.Len()
}

// Written 是否已写入
func (w *ResponseInterceptor) Written() bool {
	return w.written
}

// HeaderHelper 头部操作助手
type HeaderHelper struct {
	ctx *Context
}

// Set 设置header
func (h *HeaderHelper) Set(key, value string) {
	h.ctx.Request.Header.Set(key, value)
}

// Add 添加header
func (h *HeaderHelper) Add(key, value string) {
	h.ctx.Request.Header.Add(key, value)
}

// Del 删除header
func (h *HeaderHelper) Del(key string) {
	h.ctx.Request.Header.Del(key)
}

// Get 获取header
func (h *HeaderHelper) Get(key string) string {
	return h.ctx.Request.Header.Get(key)
}

// Filter 过滤header，只保留指定的header
func (h *HeaderHelper) Filter(keepHeaders ...string) {
	newHeader := make(http.Header)
	for _, key := range keepHeaders {
		if values := h.ctx.Request.Header.Values(key); len(values) > 0 {
			for _, value := range values {
				newHeader.Add(key, value)
			}
		}
	}
	h.ctx.Request.Header = newHeader
}

// RequestHelper 请求助手
type RequestHelper struct {
	ctx *Context
}

// SetBody 设置请求体
func (r *RequestHelper) SetBody(body []byte) {
	r.ctx.Request.Body = io.NopCloser(bytes.NewReader(body))
	r.ctx.Request.ContentLength = int64(len(body))
}

// GetBody 获取请求体
func (r *RequestHelper) GetBody() ([]byte, error) {
	if r.ctx.Request.Body == nil {
		return nil, nil
	}

	body, err := io.ReadAll(r.ctx.Request.Body)
	if err != nil {
		return nil, err
	}

	// 重新设置body，确保后续可以读取
	r.ctx.Request.Body = io.NopCloser(bytes.NewReader(body))
	return body, nil
}

// SetQuery 设置查询参数
func (r *RequestHelper) SetQuery(key, value string) {
	q := r.ctx.Request.URL.Query()
	q.Set(key, value)
	r.ctx.Request.URL.RawQuery = q.Encode()
}

// AddQuery 添加查询参数
func (r *RequestHelper) AddQuery(key, value string) {
	q := r.ctx.Request.URL.Query()
	q.Add(key, value)
	r.ctx.Request.URL.RawQuery = q.Encode()
}

// DelQuery 删除查询参数
func (r *RequestHelper) DelQuery(key string) {
	q := r.ctx.Request.URL.Query()
	q.Del(key)
	r.ctx.Request.URL.RawQuery = q.Encode()
}

// SetPath 设置请求路径
func (r *RequestHelper) SetPath(path string) {
	r.ctx.Request.URL.Path = path
	r.ctx.Path = path
}

// ResponseHelper 响应助手
type ResponseHelper struct {
	ctx *Context
}

// SetHeader 设置响应头
func (r *ResponseHelper) SetHeader(key, value string) {
	r.ctx.Response.Header().Set(key, value)
}

// AddHeader 添加响应头
func (r *ResponseHelper) AddHeader(key, value string) {
	r.ctx.Response.Header().Add(key, value)
}

// DelHeader 删除响应头
func (r *ResponseHelper) DelHeader(key string) {
	r.ctx.Response.Header().Del(key)
}

// SetStatus 设置状态码
func (r *ResponseHelper) SetStatus(code int) {
	r.ctx.Status(code)
}

// SetJSON 设置JSON响应
func (r *ResponseHelper) SetJSON(obj interface{}) {
	r.ctx.JSON(http.StatusOK, obj)
}

// SetJSONWithStatus 设置带状态码的JSON响应
func (r *ResponseHelper) SetJSONWithStatus(code int, obj interface{}) {
	r.ctx.JSON(code, obj)
}

// SetString 设置字符串响应
func (r *ResponseHelper) SetString(format string, values ...interface{}) {
	r.ctx.String(http.StatusOK, format, values...)
}

// SetStringWithStatus 设置带状态码的字符串响应
func (r *ResponseHelper) SetStringWithStatus(code int, format string, values ...interface{}) {
	r.ctx.String(code, format, values...)
}

// GetResponseBody 获取响应体内容
func (r *ResponseHelper) GetResponseBody() []byte {
	if r.ctx.interceptWriter != nil {
		return r.ctx.interceptWriter.body.Bytes()
	}
	return nil
}

// SetResponseBody 设置响应体内容
func (r *ResponseHelper) SetResponseBody(body []byte) {
	if r.ctx.interceptWriter != nil {
		r.ctx.interceptWriter.body.Reset()
		r.ctx.interceptWriter.body.Write(body)
	}
}

// GetResponseStatus 获取响应状态码
func (r *ResponseHelper) GetResponseStatus() int {
	if r.ctx.interceptWriter != nil {
		return r.ctx.interceptWriter.statusCode
	}
	return 0
}

// SetResponseStatus 设置响应状态码
func (r *ResponseHelper) SetResponseStatus(code int) {
	if r.ctx.interceptWriter != nil {
		r.ctx.interceptWriter.statusCode = code
	}
}

// Registry 拦截器注册表
type Registry struct {
	interceptors map[string]*Interceptor
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

// NewContext 创建拦截器上下文
func NewContext(c *gin.Context) *Context {
	ctx := &Context{
		Context:        c,
		Request:        c.Request,
		Response:       c.Writer,
		Path:           c.Request.URL.Path,
		Method:         c.Request.Method,
		originalWriter: c.Writer,
	}

	ctx.Headers = HeaderHelper{ctx: ctx}
	ctx.RequestHelper = RequestHelper{ctx: ctx}
	ctx.ResponseHelper = ResponseHelper{ctx: ctx}

	return ctx
}

// EnableResponseIntercept 启用响应拦截
func (ctx *Context) EnableResponseIntercept() {
	if ctx.interceptWriter == nil {
		ctx.interceptWriter = &ResponseInterceptor{
			ResponseWriter: ctx.originalWriter,
			body:           &bytes.Buffer{},
			statusCode:     200,
		}
		ctx.Writer = ctx.interceptWriter
		ctx.Response = ctx.interceptWriter
	}
}

// FlushResponse 将拦截的响应写入到真实的Writer
func (ctx *Context) FlushResponse() {
	if ctx.interceptWriter != nil {
		// 写入状态码
		if ctx.interceptWriter.statusCode != 0 {
			ctx.originalWriter.WriteHeader(ctx.interceptWriter.statusCode)
		}

		// 写入响应体
		if ctx.interceptWriter.body.Len() > 0 {
			ctx.originalWriter.Write(ctx.interceptWriter.body.Bytes())
		}
	}
}
