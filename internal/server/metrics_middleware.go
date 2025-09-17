package server

import (
	"net/http"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var (
	httpRequestsTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "habits_http_requests_total",
			Help: "Total number of HTTP requests by endpoint and method",
		},
		[]string{"endpoint", "method"},
	)

	httpRequestDuration = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "habits_http_request_duration_seconds",
			Help:    "Duration of HTTP requests in seconds",
			Buckets: prometheus.DefBuckets,
		},
		[]string{"endpoint", "method"},
	)

	activeHabits = promauto.NewGauge(
		prometheus.GaugeOpts{
			Name: "habits_active_habits_total",
			Help: "Total number of active habits",
		},
	)
)

func metricsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		next.ServeHTTP(w, r)

		duration := time.Since(start).Seconds()
		httpRequestsTotal.WithLabelValues(r.URL.Path, r.Method).Inc()
		httpRequestDuration.WithLabelValues(r.URL.Path, r.Method).Observe(duration)
	})
}
