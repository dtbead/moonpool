package db

import (
	"context"
	"database/sql"
	_ "embed"
	"errors"
	"io"
	"os"

	"github.com/dtbead/moonpool/db/sqlc"
	"modernc.org/sqlite"
)

type service struct {
	query *sqlc.Queries
	db    *sql.DB
}

type TX interface {
	Commit() error
	Rollback() error
}

type Servicer interface {
	NewEntry(ctx context.Context, path, extension string) (int64, error)
	GetEntry(ctx context.Context, archive_id int64) (sqlc.Archive, error)
	GetTags(ctx context.Context, archive_id int64) ([]string, error)
	GetFile(ctx context.Context, archive_id int64, baseDirectory string) (io.ReadCloser, error)
	SetTimestamps(ctx context.Context, archive_id int64, t Timestamp) error
	GetTimestamps(ctx context.Context, archive_id int64) (Timestamp, error)
	NewTag(ctx context.Context, tag string) (int64, error)
	SetTag(ctx context.Context, archive_id int64, tag string) error
	RemoveTag(ctx context.Context, archive_id int64, tag string) error
	GetTagID(ctx context.Context, tag string) (sqlc.Tag, error)
	GetTagCount(ctx context.Context, tag string) (sqlc.TagCount, error)
	SearchTag(ctx context.Context, tag string) ([]sqlc.SearchTagRow, error)
	GetHashes(ctx context.Context, archive_id int64) (sqlc.HashesChksum, error)
	SetHashes(ctx context.Context, archive_id int64, h Hashes) error
	GetPerceptualHash(ctx context.Context, archive_id int64, hashType string) (uint64, error)
	SetPerceptualHash(ctx context.Context, archive_id int64, hashType string, hash uint64) error
	// Import(ctx context.Context, e Entry, tags []string) (int64, error)
	DeleteTag(ctx context.Context, tag string) error
	GetMostRecentArchiveID(ctx context.Context) (int64, error)
	GetMostRecentTagID(ctx context.Context) (int64, error)
	DoesArchiveIDExist(ctx context.Context, id int64) bool
	NewTx(ctx context.Context, opt *sql.TxOptions) (sqlc.Querier, TX, error)
	NewSavepoint(ctx context.Context, name string) error
	ReleaseSavepoint(ctx context.Context, name string) error
	Rollback(ctx context.Context, name string) error
	ForceCheckpoint(ctx context.Context) error
}

// BeginTx initiates a transaction.
func (s service) NewTx(ctx context.Context, opt *sql.TxOptions) (sqlc.Querier, TX, error) {
	tx, err := s.db.BeginTx(ctx, opt)
	if err != nil {
		return nil, nil, err
	}
	q := sqlc.New(tx)
	return q, tx, nil
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

func NewService(q *sqlc.Queries, db *sql.DB) Servicer {
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

	if err := s.query.NewEntry(ctx, sqlc.NewEntryParams{
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

func (s service) GetEntry(ctx context.Context, archive_id int64) (sqlc.Archive, error) {
	a, err := s.query.GetEntry(ctx, archive_id)
	if err != nil {
		return sqlc.Archive{}, err
	}

	return a, nil
}

func (s service) GetFile(ctx context.Context, archive_id int64, baseDirectory string) (io.ReadCloser, error) {
	e, err := s.GetEntry(ctx, archive_id)
	if err != nil {
		return nil, err
	}

	if []rune(baseDirectory)[0] != '/' {
		baseDirectory = baseDirectory + "/"
	}

	f, err := os.Open(baseDirectory + e.Path)
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
	err := s.query.SetTimestamps(ctx, sqlc.SetTimestampsParams{
		ArchiveID:    archive_id,
		DateCreated:  timeToRFC3339_UTC(t.DateCreated),
		DateModified: timeToRFC3339_UTC(t.DateModified),
		DateImported: timeToRFC3339_UTC(t.DateImported),
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

// NewTag creates a new tag in the database that can be later mapped to an entry. NewTag will return a
// tag_id if tag already exists.
func (s service) NewTag(ctx context.Context, tag string) (int64, error) {
	err := s.query.NewTag(ctx, tag)
	if err != nil && isErrorConstraint(err) {
		tag, err := s.GetTagID(ctx, tag)
		if err != nil {
			return -1, err
		}
		return tag.TagID, nil
	}

	if err != nil {
		return -1, err
	}

	tag_id, err := s.GetMostRecentTagID(ctx)
	if err != nil {
		return -1, err
	}

	return tag_id, nil
}

// SetTag assigns a tag to a given archive_id and returns an error if tag does not already exist.
func (s service) SetTag(ctx context.Context, archive_id int64, tag string) error {
	tag_id, err := s.NewTag(ctx, tag)
	if err != nil {
		return err
	}

	// tag has already been assigned to archive_id
	err = s.query.SetTag(ctx, sqlc.SetTagParams{ArchiveID: archive_id, TagID: tag_id})
	if err != nil && isErrorConstraint(err) {
		return nil
	}

	if err != nil {
		return err
	}

	return nil
}

func (s service) RemoveTag(ctx context.Context, archive_id int64, tag string) error {
	err := s.query.RemoveTag(ctx, sqlc.RemoveTagParams{ArchiveID: archive_id, Text: tag})
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

func (s service) GetHashes(ctx context.Context, archive_id int64) (sqlc.HashesChksum, error) {
	h, err := s.query.GetHashes(ctx, archive_id)
	if err != nil {
		return sqlc.HashesChksum{}, err
	}

	return h, nil
}

func (s service) SetHashes(ctx context.Context, archive_id int64, h Hashes) error {
	err := s.query.SetHashes(ctx,
		sqlc.SetHashesParams{
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

	if err := s.query.NewEntry(ctx, sqlc.NewEntryParams{
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

	if err := s.query.SetHashes(ctx, sqlc.SetHashesParams{
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

		if err := s.query.SetTag(ctx, sqlc.SetTagParams{
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

// GetMostRecentTagID returns the most recently created tag_id in tags
func (s service) GetMostRecentTagID(ctx context.Context) (int64, error) {
	a, err := s.query.GetMostRecentTagID(ctx)
	if err != nil {
		return -1, err
	}

	return a, nil
}

// GetTagID searches for a tag that exists in database, regardless of whether
// it is mapped to an entry or not
func (s service) GetTagID(ctx context.Context, tag string) (sqlc.Tag, error) {
	t, err := s.query.GetTagID(ctx, tag)
	if err != nil {
		return sqlc.Tag{}, err
	}

	return t, nil
}

func (s service) GetTagCount(ctx context.Context, tag string) (sqlc.TagCount, error) {
	t, err := s.query.GetTagCount(ctx, tag)
	if err != nil {
		return sqlc.TagCount{}, err
	}

	return t, nil
}

func (s service) SearchTag(ctx context.Context, tag string) ([]sqlc.SearchTagRow, error) {
	t, err := s.query.SearchTag(ctx, tag)
	if err != nil {
		return []sqlc.SearchTagRow{}, err
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
	phash, err := s.query.GetPerceptualHash(ctx, sqlc.GetPerceptualHashParams{ArchiveID: archive_id, HashType: hashType})
	if err != nil {
		return 0, err
	}
	return uint64(phash), nil
}

func (s service) SetPerceptualHash(ctx context.Context, archive_id int64, hashType string, hash uint64) error {
	err := s.query.SetPerceptualHash(ctx, sqlc.SetPerceptualHashParams{ArchiveID: archive_id, HashType: hashType, Hash: int64(hash)})
	if err != nil {
		return err
	}
	return nil
}

func isErrorConstraint(err error) bool {
	if liteErr, ok := err.(*sqlite.Error); ok {
		if liteErr.Code() == 19 { // https://pkg.go.dev/modernc.org/sqlite@v1.28.0/lib#SQLITE_CONSTRAINT
			return true
		}
	}

	return false
}
