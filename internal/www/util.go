package www

import (
	"path/filepath"
	"runtime"
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
