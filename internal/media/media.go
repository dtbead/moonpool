package media

import (
	"image"
	"io"
	"math"

	_ "image/gif"
	"image/jpeg"
	_ "image/jpeg"
	_ "image/png"

	"github.com/dtbead/moonpool/entry"
	"github.com/kolesa-team/go-webp/encoder"
	"github.com/kolesa-team/go-webp/webp"
	"github.com/nfnt/resize"
)

func EncodeWebp(i *image.Image, w io.Writer) error {
	webpOptions, _ := encoder.NewLossyEncoderOptions(encoder.PresetDefault, 75)
	return webp.Encode(w, *i, webpOptions)
}

func EncodeJpeg(i *image.Image, w io.Writer) error {
	return jpeg.Encode(w, *i, nil)
}

func GenerateIcons(i *image.Image) (entry.Icons, error) {
	small, err := rescaleImage(*i, "small")
	if err != nil {
		return entry.Icons{}, err
	}

	medium, err := rescaleImage(*i, "medium")
	if err != nil {
		return entry.Icons{}, err
	}

	large, err := rescaleImage(*i, "large")
	if err != nil {
		return entry.Icons{}, err
	}

	return entry.Icons{
		Small:  small,
		Medium: medium,
		Large:  large,
	}, nil
}

func rescaleImage(i image.Image, size string) (*image.Image, error) {
	var scaleFactor = 0
	switch size {
	case "small":
		scaleFactor = 6
	case "medium":
		scaleFactor = 4
	case "large":
		scaleFactor = 2
	}

	NewResolution := calculateAspectRatioFit(i.Bounds().Dx(), i.Bounds().Dy(), float64(scaleFactor))

	resized := resize.Resize(uint(NewResolution[0]), uint(NewResolution[1]), i, resize.Lanczos3)
	return &resized, nil
}

func calculateAspectRatioFit(width, height int, scaleFactor float64) [2]int {
	ratio := int(math.Min((float64(width)/scaleFactor)/float64(height),
		float64(height)/scaleFactor))

	return [2]int{int(width * ratio), int(height * ratio)}
}
