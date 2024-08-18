package db

import (
	"io"
	"os"
	"testing"

	"github.com/go-test/deep"
)

func TestNew(t *testing.T) {
	f, err := os.Open("testdata/1998a30583dd5112bbefc59fd5e8dbbd.jpg")
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
		want    Entry
		wantErr bool
	}{
		{"generic", args{f, ".jpg"}, Entry{
			file: nil,
			Metadata: Metadata{
				Hash: Hashes{
					MD5:    decodeHexString("1998a30583dd5112bbefc59fd5e8dbbd"),
					SHA1:   decodeHexString("794d84958c6675a1b77be5f33c0bbd2996948db3c83d522f0ab6d63ead116e73"),
					SHA256: decodeHexString("fd9858bada2040b7020b2109abc6b63c99105138d1bb13c0fd16ea1e538a1975350839f54c85adfc46e9b4e6165e9d0cc8a9d067f0bb387b6702a97d0ed221c8"),
				},
				Timestamp:    Timestamp{},
				PathRelative: "19/1998a30583dd5112bbefc59fd5e8dbbd.jpg",
				Extension:    ".jpg",
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
			defer got.DeleteTemp()

			if diff := deep.Equal(got, tt.want); diff != nil {
				t.Errorf("New() = %v, want %v", got, tt.want)
			}
		})
	}
}
