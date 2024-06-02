package api

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
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
	sql     *sql.DB
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

type Path struct {
	Filepath  string
	Extension string
}

func New(l log.Logger, s *sql.DB) *API {
	dbQueries := db.New(s)
	a := archive.NewService(dbQueries)

	return &API{
		log:     l,
		service: a,
		sql:     s,
	}
}

func (a *API) Close() error {
	return a.sql.Close()
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
	if t.DateImported.IsZero() {
		ts, err := a.service.GetTimestamps(ctx, archive_id)
		if err != nil {
			a.log.Error("SetTimestamps: failed to get timestamp.", err)
		}

		t.DateImported = ts.DateImported
	}

	t = archive.Timestamp{
		DateImported: cleanTimestamp(t.DateImported),
		DateModified: cleanTimestamp(t.DateModified),
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

// GetFile returns the file contents of an entry. The caller is expected to close io.ReadCloser
func (a *API) GetFile(ctx context.Context, archive_id int64) (io.ReadCloser, error) {
	rc, err := a.service.GetFile(ctx, archive_id)
	if err != nil {
		a.log.Error("GetFile: unable to open file.", err)
		return nil, err
	}

	return rc, nil
}

func (a *API) SetTags(ctx context.Context, archive_id int64, tags []string) error {
	for _, v := range tags {
		_, err := a.service.GetTagID(ctx, v)
		if err == sql.ErrNoRows {
			if err := a.service.NewTag(ctx, v); err != nil {
				a.log.Error("SetTags: unable to create new tag '", v, "'.", err)
			}
		} else {
			if err != nil {
				a.log.Error("SetTags: unable to search for tag.", err)
				return err
			}
		}

		if err := a.service.SetTag(ctx, archive_id, v); err != nil {
			a.log.Error("SetTags: unable to set tag '%v', %v", v, err)
			return err
		}
	}
	return nil
}

func (a *API) GetTags(ctx context.Context, archive_id int64) ([]string, error) {
	tags, err := a.service.GetTags(ctx, archive_id)
	if err != nil {
		a.log.Error(fmt.Sprintf("GetTags: unable to get tags for archive_id %d. %v", archive_id, err))
		return nil, err
	}

	return tags, nil
}

// RemoveTags removes a tag from an entry
func (a *API) RemoveTags(ctx context.Context, archive_id int64, tags []string) error {
	tx, err := a.BeginTX(ctx)
	if err != nil {
		a.log.Error(fmt.Sprintf("RemoveTags: failed to begin transaction. %v", err))
		return err
	}

	for _, tag := range tags {
		if err := tx.q.RemoveTag(ctx, db.RemoveTagParams{ArchiveID: archive_id, Text: tag}); err != nil {
			a.log.Error(fmt.Sprintf("RemoveTags: failed to remove tag %v for archive_id %d. %v", tag, archive_id, err))
			tx.tx.Rollback()
			return err
		}

		t, err := tx.q.SearchTag(ctx, tag)
		if err == sql.ErrNoRows || len(t) == 0 {
			if err := tx.q.DeleteTag(ctx, tag); err != nil {
				a.log.Warn(fmt.Sprintf("RemoveTags: failed to completely delete tag %v from archive. %v", tag, err))
				tx.tx.Rollback()
				return err
			}
		} else {
			if err != nil {
				a.log.Error(fmt.Sprintf("RemoveTags: failed to search tag %v. %v", tag, err))
				tx.tx.Rollback()
				return err
			}
		}
	}

	if err := tx.tx.Commit(); err != nil {
		a.log.Error(fmt.Sprintf("RemoveTags: failed to commit transaction. %v", err))
		return err
	}

	return nil
}

func (a *API) GetTagID(ctx context.Context, tag string) (archive.Tag, error) {
	res, err := a.service.GetTagID(ctx, tag)

	if err == sql.ErrNoRows {
		return archive.Tag{}, nil
	}

	if err != nil {
		a.log.Error("GetTagID: failed to get tag ID.", err)
		return archive.Tag{}, err
	}

	return archive.Tag{
		TagID: int(res.TagID),
		Text:  res.Text,
	}, nil
}

func (a *API) GetPath(ctx context.Context, archive_id int64) (Path, error) {
	entry, err := a.service.GetEntry(ctx, archive_id)
	if err != nil {
		a.log.Error("GetPath: failed to get entry.", err)
		return Path{}, err
	}

	return Path{
		Filepath:  entry.Path,
		Extension: entry.Extension.String,
	}, nil
}

func (a *API) SearchTag(ctx context.Context, tag string) ([]archive.EntryTags, error) {
	res, err := a.service.SearchTag(ctx, tag)

	if err == sql.ErrNoRows {
		return []archive.EntryTags{}, nil
	}

	if err != nil {
		a.log.Error("GetTagID: failed to get tag ID.", err)
		return []archive.EntryTags{}, err
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
