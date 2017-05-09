package tools

import "testing"

// TestLowerAndStripNonAlphaNumeric checks that the method:
// leaves only spaces, numbers, and letters as well as
// making hyphens into space characters (U+0020)
func TestLowerAndStripNonAlphaNumeric(T *testing.T) {
	strs := map[string]string{
		",.h/;el'[]Lo=": "hello",
		"1'000'000.00":  "100000000",
		"false":         "false",
		"nil":           "nil",
		// should replace hyphen with space.
		"tHe-best one isThIs": "the best one isthis",
	}
	for key, value := range strs {
		s := LowerAndStripNonAlphaNumeric(key)
		if s != value {
			T.Error("String", key, "was supposed to be changed to", value+".",
				"It was changed to", s, "instead.")
		}
	}
}
