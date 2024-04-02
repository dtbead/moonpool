package file

import (
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"io"
	"os"
	"reflect"
	"testing"
	"time"

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
		{"valid", args{testFile.hashString, ".png"}, "d4/d41d8cd98f00b204e9800998ecf8427e.png"},
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

func TestCopyAndHash(t *testing.T) {
	type args struct {
		destination string
		extension   string
		r           io.Reader
	}
	tests := []struct {
		name    string
		args    args
		want    media.Entry
		wantErr bool
	}{
		{"valid", args{"testdata/tmp/h/hawk.png", ".png", testFile.f}, media.Entry{ArchiveID: 0, Metadata: media.Metadata{
			MD5Hash:      "d41d8cd98f00b204e9800998ecf8427e",
			PathDirect:   "testdata/tmp/h/hawk.png/d4/d41d8cd98f00b204e9800998ecf8427e.png",
			PathRelative: "d4/d41d8cd98f00b204e9800998ecf8427e.png",
			Extension:    ".png",
			Hash: media.Hashes{
				MD5:    []byte{212, 29, 140, 217, 143, 0, 178, 4, 233, 128, 9, 152, 236, 248, 66, 126},
				SHA1:   []byte{218, 57, 163, 238, 94, 107, 75, 13, 50, 85, 191, 239, 149, 96, 24, 144, 175, 216, 7, 9},
				SHA256: []byte{227, 176, 196, 66, 152, 252, 28, 20, 154, 251, 244, 200, 153, 111, 185, 36, 39, 174, 65, 228, 100, 155, 147, 76, 164, 149, 153, 27, 120, 82, 184, 85},
			},
			Timestamp: media.Timestamp{DateImported: time.Now()},
		}}, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := CopyAndHash(tt.args.destination, tt.args.extension, tt.args.r)
			if (err != nil) != tt.wantErr {
				t.Errorf("CopyAndHash() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("CopyAndHash() = %v, want %v", got, tt.want)
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
			if got, err := exists(p); (got == false) && tt.wantErr == true || (err != nil) != tt.wantErr {
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

func exists(path string) (bool, error) {
	_, err := os.Stat(path)
	if err == nil {
		return true, nil
	}
	if os.IsNotExist(err) {
		return false, nil
	}
	return false, err
}
