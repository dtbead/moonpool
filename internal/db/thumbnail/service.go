package thumbnail

import (
	"context"
	"database/sql"
	_ "embed"
)

type thumbnail struct {
	query *Queries
	db    *sql.DB
}

type TX interface {
	Commit() error
	Rollback() error
}

type Thumbnailer interface {
	NewThumbnail(ctx context.Context, archive_id int64) error
	NewJpeg(ctx context.Context, archive_id int64, s Sizes) error
	NewWebp(ctx context.Context, archive_id int64, s Sizes) error
}

type Sizes struct {
	Small, Medium, Large []byte
}

func (t thumbnail) NewThumbnail(ctx context.Context, archive_id int64) error {
	return t.query.NewThumbnail(ctx, archive_id)
}

func (t thumbnail) NewJpeg(ctx context.Context, archive_id int64, s *Sizes) error {
	args := NewJpegParams{
		ArchiveID: archive_id,
		Small:     s.Small,
		Medium:    s.Medium,
		Large:     s.Large,
	}

	return t.query.NewJpeg(ctx, args)
}

func (t thumbnail) NewWebp(ctx context.Context, archive_id int64, s *Sizes) error {
	args := NewWebpParams{
		ArchiveID: archive_id,
		Small:     s.Small,
		Medium:    s.Medium,
		Large:     s.Large,
	}

	return t.query.NewWebp(ctx, args)
}
