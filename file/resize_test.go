package file

import (
	"image"
	"image/png"
	"os"
	"reflect"
	"testing"
)

func TestResize(t *testing.T) {
	f, _ := os.Open("E:/Programming/go/src/github.com/dtbead/moonpool/file/9dcd819786f7b8804531bcfbfa729e48c98dd9db25a8f1195ea0acec4b8e17c3.png")
	defer f.Close()
	i, _ := png.Decode(f)

	type args struct {
		i image.Image
	}
	tests := []struct {
		name string
		args args
		want *image.RGBA
	}{
		{"output", args{i}, nil},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := Resize(tt.args.i)
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Resize() want %v", tt.want)
			}

			g, err := os.Create("resized.png")
			if err != nil {
				t.Errorf("os.Create() = %v", err)
			}

			defer f.Close()
			if err := png.Encode(g, got); err != nil {
				t.Errorf("png.Encode() = %v", err)
			}

		})
	}
}

func TestCalculateAspectRatioFit(t *testing.T) {
	type args struct {
		width       float64
		height      float64
		scaleFactor float64
	}
	tests := []struct {
		name string
		args args
		want [2]int
	}{
		{"generic", args{1024, 1024, 2}, [2]int{512, 512}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := CalculateAspectRatioFit(tt.args.width, tt.args.height, tt.args.scaleFactor); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("CalculateAspectRatioFit() = %v, want %v", got, tt.want)
			}
		})
	}
}
