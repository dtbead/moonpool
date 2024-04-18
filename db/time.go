package db

import (
	"fmt"
	"time"
)

func ParseTimestamp(s string) time.Time {
	t, err := time.Parse("2006-01-02 15:04:05.000", s)
	if err != nil {
		fmt.Println(err)
		return time.Time{}
	}
	return t
}
