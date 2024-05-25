package api

import "time"

// cleanTimestamp converts a time.Time to a UTC timestamp with no milisecond precision
func cleanTimestamp(t time.Time) time.Time {
	return t.UTC().Truncate(time.Second * 1)
}
