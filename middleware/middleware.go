package middleware

import "net/http"

type Adapter func(http.Handler) http.Handler

// Chain takes a list of middleware adapters and chains them together
func Chain(h http.Handler, adapters ...Adapter) http.Handler {
	for _, adapter := range adapters {
		h = adapter(h)
	}

	return h
}
