package thumbnail

import (
	"context"
	"database/sql"
	_ "embed"
	"errors"
	"fmt"

	"github.com/dtbead/moonpool/internal/db"
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
	GetJpeg(ctx context.Context, archive_id int64, size string) ([]byte, error)
	GetBlurHash(ctx context.Context, archive_id int64) (string, error)
	NewBlurHash(ctx context.Context, archive_id int64, hash string) error
	DeleteThumbnail(ctx context.Context, archive_id int64) error
	ForceCheckpoint(ctx context.Context) error
	NewSavepoint(ctx context.Context, name string) error
	ReleaseSavepoint(ctx context.Context, name string) error
	Close() error
}

const DEFAULT_BLURHASH string = "L4A+BoW;00n%?_n%pfo28^aycYW;"

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

	if err := t.query.NewJpeg(ctx, args); err != nil {
		return err
	}

	if err := t.setJpeg(ctx, archive_id, true); err != nil {
		return err
	}

	return nil
}

func (t thumbnail) GetJpeg(ctx context.Context, archive_id int64, size string) ([]byte, error) {
	switch size {
	case "small":
		return t.query.GetJpegsmall(ctx, archive_id)
	case "medium":
		return t.query.GetJpegMedium(ctx, archive_id)
	case "large":
		return t.query.GetJpeglarge(ctx, archive_id)
	default:
		return nil, errors.New("invalid size")
	}
}

func (t thumbnail) DeleteThumbnail(ctx context.Context, archive_id int64) error {
	err := t.NewSavepoint(ctx, "delete")
	if err != nil {
		return err
	}
	defer t.Rollback(ctx, "delete")

	err = t.query.DeleteThumbnail(ctx, archive_id)
	if err != nil {
		return err
	}

	err = t.setJpeg(ctx, archive_id, false)
	if err != nil {
		return err
	}

	err = t.ReleaseSavepoint(ctx, "delete")
	if err != nil {
		return err
	}

	return nil
}

func (t thumbnail) setJpeg(ctx context.Context, archive_id int64, hasThumbnail bool) error {
	if hasThumbnail {
		if _, err := t.db.ExecContext(ctx, fmt.Sprintf("UPDATE thumbnail SET has_jpeg = 1 WHERE archive_id == %d;", archive_id)); err != nil {
			return err
		}
	} else {
		if _, err := t.db.ExecContext(ctx, fmt.Sprintf("UPDATE thumbnail SET has_jpeg = 0 WHERE archive_id == %d;", archive_id)); err != nil {
			return err
		}
	}

	return nil
}

func (t thumbnail) NewBlurHash(ctx context.Context, archive_id int64, hash string) error {
	return t.query.NewBlurHash(ctx, NewBlurHashParams{archive_id, hash})
}

func (t thumbnail) GetBlurHash(ctx context.Context, archive_id int64) (string, error) {
	hash, err := t.query.GetBlurHash(ctx, archive_id)
	if err != nil {
		return DEFAULT_BLURHASH, err
	}

	return hash, nil
}

func (t thumbnail) ForceCheckpoint(ctx context.Context) error {
	_, err := t.db.ExecContext(ctx, "PRAGMA wal_checkpoint;")
	return err
}

func (t thumbnail) NewSavepoint(ctx context.Context, name string) error {
	if !db.IsClean(name) {
		return errors.New("invalid name")
	}

	if _, err := t.db.ExecContext(ctx, `SAVEPOINT "`+name+`";`); err != nil {
		return err
	}
	return nil
}

func (t thumbnail) ReleaseSavepoint(ctx context.Context, name string) error {
	if !db.IsClean(name) {
		return errors.New("invalid name")
	}

	if _, err := t.db.ExecContext(ctx, `RELEASE "`+name+`";`); err != nil {
		return err
	}

	return nil
}

func (t thumbnail) Rollback(ctx context.Context, name string) error {
	if !db.IsClean(name) {
		return errors.New("invalid name")
	}

	_, err := t.db.ExecContext(ctx, `ROLLBACK TO "`+name+`";`)
	return err
}

func (t thumbnail) Close() error {
	if err := t.ForceCheckpoint(context.Background()); err != nil {
		return err
	}
	return t.db.Close()
}
