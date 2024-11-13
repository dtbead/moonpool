package archive

import (
	"context"
	"database/sql"
	_ "embed"
	"errors"
	"io"
	"os"
	"strings"
	"time"

	"github.com/dtbead/moonpool/entry"
	"github.com/dtbead/moonpool/internal/db"
	"modernc.org/sqlite"
)

type archive struct {
	query *Queries
	db    *sql.DB
}

type TX interface {
	Commit() error
	Rollback() error
}

type Archiver interface {
	NewEntry(ctx context.Context, path, extension string) (int64, error)
	GetEntry(ctx context.Context, archive_id int64) (Archive, error)
	GetPage(ctx context.Context, sort string, limit, offset int) ([]Archive, error)
	DeleteEntry(ctx context.Context, archive_id int64) error
	RemoveTags(ctx context.Context, archive_id int64) error
	GetFile(ctx context.Context, archive_id int64, baseDirectory string) (io.ReadCloser, error)
	GetTags(ctx context.Context, archive_id int64) ([]string, error)
	GetTagCount(ctx context.Context, tag string) (int64, error)
	GetTagCountByRange(ctx context.Context, start, end, limit, offset int64) (entry.TagCount, error)
	GetTagCountByList(ctx context.Context, archive_ids []int64, limit int) (entry.TagCount, error)
	SetTimestamps(ctx context.Context, archive_id int64, t db.Timestamp) error
	GetTimestamps(ctx context.Context, archive_id int64) (db.Timestamp, error)
	NewTag(ctx context.Context, tag string) (int64, error)
	SetTag(ctx context.Context, archive_id int64, tag string) error
	RemoveTag(ctx context.Context, archive_id int64, tag string) error
	GetTagID(ctx context.Context, tag string) (Tag, error)
	SearchTag(ctx context.Context, tag string) ([]SearchTagRow, error)
	GetHashes(ctx context.Context, archive_id int64) (HashesChksum, error)
	SetHashes(ctx context.Context, archive_id int64, h Hashes) error
	GetPerceptualHash(ctx context.Context, archive_id int64, hashType string) (uint64, error)
	SetPerceptualHash(ctx context.Context, archive_id int64, hashType string, hash uint64) error
	DeleteTag(ctx context.Context, tag string) error
	GetMostRecentArchiveID(ctx context.Context) (int64, error)
	GetMostRecentTagID(ctx context.Context) (int64, error)
	DoesArchiveIDExist(ctx context.Context, id int64) bool
	NewTx(ctx context.Context, opt *sql.TxOptions) (Querier, TX, error)
	NewSavepoint(ctx context.Context, name string) error
	ReleaseSavepoint(ctx context.Context, name string) error
	Rollback(ctx context.Context, name string) error
	ForceCheckpoint(ctx context.Context) error
}

type Hashes struct {
	MD5, SHA1, SHA256 []byte
}

// BeginTx() initiates a transaction.
func (a archive) NewTx(ctx context.Context, opt *sql.TxOptions) (Querier, TX, error) {
	tx, err := a.db.BeginTx(ctx, opt)
	if err != nil {
		return nil, nil, err
	}
	q := New(tx)
	return q, tx, nil
}

func (a archive) NewSavepoint(ctx context.Context, name string) error {
	if !db.IsClean(name) {
		return errors.New("invalid name")
	}

	if _, err := a.db.ExecContext(ctx, "SAVEPOINT "+name); err != nil {
		return err
	}
	return nil
}

func (a archive) ReleaseSavepoint(ctx context.Context, name string) error {
	if !db.IsClean(name) {
		return errors.New("invalid name")
	}

	if _, err := a.db.ExecContext(ctx, "RELEASE "+name); err != nil {
		return err
	}

	return nil
}

func (a archive) Rollback(ctx context.Context, name string) error {
	if !db.IsClean(name) {
		return errors.New("invalid name")
	}

	_, err := a.db.ExecContext(ctx, "ROLLBACK TO "+name)
	return err
}

func (a archive) ForceCheckpoint(ctx context.Context) error {
	_, err := a.db.ExecContext(ctx, "PRAGMA wal_checkpoint;")
	return err
}

func NewArchiver(q *Queries, db *sql.DB) Archiver {
	return &archive{
		query: q,
		db:    db,
	}
}

// NewEntry() inserts a new entry into the archive. Each new entry is given a 'archive_id' that externally referred to as "id"
// to the user. NewEntry returns an integer archive_id if successful; returns -1 and error otherwise
func (a archive) NewEntry(ctx context.Context, path, extension string) (int64, error) {
	var archive_id int64

	if extension == "" || path == "" {
		return -1, errors.New("empty path or extension")
	}

	if err := a.query.NewEntry(ctx, NewEntryParams{
		Path:      path,
		Extension: extension,
	}); err != nil {
		return -1, err
	}

	archive_id, err := a.GetMostRecentArchiveID(ctx)
	if err != nil {
		return -1, err
	}

	return archive_id, nil
}

func (a archive) GetEntry(ctx context.Context, archive_id int64) (Archive, error) {
	e, err := a.query.GetEntry(ctx, archive_id)
	if err != nil {
		return Archive{}, err
	}

	return e, nil
}

func (a archive) DeleteEntry(ctx context.Context, archive_id int64) error {
	return a.query.DeleteEntry(ctx, archive_id)
}

func (a archive) GetFile(ctx context.Context, archive_id int64, baseDirectory string) (io.ReadCloser, error) {
	e, err := a.GetEntry(ctx, archive_id)
	if err != nil {
		return nil, err
	}

	if !strings.HasSuffix(baseDirectory, "/") {
		baseDirectory += "/"
	}

	f, err := os.OpenFile(baseDirectory+e.Path, os.O_RDONLY, os.ModeType)
	if err != nil {
		return nil, err
	}

	return f, nil
}

func (a archive) GetTags(ctx context.Context, archive_id int64) ([]string, error) {
	return a.query.GetTagsFromArchiveID(ctx, archive_id)
}

func (a archive) RemoveTags(ctx context.Context, archive_id int64) error {
	return a.query.RemoveTagsFromArchiveID(ctx, archive_id)
}

func (a archive) SetTimestamps(ctx context.Context, archive_id int64, t db.Timestamp) error {
	return a.query.SetTimestamps(ctx, SetTimestampsParams{
		ArchiveID:    archive_id,
		DateCreated:  t.DateCreated.UTC().UnixMilli(),
		DateModified: t.DateModified.UTC().UnixMilli(),
		DateImported: t.DateImported.UTC().UnixMilli(),
	})
}

func (a archive) GetTimestamps(ctx context.Context, archive_id int64) (db.Timestamp, error) {
	t, err := a.query.GetTimestamps(ctx, archive_id)
	if err != nil {
		return db.Timestamp{}, err
	}

	return db.Timestamp{
		DateCreated:  time.UnixMilli(t.DateCreated),
		DateModified: time.UnixMilli(t.DateModified),
		DateImported: time.UnixMilli(t.DateImported),
	}, nil
}

// NewTag() creates a new tag in the database that can be later mapped to an entry. NewTag() will return a
// tag_id if tag already exists.
func (a archive) NewTag(ctx context.Context, tag string) (int64, error) {
	err := a.query.NewTag(ctx, tag)
	if err != nil && IsErrorConstraint(err) {
		tag, err := a.GetTagID(ctx, tag)
		if err != nil {
			return -1, err
		}
		return tag.TagID, nil
	}

	if err != nil {
		return -1, err
	}

	tag_id, err := a.GetMostRecentTagID(ctx)
	if err != nil {
		return -1, err
	}

	return tag_id, nil
}

// SetTag() assigns a tag to a given archive_id and returns an error if tag does not already exist.
func (a archive) SetTag(ctx context.Context, archive_id int64, tag string) error {
	tag_id, err := a.NewTag(ctx, tag)
	if !IsErrorConstraint(err) && err != nil {
		return err
	}

	// tag has already been assigned to archive_id
	err = a.query.SetTag(ctx, SetTagParams{ArchiveID: archive_id, TagID: tag_id})
	if IsErrorConstraint(err) {
		return nil
	}

	if err != nil {
		return err
	}

	return nil
}

func (a archive) RemoveTag(ctx context.Context, archive_id int64, tag string) error {
	err := a.query.RemoveTag(ctx, RemoveTagParams{ArchiveID: archive_id, Text: tag})
	if err != nil {
		return err
	}

	return nil
}

func (a archive) DeleteTag(ctx context.Context, tag string) error {
	err := a.query.DeleteTag(ctx, tag)
	if err != nil {
		return err
	}

	return nil
}

func (a archive) DeleteTagMap(ctx context.Context, tag_id int64) error {
	err := a.query.DeleteTagMap(ctx, tag_id)
	if err != nil {
		return err
	}

	return nil
}

func (a archive) GetHashes(ctx context.Context, archive_id int64) (HashesChksum, error) {
	h, err := a.query.GetHashes(ctx, archive_id)
	if err != nil {
		return HashesChksum{}, err
	}

	return h, nil
}

func (a archive) SetHashes(ctx context.Context, archive_id int64, h Hashes) error {
	err := a.query.SetHashes(ctx,
		SetHashesParams{
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

// GetMostRecentArchiveID() returns the most recently inserted entry in the archive
func (a archive) GetMostRecentArchiveID(ctx context.Context) (int64, error) {
	archive_id, err := a.query.GetMostRecentArchiveID(ctx)
	if err != nil {
		return -1, err
	}

	return archive_id, nil
}

// GetMostRecentTagID() returns the most recently created tag_id in tags
func (a archive) GetMostRecentTagID(ctx context.Context) (int64, error) {
	tag_id, err := a.query.GetMostRecentTagID(ctx)
	if err != nil {
		return -1, err
	}

	return tag_id, nil
}

// GetTagID() searches for a tag that exists in database, regardless of whether
// it is mapped to an entry or not
func (a archive) GetTagID(ctx context.Context, tag string) (Tag, error) {
	t, err := a.query.GetTagID(ctx, tag)
	if err != nil {
		return Tag{}, err
	}

	return t, nil
}

// GetTagCount counts the total amount of archive_id's that are assigned to a tag
func (a archive) GetTagCount(ctx context.Context, tag string) (int64, error) {
	t, err := a.query.GetTagCountByTag(ctx, tag)
	if err != nil {
		return -1, err
	}

	return t.Total, nil
}

// GetTagCountByList groups the total amount of tags that are assigned to a list of archive_id's.
// entry.TagCount is implicitly sorted from largest to smallest
func (a archive) GetTagCountByList(ctx context.Context, archive_ids []int64, limit int) (entry.TagCount, error) {
	t, err := a.query.GetTagCountByList(ctx, GetTagCountByListParams{archive_ids, limit})
	if err != nil {
		return nil, err
	}

	e := make(entry.TagCount, len(t))
	for i, v := range t {
		e[i] = v
	}

	return e, nil
}

// GetTagCountByList groups the total amount of tags that are within a range of archive_id's.
// offset is the starting point in which to begin grouping each archive_id.
// entry.TagCount is implicitly sorted from largest to smallest
func (a archive) GetTagCountByRange(ctx context.Context, start, end, limit, offset int64) (entry.TagCount, error) {
	t, err := a.query.GetTagCountByRange(ctx, GetTagCountByRangeParams{
		Start:  start,
		End:    end,
		Limit:  limit,
		Offset: offset,
	})

	if err != nil {
		return nil, err
	}

	e := make(entry.TagCount, len(t))
	for i, v := range t {
		e[i] = v
	}

	return e, nil
}

func (a archive) GetPage(ctx context.Context, sort string, limit, offset int) ([]Archive, error) {
	var err error
	var p []Archive

	switch sort {
	case "imported":
		p, err = a.query.GetPagesByDateImported(ctx, GetPagesByDateImportedParams{offset, limit})
	case "created":
		p, err = a.query.GetPagesByDateImported(ctx, GetPagesByDateImportedParams{offset, limit})
	case "modified":
		p, err = a.query.GetPagesByDateImported(ctx, GetPagesByDateImportedParams{offset, limit})
	default:
		return nil, errors.New("invalid sort argument")
	}
	if err != nil {
		return nil, err
	}

	return p, nil
}

func (a archive) SearchTag(ctx context.Context, tag string) ([]SearchTagRow, error) {
	t, err := a.query.SearchTag(ctx, tag)
	if err != nil {
		return []SearchTagRow{}, err
	}

	return t, nil
}

func (a archive) DoesArchiveIDExist(ctx context.Context, archive_id int64) bool {
	res := a.db.QueryRowContext(ctx, `SELECT EXISTS(SELECT 1 FROM archive WHERE id == ? LIMIT 1);`, archive_id)

	var ret = 0
	res.Scan(&ret)

	return ret == 1
}

func (a archive) GetPerceptualHash(ctx context.Context, archive_id int64, hashType string) (uint64, error) {
	phash, err := a.query.GetPerceptualHash(ctx, GetPerceptualHashParams{ArchiveID: archive_id, HashType: hashType})
	if err != nil {
		return 0, err
	}
	return uint64(phash), nil
}

func (a archive) SetPerceptualHash(ctx context.Context, archive_id int64, hashType string, hash uint64) error {
	err := a.query.SetPerceptualHash(ctx, SetPerceptualHashParams{ArchiveID: archive_id, HashType: hashType, Hash: int64(hash)})
	if err != nil {
		return err
	}
	return nil
}

func IsErrorConstraint(err error) bool {
	if liteErr, ok := err.(*sqlite.Error); ok {
		if liteErr.Code() == 19 || liteErr.Code() == 2067 { // https://pkg.go.dev/modernc.org/sqlite@v1.28.0/lib#SQLITE_CONSTRAINT
			return true
		}
	}

	return false
}
