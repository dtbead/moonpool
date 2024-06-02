package archive

import (
	"context"
	"database/sql"
	_ "embed"
	"io"
	"os"

	"github.com/dtbead/moonpool/archive/db"
)

type service struct {
	query *db.Queries
}
type TX interface {
	Commit() error
	Rollback() error
}

type Servicer interface {
	NewEntry(ctx context.Context, path, extension string) (int64, error)
	GetEntry(ctx context.Context, archive_id int64) (db.Archive, error)
	GetTags(ctx context.Context, archive_id int64) ([]string, error)
	GetFile(ctx context.Context, archive_id int64) (io.ReadCloser, error)
	SetTimestamps(ctx context.Context, archive_id int64, t Timestamp) error
	GetTimestamps(ctx context.Context, archive_id int64) (Timestamp, error)
	NewTag(ctx context.Context, tag string) error
	SetTag(ctx context.Context, archive_id int64, tag string) error
	RemoveTag(ctx context.Context, archive_id int64, tag string) error
	GetTagID(ctx context.Context, tag string) (db.Tag, error)
	SearchTag(ctx context.Context, tag string) ([]db.SearchTagRow, error)
	GetHashes(ctx context.Context, archive_id int64) (db.Hash, error)
	SetHashes(ctx context.Context, archive_id int64, h Hashes) error
	Import(ctx context.Context, e Entry, tags []string) (int64, error)
	GetMostRecentArchiveID(ctx context.Context) (int64, error)
	NewTx(ctx context.Context, opt *sql.TxOptions) (db.Querier, TX, error)
}

// BeginTx initiates a transaction.
func (s service) NewTx(ctx context.Context, opt *sql.TxOptions) (db.Querier, TX, error) {
	return s.query.BeginTx(ctx, nil)
}

func NewService(q *db.Queries) Servicer {
	return &service{
		query: q,
	}
}

// NewEntry inserts a new entry into the archive. Each new entry is given a 'archive_id' that externally referred to as "id"
// to the user. NewEntry returns an integer archive_id if successful; returns -1 and error otherwise
func (s service) NewEntry(ctx context.Context, path, extension string) (int64, error) {
	var archive_id int64

	isValidString := func(s string) bool {
		return s != ""
	}

	if err := s.query.NewEntry(ctx, db.NewEntryParams{
		Path: path,
		Extension: sql.NullString{ // TODO: is setting Valid even necessary?
			String: extension, Valid: isValidString(extension),
		},
	}); err != nil {
		return -1, err
	}

	archive_id, err := s.GetMostRecentArchiveID(ctx)
	if err != nil {
		return -1, err
	}

	return archive_id, nil
}

func (s service) GetEntry(ctx context.Context, archive_id int64) (db.Archive, error) {
	a, err := s.query.GetEntry(ctx, archive_id)
	if err != nil {
		return db.Archive{}, err
	}

	return a, nil
}

// GetFile returns the full file content of an entry. The caller is expected to handle closing the io interface
func (s service) GetFile(ctx context.Context, archive_id int64) (io.ReadCloser, error) {
	e, err := s.GetEntry(ctx, archive_id)
	if err != nil {
		return nil, err
	}

	f, err := os.Open(e.Path)
	if err != nil {
		return nil, err
	}

	return f, nil
}

func (s service) GetTags(ctx context.Context, archive_id int64) ([]string, error) {
	t, err := s.query.GetTagsFromArchiveID(ctx, archive_id)
	if err != nil {
		return nil, err
	}

	return t, err
}

func (s service) SetTimestamps(ctx context.Context, archive_id int64, t Timestamp) error {
	err := s.query.SetTimestamps(ctx, db.SetTimestampsParams{
		ArchiveID:    archive_id,
		DateModified: ToRFC3339_UTC_Timestamp(t.DateModified),
		DateImported: ToRFC3339_UTC_Timestamp(t.DateImported),
	})
	if err != nil {
		return err
	}

	return nil
}

func (s service) GetTimestamps(ctx context.Context, archive_id int64) (Timestamp, error) {
	t, err := s.query.GetTimestamps(ctx, archive_id)
	if err != nil {
		return Timestamp{}, err
	}

	dateImported, err := ParseTimestamp(t.DateImported)
	if err != nil {
		return Timestamp{}, err
	}

	dateModified, err := ParseTimestamp(t.DateModified)
	if err != nil {
		return Timestamp{}, err
	}

	return Timestamp{
		DateImported: dateImported,
		DateModified: dateModified,
	}, nil
}

// AddTag creates a new tag in the archive and maps it to an entry (archive_id)
func (s service) NewTag(ctx context.Context, tag string) error {
	err := s.query.NewTag(ctx, tag)
	if err != nil {
		return err
	}

	return nil
}

func (s service) SetTag(ctx context.Context, archive_id int64, tag string) error {
	err := s.query.SetTag(ctx, db.SetTagParams{ArchiveID: archive_id, Tag: tag})
	if err != nil {
		return err
	}

	return nil
}

func (s service) RemoveTag(ctx context.Context, archive_id int64, tag string) error {
	err := s.query.RemoveTag(ctx, db.RemoveTagParams{ArchiveID: archive_id, Text: tag})
	if err != nil {
		return err
	}

	return nil
}

func (s service) DeleteTag(ctx context.Context, tag string) error {
	err := s.query.DeleteTag(ctx, tag)
	if err != nil {
		return err
	}

	return nil
}

func (s service) DeleteTagMap(ctx context.Context, tag_id int64) error {
	err := s.query.DeleteTagMap(ctx, tag_id)
	if err != nil {
		return err
	}

	return nil
}

func (s service) GetHashes(ctx context.Context, archive_id int64) (db.Hash, error) {
	h, err := s.query.GetHashes(ctx, archive_id)
	if err != nil {
		return db.Hash{}, err
	}

	return h, nil
}

func (s service) SetHashes(ctx context.Context, archive_id int64, h Hashes) error {
	err := s.query.SetHashes(ctx,
		db.SetHashesParams{
			ArchiveID: archive_id,
			Md5:       h.MD5,
			Sha1:      h.SHA1,
			Sha256:    h.SHA256,
		},
	)
	if err != nil {
		return err
	}

	return nil
}

// Import imports a new entry to the entirety of the archive
// TODO: should this function even exist? API should be responsible for importing
func (s service) Import(ctx context.Context, e Entry, tags []string) (int64, error) {
	isValidString := func(s string) bool {
		return s != ""
	}

	if err := s.query.NewEntry(ctx, db.NewEntryParams{
		Path: e.Metadata.PathRelative,
		Extension: sql.NullString{ // TODO: is setting Valid even necessary?
			String: e.Metadata.Extension, Valid: isValidString(e.Metadata.Extension),
		},
	}); err != nil {
		return -1, err
	}

	archive_id, err := s.query.GetMostRecentArchiveID(ctx)
	if err != nil {
		return -1, err
	}

	if err := s.query.SetHashes(ctx, db.SetHashesParams{
		ArchiveID: archive_id,
		Md5:       e.Metadata.Hash.MD5,
		Sha1:      e.Metadata.Hash.SHA1,
		Sha256:    e.Metadata.Hash.SHA256,
	}); err != nil {
		return -1, err
	}

	// TODO: use sql transactions instead
	for _, v := range tags {
		if err := s.query.NewTag(ctx, v); err != nil {
			return archive_id, err
		}

		if err := s.query.SetTag(ctx, db.SetTagParams{
			ArchiveID: archive_id,
			Tag:       v,
		}); err != nil {
			return archive_id, err
		}
	}

	return archive_id, nil
}

// GetMostRecentArchiveID returns the most recently inserted entry in the archive
func (s service) GetMostRecentArchiveID(ctx context.Context) (int64, error) {
	a, err := s.query.GetMostRecentArchiveID(ctx)
	if err != nil {
		return -1, err
	}

	return a, nil
}

// GetTagID searches for a tag that exists in database, regardless of whether
// it's mapped to an archive or not
func (s service) GetTagID(ctx context.Context, tag string) (db.Tag, error) {
	t, err := s.query.GetTagID(ctx, tag)
	if err != nil {
		return db.Tag{}, err
	}

	return t, nil
}

func (s service) SearchTag(ctx context.Context, tag string) ([]db.SearchTagRow, error) {
	t, err := s.query.SearchTag(ctx, tag)
	if err != nil {
		return []db.SearchTagRow{}, err
	}

	return t, nil
}
