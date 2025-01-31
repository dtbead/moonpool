package media

import (
	"bytes"
	"errors"
	"image"
	"io"
	"os"

	_ "image/gif"
	"image/jpeg"
	_ "image/png"

	"golang.org/x/image/webp"

	"github.com/bbrks/go-blurhash"
	"github.com/dtbead/moonpool/entry"
	"github.com/dtbead/moonpool/internal/db/thumbnail"

	"github.com/nfnt/resize"
)

const (
	ORIENTATION_NONE      = 0
	ORIENTATION_LANDSCAPE = 1
	ORIENTATION_PORTRAIT  = 2
	ORIENTATION_SQUARE    = 3
)

// GetOrientation returns the orientation a given media is a landscape type. It returns an error
// if the given io.Reader can't be interpreted as a graphic media.
func GetOrientation(media io.Reader) (ORIENTATION int, err error) {
	if media == nil {
		return -1, errors.New("given nil media")
	}

	i, _, err := image.Decode(media)
	if err != nil {
		return -1, err
	}

	switch {
	case i.Bounds().Dx() > i.Bounds().Dy():
		return ORIENTATION_LANDSCAPE, nil
	case i.Bounds().Dx() < i.Bounds().Dy():
		return ORIENTATION_PORTRAIT, nil
	case i.Bounds().Dx() == i.Bounds().Dy():
		return ORIENTATION_SQUARE, nil
	}

	return -1, errors.New("unknown error")
}

// GetDimensions returns a width and height of a given graphic. It returns an error
// if io.Reader can't be interpreted as a graphic media.
//
// TODO: This does not support video or many other image formats. Maybe replace with an interface.
func GetDimensions(media io.Reader) (struct{ Width, Height int }, error) {
	if media == nil {
		return struct{ Width, Height int }{}, errors.New("given nil media")
	}

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

// DecodeMedia takes graphical file format and returns an image.Image
func DecodeMedia(r io.Reader) (image.Image, error) {
	img, _, err := image.Decode(r)
	if err != nil && err.Error() == "webp: invalid format" {
		f, ok := r.(*os.File)
		if ok {
			f.Seek(0, io.SeekStart)
		}

		img, err = webp.Decode(r)
		if err != nil {
			return nil, err
		}
	}

	if err != nil {
		return nil, err

	}

	return img, nil
}

func calculateAspectRatioFit(width, height int64, scaleFactor float64) [2]int64 {
	return [2]int64{
		int64(float64(width) * scaleFactor), int64(float64(height) * scaleFactor),
	}
}
