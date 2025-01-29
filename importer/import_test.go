package importer

import (
	"encoding/hex"
	"fmt"
	"io"
	"os"
	"strings"
	"testing"

	"github.com/dtbead/moonpool/entry"
	"github.com/dtbead/moonpool/internal/file"
	"github.com/go-test/deep"
)

func Test_New(t *testing.T) {
	f, err := os.Open("testdata/6ba11adbdb35ee10f9353608a7b97ef248733a72.jpg")
	if err != nil {
		t.Fatalf("failed to open test file. %v", err)
	}

	type args struct {
		r         io.Reader
		extension string
	}
	tests := []struct {
		name    string
		args    args
		want    Importer
		wantErr bool
	}{
		{"generic", args{f, ".jpg"}, Importer{
			file: nil,
			e: entry.Entry{
				Metadata: entry.Metadata{
					Hash: entry.Hashes{
						MD5:    stringToHex("2ec268313d4d0bbc765144b6334df68b"),
						SHA1:   stringToHex("6ba11adbdb35ee10f9353608a7b97ef248733a72"),
						SHA256: stringToHex("7aaa7471fed00d0bcb416f123d364ec28a9080708601bd308cc4301d3fadb0e1"),
					},
					Timestamp: entry.Timestamp{},
					Paths: entry.Path{
						FileRelative:  "2e/2ec268313d4d0bbc765144b6334df68b.jpg",
						FileExtension: ".jpg",
					},
				},
			},
		}, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := New(tt.args.r, tt.args.extension)
			if (err != nil) != tt.wantErr {
				t.Errorf("New() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if diff := deep.Equal(got, tt.want); diff != nil {
				t.Errorf("New() diff = %s", strings.Join(diff, "\n"))
			}
		})
	}
}

func stringToHex(s string) []byte {
	b, _ := hex.DecodeString(s)
	return b
}

func TestImporter_Store(t *testing.T) {
	f, err := os.Open("testdata/6ba11adbdb35ee10f9353608a7b97ef248733a72.jpg")
	if err != nil {
		t.Fatalf("failed to open test file. %v", err)
	}
	defer f.Close()

	importer, err := New(f, ".jpg")
	if err != nil {
		t.Fatalf("failed to create new importer, %v", err)
	}

	type args struct {
		baseDirectory string
	}
	tests := []struct {
		name    string
		i       Importer
		args    args
		wantErr bool
	}{
		{"generic", importer, args{t.TempDir()}, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if _, err := tt.i.Store(tt.args.baseDirectory); (err != nil) != tt.wantErr {
				t.Errorf("Importer.Store() error = %v, wantErr %v", err, tt.wantErr)
			}

			dest := file.CleanPath(fmt.Sprintf("%s/%s", tt.args.baseDirectory, tt.i.e.Metadata.Paths.FileRelative))

			file, err := os.Open(dest)
			if err != nil {
				t.Fatalf("failed to open stored file at %s. %v", dest, err)
			}
			defer file.Close()

			fileStat, err := file.Stat()
			if err != nil {
				t.Fatalf("failed to get stored file stat, %v", err)
			}
			gotBytes := fileStat.Size()

			if f, ok := tt.i.file.(*os.File); ok {
				f.Seek(0, io.SeekStart)
			}

			wantBytes, err := io.Copy(io.Discard, tt.i.file)
			if err != nil {
				t.Fatalf("failed to get total size from importer file, %v", err)
			}

			if gotBytes != wantBytes {
				t.Errorf("importer returned %d bytes, expected %d bytes", gotBytes, wantBytes)
			}
		})
	}
}
