// Code generated by sqlc. DO NOT EDIT.
// versions:
//   sqlc v1.28.0

package archive

import (
	"database/sql"
)

type Archive struct {
	ID        int64
	Path      string
	Extension string
}

type ArchiveMetadatum struct {
	ArchiveID        int64
	FileSize         int64
	FileMimetype     interface{}
	MediaWidth       sql.NullInt64
	MediaHeight      sql.NullInt64
	MediaOrientation sql.NullString
}

type ArchiveTimestamp struct {
	ArchiveID    int64
	DateModified int64
	DateImported int64
	DateCreated  int64
}

type HashesChksum struct {
	ArchiveID int64
	Md5       []byte
	Sha1      []byte
	Sha256    []byte
}

type HashesPerceptual struct {
	ArchiveID int64
	HashType  string
	Hash      int64
}

type Note struct {
	ArchiveID int64
	Title     string
	Text      string
}

type Tag struct {
	TagID int64
	Text  string
}

type TagCount struct {
	TagID int64
	Total int64
}

type TagMap struct {
	TagID     int64
	ArchiveID int64
}

type TagsAlias struct {
	TagID sql.NullInt64
	Text  string
}
