package main

import (
	"fmt"
	"log"
	"net/http"
)

// middlware increase hits
func (cfg *apiConfig) middlewareMetricsInc(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		log.Printf("%s %s", r.Method, r.URL.Path)
		cfg.fileserverHits.Add(1)
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
returnFileserverHits serves an HTML page reporting the number of
fileserver requests recorded by middlewareMetricsInc.
*/
func (cfg *apiConfig) returnFileserverHits(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	if _, err := fmt.Fprintf(w, metricsPageTemplate, cfg.fileserverHits.Load()); err != nil {
		log.Printf("failed to write metrics response: %v", err)
	}
}

func (cfg *apiConfig) resetFileserverHits(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	cfg.fileserverHits.Store(0)
	w.Write([]byte("Hits reset to 0"))

}
