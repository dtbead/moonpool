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

// timeToRFC3339_UTC returns a RFC3339 string-formatted UTC timestamp rounded to the nearest second
func TimeToRFC3339_UTC(t time.Time) string {
	return t.UTC().Round(time.Second * 1).Format(time.RFC3339)
}
