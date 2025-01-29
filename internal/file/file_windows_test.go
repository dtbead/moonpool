//go:build windows

package file

import (
	"os"
	"reflect"
	"runtime"
	"syscall"
	"testing"
	"time"
)

func Test_DateCreated(t *testing.T) {
	tests := []struct {
		name    string
		want    time.Time
		wantErr bool
	}{
		{"generic", time.Unix(1729369799, 0).UTC(), false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if runtime.GOOS != "windows" {
				t.Skip("not running on Windows")
			}

			f, err := os.CreateTemp(t.TempDir(), "")
			if err != nil {
				t.Fatalf("failed to create temp file, %v", err)
			}
			defer f.Close()

			winTicks := unixTimeToFileTime(uint64(tt.want.Unix()))

			err = syscall.SetFileTime(syscall.Handle(f.Fd()), &winTicks, &syscall.Filetime{}, &syscall.Filetime{})
			if err != nil {
				t.Fatalf("failed to set date created, %v", err)
			}

			got, err := DateCreated(f)
			if (err != nil) != tt.wantErr {
				t.Errorf("DateCreated() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("DateCreated() = %v, want %v", got, tt.want)
			}
		})
	}
}
