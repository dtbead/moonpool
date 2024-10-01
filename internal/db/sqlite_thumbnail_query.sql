-- Name: NewThumbnailJPEG :exec
INSERT INTO jpeg (archive_id, low, medium, high) VALUES (:archive_id, :low, :medium, :high);

-- Name: NewThumbnailWebP :exec
INSERT INTO webp (archive_id, low, medium, high) VALUES (:archive_id, :low, :medium, :high);