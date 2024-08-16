package api

import (
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
	expr, _ := regexp.Compile(`[\t\r\n ]+`)
	return expr.ReplaceAllString(s, " ")
}
