package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"strings"
)

type OCRResponse struct {
	ParsedResults []struct {
		ParsedText string `json:"ParsedText"`
	} `json:"ParsedResults"`
	OCRExitCode           int    `json:"OCRExitCode"`
	IsErroredOnProcessing bool   `json:"IsErroredOnProcessing"`
	ErrorMessage          string `json:"ErrorMessage"`
}

type SubmitFormResponse struct {
	Success boolish `json:"success"`
	Href    string  `json:"href"`
	Error   string  `json:"error"`
	Details *ResultDetails
}

type boolish bool

type ResultDetails struct {
	Message    string      `json:"message,omitempty"`
	Violations []Violation `json:"violations,omitempty"`
}

type Violation struct {
	LicensePlate    string `json:"license_plate"`
	PlateColor      string `json:"plate_color"`
	VehicleType     string `json:"vehicle_type"`
	ViolationTime   string `json:"violation_time"`
	Location        string `json:"location"`
	Behavior        string `json:"behavior"`
	Status          string `json:"status"`
	DetectingUnit   string `json:"detecting_unit"`
	ResolutionPoint string `json:"resolution_point"`
}

func (b *boolish) UnmarshalJSON(data []byte) error {
	data = bytes.TrimSpace(data)
	if len(data) == 0 {
		*b = boolish(false)
		return nil
	}

	if data[0] == '"' {
		var s string
		if err := json.Unmarshal(data, &s); err != nil {
			return err
		}
		s = strings.TrimSpace(strings.ToLower(s))
		switch s {
		case "true", "1", "yes", "y":
			*b = boolish(true)
		default:
			*b = boolish(false)
		}
		return nil
	}

	var boolVal bool
	if err := json.Unmarshal(data, &boolVal); err == nil {
		*b = boolish(boolVal)
		return nil
	}

	var intVal int
	if err := json.Unmarshal(data, &intVal); err == nil {
		*b = boolish(intVal != 0)
		return nil
	}

	return fmt.Errorf("boolish: cannot parse %s", string(data))
}

func (b boolish) MarshalJSON() ([]byte, error) {
	if b {
		return []byte("true"), nil
	}
	return []byte("false"), nil
}

func (b boolish) Bool() bool {
	return bool(b)
}
