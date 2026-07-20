package vchiq

import (
	"regexp"
	"strings"
)

// clean removes specified substrings from the input string and trims the result.
func clean(str string, args ...string) string {
	for _, arg := range args {
		str = strings.ReplaceAll(str, arg, "")
	}
	return strings.TrimSpace(str)
}

// extractMemorySize extracts the number and unit (M, G, or K) from a string.
// Returns a single string containing the number and letter.
func extractMemorySize(input string) string {
	re := regexp.MustCompile(`(\d+)([MGK])`)
	match := re.FindStringSubmatch(input)
	if len(match) >= 3 {
		return match[1] + match[2]
	}
	return ""
}