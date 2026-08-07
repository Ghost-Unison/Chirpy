// Package admin 提供管理后台处理器: 访问量统计、用户重置与健康检查。
package admin

import (
	"fmt"
	"log"
	"net/http"
	"sync/atomic"

	"github.com/Ghost-Unison/Chirpy/internal/database"
)

// Handler 持有管理后台处理器所需的依赖。
// fileserverHits 为值类型，因 Handler 始终以指针使用（构造函数返回 *Handler，
// 所有方法均为指针接收者），不会被复制，故 atomic 的竞态保证依然成立。
type Handler struct {
	db             *database.Queries
	platform       string
	fileserverHits atomic.Int32
}

// NewHandler 构造一个持有 db 与 platform 标识的 admin.Handler。
func NewHandler(db *database.Queries, platform string) *Handler {
	return &Handler{db: db, platform: platform}
}

// MiddlewareMetricsInc 记录每次文件服务请求的访问量并打日志。
func (h *Handler) MiddlewareMetricsInc(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		log.Printf("%s %s", r.Method, r.URL.Path)
		h.fileserverHits.Add(1)
		next.ServeHTTP(w, r)
	})
}

// metricsPageTemplate is the HTML page served on /admin/metrics.
const metricsPageTemplate = `<html>
  <body>
    <h1>Welcome, Chirpy Admin</h1>
    <p>Chirpy has been visited %d times!</p>
  </body>
</html>`

/*
Metrics serves an HTML page reporting the number of
fileserver requests recorded by MiddlewareMetricsInc.
*/
func (h *Handler) Metrics(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	if _, err := fmt.Fprintf(w, metricsPageTemplate, h.fileserverHits.Load()); err != nil {
		log.Printf("failed to write metrics response: %v", err)
	}
}

// ResetHits 重置访问量计数为 0。
func (h *Handler) ResetHits(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	h.fileserverHits.Store(0)
	w.Write([]byte("Hits reset to 0"))
}

// ResetUsers 清空用户表，仅在 dev 平台下可用。
func (h *Handler) ResetUsers(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")

	//platform check
	if h.platform != "dev" {
		w.WriteHeader(http.StatusForbidden)
		w.Write([]byte("Forbidden"))
		return
	}

	//delete users
	err := h.db.DeleteUser(r.Context())
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("Delete users failed"))
		return
	}
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("Users reset"))
}

// Healthz 检查服务是否可用。
func Healthz(w http.ResponseWriter, r *http.Request) {
	//先设置header再调用writeHeader
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	//w.Write([]byte("OK"))
	w.Write([]byte(http.StatusText(http.StatusOK)))
}
