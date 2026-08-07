package main

import (
	"database/sql"
	"log"
	"net/http"
	"os"
	"sync/atomic"

	"github.com/Ghost-Unison/Chirpy/internal/database"
	"github.com/joho/godotenv"

	_ "github.com/lib/pq"
)

type apiConfig struct {
	fileserverHits atomic.Int32
	dbQueries      *database.Queries
	platform       string
	secretKey      string
	polkaKey       string
}

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
	//借助sqlc生成的代码创建数据库查询对象,并放到apiConfig中供handler使用
	dbQueries := database.New(db)
	apicfg := apiConfig{
		fileserverHits: atomic.Int32{},
		dbQueries:      dbQueries,
		platform:       platform,
		secretKey:      secretKey,
		polkaKey:       polkaKey,
	}

	// handler to route request
	mux := http.NewServeMux()
	mux.Handle("/app/", apicfg.middlewareMetricsInc(http.StripPrefix("/app", http.FileServer(http.Dir(filepathRoot)))))

	mux.HandleFunc("GET /api/healthz", healthCheckFunc)
	mux.HandleFunc("GET /admin/metrics", apicfg.returnFileserverHits)
	mux.HandleFunc("POST /admin/resetHits", apicfg.resetFileserverHits)
	mux.HandleFunc("POST /admin/reset", apicfg.resetUsersHandler)

	mux.HandleFunc("POST /api/users", apicfg.createUserHandler)
	mux.HandleFunc("PUT /api/users", apicfg.updateUserHandler)

	mux.HandleFunc("POST /api/chirps", apicfg.createChirpHandler)
	mux.HandleFunc("GET /api/chirps", apicfg.getChirpsHandler)
	mux.HandleFunc("GET /api/chirps/{chirpID}", apicfg.getChirpHandler)
	mux.HandleFunc("DELETE /api/chirps/{chirpID}", apicfg.deleteChirpHandler)

	mux.HandleFunc("POST /api/login", apicfg.loginHandler)
	mux.HandleFunc("POST /api/refresh", apicfg.refreshHandler)
	mux.HandleFunc("POST /api/revoke", apicfg.revokeHandler)

	mux.HandleFunc("POST /api/polka/webhooks", apicfg.polkaWebhooksHandler)

	//create new httpServer
	srv := &http.Server{
		Addr:    ":" + port,
		Handler: mux,
	}

	//start the server
	log.Printf("Serving files from %s on port: %s\n", filepathRoot, port)
	log.Fatal(srv.ListenAndServe())

}
