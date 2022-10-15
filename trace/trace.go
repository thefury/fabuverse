// Package trace provides a net/http middleware to trace reaquests from service to service.
package trace

// liberally copied from https://github.com/dmytrohridin/correlation-id/blob/master/correlation_id.go
// major additions to add span and parent for a better distributed view.

import (
	"context"
	"net/http"

	"github.com/google/uuid"
)

const (
	// HTTP header for Trace IDs
	DefaultTraceHeader = "Trace-Id"
	// HTTP header for Span IDs
	DefaultSpanHeader = "Span-Id"
	// HTTP header for Parent Span IDs
	DefaultParentHeader = "Parent-Id"

	ContextKey = "TraceIds"
)

// TODO decide if I really need this or not.
type TraceContext struct {
	TraceId  string
	SpanId   string
	ParentId string
}

type Middleware struct {
	TraceHeader  string
	SpanHeader   string
	ParentHeader string
	IdGenerator  func() string
}

func New() Middleware {
	return Middleware{
		TraceHeader:  DefaultTraceHeader,
		SpanHeader:   DefaultSpanHeader,
		ParentHeader: DefaultParentHeader,
		IdGenerator:  defaultGenerator,
	}
}

func (m *Middleware) Handle(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		traceHeaderName := m.getTraceName()
		spanHeaderName := m.getSpanName()
		parentHeaderName := m.getParentName()

		// trace isthe same for the entire chain of events
		traceId := r.Header.Get(traceHeaderName)

		if traceId == "" {
			traceId = m.IdGenerator()
		}

		w.Header().Set(traceHeaderName, traceId)

		// we always get a new span ID per request
		spanId := m.IdGenerator()
		w.Header().Set(spanHeaderName, spanId)

		// the parent is the callers span, so we can trace back
		parentId := r.Header.Get(spanHeaderName)
		w.Header().Set(parentHeaderName, parentId)

		newCtx := WithTraceData(r.Context(), traceId, spanId, parentId)
		next.ServeHTTP(w, r.WithContext(newCtx))
	})
}

func WithTraceData(ctx context.Context, traceId, spanId, parentId string) context.Context {
	return context.WithValue(ctx, ContextKey, TraceContext{TraceId: traceId, SpanId: spanId, ParentId: parentId})
}

func (m *Middleware) getTraceName() string {
	if m.TraceHeader == "" {
		return DefaultTraceHeader
	}

	return m.TraceHeader
}

func (m *Middleware) getParentName() string {
	if m.ParentHeader == "" {
		return DefaultParentHeader
	}

	return m.ParentHeader
}

func (m *Middleware) getSpanName() string {
	if m.SpanHeader == "" {
		return DefaultSpanHeader
	}

	return m.SpanHeader
}

func defaultGenerator() string {
	return uuid.NewString()
}
