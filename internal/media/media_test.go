package media

import (
	_ "image/gif"
	_ "image/png"
	"io"
	"os"
	"reflect"
	"testing"
)

func TestGetMediaDimensions(t *testing.T) {
	fileLandscape, err := os.Open("testdata/6ba11adbdb35ee10f9353608a7b97ef248733a72.jpg")
	if err != nil {
		t.Fatalf("failed to open test file, %v", err)
	}

	filePortrait, err := os.Open("testdata/b1c4670ded937f2a61cccb3e0d95117d18e7cfd8.jpg")
	if err != nil {
		t.Fatalf("failed to open test file, %v", err)
	}

	type args struct {
		media io.Reader
	}
	tests := []struct {
		name    string
		args    args
		want    struct{ Width, Height int }
		wantErr bool
	}{
		{"generic jpg landscape", args{fileLandscape}, struct {
			Width  int
			Height int
		}{3000, 1993}, false},
		{"generic jpg portrait", args{filePortrait}, struct {
			Width  int
			Height int
		}{853, 1280}, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := GetMediaDimensions(tt.args.media)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetMediaDimensions() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("GetMediaDimensions() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestIsLandscape(t *testing.T) {
	fileLandscape, err := os.Open("testdata/6ba11adbdb35ee10f9353608a7b97ef248733a72.jpg")
	if err != nil {
		t.Fatalf("failed to open test file, %v", err)
	}

	filePortrait, err := os.Open("testdata/b1c4670ded937f2a61cccb3e0d95117d18e7cfd8.jpg")
	if err != nil {
		t.Fatalf("failed to open test file, %v", err)
	}

	type args struct {
		media io.Reader
	}
	tests := []struct {
		name    string
		args    args
		want    bool
		wantErr bool
	}{
		{"generic jpg landscape", args{fileLandscape}, true, false},
		{"generic jpg portrait", args{filePortrait}, false, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := IsLandscape(tt.args.media)
			if (err != nil) != tt.wantErr {
				t.Errorf("IsLandscape() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("IsLandscape() = %v, want %v", got, tt.want)
			}
		})
	}
}
