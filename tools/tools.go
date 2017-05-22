// Package tools holds various utility functions that can be used for
// multiple purposes as well as constant values to be used elsewhere in the
// application.
package tools

import (
	"bytes"
	"encoding/base64"
	"encoding/gob"
	"errors"
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

// ToBase64 takes an interface struct and converts it to a base64 string
// encoding a gob representation of the struct passed.
// Reference encoding/gob and encoding/base64 for more information.
// Returns an error if anything went wrong during the process.
func ToBase64(s interface{}) (string, error) {
	// Create a gob encoder and encode the songs
	buf := new(bytes.Buffer)
	g := gob.NewEncoder(buf)
	err := g.Encode(s)

	if err != nil {
		return "",
			errors.New("Couldn't encode the interface as a gob: " +
				err.Error())
	}

	// Now do the same with base64, using the gob encoding as input
	b64 := new(bytes.Buffer)
	e := base64.NewEncoder(base64.StdEncoding, b64)
	defer e.Close()
	_, err = e.Write(buf.Bytes())

	if err != nil {
		return "",
			errors.New("Couldn't encode the interface as a base64 string: " +
				err.Error())
	}

	return b64.String(), nil
}

// FromBase64 takes a base64 string and converts it into a struct
// representation, if possible, first decoding into a gob
// and then an interface passed in.
// Returns an error if something goes wrong.
func FromBase64(b64 string, s interface{}) error {
	// Decode the base64 string into a gob representation of a struct.
	b, err := base64.StdEncoding.DecodeString(b64)
	// Create a buffer
	buf := bytes.NewBuffer(b)

	if err != nil {
		return errors.New("Couldn't decode the base64 string: " +
			err.Error())
	}

	gd := gob.NewDecoder(buf)
	err = gd.Decode(s)

	if err != nil {
		return errors.New("Couldn't decode the gob into an interface: " +
			err.Error())
	}

	return nil
}
