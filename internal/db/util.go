package db

import (
	"database/sql"
	_ "embed"
	"regexp"

	_ "modernc.org/sqlite"
)

//go:embed archive_schema.sql
var archive_SQL_Schema string

//go:embed thumbnail_schema.sql
var thumbnail_SQL_Schema string

const SQL_INIT_PRAGMA = `
	PRAGMA foreign_keys = ON;
	PRAGMA journal_mode = WAL;
	PRAGMA synchronous = normal;
`

var (
	regex_trailingWhiteSpace = regexp.MustCompile(`^[ \t]+|[ \t]+$`)
	regex_excessSpaces       = regexp.MustCompile(`[ ]{2,}`)
	regex_newlinesAndTabs    = regexp.MustCompile(`[\n\t\r]+`)
)

func OpenSQLite3(filepath string) (*sql.DB, error) {
	s, err := sql.Open("sqlite", filepath+"?&mode=rwc")
	if err != nil {
		return nil, err
	}

	if _, err := s.Exec(SQL_INIT_PRAGMA); err != nil {
		s.Close()
		return nil, err
	}
	return s, nil
}

func OpenSQLite3Memory() (*sql.DB, error) {
	s, err := sql.Open("sqlite", ":memory:?_journal_mode=WAL")
	if err != nil {
		return nil, err
	}

	if _, err := s.Exec(SQL_INIT_PRAGMA); err != nil {
		s.Close()
		return nil, err
	}
	return s, nil
}

func InitializeArchive(db *sql.DB) error {
	if _, err := db.Exec(archive_SQL_Schema); err != nil {
		return err
	}

	return nil
}

func InitializeThumbnail(db *sql.DB) error {
	if _, err := db.Exec(thumbnail_SQL_Schema); err != nil {
		return err
	}

	return nil
}

// IsClean checks if a string is alphanumerical and is within [3-24] characters
func IsClean(s string) bool {
	clean, err := regexp.MatchString("^[a-zA-Z0-9]{3,24}$", s)
	if err != nil {
		return false
	}

	return clean
}

// DeleteWhitespace removes excess whitespaces from a given string.
// 1. Newlines and tabs are replaced with spaces.
// 2. Trailing whitespaces are removed
// 3. Consecutivity spaces are replaced with a single space.
func DeleteWhitespace(s string) string {
	s = regex_newlinesAndTabs.ReplaceAllLiteralString(s, " ")
	s = regex_trailingWhiteSpace.ReplaceAllLiteralString(s, "")
	s = regex_excessSpaces.ReplaceAllLiteralString(s, " ")

	if s == " " {
		return ""
	}
	return s
}
