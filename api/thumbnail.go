package api

import (
	"context"
	"fmt"
	"image"
	"log/slog"

	"github.com/dtbead/moonpool/internal/db/thumbnail"
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

	imageSrc, _, err := image.Decode(file)
	if err != nil {
		return err
	}

	icons, err := media.GenerateIcons(&imageSrc)
	if err != nil {
		return err
	}

	type sizeError struct {
		t       thumbnail.Sizes
		encoder string
		err     error
	}

	thumb := make(chan sizeError)

	go func() {
		t, err := media.EncodeWebpIcons(icons)
		if err != nil {
			thumb <- sizeError{t: thumbnail.Sizes{}, encoder: "webp", err: err}
		}
		thumb <- sizeError{t: t, encoder: "webp", err: err}
	}()

	go func() {
		t, err := media.EncodeJpegIcons(icons)
		if err != nil {
			thumb <- sizeError{t: thumbnail.Sizes{}, encoder: "jpeg", err: err}
		}
		thumb <- sizeError{t: t, encoder: "jpeg", err: nil}
	}()

	enc1 := <-thumb
	enc2 := <-thumb

	if enc1.encoder == "webp" {
		if err := a.thumbnail.NewWebp(ctx, archive_id, enc1.t); err != nil {
			return err
		}
	} else {
		if err := a.thumbnail.NewJpeg(ctx, archive_id, enc1.t); err != nil {
			return err
		}
	}

	if enc2.encoder == "webp" {
		if err := a.thumbnail.NewWebp(ctx, archive_id, enc2.t); err != nil {
			return err
		}
	} else {
		if err := a.thumbnail.NewJpeg(ctx, archive_id, enc2.t); err != nil {
			return err
		}
	}

	if err := a.thumbnail.ReleaseSavepoint(ctx, "thumbnail"); err != nil {
		return err
	}

	a.log.LogAttrs(ctx, log.LogLevelInfo,
		fmt.Sprintf("generated thumbnails for archive_id %d", archive_id),
		slog.Int64("archive_id", archive_id))

	return nil
}
