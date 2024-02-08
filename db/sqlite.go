package db

import (
	"database/sql"
	"errors"
	"fmt"
	"log/slog"

	"github.com/dtbead/moonpool/tags"
	_ "modernc.org/sqlite"
)

type SQLite3 struct {
	db *sql.DB
}

/*
  || possibly make a function that batches multiple queries together instead of doing it one by one?
*/

// NewDatabase creates a new SQLite3 database with predefined tables for "archive"
func NewDatabase(filepath string) (SQLite3, error) {
	db, err := sql.Open("sqlite", filepath)
	if err != nil {
		slog.Error("unable to create new database. %v'", err)
		return SQLite3{nil}, err
	}
	sdb := SQLite3{db}
	sdb.createArchiveDatabaseScheme("archive")

	return sdb, nil
}

func OpenConnection(filepath string) (SQLite3, error) {
	db, err := sql.Open("sqlite", filepath)
	if err != nil {
		return SQLite3{}, err
	}
	return SQLite3{db}, nil
}

func (s *SQLite3) Close() error {
	return s.db.Close()
}

// TODO: figure out more necessary tables
func (s *SQLite3) createArchiveDatabaseScheme(TableName string) error {
	query := fmt.Sprintf(`CREATE TABLE "%s" (
		"id"	INTEGER NOT NULL UNIQUE,
		"md5"	BLOB NOT NULL UNIQUE,
		"path"	TEXT NOT NULL,
		"extension" TEXT,
		PRIMARY KEY("id" AUTOINCREMENT)
	)`, sanitizeString(TableName))
	slog.Debug("sqlite: creating database scheme with '", query, "'")

	_, err := s.db.Exec(query, TableName)
	if err != nil {
		return err
	}

	query = `CREATE TABLE "tags" (
		"tag_id"	INTEGER NOT NULL UNIQUE,
		"text"	TEXT NOT NULL UNIQUE,
		PRIMARY KEY("tag_id" AUTOINCREMENT)
	)`
	_, err = s.db.Exec(query, TableName)
	if err != nil {
		return err
	}

	query = fmt.Sprintf(`CREATE TABLE "tagmap" (
		"%s_id"	INTEGER NOT NULL UNIQUE,
		"tag_id"	INTEGER NOT NULL UNIQUE,
		FOREIGN KEY("tag_id") REFERENCES "tag"("tag_id"),
		FOREIGN KEY("%s_id") REFERENCES "%s"("id")
	)`, sanitizeString(TableName), sanitizeString(TableName), sanitizeString(TableName))
	_, err = s.db.Exec(query, TableName)
	if err != nil {
		return err
	}

	return nil
}

func (s *SQLite3) AddTag(table, tag string) error {
	query := fmt.Sprintf(`INSERT INTO %s (text) VALUES (?);`, table)
	res, err := s.db.Exec(query, tag)
	if err != nil {
		return err
	}

	slog.Debug("sqlite: inserted tag with result:'", res, "'")
	return nil
}

func (s *SQLite3) AddTags(tags []tags.Tag) error {
	query := `INSERT INTO tags (text) VALUES (?);`
	stmt, err := s.db.Prepare(query)
	if err != nil {
		return err
	}
	defer stmt.Close()

	for i := 0; i < len(tags); i++ {
		res, err := stmt.Exec(tags[i])
		if err != nil {
			return err
		}
		slog.Debug("sqlite: inserted tag with result:'", res, "'")
	}
	return nil
}

func (s *SQLite3) FindTagID(tag string) uint {
	query := `SELECT tag_id FROM tags WHERE text == (?);`
	stmt, err := s.db.Query(query, tag)
	if err != nil {
		slog.Debug("sqlite: unexpected error when searching tag. %v", err)
		return 0
	}
	var res uint
	for stmt.Next() {
		stmt.Scan(&res)
	}
	return uint(res)
}

func (s *SQLite3) MapTags(archiveID uint, tags []tags.Tag) error {
	query := `INSERT OR IGNORE INTO tagmap (archive_id, tag_id)
	VALUES (?, (SELECT tag_id FROM tags WHERE text == ?));`

	stmt, err := s.db.Prepare(query)
	if err != nil {
		return err
	}
	defer stmt.Close()

	for i := 0; i < len(tags); i++ {
		res, err := stmt.Exec(archiveID, tags[i])
		if err != nil {
			return err
		}
		slog.Debug("sqlite: mapped tag with result:'", res, "'")
	}
	return nil
}

func (s *SQLite3) SearchTag(tag string) uint {
	query := (`SELECT tag_id FROM tags WHERE text == ?`)

	stmt, err := s.db.Query(query, tag)
	if err != nil {
		return 0
	}
	defer stmt.Close()

	var res uint
	for stmt.Next() {
		stmt.Scan(&res) // should probably check for errors
		return res
	}

	return 0
}

func (s *SQLite3) InsertToArchive(table, md5hash, path, extension string) error {
	query := fmt.Sprintf(`INSERT INTO %s (md5, path, extension) VALUES (?, ?, ?);`, sanitizeString(table))
	stmt, err := s.db.Exec(query, md5hash, path, extension)
	if err != nil {
		return err
	}
	if i, _ := stmt.RowsAffected(); i == 0 {
		return errors.New("InsertToArchive didn't insert any new rows")
	}
	return nil
}

func (s *SQLite3) SingleQuery(table, row, value string) (string, error) {
	if !s.doesTableExist(table) {
		return "", errors.New("sqlite: refusing to query for non-existant table")
	}

	if !s.doesColumnExist(table, row) {
		return "", errors.New("sqlite: refusing to query for non-existant row")
	}

	query := fmt.Sprintf(`SELECT * FROM %s WHERE %s == (?);`, table, row)
	stmt, err := s.db.Query(query, value)
	if err != nil {
		return "", err
	}
	defer stmt.Close()

	if c, _ := stmt.Columns(); len(c) > 1 {
		return "", errors.New("sqlite: query has more than one row")
	}

	var res string
	for stmt.Next() {
		if err = stmt.Scan(&res); err != nil {
			return "", err
		}
	}

	return res, nil
}

func (s *SQLite3) doesColumnExist(table, row string) bool {
	query := fmt.Sprintf(`SELECT COUNT(*) AS CNTREC FROM pragma_table_info('%s') WHERE name='%s'
	`, table, row)
	stmt, err := s.db.Query(query, row)
	if err != nil {
		return false
	}
	defer stmt.Close()

	var res int
	for stmt.Next() {
		stmt.Scan(&res) // should probably check for errors
		if res == 1 {
			return true
		}
	}

	return false
}

func (s *SQLite3) doesTableExist(table string) bool {
	query := `SELECT name FROM sqlite_master WHERE type='table';`
	stmt, err := s.db.Query(query, table)
	if err != nil {
		return false
	}
	defer stmt.Close()

	var res string
	for stmt.Next() {
		stmt.Scan(&res) // should probably check for errors
		if res == table {
			return true
		}
	}

	return false
}
