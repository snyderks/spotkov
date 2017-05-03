// Package tools holds various utility functions that can be used for
// multiple purposes as well as constant values to be used elsewhere in the
// application.
package tools

import (
	"strings"
	"unicode"
)

// LowerAndStripPunct strips any Unicode-defined punctuation from
// an input string and then sets all alpha characters to their
// lowercase form.
func LowerAndStripPunct(s string) string {
	return strings.Map(func(r rune) rune {
		if unicode.IsPunct(r) {
			return -1
		}
		return r
	}, strings.ToLower(s))
}
