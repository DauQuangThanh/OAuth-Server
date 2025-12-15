package tracing

import (
	"context"
	"crypto/rand"
	"fmt"
	"time"
)

// TraceID represents a unique trace identifier
type TraceID string

// SpanID represents a unique span identifier
type SpanID string

// TraceContext contains tracing information
type TraceContext struct {
	TraceID   TraceID
	SpanID    SpanID
	ParentID  SpanID
	StartTime time.Time
	Tags      map[string]string
}

// ContextKey is used for context keys to avoid collisions
type ContextKey string

const (
	// TraceContextKey is the context key for trace information
	TraceContextKey ContextKey = "trace_context"
)

// GenerateTraceID generates a new trace ID
func GenerateTraceID() TraceID {
	bytes := make([]byte, 16)
	rand.Read(bytes)
	return TraceID(fmt.Sprintf("%x", bytes))
}

// GenerateSpanID generates a new span ID
func GenerateSpanID() SpanID {
	bytes := make([]byte, 8)
	rand.Read(bytes)
	return SpanID(fmt.Sprintf("%x", bytes))
}

// NewTraceContext creates a new trace context
func NewTraceContext() *TraceContext {
	return &TraceContext{
		TraceID:   GenerateTraceID(),
		SpanID:    GenerateSpanID(),
		StartTime: time.Now(),
		Tags:      make(map[string]string),
	}
}

// NewChildSpan creates a child span from the current trace context
func (tc *TraceContext) NewChildSpan() *TraceContext {
	return &TraceContext{
		TraceID:   tc.TraceID,
		SpanID:    GenerateSpanID(),
		ParentID:  tc.SpanID,
		StartTime: time.Now(),
		Tags:      make(map[string]string),
	}
}

// AddTag adds a tag to the trace context
func (tc *TraceContext) AddTag(key, value string) {
	tc.Tags[key] = value
}

// Duration returns the duration since the span started
func (tc *TraceContext) Duration() time.Duration {
	return time.Since(tc.StartTime)
}

// WithTraceContext adds trace context to the given context
func WithTraceContext(ctx context.Context, tc *TraceContext) context.Context {
	return context.WithValue(ctx, TraceContextKey, tc)
}

// FromContext extracts trace context from the given context
func FromContext(ctx context.Context) (*TraceContext, bool) {
	tc, ok := ctx.Value(TraceContextKey).(*TraceContext)
	return tc, ok
}

// StartSpan starts a new span in the current trace or creates a new trace if none exists
func StartSpan(ctx context.Context, operationName string) (context.Context, *TraceContext) {
	if tc, ok := FromContext(ctx); ok {
		// Create child span
		childSpan := tc.NewChildSpan()
		childSpan.AddTag("operation", operationName)
		return WithTraceContext(ctx, childSpan), childSpan
	}

	// Create new trace
	newTrace := NewTraceContext()
	newTrace.AddTag("operation", operationName)
	return WithTraceContext(ctx, newTrace), newTrace
}

// FinishSpan marks the span as finished and can be used for logging
func FinishSpan(tc *TraceContext, tags map[string]string) {
	if tc == nil {
		return
	}

	// Add any final tags
	for key, value := range tags {
		tc.AddTag(key, value)
	}

	// In a real implementation, this would send the span to a tracing backend
	// For now, we just add the duration
	tc.AddTag("duration_ms", fmt.Sprintf("%.2f", tc.Duration().Seconds()*1000))
}
