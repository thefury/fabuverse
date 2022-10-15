package middleware

import (
	"context"
	"net/http"
	"time"

	"go.uber.org/zap"
)

// logging middleware
// from: https://blog.questionable.services/article/guide-logging-middleware-go/

type responseData struct {
	status int
	size   int
}

type loggingResponseWriter struct {
	http.ResponseWriter // composition
	responseData        *responseData
}

func (w *loggingResponseWriter) Write(b []byte) (int, error) {
	size, err := w.ResponseWriter.Write(b)
	w.responseData.size += size

	return size, err
}

func (w *loggingResponseWriter) WriteHeader(statusCode int) {
	w.ResponseWriter.WriteHeader(statusCode)
	w.responseData.status = statusCode
}

// WithLogging implements an opinionated logging middleware using the zap
// logging framework.
func WithLogging(logger *zap.SugaredLogger) Adapter {
	return func(next http.Handler) http.Handler {
		fn := func(w http.ResponseWriter, r *http.Request) {

			// log even if there is a failure.
			defer func() {
				if err := recover(); err != nil {
					w.WriteHeader(http.StatusInternalServerError)
					logger.Errorw("http request error", "err", err)
				}
			}()

			start := time.Now()

			// add tracing from eariler middleware, will display NULL if nothing present
			tracedLogger := logger.With(
				"trace-id", r.Context().Value("TraceId"),
				"span-id", r.Context().Value("SpanId"),
				"parent-id", r.Context().Value("ParentId"),
			)

			responseData := &responseData{status: 0, size: 0}
			rw := loggingResponseWriter{ResponseWriter: w, responseData: responseData}

			newCtx := buildLoggingContext(r.Context(), tracedLogger)
			next.ServeHTTP(&rw, r.WithContext(newCtx))

			tracedLogger.Infow("http request",
				"status", responseData.status,
				"method", r.Method,
				"path", r.URL.EscapedPath(),
				"duration", time.Since(start),
				"size", responseData.size,
			)
		}

		return http.HandlerFunc(fn)
	}
}

func buildLoggingContext(ctx context.Context, logger *zap.SugaredLogger) context.Context {
	newctx := context.WithValue(ctx, "Logger", logger)

	return newctx
}
