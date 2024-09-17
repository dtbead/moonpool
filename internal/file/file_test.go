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
	testFile, err := os.Open("testdata/532e58065afad25d587073caf3236af9eb47ceba5ed0c0daaf8b33d8ed50a82b.png")
	if err != nil {
		t.Fatalf("failed to open test file, %v\n", err)
	}

	type args struct {
		f *os.File
	}
	tests := []struct {
		name    string
		args    args
		want    time.Time
		wantErr bool
	}{
		{"generic", args{testFile}, time.Unix(1515930174, 0), false}, // 2018 Jan 14th 11:42:54 UTC
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := DateModified(tt.args.f)
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
	f, err := os.Create("testdata/hawk.png")
	if err != nil {
		t.Fatalf("doesPathExist() error. unable to create temporary file. %v", err)
	}
	defer f.Close()

	type args struct {
		path string
	}
	tests := []struct {
		name string
		args args
		want bool
	}{
		{"exists directory", args{path: "testdata"}, true},
		{"exists file", args{path: "testdata/hawk.png"}, true},
		{"not exists directory", args{path: "thispathdoesntexist"}, false},
		{"not exists file", args{path: "testdata/thisfiledoesnotexist.png"}, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := doesPathExist(tt.args.path); got != tt.want {
				t.Errorf("doesPathExist() = %v, want %v", got, tt.want)
			}
		})
	}

	t.Cleanup(func() {
		if err := os.Remove("testdata/hawk.png"); err != nil {
			t.Fatalf("doesPathExist() error = unable to delete temporary testdata, %v", err)
		}
	})
}
func TestNewStorage(t *testing.T) {
	type args struct {
		rootPath string
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{"valid", args{"testdata/tmp/testpath"}, false},
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

			t.Cleanup(func() {
				err := os.Remove(p)
				if err != nil {
					t.Fatal(err)
				}
			})
		})
	}

}

func TestBuildPath(t *testing.T) {
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

func TestGetHash(t *testing.T) {
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
