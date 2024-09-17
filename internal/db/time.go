package db

import (
	"time"
)

type Timestamp struct {
	DateCreated  time.Time
	DateModified time.Time
	DateImported time.Time
}

func (t Timestamp) UTC() Timestamp {
	return Timestamp{
		DateCreated:  t.DateCreated.UTC(),
		DateModified: t.DateModified.UTC(),
		DateImported: t.DateImported.UTC(),
	}
}

// ParseTimestamp parses a RFC3339 timestamp to time.Time
func ParseTimestamp(s string) (time.Time, error) {
	t, err := time.Parse(time.RFC3339, s)
	if err != nil {
		return time.Time{}, err
	}
	return t, nil
}

// returns a RFC3339 string-formatted UTC timestamp
func timeToRFC3339_UTC(t time.Time) string {
	return t.UTC().Round(time.Second * 1).Format(time.RFC3339)
}
