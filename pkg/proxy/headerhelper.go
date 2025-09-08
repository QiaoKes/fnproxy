package proxy

import "net/http"

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

// Contains 检查header是否存在
func (h *HeaderHelper) Contains(key string) bool {
	val, ok := h.ctx.Request.Header[key]
	return ok && len(val) > 0
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
