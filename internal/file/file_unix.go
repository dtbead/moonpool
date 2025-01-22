//go:build unix

package file

import (
	"os"
	"time"
)

// DateCreated returns the UTC time of a date created on a file. If not running on Windows, DateCreated simply returns
// DateModified instead
func DateCreated(f *os.File) (time.Time, error) {
	return DateModified(f)
}
