package entry

import (
	"time"
)

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
	FileRelative, FilExtension            string
	ThumbnailRelative, ThumbnailExtension string
}

type Tag struct {
	Text  string
	TagID int
}
