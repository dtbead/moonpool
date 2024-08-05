package api

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"image"
	"io"
	"log/slog"
	"time"

	"github.com/dtbead/moonpool/archive"
	"github.com/dtbead/moonpool/archive/db"
	"github.com/dtbead/moonpool/config"
	"github.com/dtbead/moonpool/file"
	"github.com/dtbead/moonpool/log"
)

type API struct {
	log     slog.Logger
	service archive.Servicer
	conf    config.Config
}

type WithTX struct {
	q  db.Querier
	tx archive.TX
}

type Importer interface {
	Timestamp() archive.Timestamp
	Store(baseDirectory string) error
	Path() string
	Extension() string
	Hash() archive.Hashes
}

type Path struct {
	Filepath  string
	Extension string
}

func New(s *sql.DB, l *slog.Logger, config config.Config) *API {
	dbQueries := db.New(s)
	a := archive.NewService(dbQueries, s)

	return &API{
		log:     *l,
		service: a,
		conf:    config,
	}
}

func (a *API) BeginTX(ctx context.Context) (*WithTX, error) {
	q, tx, err := a.service.NewTx(ctx, nil)
	if err != nil {
		a.log.LogAttrs(context.Background(), log.LogLevelError, "failed to begin db transaction", slog.Any("error", err))
		return &WithTX{}, err
	}

	a.log.LogAttrs(context.Background(), log.LogLevelVerbose, "began db transaction")
	return &WithTX{
		q:  q,
		tx: tx,
	}, nil
}

// Import takes an Importer interface and returns an archive_id and nil error on success
// or an archive_id and non-nil error on partial success. Partial success currently only
// occurs when every other import routine (entry creation, hashing, file storing, timestamp)
// is successful besides tag importing. You should always check both 'int64 <=1 &&
// error != nil'
func (a *API) Import(ctx context.Context, i Importer, tags []string) (int64, error) {
	apiWithTX, err := a.BeginTX(ctx)
	if err != nil {
		a.log.LogAttrs(context.Background(), log.LogLevelError, "failed to begin import transaction", slog.Any("error", err))
		return -1, err
	}
	defer apiWithTX.tx.Rollback()

	// finalizeImport commits a db transaction, finalizing the entire import
	finalizeImport := func() error {
		if err := apiWithTX.tx.Commit(); err != nil {
			a.log.LogAttrs(context.Background(), log.LogLevelError, "failed to commit import transaction", slog.Any("error", err))
			return err
		}
		return nil
	}

	hashes := i.Hash()
	if !isValidHash(hashes.MD5, 16) {
		a.log.LogAttrs(context.Background(), log.LogLevelError, "got invalid md5 hash", slog.Any("error", errors.New("invalid md5 hash")), slog.String("md5", string(hashes.MD5)))
		return -1, errors.New("invalid md5 hash")
	}
	if !isValidHash(hashes.SHA1, 20) {
		a.log.LogAttrs(context.Background(), log.LogLevelError, "got invalid sha1 hash", slog.Any("error", errors.New("invalid sha1 hash")), slog.String("sha1", string(hashes.SHA1)))
		return -1, errors.New("invalid sha1 hash")
	}
	if !isValidHash(hashes.SHA256, 32) {
		a.log.LogAttrs(context.Background(), log.LogLevelError, "got invalid sha256 hash", slog.Any("error", errors.New("invalid sha256 hash")), slog.String("sha256", string(hashes.SHA256)))
		return -1, errors.New("invalid sha256 hash")
	}

	a.log.LogAttrs(context.Background(), log.LogLevelVerbose, "calculated hash",
		slog.String("md5", file.ByteToHexString(hashes.MD5)),
		slog.String("sha1", file.ByteToHexString(hashes.SHA1)),
		slog.String("SHA256", file.ByteToHexString(hashes.SHA256)))

	if err := apiWithTX.q.NewEntry(ctx, db.NewEntryParams{
		Path:      file.BuildPath(hashes.MD5, i.Extension()),
		Extension: sql.NullString{String: i.Extension(), Valid: true},
	}); err != nil {
		a.log.LogAttrs(context.Background(), log.LogLevelError, "failed to create new entry",
			slog.Any("error", err),
			slog.Group("media",
				slog.String("path", file.BuildPath(hashes.MD5, i.Extension())),
				slog.String("extension", i.Extension()),
				slog.Group("hashes",
					slog.String("md5", file.ByteToHexString(hashes.MD5)),
					slog.String("sha1", file.ByteToHexString(hashes.SHA1)),
					slog.String("SHA256", file.ByteToHexString(hashes.SHA256)))),
		)
		return -1, err
	}

	// archive_id, err := a.service.NewEntry(ctx, file.BuildPath(hashes.MD5, i.Extension()), i.Extension())
	archive_id, err := apiWithTX.q.GetMostRecentArchiveID(ctx)
	if err != nil {
		a.log.LogAttrs(context.Background(), log.LogLevelError, "failed to get most recent media id", slog.Any("error", err))
		return -1, err
	}

	if err := apiWithTX.q.SetHashes(ctx, db.SetHashesParams{
		ArchiveID: archive_id,
		Md5:       hashes.MD5,
		Sha1:      hashes.SHA1,
		Sha256:    hashes.SHA256,
	}); err != nil {
		a.log.LogAttrs(context.Background(), log.LogLevelError, "failed to set hashes for media", slog.Any("error", err))
		return -1, err
	}

	switch a.conf.MediaPath {
	default:
		a.log.LogAttrs(context.Background(), log.LogLevelInfo, fmt.Sprintf("copying media to %s", a.conf.MediaPath))
		err = i.Store(a.conf.MediaPath)
	case "":
		a.log.LogAttrs(context.Background(), log.LogLevelWarn, "config had no path to store media to. copying media to current directory instead")
		err = i.Store(a.conf.MediaPath)
	}

	if err != nil {
		a.log.LogAttrs(context.Background(), log.LogLevelError, fmt.Sprintf("failed to store media to %s", a.conf.MediaPath), slog.Any("error", err))
		return -1, err
	}

	if err := apiWithTX.q.SetTimestamps(ctx, db.SetTimestampsParams{
		ArchiveID:    archive_id,
		DateModified: archive.ToRFC3339_UTC_Timestamp(timeToUnixEpoch(i.Timestamp().DateModified)),
		DateImported: archive.ToRFC3339_UTC_Timestamp(timeToUnixEpoch(time.Now())),
	}); err != nil {
		a.log.LogAttrs(context.Background(), log.LogLevelWarn, "failed to set timestamps", slog.Any("error", err))
	} else {
		a.log.LogAttrs(context.Background(), log.LogLevelVerbose, "timestamp set",
			slog.String("date_modified", archive.ToRFC3339_UTC_Timestamp(timeToUnixEpoch(i.Timestamp().DateModified))),
			slog.String("date_imported", archive.ToRFC3339_UTC_Timestamp(timeToUnixEpoch(time.Now()))),
		)
	}

	// everything else imported fine at this point. finish the import if possible and let the caller worry about tags instead
	if err := a.service.NewSavepoint(ctx, "tags"); err != nil {
		a.log.LogAttrs(context.Background(), log.LogLevelWarn, "failed to begin new transaction for tags. importing regardless...", slog.Any("error", err))

		if err := finalizeImport(); err != nil {
			return -1, err
		}

		return archive_id, err
	}
	defer a.service.Rollback(ctx, "tags")

	for _, tag := range tags {
		if err := apiWithTX.q.NewTag(ctx, tag); err != nil {
			if err := finalizeImport(); err != nil {
				return -1, err
			}

			a.log.LogAttrs(context.Background(), log.LogLevelWarn, fmt.Sprintf("failed to create tag '%s'", tag), slog.Any("error", err))
			return archive_id, err
		}

		if err := apiWithTX.q.SetTag(ctx, db.SetTagParams{ArchiveID: archive_id, Tag: tag}); err != nil {
			if err := finalizeImport(); err != nil {
				return -1, err
			}
			a.log.LogAttrs(context.Background(), log.LogLevelWarn, fmt.Sprintf("failed to set tag '%s'", tag), slog.Any("error", err))
			return archive_id, err
		}
	}

	if err := a.service.ReleaseSavepoint(ctx, "tags"); err != nil {
		// failed to release savepoint and finalize import
		if err := finalizeImport(); err != nil {
			return -1, err
		}

		// only failed to release savepoint
		a.log.LogAttrs(context.Background(), log.LogLevelError, "failed to commit tags to db", slog.Any("error", err))
		return archive_id, err
	}

	if err := finalizeImport(); err != nil {
		return -1, err
	}

	a.log.LogAttrs(context.Background(), log.LogLevelInfo, fmt.Sprintf("imported new media with archive_id %d)", archive_id), slog.Int64("archive_id", archive_id))
	return archive_id, nil
}

func (a *API) GetHashes(ctx context.Context, archive_id int64) (archive.Hashes, error) {
	h, err := a.service.GetHashes(ctx, archive_id)
	if err != nil {
		a.log.LogAttrs(context.Background(), log.LogLevelError, fmt.Sprintf("failed to fetch hashes for archive_id %d", archive_id), slog.Any("error", err),
			slog.Int64("archive_id", archive_id),
		)
		return archive.Hashes{}, err
	}

	return archive.Hashes{
		MD5:    h.Md5,
		SHA1:   h.Sha1,
		SHA256: h.Sha256,
	}, nil

}

func (a *API) SetHashes(ctx context.Context, archive_id int64, h archive.Hashes) error {
	if err := a.service.SetHashes(ctx, archive_id, h); err != nil {
		a.log.LogAttrs(context.Background(), log.LogLevelError, fmt.Sprintf("failed to set hashes for archive_id %d", archive_id), slog.Any("error", err),
			slog.Int64("archive_id", archive_id),
			slog.Group("hashes",
				slog.String("md5", file.ByteToHexString(h.MD5)),
				slog.String("sha1", file.ByteToHexString(h.SHA1)),
				slog.String("sha256", file.ByteToHexString(h.SHA256))),
		)
		return err
	}

	return nil
}

func (a *API) SetTimestamps(ctx context.Context, archive_id int64, t archive.Timestamp) error {
	if err := a.service.SetTimestamps(ctx, archive_id, t); err != nil {
		a.log.LogAttrs(context.Background(), log.LogLevelError, fmt.Sprintf("failed to set timestamp for archive_id %d", archive_id), slog.Any("error", err),
			slog.Int64("archive_id", archive_id),
			slog.Any("timestamps", t))
		return err
	}
	return nil
}

// GetTimestamps returns a type Timestamp of an assoicated archive_id. If only partial timestamp information
// exists, GetTimestamps will still return a Timestamp and an error. You should ALWAYS check whether a Timestamp
// is empty or not, regardless of any errors.
func (a *API) GetTimestamps(ctx context.Context, archive_id int64) (archive.Timestamp, error) {
	t, err := a.service.GetTimestamps(ctx, archive_id)
	if err != nil {
		a.log.LogAttrs(context.Background(), log.LogLevelError, fmt.Sprintf("failed to fetch timestamp(s) for archive_id %d", archive_id), slog.Any("error", err),
			slog.Any("timestamps", t),
			slog.Int64("archive_id", archive_id),
		)
		return t, err
	}

	return t, nil
}

// GetFile returns a ReadCloser of an associated archive_id. The caller is always responsible
// for closing the ReadCloser after they are finished using it.
func (a *API) GetFile(ctx context.Context, archive_id int64) (io.ReadCloser, error) {
	rc, err := a.service.GetFile(ctx, archive_id)
	if err != nil {
		a.log.LogAttrs(context.Background(), log.LogLevelError, fmt.Sprintf("failed to fetch media file for archive_id %d", archive_id), slog.Any("error", err),
			slog.Int64("archive_id", archive_id),
		)
		return nil, err
	}

	return rc, nil
}

func (a *API) SetTags(ctx context.Context, archive_id int64, tags []string) error {
	if err := a.service.NewSavepoint(ctx, "settags"); err != nil {
		a.log.LogAttrs(context.Background(), log.LogLevelError, fmt.Sprintf("failed to begin db transaction to set tags for archive_id %d", +archive_id), slog.Any("error", err),
			slog.Int64("archive_id", archive_id),
		)
		return err
	}
	defer a.service.Rollback(ctx, "settags")

	for _, tag := range tags {
		if err := a.service.NewTag(ctx, tag); err != nil {
			a.log.LogAttrs(context.Background(), log.LogLevelWarn, fmt.Sprintf("failed to create new tag '%s'", tag), slog.Any("error", err),
				slog.Int64("archive_id", archive_id),
			)
			return err
		}

		if err := a.service.SetTag(ctx, archive_id, tag); err != nil {
			a.log.LogAttrs(context.Background(), log.LogLevelWarn, fmt.Sprintf("failed to set tag '%s' for archive_id %d", tag, archive_id), slog.Any("error", err),
				slog.Int64("archive_id", archive_id))
			return err
		}
	}

	if err := a.service.ReleaseSavepoint(ctx, "settags"); err != nil {
		a.log.LogAttrs(context.Background(), log.LogLevelError, fmt.Sprintf("failed to set commit transaction to set tags for archive_id %d", archive_id), slog.Any("error", err),
			slog.Int64("archive_id", archive_id))
		return err
	}

	return nil
}

func (a *API) GetTags(ctx context.Context, archive_id int64) ([]string, error) {
	tags, err := a.service.GetTags(ctx, archive_id)
	if err != nil {
		a.log.LogAttrs(context.Background(), log.LogLevelError, fmt.Sprintf("failed to fetch tags for archive_id %d", archive_id), slog.Any("error", err),
			slog.Int64("archive_id", archive_id))
		return nil, err
	}

	return tags, nil
}

// RemoveTags removes a tag from an entry
func (a *API) RemoveTags(ctx context.Context, archive_id int64, tags []string) error {
	if err := a.service.NewSavepoint(ctx, "removetags"); err != nil {
		a.log.LogAttrs(context.Background(), log.LogLevelError, fmt.Sprintf("failed to begin db transaction to remove tags for archive_id %d", archive_id), slog.Any("error", err),
			slog.Int64("archive_id", archive_id))
		return err
	}
	defer a.service.Rollback(ctx, "removetags")

	for _, tag := range tags {
		if err := a.service.RemoveTag(ctx, archive_id, tag); err != nil {
			a.log.LogAttrs(context.Background(), log.LogLevelError, fmt.Sprintf("failed to unmap tag '%s' for archive_id %d", tag, archive_id), slog.Any("error", err),
				slog.Int64("archive_id", archive_id))
			return err
		}

		t, err := a.service.SearchTag(ctx, tag)
		if err == sql.ErrNoRows || len(t) == 0 {
			if err := a.service.DeleteTag(ctx, tag); err != nil {
				a.log.LogAttrs(context.Background(), log.LogLevelError, fmt.Sprintf("failed to fully delete tag '%s' with no map references", tag), slog.Any("error", err),
					slog.Int64("archive_id", archive_id))
				return err
			}
		} else {
			if err != nil {
				a.log.LogAttrs(context.Background(), log.LogLevelError, fmt.Sprintf("failed to fully delete tag '%s' with no map references", tag), slog.Any("error", err),
					slog.Int64("archive_id", archive_id))

				return err
			}
		}
	}

	if err := a.service.ReleaseSavepoint(ctx, "removetags"); err != nil {
		a.log.LogAttrs(context.Background(), log.LogLevelError, fmt.Sprintf("failed to commit db transaction to remove tags for archive_id %d", archive_id), slog.Any("error", err),
			slog.Int64("archive_id", archive_id))
		return err
	}

	return nil
}

func (a *API) GetTagID(ctx context.Context, tag string) (archive.Tag, error) {
	res, err := a.service.GetTagID(ctx, tag)
	if err == sql.ErrNoRows {
		return archive.Tag{}, nil
	} else {
		if err != nil {
			a.log.LogAttrs(context.Background(), log.LogLevelError, fmt.Sprintf("failed to fetch tag id for tag '%s'", tag), slog.Any("error", err),
				slog.String("tag", tag))
			return archive.Tag{}, err
		}
	}

	return archive.Tag{Text: res.Text, TagID: int(res.TagID)}, nil
}

func (a *API) GetPath(ctx context.Context, archive_id int64) (Path, error) {
	entry, err := a.service.GetEntry(ctx, archive_id)
	if err != nil {
		a.log.LogAttrs(context.Background(), log.LogLevelError, fmt.Sprintf("failed to fetch file path for archive_id %d", archive_id), slog.Any("error", err),
			slog.Int64("archive_id", archive_id),
		)

		return Path{}, err
	}

	return Path{Filepath: entry.Path, Extension: entry.Extension.String}, nil
}

func (a *API) SearchTag(ctx context.Context, tag string) ([]archive.EntryTags, error) {
	res, err := a.service.SearchTag(ctx, tag)
	if err == sql.ErrNoRows {
		return nil, nil
	} else {
		if err != nil {
			a.log.LogAttrs(context.Background(), log.LogLevelError, fmt.Sprintf("failed to search for tag '%s'", tag), slog.Any("error", err),
				slog.String("tag", tag),
			)
			return nil, err
		}
	}

	et := make([]archive.EntryTags, len(res))
	for i := 0; i < len(res); i++ {
		et[i] = archive.EntryTags{
			ArchiveID: res[i].ID,
			Tags: []archive.Tag{
				{
					Text:  res[i].Text,
					TagID: int(res[i].TagID),
				},
			},
		}
	}

	return et, nil
}

func (a *API) GetMostRecentArchiveID(ctx context.Context) (int64, error) {
	ctxChild, cancel := context.WithTimeout(ctx, time.Millisecond*200)
	defer cancel()

	type wrapper struct {
		result int64
		err    error
	}
	ch := make(chan wrapper, 1)

	go func() {
		archive_id, err := a.service.GetMostRecentArchiveID(ctxChild)
		if err != nil {
			ch <- wrapper{-1, err}
		}

		ch <- wrapper{archive_id, nil}
	}()

	select {
	case data := <-ch:
		if data.err != nil {
			a.log.LogAttrs(context.Background(), log.LogLevelError, "failed to fetch more recent archive_id", slog.Any("error", data.err))
			return -1, data.err
		}

		return data.result, nil
	case <-ctx.Done():
		return -1, ctx.Err()
	}
}

func (a *API) GetPerceptualHash(ctx context.Context, archive_id int64, hashType string) (uint64, error) {
	phash, err := a.service.GetPerceptualHash(ctx, archive_id, hashType)
	if err != nil {
		a.log.LogAttrs(context.Background(), log.LogLevelError, fmt.Sprintf("failed to fetch perceptual hash for archive_id %d", archive_id), slog.Any("error", err))
		return 0, err
	}
	return phash, nil
}

func (a *API) GeneratePerceptualHash(ctx context.Context, archive_id int64, hashType string, r io.Reader) error {
	i, _, err := image.Decode(r)
	if err != nil {
		a.log.LogAttrs(context.Background(), log.LogLevelError, fmt.Sprintf("failed to decode image for perceptual hash generation on archive_id %d", archive_id), slog.Any("error", err),
			slog.String("hash_type", hashType),
		)
		return err
	}

	hash, err := file.GetPerceptualHash(i)
	if err != nil {
		a.log.LogAttrs(context.Background(), log.LogLevelError, fmt.Sprintf("failed to generate perceptual hash on archive_id %d", archive_id), slog.Any("error", err))
		return err
	}

	a.log.LogAttrs(context.Background(), log.LogLevelVerbose, fmt.Sprintf("generated perceptual hash as %d for archive_id %d", archive_id, hash.Hash),
		slog.Int64("archive_id", archive_id),
		slog.Any("perceptual hash", hash))

	if err := a.service.SetPerceptualHash(ctx, archive_id, hash.Type, hash.Hash); err != nil {
		a.log.LogAttrs(context.Background(), log.LogLevelError, fmt.Sprintf("failed to set perceptual hash on archive_id %d", archive_id), slog.Any("error", err))
		return err
	}

	return nil
}

func (a *API) NewSavepoint(ctx context.Context, name string) error {
	if err := a.service.NewSavepoint(ctx, name); err != nil {
		return err
	}
	return nil
}

func (a *API) DoesEntryExist(ctx context.Context, id int64) bool {
	return a.service.DoesArchiveIDExist(ctx, id)
}
