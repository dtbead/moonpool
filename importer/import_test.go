package importer

import (
	"encoding/hex"
	"io"
	"os"
	"strings"
	"testing"

	"github.com/dtbead/moonpool/entry"
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
				Hash: entry.Hashes{
					MD5:    stringToHex("1998a30583dd5112bbefc59fd5e8dbbd"),
					SHA1:   stringToHex("b74e160f21cbf37a6737ad20b8c057b090fd9003"),
					SHA256: stringToHex("794d84958c6675a1b77be5f33c0bbd2996948db3c83d522f0ab6d63ead116e73"),
				},
				Timestamp: entry.Timestamp{},
				Path: entry.Path{
					FileRelative: "19/1998a30583dd5112bbefc59fd5e8dbbd.jpg",
					FilExtension: ".jpg",
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
			defer got.DeleteTemp()

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
