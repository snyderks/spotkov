// Package tools holds various utility functions that can be used for
// multiple purposes as well as constant values to be used elsewhere in the
// application.
package tools

import (
	"strings"
	"unicode"
)

// LowerAndStripNonAlphaNumeric strips any Unicode-defined characters that
// are not letters from an input string and then sets all
// alpha characters to their lowercase form.
// It also replaces hyphens with spaces. This is search-specific, so that
// portions of a hyphenated word do not get stuck together.
func LowerAndStripNonAlphaNumeric(s string) string {
	return strings.Map(func(r rune) rune {
		if r == '-' {
			return ' '
		} else if !unicode.IsLetter(r) && !unicode.IsDigit(r) && !unicode.IsSpace(r) {
			return -1
		}
		return r
	}, strings.ToLower(s))
}
