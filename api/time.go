package api

import "time"

// timeToUnixEpoch converts a time.Time to a Unix Epoch timestamp rounded to the nearest second
func timeToUnixEpoch(t time.Time) time.Time {
	return t.UTC().Round(time.Second * 1)
}

// returns a RFC3339 string-formatted timestamp
func timeToRFC3339_UTC(t time.Time) string {
	return t.UTC().Format(time.RFC3339)
}

type Timestamp struct {
	DateCreated  time.Time
	DateModified time.Time
	DateImported time.Time
}
