package monitoring

import (
	"context"
	"sync"
	"time"
)

// Metrics represents application metrics
type Metrics struct {
	// Request metrics
	RequestCount    int64
	ErrorCount      int64
	RequestDuration time.Duration

	// Authentication metrics
	LoginAttempts    int64
	SuccessfulLogins int64
	FailedLogins     int64

	// User metrics
	TotalUsers  int64
	ActiveUsers int64

	// System metrics
	Uptime         time.Time
	MemoryUsage    int64
	GoroutineCount int64
}

// MetricsCollector collects and manages application metrics
type MetricsCollector struct {
	metrics *Metrics
	mu      sync.RWMutex
}

// NewMetricsCollector creates a new metrics collector
func NewMetricsCollector() *MetricsCollector {
	return &MetricsCollector{
		metrics: &Metrics{
			Uptime: time.Now(),
		},
	}
}

// IncRequestCount increments the request counter
func (m *MetricsCollector) IncRequestCount() {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.metrics.RequestCount++
}

// IncErrorCount increments the error counter
func (m *MetricsCollector) IncErrorCount() {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.metrics.ErrorCount++
}

// RecordRequestDuration records request duration
func (m *MetricsCollector) RecordRequestDuration(duration time.Duration) {
	m.mu.Lock()
	defer m.mu.Unlock()
	// Simple moving average for demo - in production use proper metrics
	m.metrics.RequestDuration = (m.metrics.RequestDuration + duration) / 2
}

// IncLoginAttempt increments login attempt counter
func (m *MetricsCollector) IncLoginAttempt() {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.metrics.LoginAttempts++
}

// IncSuccessfulLogin increments successful login counter
func (m *MetricsCollector) IncSuccessfulLogin() {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.metrics.SuccessfulLogins++
}

// IncFailedLogin increments failed login counter
func (m *MetricsCollector) IncFailedLogin() {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.metrics.FailedLogins++
}

// SetUserCounts updates user count metrics
func (m *MetricsCollector) SetUserCounts(total, active int64) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.metrics.TotalUsers = total
	m.metrics.ActiveUsers = active
}

// GetMetrics returns a copy of current metrics
func (m *MetricsCollector) GetMetrics() Metrics {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return *m.metrics
}

// GetMetricsMap returns metrics as a map for JSON serialization
func (m *MetricsCollector) GetMetricsMap() map[string]interface{} {
	metrics := m.GetMetrics()

	return map[string]interface{}{
		"requests": map[string]interface{}{
			"total":           metrics.RequestCount,
			"errors":          metrics.ErrorCount,
			"error_rate":      float64(metrics.ErrorCount) / float64(metrics.RequestCount+1),
			"avg_duration_ms": metrics.RequestDuration.Milliseconds(),
		},
		"authentication": map[string]interface{}{
			"login_attempts":    metrics.LoginAttempts,
			"successful_logins": metrics.SuccessfulLogins,
			"failed_logins":     metrics.FailedLogins,
			"success_rate":      float64(metrics.SuccessfulLogins) / float64(metrics.LoginAttempts+1),
		},
		"users": map[string]interface{}{
			"total":  metrics.TotalUsers,
			"active": metrics.ActiveUsers,
		},
		"system": map[string]interface{}{
			"uptime_seconds": time.Since(metrics.Uptime).Seconds(),
			"memory_mb":      metrics.MemoryUsage / 1024 / 1024,
			"goroutines":     metrics.GoroutineCount,
		},
	}
}

// HealthChecker provides health check functionality
type HealthChecker struct {
	checks map[string]func(ctx context.Context) error
	mu     sync.RWMutex
}

// NewHealthChecker creates a new health checker
func NewHealthChecker() *HealthChecker {
	return &HealthChecker{
		checks: make(map[string]func(ctx context.Context) error),
	}
}

// AddCheck adds a health check
func (h *HealthChecker) AddCheck(name string, check func(ctx context.Context) error) {
	h.mu.Lock()
	defer h.mu.Unlock()
	h.checks[name] = check
}

// CheckHealth runs all health checks
func (h *HealthChecker) CheckHealth(ctx context.Context) map[string]interface{} {
	h.mu.RLock()
	checks := make(map[string]func(ctx context.Context) error)
	for name, check := range h.checks {
		checks[name] = check
	}
	h.mu.RUnlock()

	results := make(map[string]interface{})

	for name, check := range checks {
		if err := check(ctx); err != nil {
			results[name] = map[string]interface{}{
				"status": "unhealthy",
				"error":  err.Error(),
			}
		} else {
			results[name] = map[string]interface{}{
				"status": "healthy",
			}
		}
	}

	return results
}
