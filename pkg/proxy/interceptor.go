package proxy

import (
	"bytes"
	"github.com/gin-gonic/gin"
	"net/http"
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

	// 路径参数
	PathParams map[string]string

	// 响应处理相关
	responseBody    *bytes.Buffer
	originalWriter  gin.ResponseWriter
	interceptWriter *ResponseInterceptor
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
		PathParams:     make(map[string]string),
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

// GetPathParam 获取路径参数
func (ctx *Context) GetPathParam(key string) string {
	if ctx.PathParams == nil {
		return ""
	}
	return ctx.PathParams[key]
}

// GetPathParamWithDefault 获取路径参数，如果不存在则返回默认值
func (ctx *Context) GetPathParamWithDefault(key, defaultValue string) string {
	if value := ctx.GetPathParam(key); value != "" {
		return value
	}
	return defaultValue
}

// HasPathParam 检查是否存在指定的路径参数
func (ctx *Context) HasPathParam(key string) bool {
	if ctx.PathParams == nil {
		return false
	}
	_, exists := ctx.PathParams[key]
	return exists
}

// GetAllPathParams 获取所有路径参数
func (ctx *Context) GetAllPathParams() map[string]string {
	if ctx.PathParams == nil {
		return make(map[string]string)
	}
	return ctx.PathParams
}
