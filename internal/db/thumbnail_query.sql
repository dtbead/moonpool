-- name: NewThumbnail :exec
INSERT INTO "thumbnail" (archive_id, has_jpeg, has_webp) VALUES (:archive_id, 0, 0);

-- name: DeleteThumbnail :exec
DELETE FROM "thumbnail" WHERE archive_id == (:archive_id);

-- name: NewJpeg :exec
INSERT INTO "thumbnail_jpeg" (archive_id, small, medium, large) VALUES (:archive_id, :small, :medium, :large);

-- name: NewWebp :exec
INSERT INTO "thumbnail_webp" (archive_id, small, medium, large) VALUES (:archive_id, :small, :medium, :large);

-- name: GetJpegsmall :one
SELECT small FROM "thumbnail_jpeg" WHERE archive_id == (:archive_id);

-- name: GetJpegMedium :one
SELECT medium FROM "thumbnail_jpeg" WHERE archive_id == (:archive_id);

-- name: GetJpeglarge :one
SELECT large FROM "thumbnail_jpeg" WHERE archive_id == (:archive_id);

-- name: GetWebpSmall :one
SELECT small FROM "thumbnail_webp" WHERE archive_id == (:archive_id);

-- name: GetWebpMedium :one
SELECT medium FROM "thumbnail_webp" WHERE archive_id == (:archive_id);

-- name: GetWebpLarge :one
SELECT large FROM "thumbnail_webp" WHERE archive_id == (:archive_id);

-- name: DoesArchiveIDExist :one
SELECT EXISTS(SELECT archive_id FROM "thumbnail" WHERE archive_id == (:archive_id) LIMIT 1);