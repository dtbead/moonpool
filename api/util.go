package api

import "fmt"

func byteToHex(b []byte) string {
	return fmt.Sprintf("%x", b)
}

func isValidHash(b []byte, length int) bool {
	if b == nil || len(b) != length {
		return false
	}
	return true
}
