package media

import (
	"bytes"
	"image"
	"io"

	_ "image/gif"
	"image/jpeg"
	_ "image/png"

	"github.com/bbrks/go-blurhash"
	"github.com/chai2010/webp"
	"github.com/dtbead/moonpool/entry"
	"github.com/dtbead/moonpool/internal/db/thumbnail"

	"github.com/nfnt/resize"
)

func EncodeWebp(i *image.Image, w io.Writer) error {
	data, err := webp.EncodeRGB(*i, 60)
	if err != nil {
		return err
	}

	_, err = w.Write(data)
	if err != nil {
		return err
	}

	return nil
}

func EncodeJpeg(i *image.Image, w io.Writer) error {
	return jpeg.Encode(w, *i, &jpeg.Options{Quality: 65})
}

func EncodeWebpIcons(i entry.Icons) (thumbnail.Sizes, error) {
	var small, medium, large bytes.Buffer
	var err error

	err = EncodeWebp(i.Small, &small)
	if err != nil {
		return thumbnail.Sizes{}, err
	}

	err = EncodeWebp(i.Medium, &medium)
	if err != nil {
		return thumbnail.Sizes{}, err
	}

	err = EncodeWebp(i.Large, &large)
	if err != nil {
		return thumbnail.Sizes{}, err
	}

	return thumbnail.Sizes{
		Small:  small.Bytes(),
		Medium: medium.Bytes(),
		Large:  large.Bytes(),
	}, nil
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

	NewResolution := calculateAspectRatioFit(i.Bounds().Dx(), i.Bounds().Dy(), scaleFactor)

	resized := resize.Resize(uint(NewResolution[0]), uint(NewResolution[1]), i, resize.Lanczos3)
	return &resized, nil
}

func calculateAspectRatioFit(width, height int, scaleFactor float64) [2]int {
	return [2]int{
		int(float64(width) * scaleFactor), int(float64(height) * scaleFactor),
	}
}
