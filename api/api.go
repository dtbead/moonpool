package api

import (
	"context"
	"database/sql"
	"errors"
	"io"
	"time"

	"github.com/dtbead/moonpool/archive"
	"github.com/dtbead/moonpool/archive/db"
	"github.com/dtbead/moonpool/config"
	"github.com/dtbead/moonpool/file"
)

type API struct {
	service archive.Servicer
	conf    config.Config
}

type WithTX struct {
	q  db.Querier
	tx db.TX
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

func New(s *sql.DB, config config.Config) *API {
	dbQueries := db.New(s)
	a := archive.NewService(dbQueries, s)

	return &API{
		service: a,
		conf:    config,
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

// Import takes an Importer interface and returns an archive_id and nil error on success
// or an archive_id and non-nil error on partial success. Partial success currently only
// occurs when every other import routine (entry creation, hashing, file storing, timestamp)
// is successful besides tag importing. You should always check both 'int64 <=1 &&
// error != nil'
func (a *API) Import(ctx context.Context, i Importer, tags []string) (int64, error) {
	var err error
	apiWithTX, err := a.BeginTX(ctx)
	if err != nil {
		return -1, err
	}
	defer apiWithTX.tx.Rollback()

	hashes := i.Hash()
	if !isValidHash(hashes.MD5, 16) {
		return -1, errors.New("invalid md5 hash")
	}
	if !isValidHash(hashes.SHA1, 20) {
		return -1, errors.New("invalid sha1 hash")
	}
	if !isValidHash(hashes.SHA256, 32) {
		return -1, errors.New("invalid sha256 hash")
	}

	err = apiWithTX.q.NewEntry(ctx, db.NewEntryParams{
		Path:      file.BuildPath(hashes.MD5, i.Extension()),
		Extension: sql.NullString{String: i.Extension(), Valid: true},
	})
	if err != nil {
		return -1, err
	}

	// archive_id, err := a.service.NewEntry(ctx, file.BuildPath(hashes.MD5, i.Extension()), i.Extension())
	archive_id, err := apiWithTX.q.GetMostRecentArchiveID(ctx)
	if err != nil {
		return -1, err
	}

	if err := apiWithTX.q.SetHashes(ctx, db.SetHashesParams{
		ArchiveID: archive_id,
		Md5:       hashes.MD5,
		Sha1:      hashes.SHA1,
		Sha256:    hashes.SHA256,
	}); err != nil {
		return -1, err
	}

	if a.conf.MediaPath() != "" {
		err = i.Store(a.conf.MediaPath())
	} else {
		err = i.Store("")
	}
	if err != nil {
		return -1, err
	}

	if err := apiWithTX.q.SetTimestamps(ctx, db.SetTimestampsParams{
		ArchiveID:    archive_id,
		DateModified: archive.ToRFC3339_UTC_Timestamp(cleanTimestamp(i.Timestamp().DateModified)),
		DateImported: archive.ToRFC3339_UTC_Timestamp(cleanTimestamp(time.Now())),
	}); err != nil {
		return -1, err
	}

	// everything else imported fine, complete the import
	// and let the caller worry about the tags instead
	if err := a.service.NewSavepoint(ctx, "tags"); err != nil {
		return archive_id, errors.Join(err, apiWithTX.tx.Commit())
	}

	for _, v := range tags {
		if err := apiWithTX.q.NewTag(ctx, v); err != nil {
			a.service.Rollback(ctx, "tags")
			return archive_id, errors.Join(err, apiWithTX.tx.Commit())
		}

		if err := apiWithTX.q.SetTag(ctx, db.SetTagParams{ArchiveID: archive_id, Tag: v}); err != nil {
			a.service.Rollback(ctx, "tags")
			return archive_id, errors.Join(err, apiWithTX.tx.Commit())
		}
	}

	if err := a.service.ReleaseSavepoint(ctx, "tags"); err != nil {
		return archive_id, errors.Join(err, apiWithTX.tx.Commit())
	}

	return archive_id, apiWithTX.tx.Commit()
}

func (a *API) GetHashes(ctx context.Context, archive_id int64) (archive.Hashes, error) {
	child, cancel := context.WithTimeout(ctx, time.Millisecond*50)
	defer cancel()

	type wrapper struct {
		result archive.Hashes
		err    error
	}

	ch := make(chan wrapper, 1)
	go func() {
		h, err := a.service.GetHashes(child, archive_id)
		if err != nil {
			ch <- wrapper{archive.Hashes{}, err}
		}

		ch <- wrapper{archive.Hashes{
			MD5:    h.Md5,
			SHA1:   h.Sha1,
			SHA256: h.Sha256,
		}, nil}
	}()

	select {
	case data := <-ch:
		if data.err != nil {
			return archive.Hashes{}, data.err
		}

		return data.result, nil
	case <-ctx.Done():
		return archive.Hashes{}, ctx.Err()
	}
}

func (a *API) SetHashes(ctx context.Context, archive_id int64, h archive.Hashes) error {
	child, cancel := context.WithTimeout(ctx, time.Millisecond*50)
	defer cancel()

	ch := make(chan error, 1)
	go func() {
		if err := a.service.SetHashes(child, archive_id, h); err != nil {
			ch <- err
		}

		ch <- nil
	}()

	select {
	case err := <-ch:
		if err != nil {
			return err
		}

		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}

func (a *API) SetTimestamps(ctx context.Context, archive_id int64, t archive.Timestamp) error {
	child, cancel := context.WithTimeout(ctx, time.Millisecond*150)
	defer cancel()

	type wrapper struct {
		result archive.Timestamp
		err    error
	}

	ch := make(chan wrapper, 1)
	func() {
		// didn't receive DateImported, user likely trying to update DateModified instead
		if t.DateImported.IsZero() {
			ts, err := a.service.GetTimestamps(child, archive_id)

			// likely don't have an entry for this either, might be a new import instead
			// TODO: should we not ignore this error?
			if err != nil || ts.DateImported.IsZero() {
				ts.DateImported = time.Now()
				ts.DateModified = t.DateModified
				ch <- wrapper{ts, nil}
			}
		} else {
			t = archive.Timestamp{
				DateImported: cleanTimestamp(t.DateImported),
				DateModified: cleanTimestamp(t.DateModified),
			}
			ch <- wrapper{t, nil}
		}
	}()

	res := <-ch
	go func() {
		if err := a.service.SetTimestamps(child, archive_id, res.result); err != nil {
			ch <- wrapper{archive.Timestamp{}, err}
		}
		ch <- wrapper{archive.Timestamp{}, nil}
	}()

	select {
	case res := <-ch:
		if res.err != nil {
			return res.err
		}

		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}

func (a *API) GetTimestamps(ctx context.Context, archive_id int64) (archive.Timestamp, error) {
	child, cancel := context.WithTimeout(ctx, 50*time.Millisecond)
	defer cancel()

	type wrapper struct {
		result archive.Timestamp
		err    error
	}

	ch := make(chan wrapper, 1)
	go func() {
		t, err := a.service.GetTimestamps(child, archive_id)
		if err != nil {
			ch <- wrapper{archive.Timestamp{}, err}
		}

		ch <- wrapper{t, nil}
	}()

	select {
	case res := <-ch:
		if res.err != nil {
			return archive.Timestamp{}, res.err
		}

		return res.result, nil
	case <-ctx.Done():
		return archive.Timestamp{}, ctx.Err()
	}
}

// GetFile returns the file contents of an entry. The caller is expected to close io.ReadCloser
func (a *API) GetFile(ctx context.Context, archive_id int64) (io.ReadCloser, error) {
	rc, err := a.service.GetFile(ctx, archive_id)
	if err != nil {
		return nil, err
	}

	return rc, nil
}

func (a *API) SetTags(ctx context.Context, archive_id int64, tags []string) error {
	child, cancel := context.WithTimeout(ctx, 200*time.Millisecond)
	defer cancel()

	ch := make(chan error, 1)
	go func() {
		if err := a.service.NewSavepoint(child, "settags"); err != nil {
			ch <- err
		}
		defer a.service.Rollback(child, "settags")

		for _, tag := range tags {
			if err := a.service.NewTag(child, tag); err != nil {
				ch <- err
			}

			if err := a.service.SetTag(child, archive_id, tag); err != nil {
				ch <- err
			}
		}
		ch <- a.service.ReleaseSavepoint(ctx, "settags")
	}()

	select {
	case err := <-ch:
		if err != nil {
			return err
		}

		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}

func (a *API) GetTags(ctx context.Context, archive_id int64) ([]string, error) {
	child, cancel := context.WithTimeout(ctx, 150*time.Millisecond)
	defer cancel()

	type wrapper struct {
		result []string
		err    error
	}
	ch := make(chan wrapper, 1)

	go func() {
		tags, err := a.service.GetTags(child, archive_id)
		if err != nil {
			ch <- wrapper{nil, err}
		}

		ch <- wrapper{tags, nil}
	}()

	select {
	case res := <-ch:
		if res.err != nil {
			return nil, res.err
		}

		return res.result, nil
	case <-ctx.Done():
		return nil, ctx.Err()
	}
}

// RemoveTags removes a tag from an entry
func (a *API) RemoveTags(ctx context.Context, archive_id int64, tags []string) error {
	child, cancel := context.WithTimeout(ctx, 150*time.Millisecond)
	defer cancel()

	ch := make(chan error, 1)

	go func() {
		if err := a.service.NewSavepoint(child, "removetags"); err != nil {
			ch <- err
		}
		defer a.service.Rollback(ctx, "removetags")

		for _, tag := range tags {
			if err := a.service.RemoveTag(ctx, archive_id, tag); err != nil {
				ch <- err
			}

			t, err := a.service.SearchTag(ctx, tag)
			if err == sql.ErrNoRows || len(t) == 0 {
				if err := a.service.DeleteTag(ctx, tag); err != nil {
					ch <- err
				}
			} else {
				if err != nil {
					ch <- err
				}
			}
		}

		if err := a.service.ReleaseSavepoint(ctx, "removetags"); err != nil {
			ch <- err
		}
	}()

	select {
	case err := <-ch:
		if err != nil {
			return err
		}

		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}

func (a *API) GetTagID(ctx context.Context, tag string) (archive.Tag, error) {
	child, cancel := context.WithTimeout(ctx, 100*time.Millisecond)
	defer cancel()

	type wrapper struct {
		result archive.Tag
		err    error
	}

	ch := make(chan wrapper, 1)
	func() {
		res, err := a.service.GetTagID(child, tag)
		if err == sql.ErrNoRows {
			ch <- wrapper{archive.Tag{}, nil}
		} else {
			if err != nil {
				ch <- wrapper{archive.Tag{}, err}
			}
		}

		ch <- wrapper{archive.Tag{Text: res.Text, TagID: int(res.TagID)}, nil}
	}()
	select {
	case res := <-ch:
		if res.err != nil {
			return archive.Tag{}, nil
		}

		return res.result, nil
	case <-ctx.Done():
		return archive.Tag{}, ctx.Err()
	}
}

func (a *API) GetPath(ctx context.Context, archive_id int64) (Path, error) {
	child, cancel := context.WithTimeout(ctx, 100*time.Millisecond)
	defer cancel()

	type wrapper struct {
		result Path
		err    error
	}

	ch := make(chan wrapper, 1)
	go func() {
		entry, err := a.service.GetEntry(child, archive_id)
		if err != nil {
			ch <- wrapper{Path{}, err}
		}

		ch <- wrapper{Path{Filepath: entry.Path, Extension: entry.Extension.String}, nil}
	}()

	select {
	case res := <-ch:
		if res.err != nil {
			return Path{}, res.err
		}

		return res.result, nil
	case <-ctx.Done():
		return Path{}, ctx.Err()
	}
}

func (a *API) SearchTag(ctx context.Context, tag string) ([]archive.EntryTags, error) {
	child, cancel := context.WithTimeout(ctx, time.Millisecond*200)
	defer cancel()

	type wrapper struct {
		result []db.SearchTagRow
		err    error
	}
	ch := make(chan wrapper, 1)

	go func() {
		res, err := a.service.SearchTag(child, tag)
		if err == sql.ErrNoRows {
			ch <- wrapper{nil, nil}
		}

		if err != nil {
			ch <- wrapper{nil, err}
		}

		ch <- wrapper{res, nil}
	}()

	select {
	case data := <-ch:
		if data.err != nil {
			return nil, data.err
		}

		et := make([]archive.EntryTags, len(data.result))
		for i := 0; i < len(data.result); i++ {
			et[i] = archive.EntryTags{
				ArchiveID: data.result[i].ID,
				Tags: []archive.Tag{
					{
						Text:  data.result[i].Text,
						TagID: int(data.result[i].TagID),
					},
				},
			}
		}
		return et, nil
	case <-ctx.Done():
		return nil, ctx.Err()
	}
}

func (a API) GetMostRecentArchiveID(ctx context.Context) (int64, error) {
	child, cancel := context.WithTimeout(ctx, time.Millisecond*200)
	defer cancel()

	type wrapper struct {
		result int64
		err    error
	}
	ch := make(chan wrapper, 1)

	go func() {
		archive_id, err := a.service.GetMostRecentArchiveID(child)
		if err != nil {
			ch <- wrapper{-1, err}
		}

		ch <- wrapper{archive_id, nil}
	}()

	select {
	case data := <-ch:
		if data.err != nil {
			return -1, data.err
		}

		return data.result, nil
	case <-ctx.Done():
		return -1, ctx.Err()
	}
}

func (a API) NewSavepoint(ctx context.Context, name string) error {
	if err := a.service.NewSavepoint(ctx, name); err != nil {
		return err
	}
	return nil
}
