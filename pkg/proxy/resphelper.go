package proxy

import (
	"bytes"
	"errors"
	"fnproxy/pkg/utils"
	"github.com/gin-gonic/gin"
	"io"
	"net/http"
	"strconv"
)

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

// ReadFrom 重写
func (w *ResponseInterceptor) ReadFrom(r io.Reader) (int64, error) {
	if !w.written {
		w.written = true
	}
	// 直接把上游数据读入缓冲区；不要写到底层原始 writer
	return w.body.ReadFrom(r)
}

//// 可选：如果需要兼容某些 Flush 行为
//func (w *ResponseInterceptor) Flush() {
//	// 不立即向下游刷写，由 ctx.FlushResponse() 统一输出
//	// 如需与某些中间件兼容，可选择性透传到底层：
//	// if f, ok := w.ResponseWriter.(http.Flusher); ok { f.Flush() }
//}

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

func (h *ResponseHelper) SetJSONBodyWithEncoding(plain []byte, normalizedEnc string) error {
	out, err := utils.Encode(plain, normalizedEnc)
	if err != nil && !errors.Is(err, utils.ErrUnknownEncoding) {
		// 回压缩失败则降级明文
		out = plain
		normalizedEnc = ""
	}
	h.SetResponseBody(out)

	hd := h.ctx.Response.Header()
	if hd.Get("Content-Type") == "" {
		hd.Set("Content-Type", "application/json; charset=utf-8")
	}
	if normalizedEnc == "" {
		hd.Del("Content-Encoding")
	} else {
		hd.Set("Content-Encoding", normalizedEnc)
	}
	hd.Set("Content-Length", strconv.Itoa(len(out)))
	return nil
}
