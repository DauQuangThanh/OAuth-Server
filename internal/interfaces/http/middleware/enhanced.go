package middleware

import (
	"context"
	"encoding/json"
	"net/http"
	"runtime"
	"time"

	"auth0-server/internal/infrastructure/monitoring"
	"auth0-server/internal/infrastructure/tracing"
	"auth0-server/pkg/logger"
)

// MetricsMiddleware adds metrics collection to HTTP requests
func MetricsMiddleware(metrics *monitoring.MetricsCollector) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()
			metrics.IncRequestCount()

			// Wrap response writer to capture status code
			wrapper := &responseWrapper{ResponseWriter: w, statusCode: http.StatusOK}

			defer func() {
				duration := time.Since(start)
				metrics.RecordRequestDuration(duration)

				if wrapper.statusCode >= 400 {
					metrics.IncErrorCount()
				}
			}()

			next.ServeHTTP(wrapper, r)
		})
	}
}

// TracingMiddleware adds distributed tracing to HTTP requests
func TracingMiddleware(logger logger.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Extract or create trace context
			ctx, span := tracing.StartSpan(r.Context(), r.URL.Path)
			span.AddTag("http.method", r.Method)
			span.AddTag("http.url", r.URL.String())
			span.AddTag("http.user_agent", r.UserAgent())

			// Add trace ID to response headers for debugging
			if tc, ok := tracing.FromContext(ctx); ok {
				w.Header().Set("X-Trace-ID", string(tc.TraceID))
			}

			defer func() {
				tracing.FinishSpan(span, map[string]string{
					"component": "http_server",
				})
			}()

			// Update request context with tracing
			r = r.WithContext(ctx)
			next.ServeHTTP(w, r)
		})
	}
}

// HealthCheckMiddleware provides health check functionality
func HealthCheckMiddleware(health *monitoring.HealthChecker, metrics *monitoring.MetricsCollector) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.URL.Path == "/health" && r.Method == http.MethodGet {
				handleHealthCheck(w, r, health, metrics)
				return
			}

			if r.URL.Path == "/metrics" && r.Method == http.MethodGet {
				handleMetrics(w, r, metrics)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

// SecurityHeadersMiddleware adds security headers
func SecurityHeadersMiddleware() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Security headers
			w.Header().Set("X-Content-Type-Options", "nosniff")
			w.Header().Set("X-Frame-Options", "DENY")
			w.Header().Set("X-XSS-Protection", "1; mode=block")
			w.Header().Set("Strict-Transport-Security", "max-age=31536000; includeSubDomains")
			w.Header().Set("Content-Security-Policy", "default-src 'self'")
			w.Header().Set("Referrer-Policy", "strict-origin-when-cross-origin")

			next.ServeHTTP(w, r)
		})
	}
}

// CompressionMiddleware adds gzip compression (simplified implementation)
func CompressionMiddleware() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// In a real implementation, we'd add gzip compression here
			// For now, just pass through
			next.ServeHTTP(w, r)
		})
	}
}

// handleHealthCheck handles health check requests
func handleHealthCheck(w http.ResponseWriter, r *http.Request, health *monitoring.HealthChecker, metrics *monitoring.MetricsCollector) {
	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()

	healthStatus := health.CheckHealth(ctx)

	// Add system metrics to health check
	var memStats runtime.MemStats
	runtime.ReadMemStats(&memStats)

	response := map[string]interface{}{
		"status":    "healthy",
		"timestamp": time.Now().UTC().Format(time.RFC3339),
		"checks":    healthStatus,
		"system": map[string]interface{}{
			"goroutines": runtime.NumGoroutine(),
			"memory": map[string]interface{}{
				"alloc_mb": memStats.Alloc / 1024 / 1024,
				"sys_mb":   memStats.Sys / 1024 / 1024,
				"gc_runs":  memStats.NumGC,
			},
		},
		"metrics": metrics.GetMetricsMap(),
	}

	// Check if any health checks failed
	for _, check := range healthStatus {
		if checkMap, ok := check.(map[string]interface{}); ok {
			if status, exists := checkMap["status"]; exists && status == "unhealthy" {
				response["status"] = "unhealthy"
				w.WriteHeader(http.StatusServiceUnavailable)
				break
			}
		}
	}

	w.Header().Set("Content-Type", "application/json")
	writeJSONResponse(w, response)
}

// handleMetrics handles metrics requests
func handleMetrics(w http.ResponseWriter, r *http.Request, metrics *monitoring.MetricsCollector) {
	response := map[string]interface{}{
		"timestamp": time.Now().UTC().Format(time.RFC3339),
		"metrics":   metrics.GetMetricsMap(),
	}

	w.Header().Set("Content-Type", "application/json")
	writeJSONResponse(w, response)
}

// responseWrapper wraps http.ResponseWriter to capture status code
type responseWrapper struct {
	http.ResponseWriter
	statusCode int
}

func (rw *responseWrapper) WriteHeader(code int) {
	rw.statusCode = code
	rw.ResponseWriter.WriteHeader(code)
}

// writeJSONResponse writes a JSON response
func writeJSONResponse(w http.ResponseWriter, data interface{}) {
	jsonData, err := json.Marshal(data)
	if err != nil {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	w.Write(jsonData)
}
