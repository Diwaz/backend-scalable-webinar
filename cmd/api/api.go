package main

import (
	dlog "log"
	"net/http"
	"time"
)

type application struct {
	config config
}

type config struct {
	addr string
}

func (app *application) mount() http.Handler {
	mux := http.NewServeMux()

	mux.HandleFunc("GET /v1/health", app.checkHealth)
	mux.HandleFunc("POST /v1/upload", app.handleUpload)
	mux.HandleFunc("GET /v1/test", app.testHandler)
	mux.HandleFunc("GET /v1/stream", app.handleWS)
	return corsMiddleware(mux)
}
func corsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "POST, GET, OPTIONS, PUT, DELETE")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}

		next.ServeHTTP(w, r)
	})
}
func (app *application) run(mux http.Handler) error {

	srv := &http.Server{
		Addr:         app.config.addr,
		Handler:      mux,
		WriteTimeout: time.Second * 30,
		ReadTimeout:  time.Second * 10,
		IdleTimeout:  time.Minute * 2,
	}

	dlog.Printf("server has started  at %s", app.config.addr)
	return srv.ListenAndServe()
}
