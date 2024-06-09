-- name: NewEntry :exec
INSERT INTO archive (path, extension) VALUES (:path, :extension)
RETURNING id;

-- name: GetEntry :one 
SELECT * FROM archive WHERE id == (:archive_id);

-- name: GetEntryPath :one
SELECT path, extension FROM archive WHERE id == (:archive_id);

-- name: GetMostRecentArchiveID :one
SELECT id FROM archive ORDER BY ROWID DESC LIMIT 1;


-- name: NewTag :exec
INSERT OR IGNORE INTO tags (text) VALUES (:tag);

-- name: RemoveTag :exec
DELETE FROM tagmap
	WHERE tagmap.archive_id == (:archive_id) AND
	tagmap.tag_id IN 
		(SELECT tags.tag_id FROM tags 
			INNER JOIN tagmap ON tags.tag_id = tagmap.tag_id 
				WHERE tags.text == (:text));

-- name: DeleteTag :exec
DELETE FROM tags WHERE text == (:tag);

-- name: DeleteTagMap :exec
DELETE FROM tagmap WHERE tag_id == (:tag_id);

-- name: SetTag :exec
INSERT OR IGNORE INTO tagmap 
	(archive_id, tag_id)
VALUES(:archive_id, (SELECT tag_id FROM tags WHERE text = (:tag)));

-- name: GetTagsFromArchiveID :many
SELECT tags.text FROM tags 
	INNER JOIN tagmap ON tags.tag_id = tagmap.tag_id 
WHERE tagmap.archive_id == (:archive_id);

-- name: GetTagID :one
SELECT * FROM tags WHERE text == (:tag);

-- name: SearchTag :many
SELECT archive.id, tags.tag_id, tags.text FROM tags 
	INNER JOIN tagmap ON tagmap.tag_id = tags.tag_id
	INNER JOIN archive ON archive.id = tagmap.archive_id
WHERE tags.text == (:tag);



-- name: GetTimestamps :one
SELECT * FROM timestamps WHERE archive_id == (:archive_id);

-- name: SetTimestamps :exec
INSERT OR REPLACE INTO timestamps 
	(archive_id, date_modified, date_imported)
VALUES (:archive_id, :date_modified, :date_imported);



-- name: GetHashes :one
SELECT * FROM hashes WHERE archive_id == (:archive_id);

-- name: SetHashes :exec
INSERT OR REPLACE INTO hashes 
	(archive_id, md5, sha1, sha256) 
VALUES (:archive_id, :md5, :sha1, :sha256);
