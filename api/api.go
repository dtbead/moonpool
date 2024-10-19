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
	"os"
	"path"
	"strings"
	"time"

	"github.com/dtbead/moonpool/entry"
	mdb "github.com/dtbead/moonpool/internal/db"
	"github.com/dtbead/moonpool/internal/db/archive"
	"github.com/dtbead/moonpool/internal/db/thumbnail"
	"github.com/dtbead/moonpool/internal/file"
	"github.com/dtbead/moonpool/internal/log"
)

const (
	hash_length_md5    = 16
	hash_length_sha1   = 20
	hash_length_sha256 = 32
)

type API struct {
	log       slog.Logger
	archive   archive.Archiver
	thumbnail thumbnail.Thumbnailer
	Config    Config
	db        *sql.DB // db provides low-level access to main moonpool database
}

type WithTX struct {
	q  archive.Querier
	tx archive.TX
}

type Config struct {
	ArchiveLocation, ThumbnailLocation, MediaLocation string
}

type Importer interface {
	Timestamp() entry.Timestamp
	Hash() entry.Hashes
	Path() string
	Extension() string
	Store(baseDirectory string) error
}

func New(l *slog.Logger, c Config) (*API, error) {
	var err error
	a := new(sql.DB)
	t := new(sql.DB)

	if c.ArchiveLocation != ":memory:" {
		c.ArchiveLocation = cleanPath(c.ArchiveLocation)

		a, err = mdb.OpenSQLite3(c.ArchiveLocation)
		if err != nil {
			return &API{}, err
		}
	} else {
		a, _ = mdb.OpenSQLite3Memory()
	}

	if c.ThumbnailLocation != ":memory:" {
		c.ThumbnailLocation = cleanPath(c.ThumbnailLocation)

		t, err = mdb.OpenSQLite3(c.ThumbnailLocation)
		if err != nil {
			return &API{}, err
		}
	} else {
		t, _ = mdb.OpenSQLite3Memory()
	}

	c.MediaLocation = cleanPath(c.MediaLocation)

	archive := archive.NewArchiver(archive.New(a), a)
	thumbnail := thumbnail.NewThumbnailer(thumbnail.New(t), t)

	if err := mdb.InitializeArchive(a); err != nil {
		a.Close()
		t.Close()
		return &API{}, err
	}

	if err := mdb.InitializeThumbnail(a); err != nil {
		a.Close()
		t.Close()
		return &API{}, err
	}

	return &API{
		log:       *l,
		archive:   archive,
		thumbnail: thumbnail,
		Config:    c,
		db:        a,
	}, nil
}

func Open(c Config, l *slog.Logger) (*API, error) {
	moonpool := new(API)

	c.ArchiveLocation = cleanPath(c.ArchiveLocation)
	c.ThumbnailLocation = cleanPath(c.ThumbnailLocation)
	c.MediaLocation = cleanPath(c.MediaLocation)

	if !strings.HasSuffix(c.MediaLocation, "/") || !strings.HasSuffix(c.MediaLocation, "\\") {
		c.MediaLocation += "/"
	}

	if !file.DoesPathExist(c.ArchiveLocation) || c.ArchiveLocation == "" {
		return &API{}, errors.New("archive db path does not exist")
	}

	if !file.DoesPathExist(c.MediaLocation) || c.MediaLocation == "" {
		return &API{}, errors.New("media path does not exist")
	}

	if c.ThumbnailLocation != "" {
		if !file.DoesPathExist(c.ThumbnailLocation) {
			return &API{}, errors.New("thumbnail db path does not exist")
		}

		t, err := mdb.OpenSQLite3(c.ThumbnailLocation)
		if err != nil {
			return &API{}, err
		}

		moonpool.thumbnail = thumbnail.NewThumbnailer(thumbnail.New(t), t)
	}

	a, err := mdb.OpenSQLite3(c.ArchiveLocation)
	if err != nil {
		return &API{}, err
	}

	moonpool.db = a
	moonpool.archive = archive.NewArchiver(archive.New(a), a)
	moonpool.log = *l
	moonpool.Config = c

	return moonpool, nil
}

// Close() manually runs a SQL checkpoint and closes the API connection. Calling Close()
// will implicitly close the sql.DB connection as well
func (a *API) Close() error {
	defer a.db.Close()

	if err := a.archive.ForceCheckpoint(context.Background()); err != nil {
		return err
	}

	if a.thumbnail != nil {
		defer a.thumbnail.Close()
		if err := a.thumbnail.ForceCheckpoint(context.Background()); err != nil {
			return err
		}
	}

	return nil
}

func (a *API) BeginTX(ctx context.Context) (*WithTX, error) {
	q, tx, err := a.archive.NewTx(ctx, nil)
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

// Import() takes an Importer interface and returns an archive_id and nil error on success,
// or both an archive_id AND non-nil error on partial success. Partial success currently only
// occurs when all other import routines (archive creation, file storing, timestamps, etc)
// are successful, except for tag importing. Import() will return an ArchiveID of -1 if a non-partial success isn't possible.
//
// You should ALWAYS check if "ArchiveID <=0 && error != nil"
func (a *API) Import(ctx context.Context, i Importer) (int64, error) {
	if err := a.NewSavepoint(ctx, "import"); err != nil {
		a.log.LogAttrs(context.Background(), log.LogLevelError, "failed to begin import transaction", slog.Any("error", err))
		return -1, err
	}
	defer a.RollbackSavepoint(ctx, "import")

	// finalizeImport() is a helper function which commits a db transaction, finalizing the entire import
	finalizeImport := func() error {
		if err := a.ReleaseSavepoint(ctx, "import"); err != nil {
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
		return -1, errors.New("invalid sha1 hash")
	}
	if len(h.SHA256) != hash_length_sha256 {
		a.log.LogAttrs(context.Background(), log.LogLevelError, "got invalid or empty sha256 hash", slog.Any("error", errors.New("invalid sha256 hash")), slog.String("sha256", byteToHex(h.SHA256[:])))
		return -1, errors.New("invalid sha256 hash")
	}

	a.log.LogAttrs(context.Background(), log.LogLevelVerbose, "got hash",
		slog.Group("hash",
			slog.String("md5", byteToHex(h.MD5[:])),
			slog.String("sha1", byteToHex(h.SHA1[:])),
			slog.String("SHA256", byteToHex(h.SHA256[:]))))

	archive_id, err := a.archive.NewEntry(ctx, file.BuildPath(h.MD5[:], i.Extension()), i.Extension())
	if err != nil {
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

	if err := a.SetHashes(ctx, archive_id, i.Hash()); err != nil {
		a.log.LogAttrs(context.Background(), log.LogLevelError, "failed to set hashes for media", slog.Any("error", err))
		return -1, err
	}

	switch a.Config.MediaLocation {
	default:
		a.log.LogAttrs(context.Background(), log.LogLevelInfo, fmt.Sprintf("copying media to %s", a.Config.MediaLocation))
	case "":
		a.log.LogAttrs(context.Background(), log.LogLevelWarn, "config had no path to store media to. copying media to current directory instead")
	}

	if err := i.Store(a.Config.MediaLocation); err != nil {
		a.log.LogAttrs(context.Background(), log.LogLevelError, fmt.Sprintf("failed to store media to %s", a.Config.MediaLocation), slog.Any("error", err))
		return -1, err
	}

	if err := a.SetTimestamps(ctx, archive_id, entry.Timestamp{
		DateModified: i.Timestamp().DateModified,
		DateCreated:  i.Timestamp().DateCreated,
		DateImported: time.Now(),
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

	if err := finalizeImport(); err != nil {
		return -1, err
	}

	a.log.LogAttrs(context.Background(), log.LogLevelInfo, fmt.Sprintf("imported new media with archive_id %d", archive_id), slog.Int64("archive_id", archive_id))
	return archive_id, nil
}

func (a *API) GetHashes(ctx context.Context, archive_id int64) (entry.Hashes, error) {
	h, err := a.archive.GetHashes(ctx, archive_id)
	if err != nil {
		a.log.LogAttrs(context.Background(), log.LogLevelError, fmt.Sprintf("failed to fetch hashes for archive_id %d", archive_id), slog.Any("error", err),
			slog.Int64("archive_id", archive_id),
		)
		return entry.Hashes{}, err
	}

	return entry.Hashes{
		MD5:    h.Md5,
		SHA1:   h.Sha1,
		SHA256: h.Sha256,
	}, nil

}

func (a *API) SetHashes(ctx context.Context, archive_id int64, h entry.Hashes) error {
	if err := a.archive.SetHashes(ctx, archive_id, archive.Hashes(h)); err != nil {
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

// SetTimestamps() sets assigns or updates an existing timestamp to an entry. Timestamps are implicitly
// converted into a UTC timezone.
func (a *API) SetTimestamps(ctx context.Context, archive_id int64, t entry.Timestamp) error {
	mdbTimestamp := mdb.Timestamp{
		DateCreated:  t.DateCreated,
		DateModified: t.DateModified,
		DateImported: t.DateImported,
	}
	if err := a.archive.SetTimestamps(ctx, archive_id, mdbTimestamp); err != nil {
		a.log.LogAttrs(context.Background(), log.LogLevelError, fmt.Sprintf("failed to set timestamp for archive_id %d", archive_id), slog.Any("error", err),
			slog.Int64("archive_id", archive_id),
			slog.Any("timestamps", t))
		return err
	}
	return nil
}

// GetTimestamps() returns the UTC timestamps of an entry. If only partial timestamp information exists,
// GetTimestamps() will return a type Timestamp and an error. You should ALWAYS check whether a Timestamp
// is empty or not, regardless of any errors.
func (a *API) GetTimestamps(ctx context.Context, archive_id int64) (entry.Timestamp, error) {
	t, err := a.archive.GetTimestamps(ctx, archive_id)
	if err != nil {
		a.log.LogAttrs(context.Background(), log.LogLevelError, fmt.Sprintf("failed to fetch timestamp(s) for archive_id %d", archive_id), slog.Any("error", err),
			slog.Any("timestamps", t),
			slog.Int64("archive_id", archive_id),
		)
		return entry.Timestamp{}, err
	}

	return entry.Timestamp{
		DateModified: t.DateModified.UTC(),
		DateImported: t.DateImported.UTC(),
		DateCreated:  t.DateCreated.UTC(),
	}, nil
}

func (a *API) GetFile(ctx context.Context, archive_id int64) (io.ReadCloser, error) {
	rc, err := a.archive.GetFile(ctx, archive_id, a.Config.MediaLocation)
	if err != nil {
		a.log.LogAttrs(context.Background(), log.LogLevelError, fmt.Sprintf("failed to fetch media file for archive_id %d", archive_id), slog.Any("error", err),
			slog.Int64("archive_id", archive_id),
		)
		return nil, err
	}

	return rc, nil
}

// SetTags() assigns a slice of tags to a given archive_id. A new tag will be implicitly created if one does not exist already. No errors will be
// given if a tag is already set
func (a *API) SetTags(ctx context.Context, archive_id int64, tags []string) error {
	if err := a.archive.NewSavepoint(ctx, "settags"); err != nil {
		a.log.LogAttrs(context.Background(), log.LogLevelError, fmt.Sprintf("failed to begin db transaction to assign tags for archive_id %d", +archive_id), slog.Any("error", err),
			slog.Int64("archive_id", archive_id),
		)
		return err
	}
	defer a.archive.Rollback(ctx, "settags")

	for _, tag := range tags {
		if err := a.archive.SetTag(ctx, archive_id, tag); err != nil {
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

	if err := a.archive.ReleaseSavepoint(ctx, "settags"); err != nil {
		a.log.LogAttrs(context.Background(), log.LogLevelError, fmt.Sprintf("failed to commit transaction for assigning tags on archive_id %d", archive_id), slog.Any("error", err),
			slog.Int64("archive_id", archive_id))
		return err
	}

	return nil
}

func (a *API) GetTags(ctx context.Context, archive_id int64) ([]string, error) {
	tags, err := a.archive.GetTags(ctx, archive_id)
	if err != nil {
		a.log.LogAttrs(context.Background(), log.LogLevelError, fmt.Sprintf("failed to fetch tags for archive_id %d", archive_id), slog.Any("error", err),
			slog.Int64("archive_id", archive_id))
		return nil, err
	}

	return tags, nil
}

func (a *API) GetTagCount(ctx context.Context, tag string) (int64, error) {
	cnt, err := a.archive.GetTagCount(ctx, tag)
	if err != nil {
		a.log.LogAttrs(context.Background(), log.LogLevelError, fmt.Sprintf("failed to get total mapped tags for '%s'", tag), slog.Any("error", err))
		return -1, err
	}

	return cnt.Total, nil
}

// RemoveTags() unassigns a list of tags from an entry. If a tag is no longer in reference to any entry,
// it is completely removed from the database.
func (a *API) RemoveTags(ctx context.Context, archive_id int64, tags []string) error {
	if err := a.archive.NewSavepoint(ctx, "removetags"); err != nil {
		a.log.LogAttrs(context.Background(), log.LogLevelError, fmt.Sprintf("failed to begin db transaction to remove tags for archive_id %d", archive_id), slog.Any("error", err),
			slog.Int64("archive_id", archive_id))
		return err
	}
	defer a.archive.Rollback(ctx, "removetags")

	for _, tag := range tags {
		if err := a.archive.RemoveTag(ctx, archive_id, tag); err != nil {
			a.log.LogAttrs(context.Background(), log.LogLevelError, fmt.Sprintf("failed to unmap tag '%s' for archive_id %d", tag, archive_id), slog.Any("error", err),
				slog.Int64("archive_id", archive_id))
			return err
		}

		t, err := a.archive.SearchTag(ctx, tag)
		if err == sql.ErrNoRows || len(t) == 0 {
			if err := a.archive.DeleteTag(ctx, tag); err != nil {
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

	if err := a.archive.ReleaseSavepoint(ctx, "removetags"); err != nil {
		a.log.LogAttrs(context.Background(), log.LogLevelError, fmt.Sprintf("failed to commit db transaction to remove tags for archive_id %d", archive_id), slog.Any("error", err),
			slog.Int64("archive_id", archive_id))
		return err
	}

	return nil
}

// GetPath() takes an archive_id and returns its relative folder path that points to a file
func (a *API) GetPath(ctx context.Context, archive_id int64) (entry.Path, error) {
	archive, err := a.archive.GetEntry(ctx, archive_id)
	if err != nil {
		a.log.LogAttrs(context.Background(), log.LogLevelError, fmt.Sprintf("failed to fetch file path for archive_id %d", archive_id), slog.Any("error", err),
			slog.Int64("archive_id", archive_id),
		)

		return entry.Path{}, err
	}

	return entry.Path{FileRelative: archive.Path, FileExtension: archive.Extension.String}, nil
}

// SearchTag() takes a tag and returns a slice of archive IDs
func (a *API) SearchTag(ctx context.Context, tag string) ([]int64, error) {
	res, err := a.archive.SearchTag(ctx, tag)
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
		archive_id, err := a.archive.GetMostRecentArchiveID(ctxChild)
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
	phash, err := a.archive.GetPerceptualHash(ctx, archive_id, hashType)
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

	if err := a.archive.SetPerceptualHash(ctx, archive_id, hash.Type, hash.Hash); err != nil {
		a.log.LogAttrs(context.Background(), log.LogLevelError, fmt.Sprintf("failed to set perceptual hash on archive_id %d", archive_id), slog.Any("error", err))
		return err
	}

	return nil
}

func (a *API) RemoveArchive(ctx context.Context, archive_id int64) error {
	if err := a.archive.NewSavepoint(ctx, "remove"); err != nil {
		return err
	}
	defer a.archive.Rollback(ctx, "remove")

	entry, err := a.archive.GetEntry(ctx, archive_id)
	if err != nil {
		return err
	}

	if err := a.archive.RemoveTags(ctx, archive_id); err != nil {
		return err
	}

	if err := a.archive.DeleteEntry(ctx, archive_id); err != nil {
		return err
	}

	fullPath := a.Config.MediaLocation + "/" + entry.Path
	baseDirectory := path.Dir(fullPath)
	if err := os.Remove(fullPath); err != nil {
		return err
	}

	if err := a.archive.ReleaseSavepoint(ctx, "remove"); err != nil {
		return err
	}

	if file.IsDirectoryEmpty(baseDirectory) {
		if err := os.Remove(baseDirectory); err != nil {
			a.log.LogAttrs(context.Background(), log.LogLevelWarn,
				fmt.Sprintf("failed to remove empty media directory at '%s', %v", baseDirectory, err),
				slog.Any("error", err),
				slog.Int64("archive_id", archive_id),
			)
		}
	}

	if err := a.thumbnail.DeleteThumbnail(ctx, archive_id); err != nil {
		a.log.LogAttrs(context.Background(), log.LogLevelWarn,
			fmt.Sprintf("failed to delete thumbnail for archive_id %d, %v", archive_id, err),
			slog.Any("error", err),
			slog.Int64("archive_id", archive_id),
		)
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
	if err := a.archive.NewSavepoint(ctx, name); err != nil {
		return err
	}
	return nil
}

func (a *API) ReleaseSavepoint(ctx context.Context, name string) error {
	if err := a.archive.ReleaseSavepoint(ctx, name); err != nil {
		return err
	}
	return nil
}

func (a *API) RollbackSavepoint(ctx context.Context, name string) error {
	if err := a.archive.Rollback(ctx, name); err != nil {
		return err
	}
	return nil
}

func (a *API) DoesEntryExist(ctx context.Context, archive_id int64) bool {
	return a.archive.DoesArchiveIDExist(ctx, archive_id)
}
