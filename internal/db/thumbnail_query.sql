-- name: NewThumbnail :exec
INSERT INTO "thumbnail" (archive_id, has_jpeg, has_webp) VALUES (:archive_id, 0, 0);

-- name: DeleteThumbnail :exec
DELETE FROM "thumbnail" WHERE archive_id == (:archive_id);

-- name: NewJpeg :exec
INSERT OR REPLACE INTO "thumbnail_jpeg" (archive_id, small, medium, large) VALUES (:archive_id, :small, :medium, :large);

-- name: GetJpegsmall :one
SELECT small FROM "thumbnail_jpeg" WHERE archive_id == (:archive_id);

-- name: GetJpegMedium :one
SELECT medium FROM "thumbnail_jpeg" WHERE archive_id == (:archive_id);

-- name: GetJpeglarge :one
SELECT large FROM "thumbnail_jpeg" WHERE archive_id == (:archive_id);

-- name: NewBlurHash :exec
INSERT INTO "thumbnail_blurhash" (archive_id, hash) VALUES (:archive_id, :hash);

-- name: GetBlurHash :one
SELECT hash FROM "thumbnail_blurhash" WHERE archive_id == (:archive_id);

-- name: DoesArchiveIDExist :one
SELECT EXISTS(SELECT archive_id FROM "thumbnail" WHERE archive_id == (:archive_id) LIMIT 1);