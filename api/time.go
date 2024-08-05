package api

import "time"

// timeToUnixEpoch converts a time.Time to a UnixEpoch timestamp rounded to the nearest second
func timeToUnixEpoch(t time.Time) time.Time {
	return t.Round(time.Second * 1)
}
