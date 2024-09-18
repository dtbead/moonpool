package api

import "time"

type Timestamp struct {
	DateModified time.Time
	DateImported time.Time
	DateCreated  time.Time
}

func (t Timestamp) UTC() Timestamp {
	return Timestamp{
		DateModified: t.DateModified.UTC(),
		DateImported: t.DateImported.UTC(),
		DateCreated:  t.DateCreated.UTC(),
	}
}

// timeToRFC3339_UTC() returns a RFC3339 string-formatted timestamp in UTC timezone
func timeToRFC3339_UTC(t time.Time) string {
	return t.UTC().Format(time.RFC3339)
}
