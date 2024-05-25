package archive

import (
	"time"
)

// ParseTimestamp parses a timestamp from an RFC3339 standard time format to time.Time
func ParseTimestamp(s string) (time.Time, error) {
	t, err := time.Parse(time.RFC3339, s)
	if err != nil {
		return time.Time{}, err
	}
	return t, nil
}

func ToRFC3339_UTC_Timestamp(t time.Time) string {
	return t.UTC().Format(time.RFC3339)
}
