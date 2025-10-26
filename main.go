package main

import (
	"log"
	"net/http"
	"os"
	"time"

	"github.com/joho/godotenv"
)

func main() {
	// Load .env file
	if err := godotenv.Load(); err != nil {
		log.Println("Warning: .env file not found, using default/empty values")
	}

	// Load API key from environment
	apiKey = os.Getenv("OCR_API_KEY")
	if apiKey == "" {
		log.Println("Warning: OCR_API_KEY not set in .env, OCR.space API will not work")
	}

	http.HandleFunc("/check-license-plate", licensePlateHandler)

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	// Optimize HTTP server settings for high load
	server := &http.Server{
		Addr:           ":" + port,
		ReadTimeout:    30 * time.Second,
		WriteTimeout:   60 * time.Second,
		IdleTimeout:    120 * time.Second,
		MaxHeaderBytes: 1 << 20, // 1 MB
	}

	log.Printf("Server starting on port %s with optimized settings...", port)
	log.Printf("Global rate limit: 200 requests/second")
	log.Printf("Per-IP rate limit: 10 requests/second")
	if err := server.ListenAndServe(); err != nil {
		log.Fatalf("Server failed to start: %v", err)
	}
}
