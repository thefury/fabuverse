package main

import (
	"io"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/thefury/fabuverse/trace"
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

func WithLogging(logger *zap.SugaredLogger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		fn := func(w http.ResponseWriter, r *http.Request) {
			defer func() {
				if err := recover(); err != nil {
					w.WriteHeader(http.StatusInternalServerError)
					logger.Errorw("http request error", "err", err)
				}
			}()

			start := time.Now()

			responseData := &responseData{status: 0, size: 0}
			rw := loggingResponseWriter{ResponseWriter: w, responseData: responseData}

			next.ServeHTTP(&rw, r)

			logger.Infow("http request",
				"status", responseData.status,
				"method", r.Method,
				"path", r.URL.EscapedPath(),
				"duration", time.Since(start),
				"size", responseData.size,
				"trace-id", w.Header().Get(trace.DefaultTraceHeader),
				"span-id", w.Header().Get(trace.DefaultSpanHeader),
				"parent-id", w.Header().Get(trace.DefaultParentHeader),
			)
		}

		return http.HandlerFunc(fn)
	}
}

func Reverse(s string) string {
	runes := []rune(s)
	for i, j := 0, len(runes)-1; i < j; i, j = i+1, j-1 {
		runes[i], runes[j] = runes[j], runes[i]
	}
	return string(runes)
}

func reverseHandler(w http.ResponseWriter, r *http.Request) {
	s := r.URL.Query().Get("word")

	w.WriteHeader(http.StatusOK)
	io.WriteString(w, Reverse(s))
}

func main() {
	logger, err := zap.NewProduction()
	if err != nil {
		log.Fatalf("could not initialize zap logger: %v\n", err)
	}
	defer logger.Sync()

	hostname, _ := os.Hostname()

	sugar := logger.With(
		zap.String("app", "fabuverse"),
		zap.String("host", hostname),
	).Sugar()

	router := http.NewServeMux()
	router.HandleFunc("/reverse", reverseHandler)

	tracingMiddleware := trace.New()
	withLogging := WithLogging(sugar)
	loggedRouter := tracingMiddleware.Handle(withLogging(router))

	sugar.Info("Starting fabuverse service on :3345")
	if err := http.ListenAndServe(":3345", loggedRouter); err != nil {
		sugar.Fatal(err)
	}
}
