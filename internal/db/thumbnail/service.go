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
	NewJpeg(ctx context.Context, archive_id int64, s Sizes) error
	NewWebp(ctx context.Context, archive_id int64, s Sizes) error
	DeleteThumbnail(ctx context.Context, archive_id int64) error
}

func NewThumbnailer(q *Queries, db *sql.DB) Thumbnailer {
	return thumbnail{
		query: q,
		db:    db,
	}
}

type Sizes struct {
	Small, Medium, Large []byte
}

func (t thumbnail) NewJpeg(ctx context.Context, archive_id int64, s Sizes) error {
	id, err := t.query.DoesArchiveIDExist(ctx, archive_id)
	if err != nil {
		return err
	}

	if id <= 0 {
		if err := t.query.NewThumbnail(ctx, archive_id); err != nil {
			return err
		}
	}

	args := NewJpegParams{
		ArchiveID: archive_id,
		Small:     s.Small,
		Medium:    s.Medium,
		Large:     s.Large,
	}

	return t.query.NewJpeg(ctx, args)
}

func (t thumbnail) NewWebp(ctx context.Context, archive_id int64, s Sizes) error {
	id, err := t.query.DoesArchiveIDExist(ctx, archive_id)
	if err != nil {
		return err
	}

	if id <= 0 {
		if err := t.query.NewThumbnail(ctx, archive_id); err != nil {
			return err
		}
	}

	args := NewWebpParams{
		ArchiveID: archive_id,
		Small:     s.Small,
		Medium:    s.Medium,
		Large:     s.Large,
	}

	return t.query.NewWebp(ctx, args)
}

func (t thumbnail) DeleteThumbnail(ctx context.Context, archive_id int64) error {
	return t.query.DeleteThumbnail(ctx, archive_id)
}
