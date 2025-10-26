package main

import (
	"encoding/json"
	"net/http"
	"strings"
)

func getClientIP(r *http.Request) string {
	// Check X-Forwarded-For header first (for proxies/load balancers)
	forwarded := r.Header.Get("X-Forwarded-For")
	if forwarded != "" {
		// Get first IP in the chain
		ips := strings.Split(forwarded, ",")
		return strings.TrimSpace(ips[0])
	}
	
	// Check X-Real-IP header
	realIP := r.Header.Get("X-Real-IP")
	if realIP != "" {
		return realIP
	}
	
	// Fall back to RemoteAddr
	ip := r.RemoteAddr
	// Remove port if present
	if colon := strings.LastIndex(ip, ":"); colon != -1 {
		ip = ip[:colon]
	}
	return ip
}

func licensePlateHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)
		return
	}

	// Apply per-IP rate limiting
	clientIP := getClientIP(r)
	limiter := ipRateLimiter.GetLimiter(clientIP)
	limiter.Wait()

	var requestData struct {
		LicensePlate string `json:"license_plate"`
		VehicleType  string `json:"vehicle_type"`
	}

	if err := json.NewDecoder(r.Body).Decode(&requestData); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	result, attempts, err := checkLicensePlate(requestData.LicensePlate, requestData.VehicleType)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	response := struct {
		Success        bool           `json:"success"`
		Href           string         `json:"href"`
		Error          string         `json:"error"`
		Attempts       int            `json:"attempts"`
		ViolationCount int            `json:"violation_count"`
		Details        *ResultDetails `json:"details,omitempty"`
	}{
		Success:        result.Success.Bool(),
		Href:           result.Href,
		Error:          result.Error,
		Attempts:       attempts,
		ViolationCount: getViolationCount(result.Details),
		Details:        result.Details,
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(response); err != nil {
		http.Error(w, "Error encoding response", http.StatusInternalServerError)
	}
}

