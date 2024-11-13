package entry

import (
	"image"
	"time"
)

type Entries struct {
	ArchiveID       int64
	Path, Extension string
}

type Entry struct {
	Metadata   Metadata
	Thumbnails Thumbnail
	Tags       Tag
}

type Metadata struct {
	Hash      Hashes
	Timestamp Timestamp
	Paths     Path
}

type Tags struct {
	ArchiveID int64
	Tags      []Tag
}

type Timestamp struct {
	DateCreated  time.Time
	DateModified time.Time
	DateImported time.Time
}

type Hashes struct {
	MD5, SHA1, SHA256 []byte
}

type Path struct {
	FileRelative, FileExtension           string
	ThumbnailRelative, ThumbnailExtension string
}

type Tag struct {
	Text  string
	TagID int
}

type TagCount []struct {
	Text  string
	Count int64
}

type Thumbnail struct {
	Webp, Jpeg Icons
}
type Icons struct {
	Small, Medium, Large *image.Image
}
