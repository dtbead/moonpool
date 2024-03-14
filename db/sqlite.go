package db

import (
	"database/sql"
	"fmt"
	"log/slog"

	"github.com/dtbead/moonpool/file"
	_ "modernc.org/sqlite"
)

type SQLite3 struct {
	db *sql.DB
}

/*
  || possibly make a function that batches multiple queries together instead of doing it one by one?
*/

// NewDatabase creates a new SQLite3 database with predefined tables for "archive"
func New(filepath string) (SQLite3, error) {
	db, err := sql.Open("sqlite", filepath)
	if err != nil {
		slog.Error("unable to create new database. %v'", err)
		return SQLite3{}, err
	}

	sdb := SQLite3{db}
	if err := sdb.createArchiveSchema(); err != nil {
		return SQLite3{}, err
	}

	return sdb, nil
}

func Open(filepath string) (SQLite3, error) {
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
func (s *SQLite3) createArchiveSchema() error {
	table := `CREATE TABLE archive (
		"id"	INTEGER NOT NULL UNIQUE,
		"md5"	BLOB NOT NULL UNIQUE,
		"path"	TEXT NOT NULL,
		"extension" TEXT,
		PRIMARY KEY("id" AUTOINCREMENT)
	);`

	tags := `CREATE TABLE tags (
		"tagID"	INTEGER NOT NULL UNIQUE,
		"text"	TEXT NOT NULL UNIQUE,
		PRIMARY KEY("tagID" AUTOINCREMENT)
	);`

	tagmap := `CREATE TABLE tagmap (
		"archiveID"	INTEGER NOT NULL UNIQUE,
		"tagID"	INTEGER NOT NULL UNIQUE,
		FOREIGN KEY("tagID") REFERENCES "tag"("tagID"),
		FOREIGN KEY("archiveID") REFERENCES "archive"("id")
	);`

	query := fmt.Sprint(table, tags, tagmap)
	_, err := s.db.Exec(query)
	if err != nil {
		return err
	}

	return nil
}

func (s *SQLite3) AddTag(tag string) error {
	query := `INSERT INTO tags (text) VALUES (?);`
	res, err := s.db.Exec(query, tag)
	if err != nil {
		return err
	}

	slog.Debug("sqlite: inserted tag with result:'", res, "'")
	return nil
}

func (s *SQLite3) AddTags(tags []Tag) error {
	query := `INSERT INTO tags (text) VALUES (?)`
	stmt, err := s.db.Prepare(query)
	if err != nil {
		return err
	}
	defer stmt.Close()

	for i := 0; i < len(tags); i++ {
		res, err := stmt.Exec(tags[i].Text)
		if err != nil {
			return err
		}
		rowsAffected, err := res.RowsAffected()
		if err != nil {
			fmt.Println(err)
		}
		StatusCode, err := res.LastInsertId()
		if err != nil {
			fmt.Println(err)
		}
		slog.Info(fmt.Sprintf("sqlite: inserted tag with %d rows affected and last status code as %d", rowsAffected, StatusCode))
	}

	return nil
}

// searchTagID searches for a TagID with a given string. Returns -1 if tag does not exist.
func (s *SQLite3) searchTagID(tag string) (TagID, error) {
	query := `SELECT tagID FROM tags WHERE text == (?);`
	stmt, err := s.db.Query(query, tag)
	if err != nil {
		return -1, err
	}
	defer stmt.Close()

	var res TagID = -1
	for stmt.Next() {
		stmt.Scan(&res)
	}
	return res, nil
}

// SearchTag searches for an entry in archive if tag is mapped to said entry.
func (s *SQLite3) SearchTag(tag string) ([]file.Entry, error) {
	tagID, err := s.searchTagID(tag)
	if err != nil {
		return nil, err
	}

	archiveIDs, err := s.searchTagMaps(tagID)
	if err != nil {
		return nil, err
	}

	stmt, err := s.db.Prepare("SELECT md5, path FROM archive WHERE id == ?")
	if err != nil {
		return []file.Entry{}, err
	}
	defer stmt.Close()

	entries := make([]file.Entry, len(archiveIDs))

	cnt := 0
	for i := 0; i < len(archiveIDs); i++ {
		rows, err := stmt.Query(archiveIDs[i])
		if err != nil {
			slog.Error(err.Error())
		}

		for rows.Next() {
			rows.Scan(&entries[cnt].Metadata.MD5Hash, &entries[cnt].Metadata.Path)
			cnt++
		}
	}

	return entries, nil
}

// given tagID, return archive_id
func (s *SQLite3) searchTagMaps(t TagID) ([]int, error) {
	stmt, err := s.db.Prepare("SELECT archiveID FROM tagmap WHERE tagID == ?")
	if err != nil {
		return []int{}, err
	}
	defer stmt.Close()

	rows, err := stmt.Query(t)
	if err != nil {
		return []int{}, err
	}

	var res []int
	var tmp int
	for rows.Next() {
		rows.Scan(&tmp)
		res = append(res, tmp)
	}

	return res, nil
}

func (s *SQLite3) mapTags(a ArchiveID, tags []Tag) error {
	query := `INSERT OR IGNORE INTO tagmap (archiveID, tagID)
	VALUES(?, (SELECT  tagID FROM tags WHERE text = ?));`

	stmt, err := s.db.Prepare(query)
	if err != nil {
		return err
	}
	defer stmt.Close()

	for i := 0; i < len(tags); i++ {
		res, err := stmt.Exec(a, tags[i].Text)
		if err != nil {
			return err
		}
		slog.Debug("sqlite: mapped tag with result:'", res, "'")
	}
	return nil
}

func (s *SQLite3) InsertEntry(md5hash, path, extension string) error {
	query := `INSERT INTO archive (md5, path, extension) VALUES (?, ?, ?);`
	res, err := s.db.Exec(query, md5hash, path, extension)
	if err != nil {
		return err
	}

	rows, err := res.RowsAffected()
	if err != nil {
		slog.Error(err.Error())
	}
	slog.Info(fmt.Sprintf("inserted %d entries to archive", rows))

	return nil
}

func (s *SQLite3) getTotalResults(table, row, value string) int {
	cnt := 0

	query := fmt.Sprintf(`SELECT count(*) FROM %s WHERE %s == (?);`, table, row)
	stmt, err := s.db.Query(query, value)
	if err != nil {
		return cnt
	}
	defer stmt.Close()

	for stmt.Next() {
		if err = stmt.Scan(&cnt); err != nil {
			return cnt
		}
	}
	return cnt
}

func (s *SQLite3) doesColumnExist(table, column string) bool {
	query := fmt.Sprintf(`SELECT COUNT(*) AS CNTREC FROM pragma_table_info('%s') WHERE name='%s';`, table, column)
	stmt, err := s.db.Query(query, column)
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

func (s *SQLite3) dumpSchema() map[string][]string {
	tables := s.dumpTables()
	schema := make(map[string][]string)
	for _, v := range tables {
		schema[v] = s.dumpColumns(v)
	}

	return schema
}

func (s *SQLite3) dumpTables() []string {
	query := `SELECT name FROM sqlite_master WHERE type='table';`
	stmt, err := s.db.Query(query)
	if err != nil {
		return []string{}
	}
	defer stmt.Close()

	var res string
	ress := []string{}
	for stmt.Next() {
		stmt.Scan(&res) // should probably check for errors
		ress = append(ress, res)
	}

	return ress
}

func (s *SQLite3) dumpColumns(table string) []string {
	query := fmt.Sprintf(`SELECT name FROM pragma_table_info('%s');`, table)
	stmt, err := s.db.Query(query)
	if err != nil {
		return []string{}
	}
	defer stmt.Close()

	var res string
	ress := []string{}
	for stmt.Next() {
		stmt.Scan(&res) // should probably check for errors
		ress = append(ress, res)
	}

	return ress
}

func (s *SQLite3) dumpTable(table string) interface{} {
	query := fmt.Sprintf(`SELECT * FROM %s`, table)
	stmt, err := s.db.Query(query)
	if err != nil {
		fmt.Println(err)
		return []string{}
	}

	var res interface{}
	for stmt.Next() {
		stmt.Scan(&res)
	}

	return res
}
