package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"
)

// rootHandler 输出简单 HTML，便于确认容器内服务是否启动成功。
func rootHandler(message string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		hostname, _ := os.Hostname()
		_, _ = fmt.Fprintf(w, `
<!DOCTYPE html>
<html lang="zh-cn">
<head>
  <meta charset="utf-8">
  <title>Lingcdn Docker 示例</title>
  <style>
    body { font-family: "Segoe UI", Arial, sans-serif; margin: 40px; background: #0b1014; color: #f6f9fc; }
    h1 { font-size: 2rem; margin-bottom: 0.5rem; }
    p { margin: 0.25rem 0; }
    code { background: #1b2228; padding: 2px 6px; border-radius: 4px; }
    .card { padding: 18px; border-radius: 12px; background: #151b21; max-width: 720px; }
  </style>
</head>
<body>
  <div class="card">
    <h1>%s</h1>
    <p>容器主机：<code>%s</code></p>
    <p>刷新即可验证 VS Code Docker 插件的端口转发、日志、容器终端等功能。</p>
    <p>接口演示：<code>/healthz</code>（健康检查），<code>/info</code>（返回 JSON）。</p>
  </div>
</body>
</html>
`, message, hostname)
	}
}

// infoHandler 返回当前时间与请求方信息，方便测试端口映射。
func infoHandler(message string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json; charset=utf-8")
		resp := map[string]string{
			"message":   message,
			"time":      time.Now().Format(time.RFC3339Nano),
			"client":    r.RemoteAddr,
			"userAgent": r.UserAgent(),
			"path":      r.URL.Path,
		}
		_ = json.NewEncoder(w).Encode(resp)
	}
}

// healthHandler 用于 docker compose/探针演示。
func healthHandler(w http.ResponseWriter, _ *http.Request) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	_ = json.NewEncoder(w).Encode(map[string]string{
		"status": "ok",
	})
}

func loggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		next.ServeHTTP(w, r)
		log.Printf("%s %s %s", r.Method, r.URL.Path, time.Since(start))
	})
}

func main() {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	message := os.Getenv("DEMO_MESSAGE")
	if message == "" {
		message = "Lingcdn Docker 示例服务运行中"
	}

	mux := http.NewServeMux()
	mux.HandleFunc("/", rootHandler(message))
	mux.HandleFunc("/info", infoHandler(message))
	mux.HandleFunc("/healthz", healthHandler)

	server := &http.Server{
		Addr:              ":" + port,
		Handler:           loggingMiddleware(mux),
		ReadHeaderTimeout: 5 * time.Second,
	}

	log.Printf("demo 服务启动：http://localhost:%s  （DEMO_MESSAGE=%s）", port, message)
	if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatalf("demo 服务异常退出: %v", err)
	}
}
