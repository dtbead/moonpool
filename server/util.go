package server

import (
	"fmt"
)

func byteToHex(b []byte) string {
	return fmt.Sprintf("%x", b)
}
