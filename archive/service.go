package archive

import (
	"context"
	"database/sql"
	_ "embed"
	"errors"
	"io"
	"os"

	"github.com/dtbead/moonpool/archive/db"
)

type service struct {
	query *db.Queries
	db    *sql.DB
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
	GetPerceptualHash(ctx context.Context, archive_id int64, hashType string) (uint64, error)
	SetPerceptualHash(ctx context.Context, archive_id int64, hashType string, hash uint64) error
	// Import(ctx context.Context, e Entry, tags []string) (int64, error)
	DeleteTag(ctx context.Context, tag string) error
	GetMostRecentArchiveID(ctx context.Context) (int64, error)
	DoesArchiveIDExist(ctx context.Context, id int64) bool
	NewTx(ctx context.Context, opt *sql.TxOptions) (db.Querier, TX, error)
	NewSavepoint(ctx context.Context, name string) error
	ReleaseSavepoint(ctx context.Context, name string) error
	Rollback(ctx context.Context, name string) error
	ForceCheckpoint(ctx context.Context) error
}

// BeginTx initiates a transaction.
func (s service) NewTx(ctx context.Context, opt *sql.TxOptions) (db.Querier, TX, error) {
	return s.query.BeginTx(ctx, nil)
}

func (s service) NewSavepoint(ctx context.Context, name string) error {
	if !isClean(name) {
		return errors.New("invalid name")
	}

	if _, err := s.db.ExecContext(ctx, "SAVEPOINT "+name); err != nil {
		return err
	}
	return nil
}

func (s service) ReleaseSavepoint(ctx context.Context, name string) error {
	if !isClean(name) {
		return errors.New("invalid name")
	}

	if _, err := s.db.ExecContext(ctx, "RELEASE "+name); err != nil {
		return err
	}

	return nil
}

func (s service) Rollback(ctx context.Context, name string) error {
	if !isClean(name) {
		return errors.New("invalid name")
	}

	_, err := s.db.ExecContext(ctx, "ROLLBACK TO "+name)
	return err
}

func (s service) ForceCheckpoint(ctx context.Context) error {
	_, err := s.db.ExecContext(ctx, "PRAGMA schema.wal_checkpoint;")
	return err
}

func NewService(q *db.Queries, db *sql.DB) Servicer {
	return &service{
		query: q,
		db:    db,
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
		Extension: sql.NullString{
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
		DateCreated:  ToRFC3339_UTC_Timestamp(t.DateCreated),
		DateModified: ToRFC3339_UTC_Timestamp(t.DateModified),
		DateImported: ToRFC3339_UTC_Timestamp(t.DateImported),
	})
	if err != nil {
		return err
	}

	return nil
}

// GetTimestamps will return any available timestamp regardless of whether an error has
// occured or not. You should ALWAYS check whether a returned Timestamp is empty or not.
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
		return Timestamp{
			DateImported: dateImported,
		}, err
	}

	dateCreated, err := ParseTimestamp(t.DateCreated)
	if err != nil {
		return Timestamp{
			DateImported: dateImported,
			DateModified: dateModified,
		}, err
	}

	return Timestamp{
		DateCreated:  dateCreated,
		DateImported: dateImported,
		DateModified: dateModified,
	}, nil
}

// NewTag creates a new tag in the database that can be later mapped to an entry.
// NewTag will silently continue if given a tag that already exists in database.
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

/*
// Import imports a new entry to the entirety of the archive
// TODO: should this function even exist? API should be responsible for importing
func (s service) Import(ctx context.Context, e Entry, tags []string) (int64, error) {
	isValidString := func(s string) bool {
		return s != ""
	}

	if err := s.query.NewEntry(ctx, db.NewEntryParams{
		Path: e.Metadata.PathRelative,
		Extension: sql.NullString{
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

	// we've already imported everything but tags; no reason to abandon all
	if err := s.NewSavepoint(ctx, "tags"); err != nil {
		return archive_id, err
	}

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

	if err := s.ReleaseSavepoint(ctx, "tags"); err != nil {
		return archive_id, err
	}

	return archive_id, nil
}
*/

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

func (s service) DoesArchiveIDExist(ctx context.Context, id int64) bool {
	res := s.db.QueryRowContext(ctx, `SELECT EXISTS(SELECT 1 FROM archive WHERE id == ? LIMIT 1);`, id)

	var ret = 0
	res.Scan(&ret)

	return ret == 1
}

func (s service) GetPerceptualHash(ctx context.Context, archive_id int64, hashType string) (uint64, error) {
	phash, err := s.query.GetPerceptualHash(ctx, db.GetPerceptualHashParams{ArchiveID: archive_id, Hashtype: hashType})
	if err != nil {
		return 0, err
	}
	return uint64(phash), nil
}

func (s service) SetPerceptualHash(ctx context.Context, archive_id int64, hashType string, hash uint64) error {
	err := s.query.SetPerceptualHash(ctx, db.SetPerceptualHashParams{ArchiveID: archive_id, Hashtype: hashType, Hash: int64(hash)})
	if err != nil {
		return err
	}
	return nil
}
