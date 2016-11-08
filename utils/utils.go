package utils

import (
  "unicode"
  "strings"
)

func LowerAndStripPunct(s string) string {
  return strings.Map(func (r rune) rune {
    if unicode.IsPunct(r) == true {
      return -1
    }
    return r
  }, strings.ToLower(s))
}