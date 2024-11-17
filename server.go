package main

import (
	"context"
	"log"
	"net/http"
	"time"

	"github.com/gorilla/mux"
)

type contextKey string

// type contextKey, it won't conflict with other keys, even if they have the same string value.
const configKey contextKey = "config"

type Config struct {
	App string
}

func handleHome(w http.ResponseWriter, r *http.Request) {
	// Safely retrieve the value from the context
	config, ok := r.Context().Value(configKey).(*Config)
	if !ok || config == nil {
		http.Error(w, "Configuration not found in context", http.StatusInternalServerError)
		return
	}
	appName := config.App
	time.Sleep(2 * time.Second) // Simulate processing
	w.Write([]byte("Hello, I'm " + appName))
}

func loggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		log.Printf("Received %s request: %s from address: %s\n", r.Method, r.URL, r.RemoteAddr)
		next.ServeHTTP(w, r)
	})
}

func authenticationMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		token := r.Header.Get("X-Auth-Token")
		if token != "secretKey" {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}
		log.Println("Verified token")
		next.ServeHTTP(w, r)
	})
}

func timingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		next.ServeHTTP(w, r)
		duration := time.Since(start)
		log.Printf("Request took %s\n", duration)
	})
}

// Config middleware could be used to load configuration from a file or a database,
// and apply it to the request context.
func configMiddleware(config *Config) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ctx := r.Context()
			ctx = context.WithValue(ctx, configKey, config)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

func RESTheaderMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		next.ServeHTTP(w, r)
	})
}

func corsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Add CORS headers to all responses
		w.Header().Set("Access-Control-Allow-Origin", "*") // Adjust origin as needed
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

		// If it's a preflight request, handle it here
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}

		// For other requests, proceed to the next handler
		next.ServeHTTP(w, r)
	})
}

func main() {
	router := mux.NewRouter()

	server := &http.Server{
		Addr:    ":8080",
		Handler: router,
	}

	router.HandleFunc("/", handleHome).Methods("GET")
	// Applying middleware
	router.Use(configMiddleware(&Config{App: "MyGO(Passed from configMiddleware)"}))
	router.Use(loggingMiddleware)
	router.Use(timingMiddleware)
	router.Use(authenticationMiddleware)
	router.Use(RESTheaderMiddleware)
	router.Use(corsMiddleware)

	log.Println("Starting serving on :8080")
	log.Fatal(server.ListenAndServe())
}
