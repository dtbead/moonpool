package file

import (
	"bytes"
	"crypto/md5"
	"fmt"
	"io"
	"os"
	"reflect"
	"testing"
	"time"

	"github.com/dtbead/moonpool/media"
	"github.com/go-test/deep"
)

type mockFile struct {
	f    *os.File
	i    os.FileInfo
	data []byte
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

	// golang doesn't return anything if io.reader has already been read from.
	// this should be replaced with a better solution that allows reading
	// multiple files for when we need to start benchmarking.
	testFile.data, _ = io.ReadAll(testFile.f)

	h := md5.New()
	w, err := io.Copy(h, bytes.NewReader(testFile.data))
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	fmt.Println(w)

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
		{"valid", args{"d41d8cd98f00b204e9800998ecf8427e", ".png"}, "d4/d41d8cd98f00b204e9800998ecf8427e.png"},
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
		{"generic", args{"testdata/tmp/hawkcopy.png", bytes.NewReader(testFile.data)}, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := copy(tt.args.destination, tt.args.r); (err != nil) != tt.wantErr {
				t.Errorf("copy() error = %v, wantErr %v", err, tt.wantErr)
			}
		})

		t.Cleanup(func() {
			if err := os.RemoveAll(tt.args.destination); err != nil {
				t.Fatalf("CopyAndHash() cleanup fail! %v", err)
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
		{"valid", args{bytes.NewReader(testFile.data)}, media.Hashes{
			MD5:    []byte{55, 81, 125, 90, 38, 13, 211, 80, 153, 248, 220, 184, 100, 176, 181, 167},
			SHA1:   []byte{4, 78, 247, 0, 76, 227, 44, 154, 25, 69, 82, 229, 131, 181, 188, 150, 1, 247, 178, 90},
			SHA256: []byte{121, 137, 145, 218, 146, 180, 124, 100, 101, 250, 37, 118, 62, 172, 125, 140, 90, 243, 239, 253, 109, 70, 9, 110, 9, 137, 25, 152, 173, 202, 83, 76},
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
		{"valid", args{"testdata/tmp/copyandhash", ".png", bytes.NewReader(testFile.data)}, media.Entry{ArchiveID: 0, Metadata: media.Metadata{
			MD5Hash:      "37517d5a260dd35099f8dcb864b0b5a7",
			PathDirect:   "testdata/tmp/copyandhash/37/37517d5a260dd35099f8dcb864b0b5a7.png",
			PathRelative: "37/37517d5a260dd35099f8dcb864b0b5a7.png",
			Extension:    ".png",
			Hash: media.Hashes{
				MD5:    []byte{55, 81, 125, 90, 38, 13, 211, 80, 153, 248, 220, 184, 100, 176, 181, 167},
				SHA1:   []byte{4, 78, 247, 0, 76, 227, 44, 154, 25, 69, 82, 229, 131, 181, 188, 150, 1, 247, 178, 90},
				SHA256: []byte{121, 137, 145, 218, 146, 180, 124, 100, 101, 250, 37, 118, 62, 172, 125, 140, 90, 243, 239, 253, 109, 70, 9, 110, 9, 137, 25, 152, 173, 202, 83, 76},
			},
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
				s := deep.Equal(got, tt.want)
				t.Errorf("CopyAndHash() not equal. %v", s)
			}

			t.Cleanup(func() {
				if err := os.RemoveAll(tt.args.destination); err != nil {
					t.Fatalf("CopyAndHash() cleanup fail! %v", err)
				}

				if err := os.RemoveAll(tt.want.Metadata.PathDirect); err != nil {
					t.Fatalf("CopyAndHash() cleanup fail! %v", err)
				}
			})
		})
	}
}

func TestGetDateModified(t *testing.T) {
	testFileTime, _ := time.Parse(time.RFC3339, "2018-01-12T13:12:49Z")
	type args struct {
		f *os.File
	}
	tests := []struct {
		name    string
		args    args
		want    time.Time
		wantErr bool
	}{
		{"valid", args{testFile.f}, testFileTime.UTC(), false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := GetDateModified(tt.args.f)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetDateModified() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got.UTC(), tt.want) {
				t.Errorf("GetDateModified() = %v, want %v", got, tt.want)
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
