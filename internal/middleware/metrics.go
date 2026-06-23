package middleware

import (
	"encoding/json"
	"net/http"
	"sync/atomic"
	"time"
)

type perfMetrics struct {
	totalRequests   atomic.Int64
	activeRequests  atomic.Int64
	totalErrors     atomic.Int64
	totalDurationNs atomic.Int64
}

var globalMetrics = &perfMetrics{}

func Performance() Middleware {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()
			globalMetrics.totalRequests.Add(1)
			globalMetrics.activeRequests.Add(1)
			defer globalMetrics.activeRequests.Add(-1)

			rec := &responseRecorder{ResponseWriter: w, status: http.StatusOK}
			next.ServeHTTP(rec, r)

			globalMetrics.totalDurationNs.Add(int64(time.Since(start)))
			if rec.status >= 500 {
				globalMetrics.totalErrors.Add(1)
			}
		})
	}
}

func MetricsHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		total := globalMetrics.totalRequests.Load()
		active := globalMetrics.activeRequests.Load()
		errors := globalMetrics.totalErrors.Load()
		totalDuration := globalMetrics.totalDurationNs.Load()

		var avgMs float64
		if total > 0 {
			avgMs = float64(totalDuration) / float64(total) / float64(time.Millisecond)
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]any{
			"total_requests":  total,
			"active_requests": active,
			"total_errors":    errors,
			"avg_duration_ms": avgMs,
		})
	}
}
