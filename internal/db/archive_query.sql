-- name: NewEntry :exec
INSERT INTO archive (path, extension) VALUES (:path, :extension);

-- name: GetEntry :one 
SELECT * FROM archive WHERE id == (:archive_id);

-- name: GetEntryPath :one
SELECT path, extension FROM archive WHERE id == (:archive_id);

-- name: GetMostRecentArchiveID :one
SELECT id FROM archive ORDER BY ROWID DESC LIMIT 1;

-- name: GetMostRecentTagID :one
SELECT tag_id FROM tags ORDER BY ROWID DESC LIMIT 1;

-- name: NewTag :exec
INSERT INTO tags (text) VALUES (:tag);

-- name: NewTagAlias :exec
INSERT INTO tags_alias (tag_id, text) VALUES 
	((SELECT tag_id FROM tags WHERE tags.text == :base_tag), :alias_tag);

-- name: DeleteTagAlias :exec
DELETE FROM tags_alias WHERE text == (:tag);

-- name: ResolveTagAlias :one
SELECT tags.tag_id, tags.text, tags_alias.text FROM tags
	INNER JOIN tags_alias on tags.tag_id = tags_alias.tag_id
WHERE tags.tag_id == (SELECT tag_id FROM tags_alias WHERE tags_alias.Text == (:alias_tag));

-- name: ResolveTagAliasList :many
SELECT tags.tag_id, tags.text, tags_alias.text FROM tags
	INNER JOIN tags_alias on tags.tag_id = tags_alias.tag_id
WHERE tags.tag_id IN 
	(SELECT tag_id FROM tags_alias WHERE tags_alias.text IN (sqlc.slice('alias_tags')));

-- name: RemoveTag :exec
DELETE FROM tag_map
	WHERE tag_map.archive_id == (:archive_id) AND
	tag_map.tag_id IN 
		(SELECT tags.tag_id FROM tags 
			INNER JOIN tag_map ON tags.tag_id = tag_map.tag_id 
				WHERE tags.text == (:text));

-- name: DeleteTag :exec
DELETE FROM tags WHERE text == (:tag);

-- name: DeleteTagMap :exec
DELETE FROM tag_map WHERE tag_id == (:tag_id);

-- name: DeleteEntry :exec
DELETE from archive WHERE id == (:archive_id);

-- name: AssignTag :exec
INSERT OR IGNORE INTO tag_map 
	(archive_id, tag_id)
VALUES(:archive_id, (:tag_id));

-- name: GetTagsFromArchiveID :many
SELECT tags.text FROM tags 
	INNER JOIN tag_map ON tags.tag_id = tag_map.tag_id 
WHERE tag_map.archive_id == (:archive_id);

-- name: RemoveTagsFromArchiveID :exec
DELETE FROM tag_map WHERE archive_id == (:archive_id);

-- name: GetTagID :one
SELECT * FROM tags WHERE text == (:tag);

-- name: GetTagCountByTag :one
SELECT tag_count.tag_id, tag_count.total FROM tag_count 
JOIN tags ON tags.tag_id = tag_count.tag_id
WHERE tags.text == (:tag);

-- name: GetTagCountByRange :many
SELECT tags.text, count(tags.text) FROM tags 
INNER JOIN tag_map ON tags.tag_id = tag_map.tag_id 
WHERE tag_map.archive_id BETWEEN (:start) AND (:end)
GROUP BY tags.text
ORDER BY count(tags.text) DESC, tags.text ASC
LIMIT (:limit) OFFSET (:offset);

-- name: GetTagCountByList :many
SELECT tags.text, count(tags.text) FROM tags 
INNER JOIN tag_map ON tags.tag_id = tag_map.tag_id 
INNER JOIN archive ON tag_map.archive_id = archive.id
WHERE archive.id IN (sqlc.slice('archive_ids'))
GROUP BY tags.text
ORDER BY count(tags.text) DESC, tags.text ASC LIMIT 50;

-- name: SearchTag :many
SELECT archive.id, tags.tag_id, tags.text FROM tags 
	INNER JOIN tag_map ON tag_map.tag_id = tags.tag_id
	INNER JOIN archive ON archive.id = tag_map.archive_id
WHERE tags.text == (:tag);

-- name: SearchHash :one
SELECT archive.id FROM archive 
	INNER JOIN hashes_chksum ON hashes_chksum.archive_id = archive.id
WHERE
hashes_chksum.md5 == unhex((:hash)) OR
hashes_chksum.sha1 == unhex((:hash)) OR
hashes_chksum.sha256 == unhex((:hash));

-- name: SearchTagsByListDateCreated :many
SELECT DISTINCT archive.id, archive_timestamps.date_created FROM tags 
	INNER JOIN tag_map ON tag_map.tag_id = tags.tag_id
	INNER JOIN archive ON archive.id = tag_map.archive_id
	INNER JOIN archive_timestamps ON archive.id = archive_timestamps.archive_id
WHERE tags.text IN (sqlc.slice('tags_include'))
EXCEPT
	SELECT DISTINCT archive.id, archive_timestamps.date_created FROM tags
	INNER JOIN tag_map ON tag_map.tag_id = tags.tag_id
	INNER JOIN archive ON archive.id = tag_map.archive_id
	INNER JOIN archive_timestamps ON archive.id = archive_timestamps.archive_id
	WHERE tags.text IN (sqlc.slice('tags_exclude'))
ORDER BY archive_timestamps.date_created DESC;

-- name: SearchTagsByListDateImported :many
SELECT DISTINCT archive.id, archive_timestamps.date_imported FROM tags 
	INNER JOIN tag_map ON tag_map.tag_id = tags.tag_id
	INNER JOIN archive ON archive.id = tag_map.archive_id
	INNER JOIN archive_timestamps ON archive.id = archive_timestamps.archive_id
WHERE tags.text IN (sqlc.slice('tags_include'))
EXCEPT
	SELECT DISTINCT archive.id, archive_timestamps.date_imported FROM tags
	INNER JOIN tag_map ON tag_map.tag_id = tags.tag_id
	INNER JOIN archive ON archive.id = tag_map.archive_id
	INNER JOIN archive_timestamps ON archive.id = archive_timestamps.archive_id
	WHERE tags.text IN (sqlc.slice('tags_exclude'))
ORDER BY archive_timestamps.date_imported DESC;

-- name: SearchTagsByListDateModified :many
SELECT DISTINCT archive.id, archive_timestamps.date_modified FROM tags 
	INNER JOIN tag_map ON tag_map.tag_id = tags.tag_id
	INNER JOIN archive ON archive.id = tag_map.archive_id
	INNER JOIN archive_timestamps ON archive.id = archive_timestamps.archive_id
WHERE tags.text IN (sqlc.slice('tags_include'))
EXCEPT
	SELECT DISTINCT archive.id, archive_timestamps.date_modified FROM tags
	INNER JOIN tag_map ON tag_map.tag_id = tags.tag_id
	INNER JOIN archive ON archive.id = tag_map.archive_id
	INNER JOIN archive_timestamps ON archive.id = archive_timestamps.archive_id
	WHERE tags.text IN (sqlc.slice('tags_exclude'))
ORDER BY archive_timestamps.date_modified DESC;

-- name: GetTimestamps :one
SELECT * FROM archive_timestamps WHERE archive_id == (:archive_id);

-- name: SetTimestamps :exec
INSERT OR REPLACE INTO archive_timestamps 
	(archive_id, date_modified, date_imported, date_created)
VALUES (:archive_id, :date_modified, :date_imported, :date_created);

-- name: GetHashes :one
SELECT * FROM hashes_chksum WHERE archive_id == (:archive_id);

-- name: SetHashes :exec
INSERT OR REPLACE INTO hashes_chksum 
	(archive_id, md5, sha1, sha256) 
VALUES (:archive_id, :md5, :sha1, :sha256);

-- name: GetPerceptualHash :one
SELECT hash FROM hashes_perceptual 
WHERE archive_id == (:archive_id) AND hash_type == (:hash_type);

-- name: SetPerceptualHash :exec
INSERT OR REPLACE INTO hashes_perceptual
	(archive_id, hash_type, hash)
VALUES (:archive_id, :hash_type, :hash);

-- name: GetPagesByDateCreated :many
SELECT id, path, extension FROM archive 
INNER JOIN archive_timestamps ON archive.id = archive_timestamps.archive_id
ORDER BY archive_timestamps.date_created LIMIT (:limit) OFFSET (:offset);

-- name: GetPagesByDateCreatedDescending :many
SELECT id, path, extension FROM archive 
INNER JOIN archive_timestamps ON archive.id = archive_timestamps.archive_id
ORDER BY archive_timestamps.date_created DESC LIMIT (:limit) OFFSET (:offset);

-- name: GetPagesByDateModifiedAscending :many
SELECT id, path, extension FROM archive 
INNER JOIN archive_timestamps ON archive.id = archive_timestamps.archive_id
ORDER BY archive_timestamps.date_modified ASC LIMIT (:limit) OFFSET (:offset);

-- name: GetPagesByDateModifiedDescending :many
SELECT id, path, extension FROM archive 
INNER JOIN archive_timestamps ON archive.id = archive_timestamps.archive_id
ORDER BY archive_timestamps.date_modified DESC LIMIT (:limit) OFFSET (:offset);

-- name: GetPagesByDateImportedAscending :many
SELECT id, path, extension FROM archive 
INNER JOIN archive_timestamps ON archive.id = archive_timestamps.archive_id
ORDER BY archive_timestamps.date_imported ASC LIMIT (:limit) OFFSET (:offset);

-- name: GetPagesByDateImportedDecending :many
SELECT id, path, extension FROM archive 
INNER JOIN archive_timestamps ON archive.id = archive_timestamps.archive_id
ORDER BY archive_timestamps.date_imported DESC LIMIT (:limit) OFFSET (:offset);

-- name: SetFileMetadata :exec
INSERT OR REPLACE INTO "archive_metadata"
	(archive_id, file_size, file_mimetype, media_width, media_height, media_orientation)
VALUES (:archive_id, :file_size, :file_mimetype, :media_width, :media_height, :media_orientation);

-- name: GetFileMetadata :one
SELECT * FROM "archive_metadata" WHERE archive_id == (:archive_id);