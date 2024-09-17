package db

import (
	"database/sql"
	_ "embed"
	"errors"
	"os"
	"regexp"

	_ "modernc.org/sqlite"
)

//go:embed sqlite-schema.sql
var SQLSchema string

const SQL_INIT_PRAGMA = `
	PRAGMA foreign_keys = ON;
	PRAGMA journal_mode = WAL;
	PRAGMA synchronous = normal;
`

func OpenSQLite3(filepath string) (*sql.DB, error) {
	s, err := sql.Open("sqlite", filepath+"?cache=shared&mode=rwc&journal_mode=WAL")
	if err != nil {
		return nil, err
	}

	if _, err := s.Exec(SQL_INIT_PRAGMA); err != nil {
		s.Close()
		return nil, err
	}
	return s, nil
}

func InitializeSQLite3(db *sql.DB) error {
	if _, err := db.Exec(SQLSchema); err != nil {
		return err
	}

	return nil
}

func DoesFileExist(s string) bool {
	if _, err := os.Stat(s); err == nil {
		return true
	} else if errors.Is(err, os.ErrNotExist) {
		return false
	} else {
		return false
	}
}

// isClean checks if a string is alphanumerical and is within [3-24] characters
func isClean(s string) bool {
	clean, err := regexp.MatchString("^[a-zA-Z0-9]{3,24}$", s)
	if err != nil {
		return false
	}

	return clean
}
