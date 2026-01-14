package main

import (
	"log"
	"net/http"
	"os"
	"sync"
	"time"

	"golang.org/x/time/rate"

	"github.com/secure-storage/internal/api"
	"github.com/secure-storage/internal/storage"
)

// Simple in-memory rate limiter store
type RateLimiter struct {
	visitors map[string]*rate.Limiter
	mu       sync.Mutex
}

func NewRateLimiter() *RateLimiter {
	return &RateLimiter{
		visitors: make(map[string]*rate.Limiter),
	}
}

// getLimiter creates a limiter for each IP: 1 request/sec, burst of 3
func (rl *RateLimiter) getLimiter(ip string) *rate.Limiter {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	limiter, exists := rl.visitors[ip]
	if !exists {
		limiter = rate.NewLimiter(1, 3)
		rl.visitors[ip] = limiter
	}
	return limiter
}

// Middleware to intercept requests
func (rl *RateLimiter) Limit(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ip := r.RemoteAddr // In production, use X-Forwarded-For if behind a proxy
		limiter := rl.getLimiter(ip)
		if !limiter.Allow() {
			http.Error(w, "Too Many Requests - Slow Down", http.StatusTooManyRequests)
			return
		}
		next.ServeHTTP(w, r)
	})
}

func main() {
	// Get encryption key from environment variable
	encryptionKey := os.Getenv("STORAGE_KEY")
	if encryptionKey == "" {
		log.Fatal("STORAGE_KEY environment variable is required")
	}

	// AES-256 needs a 32 byte key
	if len(encryptionKey) != 32 {
		log.Fatal("STORAGE_KEY must be exactly 32 characters for AES-256")
	}

	// Create data directory if it doesnt exist
	dataDir := "./data"
	if err := os.MkdirAll(dataDir, 0755); err != nil {
		log.Fatalf("Failed to create data directory: %v", err)
	}

	// Initialize the storage service
	storageService, err := storage.NewService(encryptionKey, dataDir)
	if err != nil {
		log.Fatalf("Failed to initialize storage service: %v", err)
	}

	// Create API handler
	handler := api.NewHandler(storageService)

	// Setup routes
	mux := http.NewServeMux()
	mux.HandleFunc("/upload", handler.UploadHandler)
	mux.HandleFunc("/download/", handler.DownloadHandler)
	mux.HandleFunc("/health", handler.HealthHandler)

	// *** SECURITY UPGRADE: Apply Rate Limiter ***
	limiter := NewRateLimiter()

	// Cleanup old visitors every minute (Prevent memory leaks)
	go func() {
		for {
			time.Sleep(time.Minute)
			limiter.mu.Lock()
			limiter.visitors = make(map[string]*rate.Limiter) // Reset map
			limiter.mu.Unlock()
		}
	}()

	// Get port from env or use default
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	log.Printf("Starting SecureStorage server on port %s", port)
	log.Printf("Endpoints: POST /upload, GET /download/{filename}, GET /health")

	if err := http.ListenAndServe(":"+port, limiter.Limit(mux)); err != nil {
		log.Fatalf("Server failed to start: %v", err)
	}
}
