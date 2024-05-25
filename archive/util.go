package archive

import (
	"database/sql"
	_ "embed"
	"errors"
	"os"

	_ "modernc.org/sqlite"
)

//go:embed sqlite-schema.sql
var SQLSchema string

func OpenSQLite3(filepath string) (*sql.DB, error) {
	s, err := sql.Open("sqlite", filepath)
	if err != nil {
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
