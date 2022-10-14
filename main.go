package main

import (
	"io"
	"net/http"
	"os"
)

func reverseHandler(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	io.WriteString(w, "This is not a reversed word\n")
}

func main() {
	router := http.NewServeMux()
	router.HandleFunc("/reverse", reverseHandler)

	if err := http.ListenAndServe(":3345", router); err != nil {
		os.Exit(1)
	}
}
