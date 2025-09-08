package proxy

import (
	"bytes"
	"io"
)

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
