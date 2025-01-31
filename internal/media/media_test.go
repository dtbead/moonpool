package media

import (
	_ "image/gif"
	_ "image/png"
	"io"
	"os"
	"reflect"
	"testing"
)

func TestGetDimensions(t *testing.T) {
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
		want    struct{ Width, Height int64 }
		wantErr bool
	}{
		{"generic jpg landscape", args{fileLandscape}, struct {
			Width  int64
			Height int64
		}{3000, 1993}, false},
		{"generic jpg portrait", args{filePortrait}, struct {
			Width  int64
			Height int64
		}{853, 1280}, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := GetDimensions(tt.args.media)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetDimensions() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("GetDimensions() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestGetOrientation(t *testing.T) {
	fileLandscape, err := os.Open("testdata/6ba11adbdb35ee10f9353608a7b97ef248733a72.jpg")
	if err != nil {
		t.Fatalf("failed to open test file, %v", err)
	}

	filePortrait, err := os.Open("testdata/b1c4670ded937f2a61cccb3e0d95117d18e7cfd8.jpg")
	if err != nil {
		t.Fatalf("failed to open test file, %v", err)
	}

	fileSquare, err := os.Open("testdata/3880f791e71d8141a71001c8678e28324fa21205.jpg")
	if err != nil {
		t.Fatalf("failed to open test file, %v", err)
	}
	defer fileSquare.Close()

	videoLandscape, err := os.Open("testdata/testsrc.mp4")
	if err != nil {
		t.Fatalf("failed to open test file, %v", err)
	}
	defer videoLandscape.Close()

	type args struct {
		media io.Reader
	}
	tests := []struct {
		name    string
		args    args
		want    int
		wantErr bool
	}{
		{"generic jpg landscape", args{fileLandscape}, ORIENTATION_LANDSCAPE, false},
		{"generic jpg portrait", args{filePortrait}, ORIENTATION_PORTRAIT, false},
		{"generic jpg square", args{fileSquare}, ORIENTATION_SQUARE, false},
		{"generic video landscape", args{videoLandscape}, ORIENTATION_LANDSCAPE, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := GetOrientation(tt.args.media)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetOrientation() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("GetOrientation() = %v, want %v", got, tt.want)
			}
		})
	}
}
