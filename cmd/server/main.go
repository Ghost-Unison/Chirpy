package main

import (
	"database/sql"
	"log"
	"net/http"
	"os"

	"github.com/Ghost-Unison/Chirpy/internal/admin"
	"github.com/Ghost-Unison/Chirpy/internal/auth"
	"github.com/Ghost-Unison/Chirpy/internal/chirps"
	"github.com/Ghost-Unison/Chirpy/internal/database"
	"github.com/Ghost-Unison/Chirpy/internal/users"
	"github.com/Ghost-Unison/Chirpy/internal/webhooks"
	"github.com/joho/godotenv"

	_ "github.com/lib/pq"
)

func main() {
	const filepathRoot = "."
	const port = "8080"

	//获取数据库连接字符串并连接数据库
	godotenv.Load()
	dbURL := os.Getenv("DB_URL")
	platform := os.Getenv("PLATFORM")
	secretKey := os.Getenv("SECRET_KEY")
	polkaKey := os.Getenv("POLKA_KEY")

	db, err := sql.Open("postgres", dbURL)
	if err != nil {
		log.Fatalf("failed to connect to db: %v\n", err)
	}
	//借助sqlc生成的代码创建数据库查询对象,并注入到各业务包的 Handler 中（显式依赖注入，替代原先的全局 apiConfig）
	dbQueries := database.New(db)

	//构造各业务包的 Handler
	chirpsH := chirps.NewHandler(dbQueries, secretKey)
	usersH := users.NewHandler(dbQueries, secretKey)
	authH := auth.NewHandler(dbQueries, secretKey)
	webhooksH := webhooks.NewHandler(dbQueries, polkaKey)
	adminH := admin.NewHandler(dbQueries, platform)

	// handler to route request
	mux := http.NewServeMux()
	mux.Handle("/app/", adminH.MiddlewareMetricsInc(http.StripPrefix("/app", http.FileServer(http.Dir(filepathRoot)))))

	mux.HandleFunc("GET /api/healthz", admin.Healthz)
	mux.HandleFunc("GET /admin/metrics", adminH.Metrics)
	mux.HandleFunc("POST /admin/resetHits", adminH.ResetHits)
	mux.HandleFunc("POST /admin/reset", adminH.ResetUsers)

	mux.HandleFunc("POST /api/users", usersH.Create)
	mux.HandleFunc("PUT /api/users", usersH.Update)
	mux.HandleFunc("POST /api/login", usersH.Login)

	mux.HandleFunc("POST /api/chirps", chirpsH.Create)
	mux.HandleFunc("GET /api/chirps", chirpsH.List)
	mux.HandleFunc("GET /api/chirps/{chirpID}", chirpsH.Get)
	mux.HandleFunc("DELETE /api/chirps/{chirpID}", chirpsH.Delete)

	mux.HandleFunc("POST /api/refresh", authH.Refresh)
	mux.HandleFunc("POST /api/revoke", authH.Revoke)

	mux.HandleFunc("POST /api/polka/webhooks", webhooksH.PolkaWebhooks)

	//create new httpServer
	srv := &http.Server{
		Addr:    ":" + port,
		Handler: mux,
	}

	//start the server
	log.Printf("Serving files from %s on port: %s\n", filepathRoot, port)
	log.Fatal(srv.ListenAndServe())
}
