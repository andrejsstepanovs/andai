package util

import (
	"regexp"
)

// ExtractCommitSHAs extracts unique Git commit SHAs from comment text
func ExtractCommitSHAs(commentText string) []string {
	// Regex pattern for matching Git commit SHAs (40-char hexadecimal)
	pattern := regexp.MustCompile(`(?i)\b[0-9a-f]{40}\b`)

	// Find all matches
	matches := pattern.FindAllString(commentText, -1)

	// Use map to ensure uniqueness
	uniqueSHAs := make(map[string]struct{})
	for _, sha := range matches {
		uniqueSHAs[sha] = struct{}{}
	}

	result := make([]string, len(uniqueSHAs))
	for sha := range uniqueSHAs {
		result = append(result, sha)
	}

	return result
}
