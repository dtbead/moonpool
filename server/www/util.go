package www

import "time"

func timeToString(t time.Time) string {
	return t.Local().String()
}
