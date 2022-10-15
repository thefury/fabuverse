// Package trace provides a net/http middleware to trace reaquests from service to service.
package middleware

import (
	"net/http"

	"github.com/google/uuid"
)

// liberally copied from https://github.com/dmytrohridin/correlation-id/blob/master/correlation_id.go
// major additions to add span and parent for a better distributed view.

const (
	// HTTP header for Trace IDs
	DefaultTraceHeaderName = "Trace-Id"
	// HTTP header for Span IDs
	DefaultSpanHeaderName = "Span-Id"
	// HTTP header for Parent Span IDs
	DefaultParentHeaderName = "Parent-Id"
)

func defaultIdGenerator() string {
	return uuid.NewString()
}

type TraceConfig struct {
	TraceHeaderName  string
	SpanHeaderName   string
	ParentHeaderName string

	IdGenerator func() string
}

func NewTraceConfig() TraceConfig {
	return TraceConfig{
		TraceHeaderName:  DefaultTraceHeaderName,
		SpanHeaderName:   DefaultSpanHeaderName,
		ParentHeaderName: DefaultParentHeaderName,
		IdGenerator:      defaultIdGenerator,
	}
}

func WithTracing(config TraceConfig) Adapter {
	return func(next http.Handler) http.Handler {
		fn := func(w http.ResponseWriter, r *http.Request) {
			// The trace id is the same for the entire chain of calls
			traceId := r.Header.Get(config.TraceHeaderName)
			if traceId == "" {
				traceId = config.IdGenerator()
			}
			w.Header().Set(config.TraceHeaderName, traceId)

			// The span id is unique to this request
			spanId := config.IdGenerator()
			w.Header().Set(config.SpanHeaderName, spanId)

			// The parent ID is the span ID of the caller
			parentId := r.Header.Get(config.SpanHeaderName)
			w.Header().Set(config.ParentHeaderName, parentId)

			// possibly set a context for handlers to read
			//		newCtx := buildTraceContext(r.Context(), traceId, spanId, parentId)
			//		next.ServeHTTP(w, r.WithContext(newCtx))

			next.ServeHTTP(w, r)
		}

		return http.HandlerFunc(fn)
	}
}

//func WithTraceData(ctx context.Context, traceId, spanId, parentId string) context.Context {
//	return context.WithValue(ctx, ContextKey, TraceContext{TraceId: traceId, SpanId: spanId, ParentId: parentId})
//}
