package importer

import (
	"errors"
	"fmt"
	"image"
	"io"
	"os"

	"github.com/dtbead/moonpool/internal/media"
)

type Thumbnail struct {
	src                  *image.Image
	small, medium, large []byte
}

func NewThumbnail(r io.Reader, format string) (Thumbnail, error) {
	img, _, err := image.Decode(r)
	if err != nil {
		return Thumbnail{}, err
	}

	var t = Thumbnail{
		src: &img,
	}

	icons, err := media.GenerateIcons(&img)
	if err != nil {
		return Thumbnail{}, err
	}

	tempSmall, _ := os.CreateTemp("", "moonpool_large_*.tmp")
	tempMedium, _ := os.CreateTemp("", "moonpool_medium_*.tmp")
	tempLarge, _ := os.CreateTemp("", "moonpool_small_*.tmp")

	debugStr := `created temp files at
		tempSmall: %s
		tempMedium: %s
		tempLarge: %s
	`
	fmt.Printf(debugStr, tempSmall.Name(), tempMedium.Name(), tempLarge.Name())

	defer tempSmall.Close()
	defer os.Remove(tempSmall.Name())

	defer tempMedium.Close()
	defer os.Remove(tempMedium.Name())

	defer tempLarge.Close()
	defer os.Remove(tempLarge.Name())

	switch format {
	default:
		return Thumbnail{}, errors.New("invalid format")
	case "webp":
		if err := media.EncodeWebp(icons.Small, tempSmall); err != nil {
			return Thumbnail{}, err
		}

		if err := media.EncodeWebp(icons.Medium, tempMedium); err != nil {
			return Thumbnail{}, err
		}

		if err := media.EncodeWebp(icons.Large, tempLarge); err != nil {
			return Thumbnail{}, err
		}
	case "jpeg":
		if err := media.EncodeJpeg(icons.Small, tempSmall); err != nil {
			return Thumbnail{}, err
		}

		if err := media.EncodeJpeg(icons.Medium, tempMedium); err != nil {
			return Thumbnail{}, err
		}

		if err := media.EncodeJpeg(icons.Large, tempLarge); err != nil {
			return Thumbnail{}, err
		}
	}

	tempSmall.Seek(0, io.SeekStart)
	tempMedium.Seek(0, io.SeekStart)
	tempLarge.Seek(0, io.SeekStart)

	small, err := io.ReadAll(tempSmall)
	if err != nil {
		return Thumbnail{}, err
	}

	medium, err := io.ReadAll(tempMedium)
	if err != nil {
		return Thumbnail{}, err
	}

	large, err := io.ReadAll(tempLarge)
	if err != nil {
		return Thumbnail{}, err
	}

	if len(small) <= 0 || len(medium) <= 0 || len(large) <= 0 {
		return Thumbnail{}, errors.New("read 0 bytes from temp file")
	}

	t.small = small
	t.medium = medium
	t.large = large

	return t, nil
}

func (t Thumbnail) Small() []byte {
	return t.small
}
func (t Thumbnail) Medium() []byte {
	return t.medium
}
func (t Thumbnail) Large() []byte {
	return t.large
}
