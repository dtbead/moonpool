package db

import (
	"github.com/dtbead/moonpool/tags"
)

const (
	DB_TYPE_SQLITE3 = iota
	DB_TYPE_POSTGRES
)

type TagMap struct {
	TagID uint
	Text  string
}

type Database interface {
	AddTags(tags []tags.Tag) error
	InsertToArchive(table, md5hash, path, extension string) error
	FindTagID(tag string) uint
	MapTags(archiveID uint, tags []tags.Tag) error
	SearchTag(tag string) uint
	SingleQuery(table, row, value string) (string, error)
	doesColumnExist(table, row string) bool
	doesTableExist(table string) bool
	Close() error
}

func OpenSQLite3(filepath string) *SQLite3 {
	SQLdb, _ := OpenConnection(filepath)
	return &SQLdb
}
