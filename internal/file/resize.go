package file

import (
	"image"
	"image/draw"
	"math"

	"github.com/anthonynsimon/bild/transform"
)

type Resizer interface {
	Resize(i *image.Image) image.Image
}

func Resize(i image.Image) *image.RGBA {
	img := redraw(i)

	resolution := CalculateAspectRatioFit(float64(img.Bounds().Dx()), float64(img.Bounds().Dy()), 2)

	return transform.Resize(i, resolution[0], resolution[1], transform.Lanczos)
}

func CalculateAspectRatioFit(width, height, scaleFactor float64) [2]int {
	ratio := math.Min((width/scaleFactor)/height, height/scaleFactor)

	return [2]int{int(width * ratio), int(height * ratio)}
}

func redraw(i image.Image) *image.RGBA {
	b := i.Bounds()
	m := image.NewRGBA(image.Rect(0, 0, b.Dx(), b.Dy()))
	draw.Draw(m, m.Bounds(), i, b.Min, draw.Src)

	return m
}
