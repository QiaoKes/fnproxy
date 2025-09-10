package proxy

import (
	"fmt"
	"fnproxy/pkg/config"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
)

// 在测试中使用的 Recorder，增加 CloseNotify 方法以兼容 gin/httputil 的断言
type closeNotifyRecorder struct {
	*httptest.ResponseRecorder
}

func (c *closeNotifyRecorder) CloseNotify() <-chan bool {
	ch := make(chan bool, 1)
	return ch
}

func TestServer_PreRequestCancel(t *testing.T) {
	// 启动一个后端测试服务，记录是否被调用
	called := false
	target := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		called = true
		w.WriteHeader(200)
		_, _ = w.Write([]byte("backend"))
	}))
	defer target.Close()

	u, err := url.Parse(target.URL)
	if err != nil {
		t.Fatalf("parse target url: %v", err)
	}

	host := u.Hostname()
	port := 0
	if p := u.Port(); p != "" {
		// parse port
		// we don't need numeric port for our config since NewServer only formats it back,
		// but config.Target.Port expects int, so parse it
		fmtPort := p
		// convert
		// simple atoi
		var pi int
		_, err := fmt.Sscanf(fmtPort, "%d", &pi)
		if err == nil {
			port = pi
		}
	}

	cfg := &config.Config{
		Server: config.ServerConfig{Listen: "127.0.0.1:0"},
		Target: config.TargetConfig{Host: host, Port: port},
		Log:    config.LogConfig{Level: 0},
	}

	srv := NewServer(cfg)

	// 注册一个在请求前直接取消并返回 403 的拦截器
	srv.RegisterPreRequest("GET", "/cancel", func(ctx *Context) InterceptorResult {
		ctx.ResponseHelper.SetStringWithStatus(403, "blocked")
		return Cancel
	})

	req := httptest.NewRequest("GET", "/cancel", nil)
	rec := &closeNotifyRecorder{httptest.NewRecorder()}

	srv.GetEngine().ServeHTTP(rec, req)

	if rec.Code != 403 {
		t.Fatalf("expected status 403, got %d", rec.Code)
	}

	b, _ := io.ReadAll(rec.Result().Body)
	if strings.TrimSpace(string(b)) != "blocked" {
		t.Fatalf("expected body 'blocked', got '%s'", string(b))
	}

	if called {
		t.Fatalf("backend should not have been called when PreRequest cancelled the request")
	}
}

func TestServer_AfterResponseModify(t *testing.T) {
	// 启动一个后端测试服务，返回固定响应
	target := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(200)
		_, _ = w.Write([]byte("original"))
	}))
	defer target.Close()

	u, err := url.Parse(target.URL)
	if err != nil {
		t.Fatalf("parse target url: %v", err)
	}

	host := u.Hostname()
	port := 0
	if p := u.Port(); p != "" {
		var pi int
		_, err := fmt.Sscanf(p, "%d", &pi)
		if err == nil {
			port = pi
		}
	}

	cfg := &config.Config{
		Server: config.ServerConfig{Listen: "127.0.0.1:0"},
		Target: config.TargetConfig{Host: host, Port: port},
		Log:    config.LogConfig{Level: 0},
	}

	srv := NewServer(cfg)

	// 注册响应后处理器，修改响应体
	srv.RegisterAfterResponse("PATCH", "/modify", func(ctx *Context) InterceptorResult {
		// 读取原始响应体
		orig := ctx.ResponseHelper.GetResponseBody()
		_ = orig // for clarity
		ctx.ResponseHelper.SetResponseBody([]byte("modified"))
		return Continue
	})

	req := httptest.NewRequest("PATCH", "/modify", nil)
	rec := &closeNotifyRecorder{httptest.NewRecorder()}

	srv.GetEngine().ServeHTTP(rec, req)

	if rec.Code != 200 {
		t.Fatalf("expected status 200, got %d", rec.Code)
	}

	b, _ := io.ReadAll(rec.Result().Body)
	if strings.TrimSpace(string(b)) != "modified" {
		t.Fatalf("expected body 'modified', got '%s'", string(b))
	}
}

// 新增：测试 PreRequest 修改请求头和请求体
func TestServer_PreRequestModifyHeadersAndBody(t *testing.T) {
	var gotHeader string
	var gotBody string
	target := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotHeader = r.Header.Get("X-Injected")
		b, _ := io.ReadAll(r.Body)
		gotBody = string(b)
		w.WriteHeader(200)
		_, _ = w.Write([]byte("ok"))
	}))
	defer target.Close()

	u, err := url.Parse(target.URL)
	if err != nil {
		t.Fatalf("parse target url: %v", err)
	}

	host := u.Hostname()
	port := 0
	if p := u.Port(); p != "" {
		var pi int
		_, err := fmt.Sscanf(p, "%d", &pi)
		if err == nil {
			port = pi
		}
	}

	cfg := &config.Config{
		Server: config.ServerConfig{Listen: "127.0.0.1:0"},
		Target: config.TargetConfig{Host: host, Port: port},
		Log:    config.LogConfig{Level: 0},
	}

	srv := NewServer(cfg)

	srv.RegisterPreRequest("POST", "/premodify", func(ctx *Context) InterceptorResult {
		// 修改请求头与请求体
		ctx.Headers.Set("X-Injected", "yes")
		ctx.RequestHelper.SetBody([]byte("replaced"))
		return Continue
	})

	req := httptest.NewRequest("POST", "/premodify", strings.NewReader("original"))
	rec := &closeNotifyRecorder{httptest.NewRecorder()}

	srv.GetEngine().ServeHTTP(rec, req)

	if rec.Code != 200 {
		t.Fatalf("expected 200, got %d", rec.Code)
	}

	if gotHeader != "yes" {
		t.Fatalf("expected backend to receive injected header, got '%s'", gotHeader)
	}

	if gotBody != "replaced" {
		t.Fatalf("expected backend to receive replaced body, got '%s'", gotBody)
	}
}

// 新增：测试 AfterResponse 修改响应头和响应体
func TestServer_AfterResponseModifyHeadersAndBody(t *testing.T) {
	// 后端返回原始头与体
	target := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("X-From-Backend", "orig")
		w.WriteHeader(200)
		_, _ = w.Write([]byte("backend-body"))
	}))
	defer target.Close()

	u, err := url.Parse(target.URL)
	if err != nil {
		t.Fatalf("parse target url: %v", err)
	}

	host := u.Hostname()
	port := 0
	if p := u.Port(); p != "" {
		var pi int
		_, err := fmt.Sscanf(p, "%d", &pi)
		if err == nil {
			port = pi
		}
	}

	cfg := &config.Config{
		Server: config.ServerConfig{Listen: "127.0.0.1:0"},
		Target: config.TargetConfig{Host: host, Port: port},
		Log:    config.LogConfig{Level: 0},
	}

	srv := NewServer(cfg)

	srv.RegisterAfterResponse("PATCH", "/aftermodify", func(ctx *Context) InterceptorResult {
		// 读取并修改响应头与体
		origBody := ctx.ResponseHelper.GetResponseBody()
		_ = origBody
		ctx.ResponseHelper.SetHeader("X-Modified", "yes")
		ctx.ResponseHelper.SetResponseBody([]byte("new-body"))
		return Continue
	})

	req := httptest.NewRequest("PATCH", "/aftermodify", nil)
	rec := &closeNotifyRecorder{httptest.NewRecorder()}

	srv.GetEngine().ServeHTTP(rec, req)

	if rec.Code != 200 {
		t.Fatalf("expected 200, got %d", rec.Code)
	}

	// 验证修改后的头与体
	if rec.Header().Get("X-Modified") != "yes" {
		t.Fatalf("expected X-Modified header, got '%s'", rec.Header().Get("X-Modified"))
	}

	b, _ := io.ReadAll(rec.Result().Body)
	if strings.TrimSpace(string(b)) != "new-body" {
		t.Fatalf("expected body 'new-body', got '%s'", string(b))
	}
}
