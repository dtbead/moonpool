package api

import (
	"context"
	"database/sql"
	"errors"
	"io"
	"time"

	"github.com/dtbead/moonpool/archive"
	"github.com/dtbead/moonpool/archive/db"
	"github.com/dtbead/moonpool/file"
	"github.com/dtbead/moonpool/log"
)

type API struct {
	log     log.Logger
	service archive.Servicer
}

type WithTX struct {
	q  db.Querier
	tx db.TX
}

type Importer interface {
	Timestamp() archive.Timestamp
	Store() error
	Path() string
	Extension() string
	Hash() archive.Hashes
}

func New(l log.Logger, s *sql.DB) *API {
	dbQueries := db.New(s)
	a := archive.NewService(dbQueries)

	return &API{
		log:     l,
		service: a,
	}
}

func (a *API) BeginTX(ctx context.Context) (*WithTX, error) {
	q, tx, err := a.service.NewTx(ctx, nil)
	if err != nil {
		return &WithTX{}, err
	}

	return &WithTX{
		q:  q,
		tx: tx,
	}, err
}

// returns -1 on error, archive_id otherwise
func (a *API) Import(ctx context.Context, i Importer, tags []string) (int64, error) {
	apiWithTX, err := a.BeginTX(ctx)
	if err != nil {
		a.log.Error("Import(): unable to begin transaction. %v", err)
		return -1, err
	}

	defer func() error {
		if err := apiWithTX.tx.Commit(); err != nil {
			a.log.Error("Import: failed to commit transaction.", err)
			return err
		}
		a.log.Info("Import: succesfully committed transaction")
		return nil
	}()

	hashes := i.Hash()

	if !isValidHash(hashes.MD5, 16) {
		a.log.Debug("Import: got invalid hash: ", byteToHex(hashes.MD5))
		return -1, errors.New("invalid md5 hash")
	}
	if !isValidHash(hashes.SHA1, 20) {
		a.log.Debug("Import: got invalid hash: ", byteToHex(hashes.SHA1))
		return -1, errors.New("invalid sha1 hash")
	}
	if !isValidHash(hashes.SHA256, 32) {
		a.log.Debug("Import: got invalid hash: ", byteToHex(hashes.SHA256))
		return -1, errors.New("invalid sha256 hash")
	}

	err = apiWithTX.q.NewEntry(ctx, db.NewEntryParams{
		Path:      file.BuildPath(hashes.MD5, i.Extension()),
		Extension: sql.NullString{String: i.Extension(), Valid: true},
	})
	if err != nil {
		a.log.Error("Import(): unable to create new entry.", err)
		apiWithTX.tx.Rollback()
		return -1, err
	}

	// archive_id, err := a.service.NewEntry(ctx, file.BuildPath(hashes.MD5, i.Extension()), i.Extension())
	archive_id, err := apiWithTX.q.GetMostRecentArchiveID(ctx)
	if err != nil {
		a.log.Error("Import(): unable to get most recent archive id.", err)
		apiWithTX.tx.Rollback()
		return -1, err
	}

	/*
		if err := a.service.SetHashes(ctx, archive_id, archive.Hashes{
			MD5:    hashes.MD5,
			SHA1:   hashes.SHA1,
			SHA256: hashes.SHA256,
		}); err != nil {
			return -1, err
		}
	*/
	if err := apiWithTX.q.SetHashes(ctx, db.SetHashesParams{
		ArchiveID: archive_id,
		Md5:       hashes.MD5,
		Sha1:      hashes.SHA1,
		Sha256:    hashes.SHA256,
	}); err != nil {
		a.log.Error("Import(): unable to set hashes.", err)
		apiWithTX.tx.Rollback()
		return -1, err
	}

	if err := i.Store(); err != nil {
		a.log.Error("Import: failed to store new entry file.", err)
		apiWithTX.tx.Rollback()
		return -1, err
	}

	// TODO: add native sqlite batch importing
	for _, v := range tags {
		if err := apiWithTX.q.NewTag(ctx, v); err != nil {
			a.log.Warn("Import: failed to import tag with archive_id", archive_id, ".", err)
		}

		if err := apiWithTX.q.SetTag(ctx, db.SetTagParams{ArchiveID: archive_id, Tag: v}); err != nil {
			a.log.Warn("Import: failed to import tag with archive_id", archive_id, ".", err)
		}
	}

	/*
		if err := a.service.SetTimestamps(ctx, archive_id, archive.Timestamp{
			DateModified: time.Now().UTC(), // TODO: use file modified or user input instead
			DateImported: time.Now().UTC(),
		}); err != nil {
			a.log.Warn("Import: failed to set timestamp for archive_id: ", archive_id, ".", err)
		}
	*/

	if err := apiWithTX.q.SetTimestamps(ctx, db.SetTimestampsParams{
		ArchiveID:    archive_id,
		DateModified: archive.ToRFC3339_UTC_Timestamp(cleanTimestamp(i.Timestamp().DateModified)),
		DateImported: archive.ToRFC3339_UTC_Timestamp(cleanTimestamp(time.Now())),
	}); err != nil {
		a.log.Warn("Import: failed to set timestamp for archive_id: ", archive_id, ".", err)
	}

	return archive_id, nil
}

func (a *API) GetHashes(ctx context.Context, archive_id int64) (archive.Hashes, error) {
	h, err := a.service.GetHashes(ctx, archive_id)
	if err != nil {
		a.log.Error("GetHashes error: %v on archive_id %d", err, archive_id)
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
		a.log.Error("GetHashes: %v on archive_id %d", err, archive_id)
		return err
	}

	return nil
}

func (a *API) SetTimestamps(ctx context.Context, archive_id int64, t archive.Timestamp) error {
	t = archive.Timestamp{
		DateImported: t.DateImported.UTC().Truncate(time.Second * 1), // millisecond precision not needed
		DateModified: t.DateModified.UTC().Truncate(time.Second * 1), // millisecond precision not needed
	}
	if err := a.service.SetTimestamps(ctx, archive_id, t); err != nil {
		a.log.Error("SetTimestamps:", err)
		return err
	}
	return nil
}

func (a *API) GetTimestamps(ctx context.Context, archive_id int64) (archive.Timestamp, error) {
	t, err := a.service.GetTimestamps(ctx, archive_id)
	if err != nil {
		a.log.Error("GetTimestamps:", err)
		return archive.Timestamp{}, err
	}

	return t, nil
}

// Get returns every information relating to an entry
func (a *API) Get(ctx context.Context, archive_id int64) (archive.Entry, error) {
	entry, err := a.service.GetEntry(ctx, archive_id)
	if err != nil {
		return archive.Entry{}, err
	}

	hashes, err := a.service.GetHashes(ctx, archive_id)
	if err != nil {
		return archive.Entry{}, err
	}

	tags, err := a.service.GetTags(ctx, archive_id)
	if err != nil {
		return archive.Entry{}, err
	}

	timestamps, err := a.service.GetTimestamps(ctx, archive_id)
	if err != nil {
		return archive.Entry{}, err
	}

	return archive.Entry{
		Metadata: archive.Metadata{
			PathRelative: entry.Path,
			Extension:    entry.Extension.String,
			Timestamp:    timestamps,
			Hash: archive.Hashes{
				MD5:    hashes.Md5,
				SHA1:   hashes.Sha1,
				SHA256: hashes.Sha256,
			},
		},
		Tags: tags,
	}, nil
}

// GetFile returns the file contents of an entry. The caller is expected to close io.ReadCloser
func (a *API) GetFile(ctx context.Context, archive_id int64) (io.ReadCloser, error) {
	rc, err := a.service.GetFile(ctx, archive_id)
	if err != nil {
		a.log.Error("GetFile: unable to open file.", err)
		return nil, err
	}

	return rc, nil
}
