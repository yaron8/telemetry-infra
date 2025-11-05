package service

import (
	"net/http"
)

// middleware logs HTTP request details and records Prometheus metrics
func (api *APIServer) middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Call the next handler
		next.ServeHTTP(w, r)
	})
}
