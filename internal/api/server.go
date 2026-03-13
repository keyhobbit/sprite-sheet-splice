package api

import (
	"net/http"
)

func corsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		origin := r.Header.Get("Origin")
		if origin == "" {
			origin = "*"
		}
		w.Header().Set("Access-Control-Allow-Origin", origin)
		w.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type")

		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}

		next.ServeHTTP(w, r)
	})
}

func NewAPIServer() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("/api/export", ExportHandler)
	mux.HandleFunc("/api/remove-bg", RemoveBGHandler)
	mux.HandleFunc("/api/export-gif", ExportGIFHandler)
	mux.HandleFunc("/api/resize", ResizeHandler)
	return corsMiddleware(mux)
}
