package archive

import (
	"context"
	"database/sql"
	_ "embed"
	"errors"
	"fmt"
	"io"
	"os"
	"slices"
	"strings"
	"time"

	"github.com/dtbead/moonpool/entry"
	"github.com/dtbead/moonpool/internal/db"
	"github.com/dtbead/moonpool/internal/db/migration"
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
	GetPage(ctx context.Context, sort string, limit, offset int64, desc bool) ([]Archive, error)
	DeleteEntry(ctx context.Context, archive_id int64) error
	RemoveTags(ctx context.Context, archive_id int64) error
	GetFile(ctx context.Context, archive_id int64, baseDirectory string) (io.ReadCloser, error)
	GetTags(ctx context.Context, archive_id int64) ([]string, error)
	GetTagCount(ctx context.Context, tag string) (int64, error)
	GetTagCountByRange(ctx context.Context, start, end, limit, offset int64) ([]entry.TagCount, error)
	GetFileMetadata(ctx context.Context, archive_id int64) (entry.FileMetadata, error)
	SetFileMetadata(ctx context.Context, archive_id int64, m entry.FileMetadata) error
	GetTagCountByList(ctx context.Context, archive_ids []int64) ([]entry.TagCount, error)
	SetTimestamps(ctx context.Context, archive_id int64, t db.Timestamp) error
	GetTimestamps(ctx context.Context, archive_id int64) (db.Timestamp, error)
	NewTag(ctx context.Context, tag string) (int64, error)
	NewTagAlias(ctx context.Context, alias_tag, base_tag string) error
	DeleteTagAlias(ctx context.Context, alias_tag string) error
	ResolveTagAlias(ctx context.Context, alias_tag string) (entry.TagAlias, error)
	ResolveTagAliasList(ctx context.Context, alias_tag []string) ([]entry.TagAlias, error)
	AssignTag(ctx context.Context, archive_id int64, tag string) error
	AssignTags(ctx context.Context, archive_id int64, tags []string) error
	RemoveTag(ctx context.Context, archive_id int64, tag string) error
	GetTagID(ctx context.Context, tag string) (Tag, error)
	SearchTag(ctx context.Context, tag string) ([]SearchTagRow, error)
	SearchTagByList(ctx context.Context, sort, order string, tags_include, tags_exclude []string) ([]int64, error)
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

type SortPageOptions struct {
	Offset, Index int64
}

// BeginTx initiates a transaction.
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

// NewEntry inserts a new entry into the archive. Each new entry is given an incremented 'archive_id' of int64.
// NewEntry returns an integer archive_id if successful; returns -1 and error otherwise.
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

func (a archive) GetFileMetadata(ctx context.Context, archive_id int64) (entry.FileMetadata, error) {
	m, err := a.query.GetFileMetadata(ctx, archive_id)
	if err != nil {
		return entry.FileMetadata{}, err
	}

	return entry.FileMetadata{
		FileSize:         m.FileSize,
		FileMimetype:     m.FileMimetype.(string),
		MediaHeight:      m.MediaHeight.Int64,
		MediaWidth:       m.MediaWidth.Int64,
		MediaOrientation: m.MediaOrientation.String,
	}, nil
}

func (a archive) SetFileMetadata(ctx context.Context, archive_id int64, m entry.FileMetadata) error {
	return a.query.SetFileMetadata(ctx, SetFileMetadataParams{
		ArchiveID:        archive_id,
		FileSize:         m.FileSize,
		FileMimetype:     m.FileMimetype,
		MediaWidth:       sql.NullInt64{Int64: m.MediaWidth, Valid: true},
		MediaHeight:      sql.NullInt64{Int64: m.MediaHeight, Valid: true},
		MediaOrientation: sql.NullString{String: m.MediaOrientation, Valid: true},
	})
}

func (a archive) GetTags(ctx context.Context, archive_id int64) ([]string, error) {
	tags, err := a.query.GetTagsFromArchiveID(ctx, archive_id)
	if !errors.Is(err, sql.ErrNoRows) && err != nil {
		return nil, err
	}
	return tags, nil
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
	if !errors.Is(err, sql.ErrNoRows) && err != nil {
		return db.Timestamp{}, err
	}

	return db.Timestamp{
		DateCreated:  time.UnixMilli(t.DateCreated),
		DateModified: time.UnixMilli(t.DateModified),
		DateImported: time.UnixMilli(t.DateImported),
	}, nil
}

// NewTag creates a new tag in the database that can be later mapped to an entry.
// NewTag will return a tag_id if tag already exists.
func (a archive) NewTag(ctx context.Context, tag string) (int64, error) {
	tag = db.DeleteWhitespace(tag)

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

// NewTagAlias creates a new tag alias that references an existing tag. alias_tag is the new alias tag to create, and
// base_tag is the existing tag that alias_tag references. A base tag cannot be an alias tag and vice versa.
func (a archive) NewTagAlias(ctx context.Context, alias_tag, base_tag string) error {
	alias_tag = db.DeleteWhitespace(alias_tag)
	base_tag = db.DeleteWhitespace(base_tag)

	t, err := a.GetTagID(ctx, alias_tag)
	if !errors.Is(err, sql.ErrNoRows) && err != nil {
		return err
	}

	if t.TagID >= 1 {
		return errors.New("tag exists as base tag")
	}

	return a.query.NewTagAlias(ctx, NewTagAliasParams{BaseTag: base_tag, AliasTag: alias_tag})
}

func (a archive) DeleteTagAlias(ctx context.Context, alias_tag string) error {
	alias_tag = db.DeleteWhitespace(alias_tag)
	return a.query.DeleteTagAlias(ctx, alias_tag)
}

// ResolveTagAlias returns the base tag that is associated to an alias tag.
func (a archive) ResolveTagAlias(ctx context.Context, alias_tag string) (entry.TagAlias, error) {
	alias_tag = db.DeleteWhitespace(alias_tag)

	res, err := a.query.ResolveTagAlias(ctx, alias_tag)
	if err != nil {
		return entry.TagAlias{}, err
	}

	return entry.TagAlias{TagID: res.TagID, BaseTag: res.Text, AliasTag: res.Text_2}, nil
}

// ResolveTagAlias returns a slice of base tag that is associated to a slice of alias tags.
func (a archive) ResolveTagAliasList(ctx context.Context, alias_tag []string) ([]entry.TagAlias, error) {
	for i, tag := range alias_tag {
		alias_tag[i] = db.DeleteWhitespace(tag)
	}

	res, err := a.query.ResolveTagAliasList(ctx, alias_tag)
	if err != nil {
		return nil, err
	}

	alias := make([]entry.TagAlias, len(res))
	for i, v := range res {
		alias[i].TagID = v.TagID
		alias[i].BaseTag = v.Text
		alias[i].AliasTag = v.Text_2
	}

	return alias, nil
}

// AssignTag assigns a tag to a given archive_id. A new tag will be created if one does not already
// exist. SetTag will automatically resolve any tag alias to a "base" tag if possible.
func (a archive) AssignTag(ctx context.Context, archive_id int64, tag string) error {
	tag = db.DeleteWhitespace(tag)

	base_tag, err := a.ResolveTagAlias(ctx, tag)
	if !errors.Is(err, sql.ErrNoRows) && err != nil {
		return err
	}

	// tag is an alias tag, no need to create new tag in 'tags' table
	if base_tag.TagID >= 1 {
		err = a.query.AssignTag(ctx, AssignTagParams{ArchiveID: archive_id, TagID: base_tag.TagID})
		if !IsErrorConstraint(err) && err != nil {
			return err
		}

		return nil
	}

	tag_id, err := a.NewTag(ctx, tag)
	if !IsErrorConstraint(err) && err != nil {
		return err
	}

	err = a.query.AssignTag(ctx, AssignTagParams{ArchiveID: archive_id, TagID: tag_id})
	if !IsErrorConstraint(err) && err != nil {
		return err
	}

	return nil
}

// AssignTags assigns a slice of tags to a given archive_id. A new tag will be created if one does not already
// exist. AssignTags will automatically resolve any tag alias to a "base" tag if possible.
func (a archive) AssignTags(ctx context.Context, archive_id int64, tags []string) error {
	err := a.NewSavepoint(ctx, "assigntags")
	if err != nil {
		return err
	}
	defer a.Rollback(ctx, "assigntags")

	var tag_id int64
	for _, tag := range tags {
		tag = db.DeleteWhitespace(tag)
		if tag != "" {
			t, err := a.GetTagID(ctx, tag)
			if !errors.Is(err, sql.ErrNoRows) && err != nil {
				return err
			}

			if t == (Tag{}) {
				tag_id, err = a.NewTag(ctx, tag)
				if err != nil {
					return err
				}
			} else {
				tag_id = t.TagID
			}

			err = a.query.AssignTag(ctx, AssignTagParams{ArchiveID: archive_id, TagID: tag_id})
			if !IsErrorConstraint(err) && err != nil {
				return err
			}
		}
	}
	return a.ReleaseSavepoint(ctx, "assigntags")
}

func (a archive) RemoveTag(ctx context.Context, archive_id int64, tag string) error {
	tag = db.DeleteWhitespace(tag)

	err := a.query.RemoveTag(ctx, RemoveTagParams{ArchiveID: archive_id, Text: tag})
	if err != nil {
		return err
	}

	return nil
}

func (a archive) DeleteTag(ctx context.Context, tag string) error {
	tag = db.DeleteWhitespace(tag)

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
	if !errors.Is(err, sql.ErrNoRows) && err != nil {
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

// GetMostRecentArchiveID returns the most recently inserted entry in an archive.
func (a archive) GetMostRecentArchiveID(ctx context.Context) (int64, error) {
	archive_id, err := a.query.GetMostRecentArchiveID(ctx)
	if err != nil {
		return -1, err
	}

	return archive_id, nil
}

// GetMostRecentTagID returns the most recently created tag_id.
func (a archive) GetMostRecentTagID(ctx context.Context) (int64, error) {
	tag_id, err := a.query.GetMostRecentTagID(ctx)
	if err != nil {
		return -1, err
	}

	return tag_id, nil
}

// GetTagID searches for an existing tag in the database, regardless of whether
// it is mapped to an entry or not. Tag aliases are automatically resolved.
func (a archive) GetTagID(ctx context.Context, tag string) (Tag, error) {
	tag = db.DeleteWhitespace(tag)

	tag_alias, _ := a.ResolveTagAlias(ctx, tag)
	if tag_alias != (entry.TagAlias{}) {
		return Tag{TagID: tag_alias.TagID, Text: tag_alias.BaseTag}, nil
	}

	t, err := a.query.GetTagID(ctx, tag)
	if err != nil {
		return Tag{}, err
	}

	return t, nil
}

// GetTagCount counts the total amount of archive_id's that are assigned to a tag
func (a archive) GetTagCount(ctx context.Context, tag string) (int64, error) {
	tag = db.DeleteWhitespace(tag)

	t, err := a.query.GetTagCountByTag(ctx, tag)
	if err != nil {
		return -1, err
	}

	return t.Total, nil
}

// GetTagCountByList groups the total amount of tags that are assigned to a list of archive_id's.
// entry. TagCount is implicitly sorted from largest to smallest.
func (a archive) GetTagCountByList(ctx context.Context, archive_ids []int64) ([]entry.TagCount, error) {
	if archive_ids == nil {
		return []entry.TagCount{}, nil
	}

	t, err := a.query.GetTagCountByList(ctx, archive_ids)
	if err != nil {
		return nil, err
	}

	e := make([]entry.TagCount, len(t))
	for i, v := range t {
		e[i].Count = v.Count
		e[i].Text = v.Text
	}

	return e, nil
}

// GetTagCountByList groups the total amount of tags that are within a range of archive_id's.
// offset is the starting point in which to begin grouping each archive_id.
// entry.TagCount is implicitly sorted from largest to smallest
func (a archive) GetTagCountByRange(ctx context.Context, start, end, limit, offset int64) ([]entry.TagCount, error) {
	t, err := a.query.GetTagCountByRange(ctx, GetTagCountByRangeParams{
		Start:  start,
		End:    end,
		Limit:  limit,
		Offset: offset,
	})

	if err != nil {
		return nil, err
	}

	e := make([]entry.TagCount, len(t))
	for i, v := range t {
		e[i].Count = v.Count
		e[i].Text = v.Text
	}

	return e, nil
}

// GetPage returns a list of entries's based on the given search options.
//
// Valid sort options are "imported", "created" and "modified".
// limit limits how many total entries to return.
// offset is the pagination of a search result.
// For example, calling GetPage with a limit of 50, and an offset of 0 would return archive_id's between 1-50.
// Calling GetPage with the same limit of 50, and an offset of 50 would return an archive_id's 50-100.
func (a archive) GetPage(ctx context.Context, sort string, limit, offset int64, desc bool) ([]Archive, error) {
	var err error
	res := new(sql.Rows)

	const query = `SELECT id, path, extension FROM archive 
INNER JOIN archive_timestamps ON archive.id = archive_timestamps.archive_id
ORDER BY archive_timestamps.%s %s LIMIT %d OFFSET %d`

	order := "DESC"
	if !desc {
		order = "ASC"
	}

	switch sort {
	case "imported":
		res, err = a.db.QueryContext(ctx, fmt.Sprintf(query, "date_imported", order, limit, offset))
	case "created":
		res, err = a.db.QueryContext(ctx, fmt.Sprintf(query, "date_created", order, limit, offset))
	case "modified":
		res, err = a.db.QueryContext(ctx, fmt.Sprintf(query, "date_modified", order, limit, offset))
	default:
		return nil, errors.New("invalid sort argument")
	}

	defer res.Close()
	if err != nil {
		return nil, err
	}

	archiveList := make([]Archive, 0, 50)
	var archive Archive

	for res.Next() {
		if err := res.Scan(&archive.ID, &archive.Path, &archive.Extension); err != nil {
			return nil, err
		}
		archiveList = append(archiveList, archive)
	}

	if err := res.Err(); err != nil {
		return nil, err
	}

	return archiveList, nil
}

func (a archive) SearchTag(ctx context.Context, tag string) ([]SearchTagRow, error) {
	tag = db.DeleteWhitespace(tag)

	t, err := a.query.SearchTag(ctx, tag)
	if !errors.Is(err, sql.ErrNoRows) && err != nil {
		return []SearchTagRow{}, err
	}

	return t, nil
}

// Valid sort options are "imported", "created", and "modified".
// Valid order options are "descending", "ascending".
func (a archive) SearchTagByList(ctx context.Context, sort, order string, tags_include, tags_exclude []string) ([]int64, error) {
	if tags_include == nil {
		return nil, nil
	}

	for i, tag := range tags_include {
		tags_include[i] = db.DeleteWhitespace(tag)
	}

	for i, tag := range tags_exclude {
		tags_exclude[i] = db.DeleteWhitespace(tag)
	}

	tags, err := a.ResolveTagAliasList(ctx, slices.Concat(tags_include, tags_exclude))
	if err != nil {
		return nil, err
	}

	// TODO: integrate this procedure globally for all other search funcs
	// or handle all resolving purely in the DB
	for _, resolved := range tags {
		for i, tag := range tags_include {
			if resolved.AliasTag == tag {
				tags_include[i] = resolved.BaseTag
			}
		}
	}

	for _, resolved := range tags {
		for i, tag := range tags_exclude {
			if resolved.AliasTag == tag {
				tags_exclude[i] = resolved.BaseTag
			}
		}
	}

	archive_ids := make([]int64, 0, 50)

	switch sort {
	default:
		return nil, errors.New("invalid sort or order option")
	case "created":
		res, err := a.query.SearchTagsByListDateCreated(ctx, SearchTagsByListDateCreatedParams{tags_include, tags_exclude})
		if err != nil {
			return nil, err
		}

		for _, v := range res {
			archive_ids = append(archive_ids, v.ID)
		}
	case "imported":
		res, err := a.query.SearchTagsByListDateImported(ctx, SearchTagsByListDateImportedParams{tags_include, tags_exclude})
		if err != nil {
			return nil, err
		}

		for _, v := range res {
			archive_ids = append(archive_ids, v.ID)
		}
	case "modified":
		res, err := a.query.SearchTagsByListDateModified(ctx, SearchTagsByListDateModifiedParams{tags_include, tags_exclude})
		if err != nil {
			return nil, err
		}

		for _, v := range res {
			archive_ids = append(archive_ids, v.ID)
		}
	}

	// SearchTagsByList* returns results in descending order by default
	if order == "ascending" {
		slices.Reverse(archive_ids)
	}

	return archive_ids, nil
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

// GetVersion returns the moonpool database version.
func (a archive) GetVersion(ctx context.Context) (int64, error) {
	const SQL_GET_USER_VERSION = `PRAGMA main.user_version;`

	res, err := a.db.QueryContext(ctx, SQL_GET_USER_VERSION)
	if err != nil {
		return -1, err
	}

	var userVersion int64
	if res.Next() {
		err = res.Scan(&userVersion)
		if err != nil {
			return -1, err
		}
	}

	return userVersion, nil
}

// SetVersion sets the moonpool database version.
func (a archive) SetVersion(ctx context.Context, version int64) error {
	const SQL_UPDATE_USER_VERSION = `PRAGMA main.user_version = (?);`

	res, err := a.db.ExecContext(ctx, SQL_UPDATE_USER_VERSION, version)
	if err != nil {
		return err
	}

	r, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if r != 1 {
		return errors.New("no rows affected")
	}

	return nil
}
func (a archive) Migrate(ctx context.Context, m migration.Migrator) error {
	tx, err := a.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	for _, q := range m.Queries() {
		_, err = tx.ExecContext(ctx, q)
		if err != nil {
			return err
		}
	}

	v, err := a.GetVersion(ctx)
	if err != nil {
		return err
	}

	err = a.SetVersion(ctx, v+1)
	if err != nil {
		return err
	}

	err = tx.Commit()
	if err != nil {
		return err
	}

	return nil
}

func IsErrorConstraint(err error) bool {
	if liteErr, ok := err.(*sqlite.Error); ok {
		if liteErr.Code() == 19 || liteErr.Code() == 2067 || liteErr.Code() == 787 { // https://pkg.go.dev/modernc.org/sqlite@v1.28.0/lib#SQLITE_CONSTRAINT
			return true
		}
	}

	return false
}
