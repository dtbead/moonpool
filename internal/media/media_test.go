package media

import (
	_ "image/gif"
	_ "image/png"
	"io"
	"os"
	"reflect"
	"testing"

	"github.com/dtbead/moonpool/internal/file"
	ffmpeg_go "github.com/u2takey/ffmpeg-go"
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
	defer fileLandscape.Close()

	filePortrait, err := os.Open("testdata/b1c4670ded937f2a61cccb3e0d95117d18e7cfd8.jpg")
	if err != nil {
		t.Fatalf("failed to open test file, %v", err)
	}
	defer filePortrait.Close()

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

func Test_generateVideoThumbnail(t *testing.T) {
	type args struct {
		filepath string
	}
	tests := []struct {
		name      string
		args      args
		wantPHash file.PerceptualHashes
		wantErr   bool
	}{
		{"video", args{"testdata/testsrc.mp4"},
			file.PerceptualHashes{Type: "PHash", Hash: 10621923439404376150}, false},
		{"image", args{"testdata/6ba11adbdb35ee10f9353608a7b97ef248733a72.jpg"},
			file.PerceptualHashes{Type: "PHash", Hash: 14274222685500242926}, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotImg, err := generateVideoThumbnail(tt.args.filepath)
			if (err != nil) != tt.wantErr {
				t.Errorf("generateVideoThumbnail() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			h, err := file.GetPerceptualHash(gotImg)
			if err != nil {
				t.Errorf("failed to genereate perceptual hash. %v", err)
			}

			if h.Hash != tt.wantPHash.Hash {
				t.Errorf("generateVideoThumbnail() got = %v, wantPHash %v", h.Hash, tt.wantPHash)
			}
		})
	}
}

func Test_unmarshalFFmpeg(t *testing.T) {
	testVideo, err := os.Open("testdata/test_video_audio.mp4")
	if err != nil {
		t.Fatalf("failed to open test file, %v", err)
	}
	defer testVideo.Close()

	ffmpegOutput, err := ffmpeg_go.ProbeReader(testVideo)
	if err != nil {
		t.Fatalf("failed to parse video in ffmpeg, %v", err)
	}

	type args struct {
		b []byte
	}
	tests := []struct {
		name    string
		args    args
		want    []ffmpegMetadata
		wantErr bool
	}{
		{"generic video", args{[]byte(ffmpegOutput)},
			[]ffmpegMetadata{
				{Width: 1280, Height: 720, Duration: 5, Framerate: "30/1"},
				{Width: 0, Height: 0, Duration: 5, Framerate: "0/0"},
			}, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := unmarshalFFmpeg(tt.args.b)
			if (err != nil) != tt.wantErr {
				t.Errorf("unmarshalFFmpeg() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("unmarshalFFmpeg() = %v, want %v", got, tt.want)
			}
		})
	}
}
