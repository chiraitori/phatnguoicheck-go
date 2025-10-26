package main

import (
	"encoding/json"
	"net/http"
)

func licensePlateHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)
		return
	}

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
