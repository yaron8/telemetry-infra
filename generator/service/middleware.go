package service

import (
	"log"
	"net/http"
	"time"
)

// middleware logs HTTP request details and records Prometheus metrics
func (api *APIServer) middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		// Log request
		log.Printf("[%s] %s %s - Started", r.Method, r.URL.Path, r.RemoteAddr)

		// Call the next handler
		next.ServeHTTP(w, r)

		// Calculate duration
		duration := time.Since(start)

		// Record metrics in Prometheus (convert to milliseconds)
		// Log response time
		log.Printf("[%s] %s %s - Completed in %v", r.Method, r.URL.Path, r.RemoteAddr, duration)
	})
}
