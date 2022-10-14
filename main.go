package main

import (
	"io"
	"net/http"
	"os"
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
	router := http.NewServeMux()
	router.HandleFunc("/reverse", reverseHandler)

	if err := http.ListenAndServe(":3345", router); err != nil {
		os.Exit(1)
	}
}
