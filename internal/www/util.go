package www

import (
	"strconv"
	"time"
)

func timeToString(t time.Time) string {
	return t.Local().String()
}

// returns -1 if given invalid string
func stringToInt64(s string) int64 {
	i, err := strconv.ParseInt(s, 10, 64)
	if err != nil {
		return -1
	}

	return i
}
