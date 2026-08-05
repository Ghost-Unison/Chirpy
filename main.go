package main

import (
	"log"
	"net/http"
	"sync/atomic"
)

// keep track of the number of requests we've received
type apiConfig struct {
	fileserverHits atomic.Int32
}

func main() {
	const filepathRoot = "."
	const port = "8080"
	var apicfg = apiConfig{atomic.Int32{}}

	// handler to route request
	mux := http.NewServeMux()
	mux.Handle("/app/", apicfg.middlewareMetricsInc(http.StripPrefix("/app", http.FileServer(http.Dir(filepathRoot)))))

	mux.HandleFunc("GET /api/healthz", healthCheckFunc)
	mux.HandleFunc("GET /admin/metrics", apicfg.returnFileserverHits)
	mux.HandleFunc("POST /admin/reset", apicfg.resetFileserverHits)

	//create new httpServer
	srv := &http.Server{
		Addr:    ":" + port,
		Handler: mux,
	}

	//start the server
	log.Printf("Serving files from %s on port: %s\n", filepathRoot, port)
	log.Fatal(srv.ListenAndServe())

}
