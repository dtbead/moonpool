package file

import (
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"io"
	"os"
	"testing"
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
