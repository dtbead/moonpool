package www

import (
	"strconv"
	"time"
)

const dateFormat = "Jan 2 2006, 3:04:05 PM"

func timeToString(t time.Time) string {
	return t.Format(dateFormat)
}

// returns -1 if given invalid string
func stringToInt64(s string) int64 {
	i, err := strconv.ParseInt(s, 10, 64)
	if err != nil {
		return -1
	}

	return i
}

func add(n, s int64) int64 {
	if n+s <= 0 {
		return 0
	}
	return n + s
}
