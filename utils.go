package main

import (
	"strings"
	"unicode"

	"golang.org/x/text/unicode/norm"
)

func normalizeLabel(s string) string {
	s = strings.ReplaceAll(s, "\u00a0", " ")
	s = strings.Join(strings.Fields(s), " ")
	s = strings.TrimSuffix(s, ":")
	s = strings.TrimSpace(s)
	s = strings.ToLower(removeDiacritics(s))
	return s
}

func normalizeMultiline(s string) string {
	s = strings.ReplaceAll(s, "\u00a0", " ")
	s = strings.ReplaceAll(s, "\r\n", "\n")
	s = strings.ReplaceAll(s, "\r", "\n")
	lines := strings.Split(s, "\n")
	cleaned := make([]string, 0, len(lines))
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		line = strings.Join(strings.Fields(line), " ")
		if line != "" {
			cleaned = append(cleaned, line)
		}
	}
	return strings.Join(cleaned, "\n")
}

func removeDiacritics(s string) string {
	// Replace Vietnamese đ/Đ first
	s = strings.ReplaceAll(s, "đ", "d")
	s = strings.ReplaceAll(s, "Đ", "D")
	
	decomposed := norm.NFD.String(s)
	builder := strings.Builder{}
	for _, r := range decomposed {
		if unicode.Is(unicode.Mn, r) {
			continue
		}
		builder.WriteRune(r)
	}
	return builder.String()
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func getViolationCount(details *ResultDetails) int {
	if details == nil {
		return 0
	}
	return len(details.Violations)
}
