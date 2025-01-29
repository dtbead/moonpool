// Code generated by sqlc. DO NOT EDIT.
// versions:
//   sqlc v1.27.0
// source: archive_query.sql

package archive

import (
	"context"
	"database/sql"
	"strings"
)

const AssignTag = `-- name: AssignTag :exec
INSERT OR IGNORE INTO tag_map 
	(archive_id, tag_id)
VALUES(?1, (?2))
`

type AssignTagParams struct {
	ArchiveID int64
	TagID     interface{}
}

func (q *Queries) AssignTag(ctx context.Context, arg AssignTagParams) error {
	_, err := q.exec(ctx, q.assignTagStmt, AssignTag, arg.ArchiveID, arg.TagID)
	return err
}

const DeleteEntry = `-- name: DeleteEntry :exec
DELETE from archive WHERE id == (?1)
`

func (q *Queries) DeleteEntry(ctx context.Context, archiveID int64) error {
	_, err := q.exec(ctx, q.deleteEntryStmt, DeleteEntry, archiveID)
	return err
}

const DeleteTag = `-- name: DeleteTag :exec
DELETE FROM tags WHERE text == (?1)
`

func (q *Queries) DeleteTag(ctx context.Context, tag string) error {
	_, err := q.exec(ctx, q.deleteTagStmt, DeleteTag, tag)
	return err
}

const DeleteTagAlias = `-- name: DeleteTagAlias :exec
DELETE FROM tags_alias WHERE text == (?1)
`

func (q *Queries) DeleteTagAlias(ctx context.Context, tag string) error {
	_, err := q.exec(ctx, q.deleteTagAliasStmt, DeleteTagAlias, tag)
	return err
}

const DeleteTagMap = `-- name: DeleteTagMap :exec
DELETE FROM tag_map WHERE tag_id == (?1)
`

func (q *Queries) DeleteTagMap(ctx context.Context, tagID int64) error {
	_, err := q.exec(ctx, q.deleteTagMapStmt, DeleteTagMap, tagID)
	return err
}

const GetEntry = `-- name: GetEntry :one
SELECT id, path, extension FROM archive WHERE id == (?1)
`

func (q *Queries) GetEntry(ctx context.Context, archiveID int64) (Archive, error) {
	row := q.queryRow(ctx, q.getEntryStmt, GetEntry, archiveID)
	var i Archive
	err := row.Scan(&i.ID, &i.Path, &i.Extension)
	return i, err
}

const GetEntryPath = `-- name: GetEntryPath :one
SELECT path, extension FROM archive WHERE id == (?1)
`

type GetEntryPathRow struct {
	Path      string
	Extension string
}

func (q *Queries) GetEntryPath(ctx context.Context, archiveID int64) (GetEntryPathRow, error) {
	row := q.queryRow(ctx, q.getEntryPathStmt, GetEntryPath, archiveID)
	var i GetEntryPathRow
	err := row.Scan(&i.Path, &i.Extension)
	return i, err
}

const GetHashes = `-- name: GetHashes :one
SELECT archive_id, md5, sha1, sha256 FROM hashes_chksum WHERE archive_id == (?1)
`

func (q *Queries) GetHashes(ctx context.Context, archiveID int64) (HashesChksum, error) {
	row := q.queryRow(ctx, q.getHashesStmt, GetHashes, archiveID)
	var i HashesChksum
	err := row.Scan(
		&i.ArchiveID,
		&i.Md5,
		&i.Sha1,
		&i.Sha256,
	)
	return i, err
}

const GetMetadata = `-- name: GetMetadata :one
SELECT archive_id, file_size, file_mimetype, media_width, media_height, media_orientation FROM "archive_metadata" WHERE archive_id == (?1)
`

func (q *Queries) GetMetadata(ctx context.Context, archiveID int64) (ArchiveMetadata, error) {
	row := q.queryRow(ctx, q.getMetadataStmt, GetMetadata, archiveID)
	var i ArchiveMetadata
	err := row.Scan(
		&i.ArchiveID,
		&i.FileSize,
		&i.FileMimetype,
		&i.MediaWidth,
		&i.MediaHeight,
		&i.MediaOrientation,
	)
	return i, err
}

const GetMostRecentArchiveID = `-- name: GetMostRecentArchiveID :one
SELECT id FROM archive ORDER BY ROWID DESC LIMIT 1
`

func (q *Queries) GetMostRecentArchiveID(ctx context.Context) (int64, error) {
	row := q.queryRow(ctx, q.getMostRecentArchiveIDStmt, GetMostRecentArchiveID)
	var id int64
	err := row.Scan(&id)
	return id, err
}

const GetMostRecentTagID = `-- name: GetMostRecentTagID :one
SELECT tag_id FROM tags ORDER BY ROWID DESC LIMIT 1
`

func (q *Queries) GetMostRecentTagID(ctx context.Context) (int64, error) {
	row := q.queryRow(ctx, q.getMostRecentTagIDStmt, GetMostRecentTagID)
	var tag_id int64
	err := row.Scan(&tag_id)
	return tag_id, err
}

const GetPagesByDateCreated = `-- name: GetPagesByDateCreated :many
SELECT id, path, extension FROM archive 
INNER JOIN archive_timestamps ON archive.id = archive_timestamps.archive_id
ORDER BY archive_timestamps.date_created LIMIT (?2) OFFSET (?1)
`

type GetPagesByDateCreatedParams struct {
	Offset interface{}
	Limit  interface{}
}

func (q *Queries) GetPagesByDateCreated(ctx context.Context, arg GetPagesByDateCreatedParams) ([]Archive, error) {
	rows, err := q.query(ctx, q.getPagesByDateCreatedStmt, GetPagesByDateCreated, arg.Offset, arg.Limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []Archive
	for rows.Next() {
		var i Archive
		if err := rows.Scan(&i.ID, &i.Path, &i.Extension); err != nil {
			return nil, err
		}
		items = append(items, i)
	}
	if err := rows.Close(); err != nil {
		return nil, err
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return items, nil
}

const GetPagesByDateCreatedDescending = `-- name: GetPagesByDateCreatedDescending :many
SELECT id, path, extension FROM archive 
INNER JOIN archive_timestamps ON archive.id = archive_timestamps.archive_id
ORDER BY archive_timestamps.date_created DESC LIMIT (?2) OFFSET (?1)
`

type GetPagesByDateCreatedDescendingParams struct {
	Offset interface{}
	Limit  interface{}
}

func (q *Queries) GetPagesByDateCreatedDescending(ctx context.Context, arg GetPagesByDateCreatedDescendingParams) ([]Archive, error) {
	rows, err := q.query(ctx, q.getPagesByDateCreatedDescendingStmt, GetPagesByDateCreatedDescending, arg.Offset, arg.Limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []Archive
	for rows.Next() {
		var i Archive
		if err := rows.Scan(&i.ID, &i.Path, &i.Extension); err != nil {
			return nil, err
		}
		items = append(items, i)
	}
	if err := rows.Close(); err != nil {
		return nil, err
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return items, nil
}

const GetPagesByDateImportedAscending = `-- name: GetPagesByDateImportedAscending :many
SELECT id, path, extension FROM archive 
INNER JOIN archive_timestamps ON archive.id = archive_timestamps.archive_id
ORDER BY archive_timestamps.date_imported ASC LIMIT (?2) OFFSET (?1)
`

type GetPagesByDateImportedAscendingParams struct {
	Offset interface{}
	Limit  interface{}
}

func (q *Queries) GetPagesByDateImportedAscending(ctx context.Context, arg GetPagesByDateImportedAscendingParams) ([]Archive, error) {
	rows, err := q.query(ctx, q.getPagesByDateImportedAscendingStmt, GetPagesByDateImportedAscending, arg.Offset, arg.Limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []Archive
	for rows.Next() {
		var i Archive
		if err := rows.Scan(&i.ID, &i.Path, &i.Extension); err != nil {
			return nil, err
		}
		items = append(items, i)
	}
	if err := rows.Close(); err != nil {
		return nil, err
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return items, nil
}

const GetPagesByDateImportedDecending = `-- name: GetPagesByDateImportedDecending :many
SELECT id, path, extension FROM archive 
INNER JOIN archive_timestamps ON archive.id = archive_timestamps.archive_id
ORDER BY archive_timestamps.date_imported DESC LIMIT (?2) OFFSET (?1)
`

type GetPagesByDateImportedDecendingParams struct {
	Offset interface{}
	Limit  interface{}
}

func (q *Queries) GetPagesByDateImportedDecending(ctx context.Context, arg GetPagesByDateImportedDecendingParams) ([]Archive, error) {
	rows, err := q.query(ctx, q.getPagesByDateImportedDecendingStmt, GetPagesByDateImportedDecending, arg.Offset, arg.Limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []Archive
	for rows.Next() {
		var i Archive
		if err := rows.Scan(&i.ID, &i.Path, &i.Extension); err != nil {
			return nil, err
		}
		items = append(items, i)
	}
	if err := rows.Close(); err != nil {
		return nil, err
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return items, nil
}

const GetPagesByDateModifiedAscending = `-- name: GetPagesByDateModifiedAscending :many
SELECT id, path, extension FROM archive 
INNER JOIN archive_timestamps ON archive.id = archive_timestamps.archive_id
ORDER BY archive_timestamps.date_modified ASC LIMIT (?2) OFFSET (?1)
`

type GetPagesByDateModifiedAscendingParams struct {
	Offset interface{}
	Limit  interface{}
}

func (q *Queries) GetPagesByDateModifiedAscending(ctx context.Context, arg GetPagesByDateModifiedAscendingParams) ([]Archive, error) {
	rows, err := q.query(ctx, q.getPagesByDateModifiedAscendingStmt, GetPagesByDateModifiedAscending, arg.Offset, arg.Limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []Archive
	for rows.Next() {
		var i Archive
		if err := rows.Scan(&i.ID, &i.Path, &i.Extension); err != nil {
			return nil, err
		}
		items = append(items, i)
	}
	if err := rows.Close(); err != nil {
		return nil, err
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return items, nil
}

const GetPagesByDateModifiedDescending = `-- name: GetPagesByDateModifiedDescending :many
SELECT id, path, extension FROM archive 
INNER JOIN archive_timestamps ON archive.id = archive_timestamps.archive_id
ORDER BY archive_timestamps.date_modified DESC LIMIT (?2) OFFSET (?1)
`

type GetPagesByDateModifiedDescendingParams struct {
	Offset interface{}
	Limit  interface{}
}

func (q *Queries) GetPagesByDateModifiedDescending(ctx context.Context, arg GetPagesByDateModifiedDescendingParams) ([]Archive, error) {
	rows, err := q.query(ctx, q.getPagesByDateModifiedDescendingStmt, GetPagesByDateModifiedDescending, arg.Offset, arg.Limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []Archive
	for rows.Next() {
		var i Archive
		if err := rows.Scan(&i.ID, &i.Path, &i.Extension); err != nil {
			return nil, err
		}
		items = append(items, i)
	}
	if err := rows.Close(); err != nil {
		return nil, err
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return items, nil
}

const GetPerceptualHash = `-- name: GetPerceptualHash :one
SELECT hash FROM hashes_perceptual 
WHERE archive_id == (?1) AND hash_type == (?2)
`

type GetPerceptualHashParams struct {
	ArchiveID int64
	HashType  string
}

func (q *Queries) GetPerceptualHash(ctx context.Context, arg GetPerceptualHashParams) (int64, error) {
	row := q.queryRow(ctx, q.getPerceptualHashStmt, GetPerceptualHash, arg.ArchiveID, arg.HashType)
	var hash int64
	err := row.Scan(&hash)
	return hash, err
}

const GetTagCountByList = `-- name: GetTagCountByList :many
SELECT tags.text, count(tags.text) FROM tags 
INNER JOIN tag_map ON tags.tag_id = tag_map.tag_id 
INNER JOIN archive ON tag_map.archive_id = archive.id
WHERE archive.id IN (/*SLICE:archive_ids*/?)
GROUP BY tags.text
ORDER BY count(tags.text) DESC, tags.text ASC LIMIT 50
`

type GetTagCountByListRow struct {
	Text  string
	Count int64
}

func (q *Queries) GetTagCountByList(ctx context.Context, archiveIds []int64) ([]GetTagCountByListRow, error) {
	query := GetTagCountByList
	var queryParams []interface{}
	if len(archiveIds) > 0 {
		for _, v := range archiveIds {
			queryParams = append(queryParams, v)
		}
		query = strings.Replace(query, "/*SLICE:archive_ids*/?", strings.Repeat(",?", len(archiveIds))[1:], 1)
	} else {
		query = strings.Replace(query, "/*SLICE:archive_ids*/?", "NULL", 1)
	}
	rows, err := q.query(ctx, nil, query, queryParams...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []GetTagCountByListRow
	for rows.Next() {
		var i GetTagCountByListRow
		if err := rows.Scan(&i.Text, &i.Count); err != nil {
			return nil, err
		}
		items = append(items, i)
	}
	if err := rows.Close(); err != nil {
		return nil, err
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return items, nil
}

const GetTagCountByRange = `-- name: GetTagCountByRange :many
SELECT tags.text, count(tags.text) FROM tags 
INNER JOIN tag_map ON tags.tag_id = tag_map.tag_id 
WHERE tag_map.archive_id BETWEEN (?1) AND (?2)
GROUP BY tags.text
ORDER BY count(tags.text) DESC, tags.text ASC
LIMIT (?4) OFFSET (?3)
`

type GetTagCountByRangeParams struct {
	Start  int64
	End    int64
	Offset interface{}
	Limit  interface{}
}

type GetTagCountByRangeRow struct {
	Text  string
	Count int64
}

func (q *Queries) GetTagCountByRange(ctx context.Context, arg GetTagCountByRangeParams) ([]GetTagCountByRangeRow, error) {
	rows, err := q.query(ctx, q.getTagCountByRangeStmt, GetTagCountByRange,
		arg.Start,
		arg.End,
		arg.Offset,
		arg.Limit,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []GetTagCountByRangeRow
	for rows.Next() {
		var i GetTagCountByRangeRow
		if err := rows.Scan(&i.Text, &i.Count); err != nil {
			return nil, err
		}
		items = append(items, i)
	}
	if err := rows.Close(); err != nil {
		return nil, err
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return items, nil
}

const GetTagCountByTag = `-- name: GetTagCountByTag :one
SELECT tag_count.tag_id, tag_count.total FROM tag_count 
JOIN tags ON tags.tag_id = tag_count.tag_id
WHERE tags.text == (?1)
`

func (q *Queries) GetTagCountByTag(ctx context.Context, tag string) (TagCount, error) {
	row := q.queryRow(ctx, q.getTagCountByTagStmt, GetTagCountByTag, tag)
	var i TagCount
	err := row.Scan(&i.TagID, &i.Total)
	return i, err
}

const GetTagID = `-- name: GetTagID :one
SELECT tag_id, text FROM tags WHERE text == (?1)
`

func (q *Queries) GetTagID(ctx context.Context, tag string) (Tag, error) {
	row := q.queryRow(ctx, q.getTagIDStmt, GetTagID, tag)
	var i Tag
	err := row.Scan(&i.TagID, &i.Text)
	return i, err
}

const GetTagsFromArchiveID = `-- name: GetTagsFromArchiveID :many
SELECT tags.text FROM tags 
	INNER JOIN tag_map ON tags.tag_id = tag_map.tag_id 
WHERE tag_map.archive_id == (?1)
`

func (q *Queries) GetTagsFromArchiveID(ctx context.Context, archiveID int64) ([]string, error) {
	rows, err := q.query(ctx, q.getTagsFromArchiveIDStmt, GetTagsFromArchiveID, archiveID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []string
	for rows.Next() {
		var text string
		if err := rows.Scan(&text); err != nil {
			return nil, err
		}
		items = append(items, text)
	}
	if err := rows.Close(); err != nil {
		return nil, err
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return items, nil
}

const GetTimestamps = `-- name: GetTimestamps :one
SELECT archive_id, date_modified, date_imported, date_created FROM archive_timestamps WHERE archive_id == (?1)
`

func (q *Queries) GetTimestamps(ctx context.Context, archiveID int64) (ArchiveTimestamp, error) {
	row := q.queryRow(ctx, q.getTimestampsStmt, GetTimestamps, archiveID)
	var i ArchiveTimestamp
	err := row.Scan(
		&i.ArchiveID,
		&i.DateModified,
		&i.DateImported,
		&i.DateCreated,
	)
	return i, err
}

const NewEntry = `-- name: NewEntry :exec
INSERT INTO archive (path, extension) VALUES (?1, ?2)
`

type NewEntryParams struct {
	Path      string
	Extension string
}

func (q *Queries) NewEntry(ctx context.Context, arg NewEntryParams) error {
	_, err := q.exec(ctx, q.newEntryStmt, NewEntry, arg.Path, arg.Extension)
	return err
}

const NewTag = `-- name: NewTag :exec
INSERT INTO tags (text) VALUES (?1)
`

func (q *Queries) NewTag(ctx context.Context, tag string) error {
	_, err := q.exec(ctx, q.newTagStmt, NewTag, tag)
	return err
}

const NewTagAlias = `-- name: NewTagAlias :exec
INSERT INTO tags_alias (tag_id, text) VALUES 
	((SELECT tag_id FROM tags WHERE tags.text == ?1), ?2)
`

type NewTagAliasParams struct {
	BaseTag  string
	AliasTag string
}

func (q *Queries) NewTagAlias(ctx context.Context, arg NewTagAliasParams) error {
	_, err := q.exec(ctx, q.newTagAliasStmt, NewTagAlias, arg.BaseTag, arg.AliasTag)
	return err
}

const RemoveTag = `-- name: RemoveTag :exec
DELETE FROM tag_map
	WHERE tag_map.archive_id == (?1) AND
	tag_map.tag_id IN 
		(SELECT tags.tag_id FROM tags 
			INNER JOIN tag_map ON tags.tag_id = tag_map.tag_id 
				WHERE tags.text == (?2))
`

type RemoveTagParams struct {
	ArchiveID int64
	Text      string
}

func (q *Queries) RemoveTag(ctx context.Context, arg RemoveTagParams) error {
	_, err := q.exec(ctx, q.removeTagStmt, RemoveTag, arg.ArchiveID, arg.Text)
	return err
}

const RemoveTagsFromArchiveID = `-- name: RemoveTagsFromArchiveID :exec
DELETE FROM tag_map WHERE archive_id == (?1)
`

func (q *Queries) RemoveTagsFromArchiveID(ctx context.Context, archiveID int64) error {
	_, err := q.exec(ctx, q.removeTagsFromArchiveIDStmt, RemoveTagsFromArchiveID, archiveID)
	return err
}

const ResolveTagAlias = `-- name: ResolveTagAlias :one
SELECT tags.tag_id, tags.text, tags_alias.text FROM tags
	INNER JOIN tags_alias on tags.tag_id = tags_alias.tag_id
WHERE tags.tag_id == (SELECT tag_id FROM tags_alias WHERE tags_alias.Text == (?1))
`

type ResolveTagAliasRow struct {
	TagID  int64
	Text   string
	Text_2 string
}

func (q *Queries) ResolveTagAlias(ctx context.Context, aliasTag string) (ResolveTagAliasRow, error) {
	row := q.queryRow(ctx, q.resolveTagAliasStmt, ResolveTagAlias, aliasTag)
	var i ResolveTagAliasRow
	err := row.Scan(&i.TagID, &i.Text, &i.Text_2)
	return i, err
}

const ResolveTagAliasList = `-- name: ResolveTagAliasList :many
SELECT tags.tag_id, tags.text, tags_alias.text FROM tags
	INNER JOIN tags_alias on tags.tag_id = tags_alias.tag_id
WHERE tags.tag_id IN 
	(SELECT tag_id FROM tags_alias WHERE tags_alias.text IN (/*SLICE:alias_tags*/?))
`

type ResolveTagAliasListRow struct {
	TagID  int64
	Text   string
	Text_2 string
}

func (q *Queries) ResolveTagAliasList(ctx context.Context, aliasTags []string) ([]ResolveTagAliasListRow, error) {
	query := ResolveTagAliasList
	var queryParams []interface{}
	if len(aliasTags) > 0 {
		for _, v := range aliasTags {
			queryParams = append(queryParams, v)
		}
		query = strings.Replace(query, "/*SLICE:alias_tags*/?", strings.Repeat(",?", len(aliasTags))[1:], 1)
	} else {
		query = strings.Replace(query, "/*SLICE:alias_tags*/?", "NULL", 1)
	}
	rows, err := q.query(ctx, nil, query, queryParams...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []ResolveTagAliasListRow
	for rows.Next() {
		var i ResolveTagAliasListRow
		if err := rows.Scan(&i.TagID, &i.Text, &i.Text_2); err != nil {
			return nil, err
		}
		items = append(items, i)
	}
	if err := rows.Close(); err != nil {
		return nil, err
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return items, nil
}

const SearchTag = `-- name: SearchTag :many
SELECT archive.id, tags.tag_id, tags.text FROM tags 
	INNER JOIN tag_map ON tag_map.tag_id = tags.tag_id
	INNER JOIN archive ON archive.id = tag_map.archive_id
WHERE tags.text == (?1)
`

type SearchTagRow struct {
	ID    int64
	TagID int64
	Text  string
}

func (q *Queries) SearchTag(ctx context.Context, tag string) ([]SearchTagRow, error) {
	rows, err := q.query(ctx, q.searchTagStmt, SearchTag, tag)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []SearchTagRow
	for rows.Next() {
		var i SearchTagRow
		if err := rows.Scan(&i.ID, &i.TagID, &i.Text); err != nil {
			return nil, err
		}
		items = append(items, i)
	}
	if err := rows.Close(); err != nil {
		return nil, err
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return items, nil
}

const SearchTagsByListDateCreated = `-- name: SearchTagsByListDateCreated :many
SELECT DISTINCT archive.id, archive_timestamps.date_created FROM tags 
	INNER JOIN tag_map ON tag_map.tag_id = tags.tag_id
	INNER JOIN archive ON archive.id = tag_map.archive_id
	INNER JOIN archive_timestamps ON archive.id = archive_timestamps.archive_id
WHERE tags.text IN (/*SLICE:tags_include*/?)
EXCEPT
	SELECT DISTINCT archive.id, archive_timestamps.date_created FROM tags
	INNER JOIN tag_map ON tag_map.tag_id = tags.tag_id
	INNER JOIN archive ON archive.id = tag_map.archive_id
	INNER JOIN archive_timestamps ON archive.id = archive_timestamps.archive_id
	WHERE tags.text IN (/*SLICE:tags_exclude*/?)
ORDER BY archive_timestamps.date_created DESC
`

type SearchTagsByListDateCreatedParams struct {
	TagsInclude []string
	TagsExclude []string
}

type SearchTagsByListDateCreatedRow struct {
	ID          int64
	DateCreated int64
}

func (q *Queries) SearchTagsByListDateCreated(ctx context.Context, arg SearchTagsByListDateCreatedParams) ([]SearchTagsByListDateCreatedRow, error) {
	query := SearchTagsByListDateCreated
	var queryParams []interface{}
	if len(arg.TagsInclude) > 0 {
		for _, v := range arg.TagsInclude {
			queryParams = append(queryParams, v)
		}
		query = strings.Replace(query, "/*SLICE:tags_include*/?", strings.Repeat(",?", len(arg.TagsInclude))[1:], 1)
	} else {
		query = strings.Replace(query, "/*SLICE:tags_include*/?", "NULL", 1)
	}
	if len(arg.TagsExclude) > 0 {
		for _, v := range arg.TagsExclude {
			queryParams = append(queryParams, v)
		}
		query = strings.Replace(query, "/*SLICE:tags_exclude*/?", strings.Repeat(",?", len(arg.TagsExclude))[1:], 1)
	} else {
		query = strings.Replace(query, "/*SLICE:tags_exclude*/?", "NULL", 1)
	}
	rows, err := q.query(ctx, nil, query, queryParams...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []SearchTagsByListDateCreatedRow
	for rows.Next() {
		var i SearchTagsByListDateCreatedRow
		if err := rows.Scan(&i.ID, &i.DateCreated); err != nil {
			return nil, err
		}
		items = append(items, i)
	}
	if err := rows.Close(); err != nil {
		return nil, err
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return items, nil
}

const SearchTagsByListDateImported = `-- name: SearchTagsByListDateImported :many
SELECT DISTINCT archive.id, archive_timestamps.date_imported FROM tags 
	INNER JOIN tag_map ON tag_map.tag_id = tags.tag_id
	INNER JOIN archive ON archive.id = tag_map.archive_id
	INNER JOIN archive_timestamps ON archive.id = archive_timestamps.archive_id
WHERE tags.text IN (/*SLICE:tags_include*/?)
EXCEPT
	SELECT DISTINCT archive.id, archive_timestamps.date_imported FROM tags
	INNER JOIN tag_map ON tag_map.tag_id = tags.tag_id
	INNER JOIN archive ON archive.id = tag_map.archive_id
	INNER JOIN archive_timestamps ON archive.id = archive_timestamps.archive_id
	WHERE tags.text IN (/*SLICE:tags_exclude*/?)
ORDER BY archive_timestamps.date_imported DESC
`

type SearchTagsByListDateImportedParams struct {
	TagsInclude []string
	TagsExclude []string
}

type SearchTagsByListDateImportedRow struct {
	ID           int64
	DateImported int64
}

func (q *Queries) SearchTagsByListDateImported(ctx context.Context, arg SearchTagsByListDateImportedParams) ([]SearchTagsByListDateImportedRow, error) {
	query := SearchTagsByListDateImported
	var queryParams []interface{}
	if len(arg.TagsInclude) > 0 {
		for _, v := range arg.TagsInclude {
			queryParams = append(queryParams, v)
		}
		query = strings.Replace(query, "/*SLICE:tags_include*/?", strings.Repeat(",?", len(arg.TagsInclude))[1:], 1)
	} else {
		query = strings.Replace(query, "/*SLICE:tags_include*/?", "NULL", 1)
	}
	if len(arg.TagsExclude) > 0 {
		for _, v := range arg.TagsExclude {
			queryParams = append(queryParams, v)
		}
		query = strings.Replace(query, "/*SLICE:tags_exclude*/?", strings.Repeat(",?", len(arg.TagsExclude))[1:], 1)
	} else {
		query = strings.Replace(query, "/*SLICE:tags_exclude*/?", "NULL", 1)
	}
	rows, err := q.query(ctx, nil, query, queryParams...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []SearchTagsByListDateImportedRow
	for rows.Next() {
		var i SearchTagsByListDateImportedRow
		if err := rows.Scan(&i.ID, &i.DateImported); err != nil {
			return nil, err
		}
		items = append(items, i)
	}
	if err := rows.Close(); err != nil {
		return nil, err
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return items, nil
}

const SearchTagsByListDateModified = `-- name: SearchTagsByListDateModified :many
SELECT DISTINCT archive.id, archive_timestamps.date_modified FROM tags 
	INNER JOIN tag_map ON tag_map.tag_id = tags.tag_id
	INNER JOIN archive ON archive.id = tag_map.archive_id
	INNER JOIN archive_timestamps ON archive.id = archive_timestamps.archive_id
WHERE tags.text IN (/*SLICE:tags_include*/?)
EXCEPT
	SELECT DISTINCT archive.id, archive_timestamps.date_modified FROM tags
	INNER JOIN tag_map ON tag_map.tag_id = tags.tag_id
	INNER JOIN archive ON archive.id = tag_map.archive_id
	INNER JOIN archive_timestamps ON archive.id = archive_timestamps.archive_id
	WHERE tags.text IN (/*SLICE:tags_exclude*/?)
ORDER BY archive_timestamps.date_modified DESC
`

type SearchTagsByListDateModifiedParams struct {
	TagsInclude []string
	TagsExclude []string
}

type SearchTagsByListDateModifiedRow struct {
	ID           int64
	DateModified int64
}

func (q *Queries) SearchTagsByListDateModified(ctx context.Context, arg SearchTagsByListDateModifiedParams) ([]SearchTagsByListDateModifiedRow, error) {
	query := SearchTagsByListDateModified
	var queryParams []interface{}
	if len(arg.TagsInclude) > 0 {
		for _, v := range arg.TagsInclude {
			queryParams = append(queryParams, v)
		}
		query = strings.Replace(query, "/*SLICE:tags_include*/?", strings.Repeat(",?", len(arg.TagsInclude))[1:], 1)
	} else {
		query = strings.Replace(query, "/*SLICE:tags_include*/?", "NULL", 1)
	}
	if len(arg.TagsExclude) > 0 {
		for _, v := range arg.TagsExclude {
			queryParams = append(queryParams, v)
		}
		query = strings.Replace(query, "/*SLICE:tags_exclude*/?", strings.Repeat(",?", len(arg.TagsExclude))[1:], 1)
	} else {
		query = strings.Replace(query, "/*SLICE:tags_exclude*/?", "NULL", 1)
	}
	rows, err := q.query(ctx, nil, query, queryParams...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []SearchTagsByListDateModifiedRow
	for rows.Next() {
		var i SearchTagsByListDateModifiedRow
		if err := rows.Scan(&i.ID, &i.DateModified); err != nil {
			return nil, err
		}
		items = append(items, i)
	}
	if err := rows.Close(); err != nil {
		return nil, err
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return items, nil
}

const SetHashes = `-- name: SetHashes :exec
INSERT OR REPLACE INTO hashes_chksum 
	(archive_id, md5, sha1, sha256) 
VALUES (?1, ?2, ?3, ?4)
`

type SetHashesParams struct {
	ArchiveID int64
	Md5       []byte
	Sha1      []byte
	Sha256    []byte
}

func (q *Queries) SetHashes(ctx context.Context, arg SetHashesParams) error {
	_, err := q.exec(ctx, q.setHashesStmt, SetHashes,
		arg.ArchiveID,
		arg.Md5,
		arg.Sha1,
		arg.Sha256,
	)
	return err
}

const SetMetadata = `-- name: SetMetadata :exec
INSERT OR REPLACE INTO "archive_metadata"
	(archive_id, file_size, file_mimetype, media_width, media_height, media_orientation)
VALUES (?1, ?2, ?3, ?4, ?5, ?6)
`

type SetMetadataParams struct {
	ArchiveID        int64
	FileSize         int64
	FileMimetype     string
	MediaWidth       sql.NullInt64
	MediaHeight      sql.NullInt64
	MediaOrientation sql.NullString
}

func (q *Queries) SetMetadata(ctx context.Context, arg SetMetadataParams) error {
	_, err := q.exec(ctx, q.setMetadataStmt, SetMetadata,
		arg.ArchiveID,
		arg.FileSize,
		arg.FileMimetype,
		arg.MediaWidth,
		arg.MediaHeight,
		arg.MediaOrientation,
	)
	return err
}

const SetPerceptualHash = `-- name: SetPerceptualHash :exec
INSERT OR REPLACE INTO hashes_perceptual
	(archive_id, hash_type, hash)
VALUES (?1, ?2, ?3)
`

type SetPerceptualHashParams struct {
	ArchiveID int64
	HashType  string
	Hash      int64
}

func (q *Queries) SetPerceptualHash(ctx context.Context, arg SetPerceptualHashParams) error {
	_, err := q.exec(ctx, q.setPerceptualHashStmt, SetPerceptualHash, arg.ArchiveID, arg.HashType, arg.Hash)
	return err
}

const SetTimestamps = `-- name: SetTimestamps :exec
INSERT OR REPLACE INTO archive_timestamps 
	(archive_id, date_modified, date_imported, date_created)
VALUES (?1, ?2, ?3, ?4)
`

type SetTimestampsParams struct {
	ArchiveID    int64
	DateModified int64
	DateImported int64
	DateCreated  int64
}

func (q *Queries) SetTimestamps(ctx context.Context, arg SetTimestampsParams) error {
	_, err := q.exec(ctx, q.setTimestampsStmt, SetTimestamps,
		arg.ArchiveID,
		arg.DateModified,
		arg.DateImported,
		arg.DateCreated,
	)
	return err
}
