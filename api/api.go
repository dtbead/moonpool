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
	"github.com/dtbead/moonpool/internal/media"
)

const (
	hash_length_md5    = 16
	hash_length_sha1   = 20
	hash_length_sha256 = 32

	UNKNOWN   Orientation = -1
	NONE      Orientation = 0
	LANDSCAPE Orientation = 1
	PORTRAIT  Orientation = 2
	SQUARE    Orientation = 3
)

type Orientation int

var (
	ErrThumbnailNotFound = errors.New("thumbnail not found")
	ErrDuplicateEntry    = errors.New("duplicate entry")
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
	FileData() io.Reader
	Timestamp() entry.Timestamp
	Hash() entry.Hashes
	Path() string
	Extension() string
	FileSize() int
	Store(baseDirectory string) (path string, err error)
}

func New(c Config, l *slog.Logger) (*API, error) {
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
		a, err = mdb.OpenSQLite3Memory()
		if err != nil {
			return nil, err
		}
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

// Close manually runs a SQL checkpoint and closes the API connection. Calling Close()
// will implicitly close the sql.DB connection as well
func (a *API) Close(ctx context.Context) error {
	defer a.db.Close()

	if err := a.archive.ForceCheckpoint(ctx); err != nil {
		return err
	}

	if a.thumbnail != nil {
		defer a.thumbnail.Close()
		if err := a.thumbnail.ForceCheckpoint(ctx); err != nil {
			return err
		}
	}

	return nil
}

func (a *API) BeginTX(ctx context.Context) (*WithTX, error) {
	q, tx, err := a.archive.NewTx(ctx, nil)
	if err != nil {
		a.log.LogAttrs(ctx, log.LogLevelError, "failed to begin db transaction", slog.Any("error", err))
		return &WithTX{}, err
	}

	a.log.LogAttrs(ctx, log.LogLevelVerbose, "began db transaction")
	return &WithTX{
		q:  q,
		tx: tx,
	}, nil
}

// Import takes an Importer and returns an archive_id to be referenced throughout all API functions,
// and an error.
func (a *API) Import(ctx context.Context, i Importer) (archive_id int64, err error) {
	err = a.NewSavepoint(ctx, "import")
	if err != nil {
		a.log.LogAttrs(ctx, log.LogLevelError, "failed to begin import transaction", slog.Any("error", err))
		return -1, err
	}
	defer a.RollbackSavepoint(ctx, "import")

	// finalizeImport is a helper function which commits a db transaction, finalizing the entire import.
	finalizeImport := func() error {
		if err := a.ReleaseSavepoint(ctx, "import"); err != nil {
			a.log.LogAttrs(ctx, log.LogLevelError, "failed to commit import transaction", slog.Any("error", err))
			return err
		}
		return nil
	}

	// get file hash
	h := i.Hash()
	if len(h.MD5) != hash_length_md5 {
		a.log.LogAttrs(ctx, log.LogLevelError, "got invalid or empty md5 hash", slog.Any("error", errors.New("invalid md5 hash")), slog.String("md5", byteToHex(h.MD5[:])))
		return -1, errors.New("invalid md5 hash")
	}
	if len(h.SHA1) != hash_length_sha1 {
		a.log.LogAttrs(ctx, log.LogLevelError, "got invalid or empty sha1 hash", slog.Any("error", errors.New("invalid sha1 hash")), slog.String("sha1", byteToHex(h.SHA1[:])))
		return -1, errors.New("invalid sha1 hash")
	}
	if len(h.SHA256) != hash_length_sha256 {
		a.log.LogAttrs(ctx, log.LogLevelError, "got invalid or empty sha256 hash", slog.Any("error", errors.New("invalid sha256 hash")), slog.String("sha256", byteToHex(h.SHA256[:])))
		return -1, errors.New("invalid sha256 hash")
	}
	a.log.LogAttrs(ctx, log.LogLevelVerbose, "got hash",
		slog.Group("hash",
			slog.String("md5", byteToHex(h.MD5[:])),
			slog.String("sha1", byteToHex(h.SHA1[:])),
			slog.String("SHA256", byteToHex(h.SHA256[:]))))

	// add new database entry
	archive_id, err = a.archive.NewEntry(ctx, file.BuildPath(h.MD5[:], i.Extension()), i.Extension())
	if err != nil && archive.IsErrorConstraint(err) {
		return -1, ErrDuplicateEntry
	}
	if err != nil {
		a.log.LogAttrs(ctx, log.LogLevelError, "failed to create new entry",
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

	// assign hashes to db entry
	err = a.SetHashes(ctx, archive_id, i.Hash())
	if err != nil {
		a.log.LogAttrs(ctx, log.LogLevelError, "failed to set hashes for media", slog.Any("error", err))
		return -1, err
	}

	// assign timestamps to entry
	err = a.SetTimestamps(ctx, archive_id, entry.Timestamp{
		DateImported: time.Now(),
	})
	if err != nil {
		a.log.LogAttrs(ctx, log.LogLevelWarn, "failed to set timestamps", slog.Any("error", err))
		return -1, err
	}

	// get file metadata
	metadata := entry.FileMetadata{
		FileMimetype: file.GetMimeTypeByExtension(i.Extension()),
		FileSize:     int64(i.FileSize()),
	}

	err = a.archive.SetFileMetadata(ctx, archive_id, metadata)
	if err != nil {
		a.log.LogAttrs(ctx, log.LogLevelError, "failed to set metadata",
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

	// copy file to moonpool managed folder
	switch a.Config.MediaLocation {
	default:
		a.log.LogAttrs(ctx, log.LogLevelInfo, "copying media to "+a.Config.MediaLocation)
	case "":
		a.log.LogAttrs(ctx, log.LogLevelWarn, "config had no path to store media to. copying media to current directory instead")
	}

	entryPath, err := i.Store(a.Config.MediaLocation)
	if err != nil {
		a.log.LogAttrs(ctx, log.LogLevelError, "failed to store media to %s"+entryPath, slog.Any("error", err))
		return -1, err
	}

	// import
	err = finalizeImport()
	if err != nil {
		return -1, err
	}

	a.log.LogAttrs(ctx, log.LogLevelInfo, "imported new media with archive_id "+int64ToString(archive_id), slog.Int64("archive_id", archive_id))
	return archive_id, nil
}

func (a *API) GetHashes(ctx context.Context, archive_id int64) (entry.Hashes, error) {
	h, err := a.archive.GetHashes(ctx, archive_id)
	if err != nil {
		a.log.LogAttrs(ctx, log.LogLevelError, "failed to fetch hashes for archive_id "+int64ToString(archive_id), slog.Any("error", err),
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
		a.log.LogAttrs(ctx, log.LogLevelError, "failed to set hashes for archive_id "+int64ToString(archive_id), slog.Any("error", err),
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

// SetTimestamps sets assigns or updates an existing timestamp to an entry. Timestamps are implicitly
// converted into a UTC timezone. If a timestamp field is empty, SetTimestamps will ignore that field
// and only update the field of an entry with a valid time.
func (a *API) SetTimestamps(ctx context.Context, archive_id int64, t entry.Timestamp) error {
	mdbTimestamp := mdb.Timestamp{
		DateCreated:  t.DateCreated,
		DateModified: t.DateModified,
		DateImported: t.DateImported,
	}

	if mdbTimestamp.DateCreated == (time.Time{}) || mdbTimestamp.DateModified == (time.Time{}) || mdbTimestamp.DateImported == (time.Time{}) {
		ts, err := a.archive.GetTimestamps(ctx, archive_id)
		if err != nil {
			a.log.LogAttrs(ctx, log.LogLevelError, "failed to fetch timestamp while trying to find non-empty timestamp",
				slog.Int64("archive_id", archive_id),
				slog.Any("error", err))
			return err
		}

		if mdbTimestamp.DateCreated == (time.Time{}) {
			mdbTimestamp.DateCreated = ts.DateCreated
		}

		if mdbTimestamp.DateModified == (time.Time{}) {
			mdbTimestamp.DateModified = ts.DateModified
		}

		if mdbTimestamp.DateImported == (time.Time{}) {
			mdbTimestamp.DateImported = ts.DateImported
		}

		a.log.LogAttrs(ctx, log.LogLevelVerbose,
			"got empty timestamp in one or more field. using an existing timestamp in archive",
			slog.Int64("archive_id", archive_id))
	}

	if err := a.archive.SetTimestamps(ctx, archive_id, mdbTimestamp); err != nil {
		a.log.LogAttrs(ctx, log.LogLevelError, "failed to set timestamp for archive_id "+int64ToString(archive_id), slog.Any("error", err),
			slog.Int64("archive_id", archive_id),
			slog.Any("timestamps", t))
		return err
	}
	return nil
}

// GetTimestamps returns the UTC timestamps of an entry. If only partial timestamp information exists,
// GetTimestamps will return a partial timestamp and an error. You should ALWAYS check whether a Timestamp
// is empty or not, regardless of any errors.
func (a *API) GetTimestamps(ctx context.Context, archive_id int64) (entry.Timestamp, error) {
	t, err := a.archive.GetTimestamps(ctx, archive_id)
	if err != nil {
		a.log.LogAttrs(ctx, log.LogLevelError, "failed to fetch timestamp(s) for archive_id "+int64ToString(archive_id), slog.Any("error", err),
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
		a.log.LogAttrs(ctx, log.LogLevelError, "failed to fetch media file for archive_id "+int64ToString(archive_id), slog.Any("error", err),
			slog.Int64("archive_id", archive_id),
		)
		return nil, err
	}

	return rc, nil
}

// GetPath takes an archive_id and returns its relative folder path that points to a file
func (a *API) GetPath(ctx context.Context, archive_id int64) (entry.Path, error) {
	archive, err := a.archive.GetEntry(ctx, archive_id)
	if err != nil {
		a.log.LogAttrs(ctx, log.LogLevelError, "failed to fetch file path for archive_id "+int64ToString(archive_id),
			slog.Any("error", err),
			slog.Int64("archive_id", archive_id),
		)
		return entry.Path{}, err
	}

	return entry.Path{FileRelative: archive.Path, FileExtension: archive.Extension}, nil
}

// GetPage returns a list of archives within a given range. Valid sort options are
// "imported", "created", and "modified".
func (a *API) GetPage(ctx context.Context, sort string, amount, pagenation int64, desc bool) ([]archive.Archive, error) {
	return a.archive.GetPage(ctx, sort, amount, pagenation, desc)
}

func (a *API) GetEntry(ctx context.Context, archive_id int64) (entry.Entries, error) {
	res, err := a.archive.GetEntry(ctx, archive_id)
	if err != nil {
		return entry.Entries{}, err
	}
	return entry.Entries{ArchiveID: archive_id, Path: res.Path, Extension: res.Extension}, nil
}

// GetMetadata returns the metadata of an entry.
func (a *API) GetFileMetadata(ctx context.Context, archive_id int64) (entry.FileMetadata, error) {
	return a.archive.GetFileMetadata(ctx, archive_id)
}

// GenerateFileMetadata automatically generates and sets a given archive_id metadata according to entry.FileMetadata.
func (a *API) GenerateFileMetadata(ctx context.Context, archive_id int64) error {
	rd, err := a.GetFile(ctx, archive_id)
	if err != nil {
		return err
	}
	defer rd.Close()

	metadata := entry.FileMetadata{}
	d, err := media.GetDimensions(rd)
	if err != nil {
		return err
	}
	metadata.MediaHeight = int64(d.Height)
	metadata.MediaWidth = int64(d.Width)

	o, err := OrientationToString(getOrientation(int64(d.Width), int64(d.Height)))
	if err != nil {
		return err
	}
	metadata.MediaOrientation = o

	s, err := file.GetSize(rd)
	if err != nil {
		return err
	}
	metadata.FileSize = s

	e, err := a.GetEntry(ctx, archive_id)
	if err != nil {
		return err
	}

	mtype := file.GetMimeTypeByExtension(e.Extension)
	if mtype == "" {
		mtype = "unknown"
	}

	metadata.FileMimetype = mtype
	return a.archive.SetFileMetadata(ctx, archive_id, metadata)
}

// GetMostRecentArchiveID gets the the most recently imported archive_id.
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
			a.log.LogAttrs(ctx, log.LogLevelError, "failed to fetch more recent archive_id",
				slog.Any("error", data.err))
			return -1, data.err
		}

		return data.result, nil
	case <-ctx.Done():
		return -1, ctx.Err()
	}
}

// hashType is unused
func (a *API) GetPerceptualHash(ctx context.Context, archive_id int64, hashType string) (uint64, error) {
	phash, err := a.archive.GetPerceptualHash(ctx, archive_id, "PHash")
	if err != nil {
		a.log.LogAttrs(ctx, log.LogLevelError,
			"failed to fetch perceptual hash for archive_id "+int64ToString(archive_id),
			slog.Any("error", err))
		return 0, err
	}
	return phash, nil
}

// hashType is unused
func (a *API) GeneratePerceptualHash(ctx context.Context, archive_id int64, hashType string, r io.Reader) error {
	if r == nil {
		return errors.New("given nil io.Reader")
	}

	i, _, err := image.Decode(r)
	if err != nil {
		a.log.LogAttrs(ctx, log.LogLevelError,
			"failed to decode image for perceptual hash generation on archive_id "+int64ToString(archive_id),
			slog.Any("error", err),
			slog.String("hash_type", hashType),
		)
		return err
	}

	hash, err := file.GetPerceptualHash(i)
	if err != nil {
		a.log.LogAttrs(ctx, log.LogLevelError, "failed to generate perceptual hash on archive_id "+int64ToString(archive_id),
			slog.Any("error", err))
		return err
	}

	a.log.LogAttrs(ctx, log.LogLevelVerbose,
		fmt.Sprintf("generated perceptual hash as '%d' for archive_id %d", hash.Hash, archive_id),
		slog.Int64("archive_id", archive_id),
		slog.Any("perceptual hash", hash))

	if err := a.archive.SetPerceptualHash(ctx, archive_id, hash.Type, hash.Hash); err != nil {
		a.log.LogAttrs(ctx, log.LogLevelError,
			"failed to set perceptual hash on archive_id "+int64ToString(archive_id),
			slog.Any("error", err))
		return err
	}

	return nil
}

// GetThumbnail returns the thumbnail data from a given entry. Valid sizes are "small", "medium", and "large".
// Valid formats are "jpeg"
func (a *API) GetThumbnail(ctx context.Context, archive_id int64, size, format string) ([]byte, error) {
	var data []byte
	var err error

	switch format {
	case "jpeg":
		data, err = a.thumbnail.GetJpeg(ctx, archive_id, size)
	default:
		return nil, errors.New("unknown format")
	}

	if errors.Is(err, sql.ErrNoRows) {
		a.log.LogAttrs(ctx, log.LogLevelWarn, "missing thumbnail for archive_id "+int64ToString(archive_id),
			slog.Int64("archive_id", archive_id),
			slog.String("thumbnail_size", size),
			slog.String("thumbnail_format", format))
		return nil, ErrThumbnailNotFound
	}

	if err != nil {
		return nil, err
	}

	return data, nil
}

// RemoveArchive completely deletes an entry from the moonpool database.
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

	if err := a.thumbnail.DeleteThumbnail(ctx, archive_id); err != nil {
		a.log.LogAttrs(ctx, log.LogLevelWarn,
			fmt.Sprintf("failed to delete thumbnail for archive_id %d, %v", archive_id, err),
			slog.Any("error", err),
			slog.Int64("archive_id", archive_id),
		)
	}

	fullPath := a.Config.MediaLocation + "/" + entry.Path
	baseDirectory := path.Dir(fullPath)
	if err := os.Remove(fullPath); err != nil {
		return err
	}

	if file.IsDirectoryEmpty(baseDirectory) {
		if err := os.Remove(baseDirectory); err != nil {
			a.log.LogAttrs(ctx, log.LogLevelWarn,
				fmt.Sprintf("failed to remove empty media directory at '%s', %v", baseDirectory, err),
				slog.Any("error", err),
				slog.Int64("archive_id", archive_id),
			)
			return err
		}
	}

	if err := a.archive.ReleaseSavepoint(ctx, "remove"); err != nil {
		return err
	}

	return nil
}

func (a *API) Vacuum(ctx context.Context) (int64, error) {
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
