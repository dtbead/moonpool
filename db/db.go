package db

import (
	"github.com/dtbead/moonpool/file"
)

const (
	DB_TYPE_SQLITE3 = iota
	DB_TYPE_POSTGRES
)

type TagID int
type ArchiveID int

type Tag struct {
	ID   TagID
	Text string
}

type Database interface {
	AddTags(tags []Tag) error
	InsertToArchive(table, md5hash, path, extension string) error
	FindTagID(tag string) uint
	MapTags(archiveID uint, tags []Tag) error
	SearchTag(table, tag string) ([]file.Entry, error)
	SingleQuery(table, row, value string) ([]string, error)
	doesColumnExist(table, row string) bool
	doesTableExist(table string) bool
	Close() error
}

func OpenSQLite3(filepath string) (*SQLite3, error) {
	SQLdb, err := Open(filepath)
	if err != nil {
		return nil, err
	}
	return &SQLdb, nil
}
