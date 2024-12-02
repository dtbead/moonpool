//go:build windows

package file

import (
	"os"
	"syscall"
	"time"
)

// DateCreated returns the UTC time of a date created on a file. If not running on Windows, DateCreated simply returns
// DateModified instead
func DateCreated(f *os.File) (time.Time, error) {
	fileInfo := new(syscall.ByHandleFileInformation)
	if err := syscall.GetFileInformationByHandle(syscall.Handle(f.Fd()), fileInfo); err != nil {
		return time.Time{}, err
	}

	return time.Unix(0, fileInfo.CreationTime.Nanoseconds()).UTC(), nil
}

func unixTimeToFileTime(unix uint64) syscall.Filetime {
	unix = (unix * 10000000) + 116444736000000000
	return syscall.Filetime{
		LowDateTime:  uint32(unix & 0xFFFFFFFF),
		HighDateTime: uint32(unix >> 32),
	}
}
