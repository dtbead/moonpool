package file

import (
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"io"
	"os"
	"reflect"
	"testing"

	"github.com/dtbead/moonpool/media"
)

type mockFile struct {
	f          *os.File
	i          os.FileInfo
	hash       []byte
	hashString string
}

const testFilePath = "testdata/hawk.png"

var testFile mockFile

func TestMain(m *testing.M) {
	var err error

	testFile.f, err = os.Open(testFilePath)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	defer testFile.f.Close()

	data := []byte{}
	testFile.f.Read(data)
	h := md5.New()
	testFile.hash = h.Sum(data)
	testFile.hashString = hex.EncodeToString(testFile.hash[:])

	testFile.i, err = os.Stat(testFilePath)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	code := m.Run()
	os.Exit(code)
}

func TestBuildPath(t *testing.T) {
	type args struct {
		HashString string
		extension  string
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{"generic", args{testFile.hashString, "png"}, "d4/d41d8cd98f00b204e9800998ecf8427e.png"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := BuildPath(tt.args.HashString, tt.args.extension); got != tt.want {
				t.Errorf("BuildPath() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_copy(t *testing.T) {
	type args struct {
		destination string
		r           io.Reader
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{"generic", args{"testdata/tmp/hawkcopy.png", testFile.f}, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := copy(tt.args.destination, tt.args.r); (err != nil) != tt.wantErr {
				t.Errorf("copy() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestGetHashes(t *testing.T) {
	type args struct {
		r io.Reader
	}
	tests := []struct {
		name    string
		args    args
		want    media.Hashes
		wantErr bool
	}{
		{"valid", args{testFile.f}, media.Hashes{
			MD5:    []byte{212, 29, 140, 217, 143, 0, 178, 4, 233, 128, 9, 152, 236, 248, 66, 126},
			SHA1:   []byte{218, 57, 163, 238, 94, 107, 75, 13, 50, 85, 191, 239, 149, 96, 24, 144, 175, 216, 7, 9},
			SHA256: []byte{227, 176, 196, 66, 152, 252, 28, 20, 154, 251, 244, 200, 153, 111, 185, 36, 39, 174, 65, 228, 100, 155, 147, 76, 164, 149, 153, 27, 120, 82, 184, 85},
		}, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := GetHashes(tt.args.r)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetHashes() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("GetHashes() = %v, want %v", got, tt.want)
			}
		})
	}
}
