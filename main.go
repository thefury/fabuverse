package main

import (
	"io"
	"log"
	"net/http"
	"os"

	"github.com/thefury/fabuverse/middleware"
	"go.uber.org/zap"
)

func Reverse(s string) string {
	runes := []rune(s)
	for i, j := 0, len(runes)-1; i < j; i, j = i+1, j-1 {
		runes[i], runes[j] = runes[j], runes[i]
	}
	return string(runes)
}

func reverseHandler(w http.ResponseWriter, r *http.Request) {
	logger, ok := r.Context().Value("Logger").(*zap.SugaredLogger)

	if ok {
		logger.Info("running reverse handler")
	}

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

	chainedRouter := middleware.Chain(
		router,
		middleware.WithLogging(sugar),
		middleware.WithTracing(middleware.NewTraceConfig()),
	)

	sugar.Info("Starting fabuverse service on :3345")
	if err := http.ListenAndServe(":3345", chainedRouter); err != nil {
		sugar.Fatal(err)
	}
}
