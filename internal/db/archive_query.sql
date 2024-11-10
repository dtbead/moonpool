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

-- name: SetTag :exec
INSERT INTO tag_map 
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

-- name: GetTagCount :one
SELECT tag_count.tag_id, tag_count.total FROM tag_count 
JOIN tags ON tags.tag_id = tag_count.tag_id
WHERE tags.text == (:tag);

-- name: GetTagRange :many
SELECT tags.text, count(tags.text) FROM tags 
INNER JOIN tag_map ON tags.tag_id = tag_map.tag_id 
WHERE tag_map.archive_id BETWEEN (:start) AND (:end)
GROUP BY tags.text
ORDER BY count(tags.text) DESC 
LIMIT (:limit) OFFSET (:offset);

-- name: SearchTag :many
SELECT archive.id, tags.tag_id, tags.text FROM tags 
	INNER JOIN tag_map ON tag_map.tag_id = tags.tag_id
	INNER JOIN archive ON archive.id = tag_map.archive_id
WHERE tags.text == (:tag);



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

-- name: SetMetadata :exec
INSERT OR REPLACE INTO "archive_metadata"
	(archive_id, file_size, file_mimetype, media_width, media_height, media_orientation)
VALUES (:archive_id, :file_size, :file_mimetype, :media_width, :media_height, :media_orientation);

-- name: GetMetadata :one
SELECT * FROM "archive_metadata" WHERE archive_id == (:archive_id);