package media

import (
	"bytes"
	"image"
	"io"

	_ "image/gif"
	"image/jpeg"
	_ "image/png"

	"github.com/bbrks/go-blurhash"
	"github.com/dtbead/moonpool/entry"
	"github.com/dtbead/moonpool/internal/db/thumbnail"

	"github.com/nfnt/resize"
)

// IsLandscape checks whether a given media is a landscape type. It returns an error
// if given io.Reader can't be interpreted as a graphic media.
func IsLandscape(media io.Reader) (bool, error) {
	i, _, err := image.Decode(media)
	if err != nil {
		return false, err
	}

	return i.Bounds().Dx() > i.Bounds().Dy(), nil
}

// GetMediaDimensions returns a width and height of a given graphic. It returns an error
// if io.Reader can't be interpreted as a graphic media.
//
// TODO: This does not support video or many other image formats. Maybe replace with an interface.
func GetMediaDimensions(media io.Reader) (struct{ Width, Height int }, error) {
	i, _, err := image.Decode(media)
	if err != nil {
		return struct {
			Width  int
			Height int
		}{}, err
	}

	return struct {
		Width  int
		Height int
	}{
		Width:  i.Bounds().Dx(),
		Height: i.Bounds().Dy()}, nil
}

func EncodeJpeg(i *image.Image, w io.Writer) error {
	return jpeg.Encode(w, *i, &jpeg.Options{Quality: 65})
}

func EncodeJpegIcons(i entry.Icons) (thumbnail.Sizes, error) {
	var small, medium, large bytes.Buffer
	var err error

	err = EncodeJpeg(i.Small, &small)
	if err != nil {
		return thumbnail.Sizes{}, err
	}

	err = EncodeJpeg(i.Medium, &medium)
	if err != nil {
		return thumbnail.Sizes{}, err
	}

	err = EncodeJpeg(i.Large, &large)
	if err != nil {
		return thumbnail.Sizes{}, err
	}

	return thumbnail.Sizes{
		Small:  small.Bytes(),
		Medium: medium.Bytes(),
		Large:  large.Bytes(),
	}, nil
}

func GenerateIcons(i *image.Image) (entry.Icons, error) {
	small, err := RescaleImage(*i, "small")
	if err != nil {
		return entry.Icons{}, err
	}

	medium, err := RescaleImage(*i, "medium")
	if err != nil {
		return entry.Icons{}, err
	}

	large, err := RescaleImage(*i, "large")
	if err != nil {
		return entry.Icons{}, err
	}

	return entry.Icons{
		Small:  small,
		Medium: medium,
		Large:  large,
	}, nil
}

func DecodeBlurHash(hash string) (image.Image, error) {
	return blurhash.Decode(hash, 512, 512, 1)
}

func EncodeBlurHash(i image.Image) (string, error) {
	hash, err := blurhash.Encode(4, 4, i)
	if err != nil {
		return "", err
	}

	return hash, nil
}

// Valid sizes are "small", "medium", and "large"
func RescaleImage(i image.Image, size string) (*image.Image, error) {
	var scaleFactor float64 = 1
	switch size {
	case "small":
		scaleFactor = 0.2
	case "medium":
		scaleFactor = 0.4
	case "large":
		scaleFactor = 0.6
	}

	NewResolution := calculateAspectRatioFit(int64(i.Bounds().Dx()), int64(i.Bounds().Dy()), scaleFactor)

	resized := resize.Resize(uint(NewResolution[0]), uint(NewResolution[1]), i, resize.Lanczos3)
	return &resized, nil
}

func calculateAspectRatioFit(width, height int64, scaleFactor float64) [2]int64 {
	return [2]int64{
		int64(float64(width) * scaleFactor), int64(float64(height) * scaleFactor),
	}
}
