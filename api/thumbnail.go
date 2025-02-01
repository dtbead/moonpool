package api

import (
	"bytes"
	"context"
	"image"
	"log/slog"

	"github.com/dtbead/moonpool/internal/log"
	"github.com/dtbead/moonpool/internal/media"
)

func (a *API) GenerateThumbnail(ctx context.Context, archive_id int64) error {
	if err := a.thumbnail.NewSavepoint(ctx, "thumbnail"); err != nil {
		return err
	}
	defer a.archive.Rollback(ctx, "thumbnail")

	file, err := a.archive.GetFile(ctx, archive_id, a.Config.MediaLocation)
	if err != nil {
		return err
	}
	defer file.Close()

	imageSrc, err := media.DecodeImage(file)
	if err != nil {
		return err
	}

	icons, err := media.GenerateIcons(&imageSrc)
	if err != nil {
		return err
	}

	t, err := media.EncodeJpegIcons(icons)
	if err != nil {
		return err
	}

	err = a.thumbnail.NewJpeg(ctx, archive_id, t)
	if err != nil {
		return err
	}

	if err := a.thumbnail.ReleaseSavepoint(ctx, "thumbnail"); err != nil {
		return err
	}

	a.log.LogAttrs(ctx, log.LogLevelInfo,
		"generated thumbnails for archive_id "+int64ToString(archive_id),
		slog.Int64("archive_id", archive_id))

	return nil
}

func (a *API) GenerateBlurHash(ctx context.Context, archive_id int64) error {
	if err := a.thumbnail.NewSavepoint(ctx, "blurhash"); err != nil {
		return err
	}
	defer a.archive.Rollback(ctx, "blurhash")

	newBlurHashFromThumbnail := func() error {
		imgData, err := a.thumbnail.GetJpeg(ctx, archive_id, "small")
		if err != nil {
			return err
		}

		r := bytes.NewReader(imgData)
		img, _, err := image.Decode(r)
		if err != nil {
			return err
		}

		hash, err := media.EncodeBlurHash(img)
		if err != nil {
			return nil
		}

		if err := a.thumbnail.NewBlurHash(ctx, archive_id, hash); err != nil {
			return err
		}
		return nil
	}

	newBlurHashFromFile := func() error {
		f, err := a.archive.GetFile(ctx, archive_id, a.Config.MediaLocation)
		if err != nil {
			return err
		}

		img, _, err := image.Decode(f)
		if err != nil {
			return err
		}

		hash, err := media.EncodeBlurHash(img)
		if err != nil {
			return err
		}

		if err := a.thumbnail.NewBlurHash(ctx, archive_id, hash); err != nil {
			return err
		}
		return nil
	}

	var err error
	err = newBlurHashFromThumbnail()
	if err != nil {
		err = newBlurHashFromFile()
		if err != nil {
			return err
		}
	}

	if err := a.thumbnail.ReleaseSavepoint(ctx, "blurhash"); err != nil {
		return err
	}

	a.log.LogAttrs(ctx, log.LogLevelVerbose,
		"generated blurhash for archive_id "+int64ToString(archive_id),
		slog.Int64("archive_id", archive_id),
		slog.String("blurhash", "hash"))
	return nil
}

func (a *API) GetBlurHashString(ctx context.Context, archive_id int64) (string, error) {
	return a.thumbnail.GetBlurHash(ctx, archive_id)
}

func (a *API) RenderBlurHash(ctx context.Context, archive_id int64) (image.Image, error) {
	hash, err := a.thumbnail.GetBlurHash(ctx, archive_id)
	if err != nil {
		return nil, err
	}

	return media.DecodeBlurHash(hash)
}
