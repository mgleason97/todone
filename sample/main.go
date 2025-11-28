package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"
)

// A simple server that serves a /ping endpoint.
// Intended to serve as a sample for todo aggregation.

func main() {
	// TODO: read port from environment variable (e.g. PORT)
	port := "8080"

	mux := http.NewServeMux()

	mux.HandleFunc("/ping", func(w http.ResponseWriter, r *http.Request) {
		resp := map[string]string{
			"message": "pong",
			"time":    time.Now().Format(time.RFC3339),
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	})

	// TODO: add healthz and readyz endpoints for k8s probes.

	srv := &http.Server{
		Addr:    ":" + port,
		Handler: loggingMiddleware(mux),
	}

	fmt.Printf("server listening on http://localhost:%s\n", port)
	if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatalf("server error: %v", err)
	}

	// TODO: implement graceful shutdown when receiving SIGINT/SIGTERM.
}

func loggingMiddleware(next http.Handler) http.Handler {
	// TODO: replace stdlib log with structured logger (zap, slog, zerolog).
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		next.ServeHTTP(w, r)

		// TODO: log request size and response status code.
		fmt.Printf("%s %s took %s\n", r.Method, r.URL.Path, time.Since(start))
	})
}

// TODO: add unit tests for /ping handler using httptest package.
