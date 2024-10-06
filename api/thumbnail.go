package api

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/dtbead/moonpool/internal/db/thumbnail"
	"github.com/dtbead/moonpool/internal/log"
)

type Encoder interface {
	Small() []byte
	Medium() []byte
	Large() []byte
}

func (a *API) GenerateThumbnailJpeg(ctx context.Context, archive_id int64, e Encoder) error {
	if err := a.thumbnail.NewJpeg(ctx, archive_id, thumbnail.Sizes{
		Small:  e.Small(),
		Medium: e.Medium(),
		Large:  e.Large(),
	}); err != nil {
		a.log.LogAttrs(ctx, log.LogLevelError,
			fmt.Sprintf("failed to generate jpeg thumbnails for archive_id %d", archive_id), slog.Any("error", err),
			slog.Int64("archive_id", archive_id))
		return err
	}

	a.log.LogAttrs(ctx, log.LogLevelInfo,
		fmt.Sprintf("generated jpeg thumbnails for archive_id %d", archive_id),
		slog.Int64("archive_id", archive_id))

	return nil
}

func (a *API) GenerateThumbnailWebp(ctx context.Context, archive_id int64, e Encoder) error {
	small, medium, large := e.Small(), e.Medium(), e.Large()

	if err := a.thumbnail.NewWebp(ctx, archive_id, thumbnail.Sizes{
		Small:  small,
		Medium: medium,
		Large:  large,
	}); err != nil {
		a.log.LogAttrs(ctx, log.LogLevelError,
			fmt.Sprintf("failed to generate webp thumbnails for archive_id %d", archive_id), slog.Any("error", err),
			slog.Int64("archive_id", archive_id))
		return err
	}

	a.log.LogAttrs(ctx, log.LogLevelInfo,
		fmt.Sprintf("generated webp thumbnails for archive_id %d", archive_id),
		slog.Int64("archive_id", archive_id))

	return nil
}
