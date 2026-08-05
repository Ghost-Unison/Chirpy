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
