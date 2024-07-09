// The MIT License (MIT)

// Copyright (c) 2017-2020 Uber Technologies Inc.

// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to deal
// in the Software without restriction, including without limitation the rights
// to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
// copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:
//
// The above copyright notice and this permission notice shall be included in all
// copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
// SOFTWARE.

package visibility

import (
	"fmt"
	"regexp"
	"strings"
)

var (
	validSearchAttributeKey = regexp.MustCompile(`^[a-zA-Z][a-zA-Z_0-9]*$`)
	nonAlphanumericRegex    = regexp.MustCompile(`[^a-zA-Z0-9]+`)
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
