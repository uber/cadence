package validator

import (
	"fmt"
	"regexp"
	"strings"
)

var (
	validSearchAttributeKey = regexp.MustCompile(`^[a-zA-Z][a-zA-Z_0-9]*$`)
	nonAlphanumericRegex = regexp.MustCompile(`[^a-zA-Z0-9]+`)
)

// ValidateSearchAttributeKey checks if the search attribute key has valid format
func ValidateSearchAttributeKey(name string) error {
	if !validSearchAttributeKey.MatchString(name) {
		return fmt.Errorf("has to be contain alphanumeric and start with a letter")
	}
	return nil
}

// SanitizeSearchAttributeKey try to sanitize the search attribute key
func SanitizeSearchAttributeKey(name string) (string, error) {
	sanitized := nonAlphanumericRegex.ReplaceAllString(name, "_")
	// remove leading and trailing underscores
	sanitized = strings.Trim(sanitized, "_")
	// remove leading numbers
	sanitized = strings.TrimLeft(sanitized, "0123456789")
	return sanitized, ValidateSearchAttributeKey(sanitized)
}
