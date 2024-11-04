package www

import (
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"time"
)

func timeToString(t time.Time) string {
	return t.Local().String()
}

func getProjectDirectory() string {
	_, b, _, _ := runtime.Caller(0)
	basepath := filepath.Dir(b)

	return strings.ReplaceAll(basepath, "\\", "/")
}

func stringToInt64(s string) int64 {
	archive_id, err := strconv.ParseInt(s, 10, 64)
	if err != nil {
		return -1
	}

	return archive_id
}
