package main

import (
	"fmt"
	"log"
	"net/http"
	"sync/atomic"
)

// keep track of the number of requests we've received
type apiConfig struct {
	fileserverHits atomic.Int32
}

// middlware increase hits
func (cfg *apiConfig) middlewareMetricsInc(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		log.Printf("%s %s", r.Method, r.URL.Path)
		cfg.fileserverHits.Add(1)
		next.ServeHTTP(w, r)
	})

}

func main() {
	const filepathRoot = "."
	const port = "8080"
	var apicfg = apiConfig{atomic.Int32{}} //atomic.Int32 的零值就是 0

	// handler to route request
	mux := http.NewServeMux() // mux 是 *ServeMux 类型
	mux.Handle("/app/", apicfg.middlewareMetricsInc(http.StripPrefix("/app", http.FileServer(http.Dir(filepathRoot)))))
	//设置endpoint只能由某一种Http方法访问 [METHOD ][HOST]/[PATH]
	mux.HandleFunc("GET /healthz", healthCheckFunc)
	mux.HandleFunc("GET /metrics", apicfg.returnFileserverHits)
	mux.HandleFunc("POST /reset", apicfg.resetFileserverHits)

	//create new httpServer
	srv := &http.Server{
		Addr:    ":" + port,
		Handler: mux,
	}

	log.Printf("Serving files from %s on port: %s\n", filepathRoot, port)
	log.Fatal(srv.ListenAndServe()) //start the server

}

func healthCheckFunc(w http.ResponseWriter, r *http.Request) {
	//先设置header再调用writeHeader
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	//w.Write([]byte("OK"))
	w.Write([]byte(http.StatusText(http.StatusOK)))
}

/*
writes the number of requests that have been counted as plain text in this format to the HTTP response:
Hits: x
*/
func (cfg *apiConfig) returnFileserverHits(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	//w.Write([]byte(fmt.Sprintf("Hits: %d", cfg.fileserverHits.Load())))
	w.Write(fmt.Appendf(nil, "Hits: %d", cfg.fileserverHits.Load()))
}

func (cfg *apiConfig) resetFileserverHits(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	cfg.fileserverHits.Store(0)
	w.Write([]byte("Hits reset to 0"))

}
