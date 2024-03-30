package db

import (
	"database/sql"
	"errors"
	"fmt"
	"log/slog"

	"github.com/dtbead/moonpool/file"
	"github.com/dtbead/moonpool/media"
	_ "modernc.org/sqlite"
)

type SQLite3 struct {
	db             *sql.DB
	tx             *sql.Tx
	HasTransaction bool
}

func New(filepath string) (SQLite3, error) {
	db, err := sql.Open("sqlite", filepath)
	if err != nil {
		slog.Error("unable to create new database. %v'", err)
		return SQLite3{}, err
	}

	sdb := SQLite3{db, nil, false}
	if err := sdb.Initialize(); err != nil {
		return SQLite3{}, err
	}

	return sdb, nil
}

func Open(filepath string) (SQLite3, error) {
	db, err := sql.Open("sqlite", filepath)
	if err != nil {
		return SQLite3{}, err
	}

	return SQLite3{db, nil, false}, nil
}

func (s *SQLite3) Close() error {
	return s.db.Close()
}

// TODO: figure out more necessary tables
func (s *SQLite3) Initialize() error {
	table := `CREATE TABLE archive (
		"id"	INTEGER NOT NULL UNIQUE,
		"path"	TEXT NOT NULL,
		"extension" TEXT,
		PRIMARY KEY("id" AUTOINCREMENT)
	);`

	hashes := `CREATE TABLE hashes (
		"archiveID"	INTEGER NOT NULL UNIQUE,
		"md5"	BLOB NOT NULL UNIQUE,
		"sha1"	BLOB NOT NULL UNIQUE,
		"sha256"	BLOB NOT NULL UNIQUE,
		FOREIGN KEY("archiveID") REFERENCES "archive"("id")
	);`

	tags := `CREATE TABLE tags (
		"tagID"	INTEGER NOT NULL UNIQUE,
		"text"	TEXT NOT NULL UNIQUE,
		PRIMARY KEY("tagID" AUTOINCREMENT)
	);`

	tagmap := `CREATE TABLE tagmap (
		"archiveID"	INTEGER NOT NULL,
		"tagID"	INTEGER NOT NULL UNIQUE,
		FOREIGN KEY("tagID") REFERENCES "tags"("tagID"),
		FOREIGN KEY("archiveID") REFERENCES "archive"("id")
	);`

	query := fmt.Sprint(table, hashes, tags, tagmap)
	_, err := s.db.Exec(query)
	if err != nil {
		return err
	}

	return nil
}

func (s *SQLite3) TXBegin() error {
	tx, err := s.db.Begin()
	if err != nil {
		return err
	}

	s.tx = tx
	s.HasTransaction = true

	return nil
}

func (s *SQLite3) TXRollback() error {
	if !s.HasTransaction {
		return errors.New("database has no transaction to rollback to")
	}

	if err := s.tx.Rollback(); err != nil {
		return err
	}

	s.HasTransaction = false
	return nil
}

func (s *SQLite3) TXCommit() error {
	if !s.HasTransaction {
		return errors.New("database has no transaction to commit to")
	}

	if err := s.tx.Commit(); err != nil {
		return err
	}

	s.HasTransaction = false

	return nil
}

func (s *SQLite3) AddTag(tag string) error {
	var res sql.Result
	var err error

	query := `INSERT OR IGNORE INTO tags (text) VALUES (?);`

	if s.HasTransaction {
		res, err = s.tx.Exec(query, tag)
	} else {
		res, err = s.db.Exec(query, tag)
	}
	if err != nil {
		return err
	}

	slog.Debug("sqlite: inserted tag with result:'", res, "'")

	return nil
}

func (s *SQLite3) AddTags(t []string) ([]media.Tag, error) {
	var stmt *sql.Stmt
	var tags []media.Tag
	var err error

	query := `INSERT OR IGNORE INTO tags (text) VALUES (?)`

	if s.HasTransaction {
		stmt, err = s.tx.Prepare(query)
	} else {
		stmt, err = s.db.Prepare(query)
	}

	if err != nil {
		return nil, err
	}
	defer stmt.Close()

	for i := 0; i < len(t); i++ {
		res, err := stmt.Exec(t[i])
		if err != nil {
			return nil, err
		}

		rowsAffected, err := res.RowsAffected()
		if err != nil {
			fmt.Println(err)
		}

		StatusCode, err := res.LastInsertId()
		if err != nil {
			fmt.Println(err)
		}

		tags = append(tags, media.Tag{Text: t[i], ID: int(StatusCode)})

		slog.Info(fmt.Sprintf("sqlite: inserted tag '%s' with %d row(s) affected and status code == %d", t[i], rowsAffected, StatusCode))
	}

	return tags, nil
}

// searchTagID searches for a TagID with a given string. Returns -1 if tag does not exist.
func (s *SQLite3) searchTagID(tag string) (int, error) {
	query := `SELECT tagID FROM tags WHERE text == (?);`
	stmt, err := s.db.Query(query, tag)
	if err != nil {
		return -1, err
	}
	defer stmt.Close()

	var res int = -1
	for stmt.Next() {
		stmt.Scan(&res)
	}

	return res, nil
}

// SearchTag searches for an entry in archive if tag is mapped to said entry. Returns ArchiveID, PathRelative, MD5, and MD5Hash
func (s *SQLite3) SearchTag(tag string) ([]media.Entry, error) {
	tagID, err := s.searchTagID(tag)
	if err != nil || tagID == -1 {
		return nil, err
	}

	archiveIDs, err := s.searchTagMaps(tagID)
	if err != nil {
		return nil, err
	}

	stmt, err := s.db.Prepare("SELECT id, path, hashes.MD5 FROM archive INNER JOIN hashes ON archive.id = hashes.archiveID;")
	if err != nil {
		return []media.Entry{}, err
	}
	defer stmt.Close()

	entries := make([]media.Entry, len(archiveIDs))

	cnt := 0
	for i := 0; i < len(archiveIDs); i++ {
		rows, err := stmt.Query(archiveIDs[i])
		if err != nil {
			slog.Warn(err.Error())
		}

		for rows.Next() {
			rows.Scan(&entries[cnt].ArchiveID, &entries[cnt].Metadata.PathRelative, &entries[cnt].Metadata.Hash.MD5)
			entries[cnt].Metadata.MD5Hash = file.ByteToString(entries[cnt].Metadata.Hash.MD5)
			cnt++
		}
	}

	return entries, nil
}

// given tagID, return archive_id
func (s *SQLite3) searchTagMaps(t int) ([]ArchiveID, error) {
	stmt, err := s.db.Prepare("SELECT archiveID FROM tagmap WHERE tagID == ?")
	if err != nil {
		return []ArchiveID{}, err
	}
	defer stmt.Close()

	rows, err := stmt.Query(t)
	if err != nil {
		return []ArchiveID{}, err
	}

	var res []ArchiveID
	var tmp ArchiveID
	for rows.Next() {
		rows.Scan(&tmp)
		res = append(res, tmp)
	}

	return res, nil
}

// returns a slice with TagMapIDs the same length as tags. A non-positive integer means a tag did not map.
func (s *SQLite3) MapTags(a ArchiveID, tags []string) ([]int, error) {
	var stmt *sql.Stmt
	var err error

	query := `INSERT OR IGNORE INTO tagmap (archiveID, tagID)
	VALUES(?, (SELECT tagID FROM tags WHERE text = ?));`

	if s.HasTransaction {
		stmt, err = s.tx.Prepare(query)
	} else {
		stmt, err = s.db.Prepare(query)
	}

	if err != nil {
		return nil, err
	}
	defer stmt.Close()

	mappedTags := make([]int, len(tags))
	var tmp int64

	for i := 0; i < len(tags); i++ {
		res, err := stmt.Exec(a, tags[i])
		if err != nil {
			return nil, err
		}

		tmp, _ = res.RowsAffected()
		mappedTags[i] = int(tmp)
	}

	return mappedTags, nil
}

func (s *SQLite3) MapTagsWithID(a ArchiveID, tags []media.Tag) error {
	var stmt *sql.Stmt
	var err error

	query := `INSERT INTO tagmap (archiveID, tagID) VALUES (?, ?);`

	if s.HasTransaction {
		stmt, err = s.tx.Prepare(query)
	} else {
		stmt, err = s.db.Prepare(query)
	}

	if err != nil {
		return err
	}
	defer stmt.Close()

	for i := 0; i < len(tags); i++ {
		res, err := stmt.Exec(a, tags[i].ID)
		if err != nil {
			return err
		}

		lastInsert, _ := res.RowsAffected()
		slog.Info(fmt.Sprintf("sqlite: mapped '%v' with result %v", tags[i].Text, int(lastInsert)))
	}

	return nil
}

// InsertEntry takes in a MD5 string, storage path, and extension to insert into our database.
// Returns the archiveID on success, otherwise it'll return -1 and an error.
func (s *SQLite3) InsertEntry(h media.Hashes, path, extension string) (int, error) {
	var res sql.Result
	var err error

	query := `INSERT INTO archive (path, extension) VALUES (?, ?);`

	if s.HasTransaction {
		res, err = s.tx.Exec(query, path, extension)
	} else {
		res, err = s.db.Exec(query, path, extension)
	}
	if err != nil {
		return -1, err
	}

	rows, err := res.RowsAffected()
	if err != nil {
		return -1, err
	}

	lastInsert, err := res.LastInsertId()
	if err != nil {
		return -1, err
	}

	if err := s.insertHashes(ArchiveID(lastInsert), h); err != nil {
		return -1, err
	}

	slog.Info(fmt.Sprintf("inserted %d entries to archive with archiveID = %d", rows, lastInsert))

	return int(lastInsert), nil
}

func (s *SQLite3) insertHashes(a ArchiveID, h media.Hashes) error {
	var err error

	query := `INSERT INTO hashes (archiveID, md5, sha1, sha256) VALUES (?, ?, ?, ?)`

	if s.HasTransaction {
		_, err = s.tx.Exec(query, a, h.MD5, h.SHA1, h.SHA256)
	} else {
		_, err = s.db.Exec(query, a, h.MD5, h.SHA1, h.SHA256)
	}
	if err != nil {
		return err
	}

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
