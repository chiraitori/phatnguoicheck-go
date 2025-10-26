package main

import (
"log"
"net/http"
"os"

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

log.Printf("Server starting on port %s...", port)
if err := http.ListenAndServe(":"+port, nil); err != nil {
log.Fatalf("Server failed to start: %v", err)
}
}

