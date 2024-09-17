// Package api provides an API for accessing/mangaging a Moonpool database.
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

	"github.com/dtbead/moonpool/entry"
	mdb "github.com/dtbead/moonpool/internal/db"
	"github.com/dtbead/moonpool/internal/db/sqlc"
	"github.com/dtbead/moonpool/internal/file"
	"github.com/dtbead/moonpool/internal/log"
)

const (
	hash_length_md5    = 16
	hash_length_sha1   = 20
	hash_length_sha256 = 32
)

type API struct {
	log     slog.Logger
	service mdb.Servicer
	Conf    Config
	db      *sql.DB
}

type WithTX struct {
	q  sqlc.Querier
	tx mdb.TX
}

type Config struct {
	MediaLocation string
}

type Importer interface {
	Timestamp() entry.Timestamp
	Hash() entry.Hashes
	Path() string
	Extension() string
	Store(baseDirectory string) error
}

func New(s *sql.DB, l *slog.Logger, config Config) *API {
	dbQueries := sqlc.New(s)
	a := mdb.NewService(dbQueries, s)

	config = Config{
		MediaLocation: cleanPath(config.MediaLocation),
	}

	return &API{
		log:     *l,
		service: a,
		Conf:    config,
		db:      s,
	}
}

// Close checkpoinst any remaining database transactions and closes the API connection. Calling Close
// will implicitly close the sql.DB connection as well.
func (a *API) Close() error {
	defer a.db.Close()

	_, err := a.db.Exec("PRAGMA wal_checkpoint;")
	if err != nil {
		return err
	}
	return nil
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

// Import takes an Importer interface and returns an archive_id and nil error on success,
// or both an archive_id AND non-nil error on partial success. Partial success currently only
// occurs when all other import routines (archive creation, file storing, timestamps, etc)
// are successful, except for tag importing. Import will return an ArchiveID of -1 if a non-partial success isn't possible.
//
// You should ALWAYS check if "ArchiveID <=0 && error != nil"
func (a *API) Import(ctx context.Context, i Importer, tags []string) (int64, error) {
	apiWithTX, err := a.BeginTX(ctx)
	if err != nil {
		a.log.LogAttrs(context.Background(), log.LogLevelError, "failed to begin import transaction", slog.Any("error", err))
		return -1, err
	}
	defer apiWithTX.tx.Rollback()

	// finalizeImport is a helper function which commits a db transaction, finalizing the entire import
	finalizeImport := func() error {
		if err := apiWithTX.tx.Commit(); err != nil {
			a.log.LogAttrs(context.Background(), log.LogLevelError, "failed to commit import transaction", slog.Any("error", err))
			return err
		}
		return nil
	}

	h := i.Hash()

	if len(h.MD5) != hash_length_md5 {
		a.log.LogAttrs(context.Background(), log.LogLevelError, "got invalid or empty md5 hash", slog.Any("error", errors.New("invalid md5 hash")), slog.String("md5", byteToHex(h.MD5[:])))
		return -1, errors.New("invalid md5 hash")
	}
	if len(h.SHA1) != hash_length_sha1 {
		a.log.LogAttrs(context.Background(), log.LogLevelError, "got invalid or empty sha1 hash", slog.Any("error", errors.New("invalid sha1 hash")), slog.String("sha1", byteToHex(h.SHA1[:])))
		return -1, errors.New("invalid md5 hash")
	}
	if len(h.SHA256) != hash_length_sha256 {
		a.log.LogAttrs(context.Background(), log.LogLevelError, "got invalid or empty sha256 hash", slog.Any("error", errors.New("invalid sha256 hash")), slog.String("sha256", byteToHex(h.SHA256[:])))
		return -1, errors.New("invalid md5 hash")
	}

	a.log.LogAttrs(context.Background(), log.LogLevelVerbose, "got hash",
		slog.Group("hash",
			slog.String("md5", byteToHex(h.MD5[:])),
			slog.String("sha1", byteToHex(h.SHA1[:])),
			slog.String("SHA256", byteToHex(h.SHA256[:]))))

	if err := apiWithTX.q.NewEntry(ctx, sqlc.NewEntryParams{
		Path:      file.BuildPath(h.MD5[:], i.Extension()),
		Extension: sql.NullString{String: i.Extension(), Valid: true},
	}); err != nil {
		a.log.LogAttrs(context.Background(), log.LogLevelError, "failed to create new entry",
			slog.Any("error", err),
			slog.Group("media",
				slog.String("path", file.BuildPath(h.MD5[:], i.Extension())),
				slog.String("extension", i.Extension()),
				slog.Group("hashes",
					slog.String("md5", byteToHex(h.MD5[:])),
					slog.String("sha1", byteToHex(h.SHA1[:])),
					slog.String("SHA256", byteToHex(h.SHA256[:])))),
		)
		return -1, err
	}

	// archive_id, err := a.service.NewEntry(ctx, file.BuildPath(hashes.MD5, i.Extension()), i.Extension())
	archive_id, err := apiWithTX.q.GetMostRecentArchiveID(ctx)
	if err != nil {
		a.log.LogAttrs(context.Background(), log.LogLevelError, "failed to get most recent media id", slog.Any("error", err))
		return -1, err
	}

	if err := apiWithTX.q.SetHashes(ctx, sqlc.SetHashesParams{
		ArchiveID: archive_id,
		Md5:       h.MD5[:],
		Sha1:      h.SHA1[:],
		Sha256:    h.SHA256[:],
	}); err != nil {
		a.log.LogAttrs(context.Background(), log.LogLevelError, "failed to set hashes for media", slog.Any("error", err))
		return -1, err
	}

	switch a.Conf.MediaLocation {
	default:
		a.log.LogAttrs(context.Background(), log.LogLevelInfo, fmt.Sprintf("copying media to %s", a.Conf.MediaLocation))
		err = i.Store(a.Conf.MediaLocation)
	case "":
		a.log.LogAttrs(context.Background(), log.LogLevelWarn, "config had no path to store media to. copying media to current directory instead")
		err = i.Store(a.Conf.MediaLocation)
	}

	if err != nil {
		a.log.LogAttrs(context.Background(), log.LogLevelError, fmt.Sprintf("failed to store media to %s", a.Conf.MediaLocation), slog.Any("error", err))
		return -1, err
	}

	if err := apiWithTX.q.SetTimestamps(ctx, sqlc.SetTimestampsParams{
		ArchiveID:    archive_id,
		DateModified: timeToRFC3339_UTC(i.Timestamp().DateModified),
		DateCreated:  timeToRFC3339_UTC(i.Timestamp().DateCreated),
		DateImported: timeToRFC3339_UTC(time.Now()),
	}); err != nil {
		a.log.LogAttrs(context.Background(), log.LogLevelWarn, "failed to set timestamps", slog.Any("error", err))
	} else {
		a.log.LogAttrs(context.Background(), log.LogLevelVerbose, "timestamp set",
			slog.Group("timestamp",
				slog.String("date_imported", timeToRFC3339_UTC(time.Now())),
				slog.String("date_modified", timeToRFC3339_UTC(i.Timestamp().DateModified)),
				slog.String("date_created", timeToRFC3339_UTC(i.Timestamp().DateCreated))),
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

		tag_id, err := apiWithTX.q.GetMostRecentTagID(ctx)
		if err != nil {
			a.log.LogAttrs(context.Background(), log.LogLevelWarn, fmt.Sprintf("failed to create tag '%s'", tag), slog.Any("error", err))
		}

		if err := apiWithTX.q.SetTag(ctx, sqlc.SetTagParams{ArchiveID: archive_id, TagID: tag_id}); err != nil {
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

	a.log.LogAttrs(context.Background(), log.LogLevelInfo, fmt.Sprintf("imported new media with archive_id %d", archive_id), slog.Int64("archive_id", archive_id))
	return archive_id, nil
}

func (a *API) GetHashes(ctx context.Context, archive_id int64) (entry.Hashes, error) {
	h, err := a.service.GetHashes(ctx, archive_id)
	if err != nil {
		a.log.LogAttrs(context.Background(), log.LogLevelError, fmt.Sprintf("failed to fetch hashes for archive_id %d", archive_id), slog.Any("error", err),
			slog.Int64("archive_id", archive_id),
		)
		return entry.Hashes{}, err
	}

	return entry.Hashes{
		MD5:    h.Md5,
		SHA1:   h.Sha1,
		SHA256: h.Md5,
	}, nil

}

func (a *API) SetHashes(ctx context.Context, archive_id int64, h entry.Hashes) error {
	if err := a.service.SetHashes(ctx, archive_id, mdb.Hashes(h)); err != nil {
		a.log.LogAttrs(context.Background(), log.LogLevelError, fmt.Sprintf("failed to set hashes for archive_id %d", archive_id), slog.Any("error", err),
			slog.Int64("archive_id", archive_id),
			slog.Group("hashes",
				slog.String("md5", byteToHex(h.MD5[:])),
				slog.String("sha1", byteToHex(h.SHA1[:])),
				slog.String("sha256", byteToHex(h.SHA256[:]))),
		)
		return err
	}

	return nil
}

// SetTimestamps sets assigns or updates an existing timestamp to an entry. Timestamps are automatically
// converted into a UTC timezone.
func (a *API) SetTimestamps(ctx context.Context, archive_id int64, t Timestamp) error {
	mdbTimestamp := mdb.Timestamp{
		DateCreated:  t.DateCreated,
		DateModified: t.DateModified,
		DateImported: t.DateImported,
	}
	if err := a.service.SetTimestamps(ctx, archive_id, mdbTimestamp); err != nil {
		a.log.LogAttrs(context.Background(), log.LogLevelError, fmt.Sprintf("failed to set timestamp for archive_id %d", archive_id), slog.Any("error", err),
			slog.Int64("archive_id", archive_id),
			slog.Any("timestamps", t))
		return err
	}
	return nil
}

// GetTimestamps returns the UTC timestamps of an entry. If only partial timestamp information exists,
// GetTimestamps will return a type Timestamp and an error. You should ALWAYS check whether a Timestamp
// is empty or not, regardless of any errors.
func (a *API) GetTimestamps(ctx context.Context, archive_id int64) (Timestamp, error) {
	t, err := a.service.GetTimestamps(ctx, archive_id)
	if err != nil {
		a.log.LogAttrs(context.Background(), log.LogLevelError, fmt.Sprintf("failed to fetch timestamp(s) for archive_id %d", archive_id), slog.Any("error", err),
			slog.Any("timestamps", t),
			slog.Int64("archive_id", archive_id),
		)
		return Timestamp{}, err
	}

	return Timestamp{
		DateModified: t.DateModified,
		DateImported: t.DateImported,
		DateCreated:  t.DateCreated,
	}, nil
}

func (a *API) GetFile(ctx context.Context, archive_id int64) (io.ReadCloser, error) {
	rc, err := a.service.GetFile(ctx, archive_id, a.Conf.MediaLocation)
	if err != nil {
		a.log.LogAttrs(context.Background(), log.LogLevelError, fmt.Sprintf("failed to fetch media file for archive_id %d", archive_id), slog.Any("error", err),
			slog.Int64("archive_id", archive_id),
		)
		return nil, err
	}

	return rc, nil
}

// SetTags assigns a slice of tags to a given archive_id. A new tag will be implicitly created if one does not exist already. No errors will be
// given if a tag is already set
func (a *API) SetTags(ctx context.Context, archive_id int64, tags []string) error {
	if err := a.service.NewSavepoint(ctx, "settags"); err != nil {
		a.log.LogAttrs(context.Background(), log.LogLevelError, fmt.Sprintf("failed to begin db transaction to assign tags for archive_id %d", +archive_id), slog.Any("error", err),
			slog.Int64("archive_id", archive_id),
		)
		return err
	}
	defer a.service.Rollback(ctx, "settags")

	for _, tag := range tags {
		if err := a.service.SetTag(ctx, archive_id, tag); err != nil {
			a.log.LogAttrs(context.Background(), log.LogLevelError, fmt.Sprintf("failed to assign tag '%s' to archive_id %d", tag, archive_id), slog.Any("error", err),
				slog.Int64("archive_id", archive_id),
				slog.Group("tag",
					slog.String("text", tag)),
			)
			return err
		}

		a.log.LogAttrs(context.Background(), log.LogLevelVerbose, fmt.Sprintf("assigned tag '%s' to archive_id %d", tag, archive_id),
			slog.Int64("archive_id", archive_id),
			slog.Group("tag",
				slog.String("text", tag)))

	}

	if err := a.service.ReleaseSavepoint(ctx, "settags"); err != nil {
		a.log.LogAttrs(context.Background(), log.LogLevelError, fmt.Sprintf("failed to commit transaction for assigning tags on archive_id %d", archive_id), slog.Any("error", err),
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

func (a *API) GetTagCount(ctx context.Context, tag string) (int64, error) {
	cnt, err := a.service.GetTagCount(ctx, tag)
	if err != nil {
		a.log.LogAttrs(context.Background(), log.LogLevelError, fmt.Sprintf("failed to get total mapped tags for '%s'", tag), slog.Any("error", err))
		return -1, err
	}

	return cnt.Total, nil
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

// GetPath takes an archive_id and returns its relative folder path that points to a file
func (a *API) GetPath(ctx context.Context, archive_id int64) (entry.Path, error) {
	archive, err := a.service.GetEntry(ctx, archive_id)
	if err != nil {
		a.log.LogAttrs(context.Background(), log.LogLevelError, fmt.Sprintf("failed to fetch file path for archive_id %d", archive_id), slog.Any("error", err),
			slog.Int64("archive_id", archive_id),
		)

		return entry.Path{}, err
	}

	return entry.Path{FileRelative: archive.Path, FilExtension: archive.Extension.String}, nil
}

// SearchTag takes a tag and returns a slice of archive IDs
func (a *API) SearchTag(ctx context.Context, tag string) ([]int64, error) {
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

	archive_ids := make([]int64, len(res))
	for i := 0; i < len(res); i++ {
		archive_ids[i] = res[i].ID
	}

	return archive_ids, nil
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

func (a *API) Vaccum(ctx context.Context) (int64, error) {
	res, err := a.db.ExecContext(ctx, "VACUUM;")
	if err != nil {
		a.log.Error("failed to vacuum archive", slog.Any("error", err))
		return -1, err
	}

	return res.LastInsertId()
}

func (a *API) NewSavepoint(ctx context.Context, name string) error {
	if err := a.service.NewSavepoint(ctx, name); err != nil {
		return err
	}
	return nil
}

func (a *API) ReleaseSavepoint(ctx context.Context, name string) error {
	if err := a.service.ReleaseSavepoint(ctx, name); err != nil {
		return err
	}
	return nil
}

func (a *API) RollbackSavepoint(ctx context.Context, name string) error {
	if err := a.service.Rollback(ctx, name); err != nil {
		return err
	}
	return nil
}

func (a *API) DoesEntryExist(ctx context.Context, id int64) bool {
	return a.service.DoesArchiveIDExist(ctx, id)
}
