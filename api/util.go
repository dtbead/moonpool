package api

import (
	"crypto/rand"
	"fmt"
	"regexp"
)

func byteToHex(b []byte) string {
	return fmt.Sprintf("%x", b)
}

func isValidHash(b []byte, length int) bool {
	if b == nil || len(b) != length {
		return false
	}
	return true
}

func deleteWhitespace(s string) string {
	whitespace, _ := regexp.Compile(`[\t\r\n]+`)
	excessSpaces, _ := regexp.Compile(`[ ]{2,}`)

	s = whitespace.ReplaceAllString(s, " ")
	s = excessSpaces.ReplaceAllString(s, " ")
	return s
}

func randomString(length int) string {
	b := make([]byte, length+2)
	rand.Read(b)
	return fmt.Sprintf("%x", b)[2 : length+2]
}
