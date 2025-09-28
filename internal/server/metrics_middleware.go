package server

import (
	"net/http"
	"strconv"
	"time"

	"github.com/brk3/habits/internal/logger"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var (
	httpRequestsTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "habits_http_requests_total",
			Help: "Total number of HTTP requests by endpoint, method, and status",
		},
		[]string{"endpoint", "method", "status_code"},
	)

	httpRequestDuration = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "habits_http_request_duration_seconds",
			Help:    "Duration of HTTP requests in seconds",
			Buckets: prometheus.DefBuckets,
		},
		[]string{"endpoint", "method", "status_code"},
	)

	userRequestsTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "habits_user_requests_total",
			Help: "Total number of authenticated requests per user",
		},
		[]string{"user_id", "endpoint", "method"},
	)

	authEventsTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "habits_auth_events_total",
			Help: "Total authentication events by type and result",
		},
		[]string{"event_type", "result", "provider"},
	)

	activeHabitsPerUser = promauto.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "habits_active_habits_per_user",
			Help: "Number of active habits per user",
		},
		[]string{"user_id"},
	)

	// Legacy metric - keep for backward compatibility
	activeHabits = promauto.NewGauge(
		prometheus.GaugeOpts{
			Name: "habits_active_habits_total",
			Help: "Total number of active habits across all users",
		},
	)
)

type responseWriter struct {
	http.ResponseWriter
	statusCode int
}

func (rw *responseWriter) WriteHeader(code int) {
	rw.statusCode = code
	rw.ResponseWriter.WriteHeader(code)
}

func metricsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		wrapped := &responseWriter{ResponseWriter: w, statusCode: http.StatusOK}

		next.ServeHTTP(wrapped, r)

		duration := time.Since(start).Seconds()
		statusCode := strconv.Itoa(wrapped.statusCode)

		httpRequestsTotal.WithLabelValues(r.URL.Path, r.Method, statusCode).Inc()
		httpRequestDuration.WithLabelValues(r.URL.Path, r.Method, statusCode).Observe(duration)
	})
}

func (s *Server) userAwareMetricsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		next.ServeHTTP(w, r)

		// Only collect user metrics for authenticated requests
		if s.cfg.AuthEnabled {
			if user, ok := r.Context().Value(userCtxKey{}).(*User); ok {
				userRequestsTotal.WithLabelValues(user.UserID, r.URL.Path, r.Method).Inc()
			}
		}
	})
}

func RecordAuthEvent(eventType, result, provider string) {
	authEventsTotal.WithLabelValues(eventType, result, provider).Inc()
	logger.Debug("Recorded auth event", "type", eventType, "result", result, "provider", provider)
}

func UpdateActiveHabitsForUser(userID string, count int) {
	activeHabitsPerUser.WithLabelValues(userID).Set(float64(count))
}

func UpdateTotalActiveHabits(count int) {
	activeHabits.Set(float64(count))
}
