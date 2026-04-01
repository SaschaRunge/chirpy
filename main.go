package main

import (
	"fmt"
	"net/http"
	"sync/atomic"
)

type apiConfig struct {
	fileServerHits atomic.Int32
}

func (cfg *apiConfig) middlewareMetricsInc(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		cfg.fileServerHits.Add(1)
		next.ServeHTTP(w, r)
	})
}

func (cfg *apiConfig) handlerReturnFileServerHits(w http.ResponseWriter, r *http.Request) {
	w.Header().Add("Content-Type", "text/plain; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	s := fmt.Sprintf("Hits: %d", cfg.fileServerHits.Load())
	w.Write([]byte(s))
}

func (cfg *apiConfig) handlerResetFileServerHits(w http.ResponseWriter, r *http.Request) {
	w.Header().Add("Content-Type", "text/plain; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	cfg.fileServerHits.Store(0)
	w.Write([]byte("Reset FileServerHits counter to 0."))
}

func main() {
	cfg := apiConfig{}
	mux := http.NewServeMux()
	mux.Handle(
		"/app/",
		http.StripPrefix("/app/", cfg.middlewareMetricsInc(http.FileServer(http.Dir(".")))),
	)
	mux.HandleFunc("GET /healthz", handlerReadiness)
	mux.HandleFunc("GET /metrics", cfg.handlerReturnFileServerHits)
	mux.HandleFunc("POST /reset", cfg.handlerResetFileServerHits)
	server := http.Server{
		Addr:                         ":8080",
		Handler:                      mux,
		DisableGeneralOptionsHandler: false,
	}

	err := server.ListenAndServe()

	fmt.Println(err)
}
