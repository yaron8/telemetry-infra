package bootstrap

import (
	"fmt"
	"log"
	"net/http"
)

type Bootstrap struct {
}

func NewBootstrap() *Bootstrap {
	return &Bootstrap{}
}

// StartServer initializes and starts the HTTP server on port 9001
func (b *Bootstrap) StartServer() error {
	// Set up HTTP handlers
	http.HandleFunc("/counters", b.countersHandler)
	http.HandleFunc("/data", b.dataHandler)

	// Start the server
	port := ":9001"
	fmt.Printf("Starting HTTP server on port %s\n", port)
	if err := http.ListenAndServe(port, nil); err != nil {
		log.Fatalf("Failed to start server: %v", err)
		return err
	}

	return nil
}

// countersHandler handles the /counters endpoint
func (b *Bootstrap) countersHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Fprint(w, "this is counters EP")
}

// dataHandler handles the /data endpoint with query parameters
func (b *Bootstrap) dataHandler(w http.ResponseWriter, r *http.Request) {
	// Get query parameters
	paramA := r.URL.Query().Get("param_a")
	paramB := r.URL.Query().Get("param_b")

	// Print the parameters
	fmt.Fprintf(w, "param_a: %s\nparam_b: %s", paramA, paramB)
}
