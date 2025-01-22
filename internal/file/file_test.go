package file

import (
	"encoding/hex"
	"fmt"
	"io"
	"os"
	"reflect"
	"strings"
	"testing"
	"time"
)

func Test_DateModified(t *testing.T) {
	tests := []struct {
		name    string
		want    time.Time
		wantErr bool
	}{
		{"generic", time.Unix(1515930174, 0), false}, // 2018 Jan 14th 11:42:54 UTC
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			f, err := os.CreateTemp(t.TempDir(), "")
			if err != nil {
				t.Fatalf("failed to set create temporary file, %v", err)
			}
			defer f.Close()
			defer os.Remove(f.Name())

			if err := os.Chtimes(f.Name(), time.Time{}, tt.want); err != nil {
				t.Fatalf("failed to set date modified, %v", err)
			}

			got, err := DateModified(f)
			if (err != nil) != tt.wantErr {
				t.Errorf("DateModified() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want.UTC()) {
				t.Errorf("DateModified() = %v, want %v", got, tt.want.UTC())
			}
		})
	}
}

func Test_doesPathExist(t *testing.T) {
	type args struct {
		path string
	}
	tests := []struct {
		name string
		args args
		want bool
	}{
		{"exists directory", args{path: "testdata"}, true},
		{"exists file", args{path: "testdata/6ba11adbdb35ee10f9353608a7b97ef248733a72.jpg"}, true},
		{"not exists directory", args{path: "thispathdoesntexist"}, false},
		{"not exists file", args{path: "testdata/thisfiledoesnotexist.png"}, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := DoesPathExist(tt.args.path); got != tt.want {
				t.Errorf("doesPathExist() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_NewStorage(t *testing.T) {
	type args struct {
		rootPath string
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{"valid", args{t.TempDir() + "/testpath"}, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := NewStorage(tt.args.rootPath); (err != nil) != tt.wantErr {
				t.Errorf("NewStorage() error = %v, wantErr %v", err, tt.wantErr)
			}

			p := fmt.Sprintf("%s/db/media/storage", tt.args.rootPath)
			if got, err := exists(t, p); (got == false) && tt.wantErr == true || (err != nil) != tt.wantErr {
				t.Errorf("NewStorage() error = %v, wantErr %v. path = %v", err, tt.wantErr, got)
			}
		})
	}

}

func Test_BuildPath(t *testing.T) {
	type args struct {
		md5       []byte
		extension string
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{"generic", args{md5: []byte{91, 115, 3, 1, 18, 87, 5, 166, 60, 160, 100, 218, 24, 159, 125, 80}, extension: ".png"}, "5b/5b730301125705a63ca064da189f7d50.png"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := BuildPath(tt.args.md5, tt.args.extension); got != tt.want {
				t.Errorf("BuildPath() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_GetHash(t *testing.T) {
	bMD5, _ := hex.DecodeString("3858f62230ac3c915f300c664312c63f")
	bSHA1, _ := hex.DecodeString("8843d7f92416211de9ebb963ff4ce28125932878")
	bSHA256, _ := hex.DecodeString("c3ab8ff13720e8ad9047dd39466b3c8974e592c2fa383d4a3960714caef0c4f2")

	h := Hashes{
		MD5:    bMD5,
		SHA1:   bSHA1,
		SHA256: bSHA256,
	}

	type args struct {
		r io.Reader
	}
	tests := []struct {
		name    string
		args    args
		want    Hashes
		wantErr bool
	}{
		{"generic", args{strings.NewReader("foobar")}, h, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := GetHash(tt.args.r)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetHash() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("GetHash() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_unixTimeToWindowsTicks(t *testing.T) {
	type args struct {
		unix uint64
	}
	tests := []struct {
		name string
		args args
		want uint64
	}{
		{"generic", args{1729397600}, 0x1DB22A6657A7000},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := unixTimeToWindowsTicks(tt.args.unix); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("unixTimeToWindowsTicks() = %v, want %v", got, tt.want)
			}
		})
	}
}

func exists(t *testing.T, path string) (bool, error) {
	t.Helper()
	_, err := os.Stat(path)
	if err == nil {
		return true, nil
	}
	if os.IsNotExist(err) {
		return false, nil
	}
	return false, err
}
