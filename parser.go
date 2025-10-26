package main

import (
	"strings"

	"github.com/PuerkitoBio/goquery"
)

func extractViolations(doc *goquery.Document) []Violation {
	groups := doc.Find("#bodyPrint123 .form-group")
	if groups.Length() == 0 {
		return nil
	}

	var violations []Violation
	current := Violation{}
	fieldsSet := 0

	commit := func() {
		if fieldsSet > 0 {
			violations = append(violations, current)
			current = Violation{}
			fieldsSet = 0
		}
	}

	groups.Each(func(_ int, group *goquery.Selection) {
		label := normalizeLabel(group.Find(".col-md-3").Text())
		value := normalizeMultiline(group.Find(".col-md-9").Text())
		if label == "" || value == "" {
			return
		}

		switch label {
		case "bien kiem soat":
			commit()
			current.LicensePlate = value
			fieldsSet++
		case "mau bien":
			current.PlateColor = value
			fieldsSet++
		case "loai phuong tien":
			current.VehicleType = value
			fieldsSet++
		case "thoi gian vi pham":
			current.ViolationTime = value
			fieldsSet++
		case "dia diem vi pham":
			current.Location = value
			fieldsSet++
		case "hanh vi vi pham":
			current.Behavior = value
			fieldsSet++
		case "trang thai":
			current.Status = value
			fieldsSet++
		case "don vi phat hien vi pham":
			current.DetectingUnit = value
			fieldsSet++
		case "noi giai quyet vu viec":
			current.ResolutionPoint = value
			fieldsSet++
		}
	})

	commit()

	// Parse resolution_point separately from full text if not found in form-group
	if len(violations) > 0 {
		fullText := doc.Find("#bodyPrint123").Text()
		parseResolutionPoints(fullText, violations)
	}

	return violations
}

func parseResolutionPoints(fullText string, violations []Violation) {
	// Split by "Biển kiểm soát:" to get each violation block
	parts := strings.Split(fullText, "Biển kiểm soát:")

	for i := 1; i < len(parts) && i <= len(violations); i++ {
		block := parts[i]

		// Find "Nơi giải quyết vụ việc:" in this block
		resolutionIdx := strings.Index(block, "Nơi giải quyết vụ việc:")
		if resolutionIdx == -1 {
			continue
		}

		// Extract text after "Nơi giải quyết vụ việc:"
		afterResolution := block[resolutionIdx+len("Nơi giải quyết vụ việc:"):]

		// Find the end - either next "Biển kiểm soát" or end of string
		endIdx := len(afterResolution)
		if nextBienIdx := strings.Index(afterResolution, "Biển kiểm soát:"); nextBienIdx != -1 {
			endIdx = nextBienIdx
		}

		resolutionText := strings.TrimSpace(afterResolution[:endIdx])

		// Extract only addresses (lines starting with "Địa chỉ:")
		addresses := extractAddresses(resolutionText)
		violations[i-1].ResolutionPoint = addresses
	}
}

func extractAddresses(text string) string {
	lines := strings.Split(text, "\n")
	var addresses []string

	for _, line := range lines {
		line = strings.TrimSpace(line)
		// Look for "Địa chỉ:" prefix
		if strings.HasPrefix(line, "Địa chỉ:") {
			addr := strings.TrimPrefix(line, "Địa chỉ:")
			addr = strings.TrimSpace(addr)
			if addr != "" {
				addresses = append(addresses, addr)
			}
		}
	}

	return strings.Join(addresses, " | ")
}
