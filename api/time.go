package api

import "time"

// cleanTimestamp converts a time.Time to a UTC timestamp with no milisecond precision
func cleanTimestamp(t time.Time) time.Time {
	return t.UTC().Round(time.Second * 1)
}
