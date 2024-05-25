package server

import (
	"fmt"
	"strconv"
)

func byteToHex(b []byte) string {
	return fmt.Sprintf("%x", b)
}

// returns -1 on invalid id string. returns int64 >= 1 otherwise
func (m Moonpool) parseArchiveID(id string) int64 {
	archive_id, err := strconv.ParseInt(id, 10, 64)
	if err != nil {
		return -1
	}

	return archive_id
}
