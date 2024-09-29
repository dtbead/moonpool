//go:build windows
// +build windows

package file

import (
	"errors"
	"os"
	"runtime"
	"syscall"
	"time"
)

// DateCreatedWindows() returns the UTC time of the date created of a file on a Windows machine only.
func DateCreatedWindows(f *os.File) (time.Time, error) {
	if runtime.GOOS != "windows" {
		return DateModified(f)
	}

	fi, err := f.Stat()
	if err != nil {
		return time.Time{}, err
	}

	d := fi.Sys().(*syscall.Win32FileAttributeData)
	if d == nil {
		return time.Time{}, errors.New("failed to get Win32FileAttributeData")
	}

	winEpochMilli := d.CreationTime.Nanoseconds() / int64(time.Millisecond)
	return time.UnixMilli(winEpochMilli).UTC(), nil
}
