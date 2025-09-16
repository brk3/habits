package server

import (
	"context"
	"net/http"
	"time"
)

type contextKey string

func authMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			http.Error(w, "Authorization header required", http.StatusUnauthorized)
			return
		}

		// TODO(pbourke): validate and extract from token
		userId := "paul"

		ctx := context.WithValue(r.Context(), contextKey("user"), userId)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func metricsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		next.ServeHTTP(w, r)

		duration := time.Since(start).Seconds()
		httpRequestsTotal.WithLabelValues(r.URL.Path, r.Method).Inc()
		httpRequestDuration.WithLabelValues(r.URL.Path, r.Method).Observe(duration)
	})
}
