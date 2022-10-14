package main

import (
	"io"
	"net/http"
	"os"

	log "github.com/sirupsen/logrus"
)

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
	log.SetFormatter(&log.JSONFormatter{})
	log.SetOutput(os.Stdout)
	log.SetLevel(log.InfoLevel)

	router := http.NewServeMux()
	router.HandleFunc("/reverse", reverseHandler)

	log.Info("Starting fabuverse service on :3345")
	if err := http.ListenAndServe(":3345", router); err != nil {
		log.Fatal(err)
	}
}
