package migration

import (
	"errors"
	"fmt"
	"io"
	"strings"
)

type Migrator struct {
	Version, Timestamp int64
	Title              string
	sql                []string
	Upgrade            bool
}

// NewMigrator creates a new Migrator to be used for updating database version. filename expects a
// string in the format: "{unix_timestamp}_{db_version}_{title}_{is_upgrade}.sql" with {unix_timestamp} and
// {version} being an int64, {title} being a string and {up|down} a string to indicate whether
// the migration is to upgrade or downgrade to the database version.
func NewMigrator(r io.Reader, filename string) (Migrator, error) {
	m, err := parseFilename(filename)
	if err != nil {
		return Migrator{}, err
	}

	buf := new(strings.Builder)
	_, err = io.Copy(buf, r)
	if err != nil {
		return Migrator{}, err
	}

	m.sql = strings.Split(buf.String(), "\n")

	return m, nil
}

func parseFilename(filename string) (Migrator, error) {
	m := Migrator{}

	var upgrade string

	scanned, err := fmt.Sscanf(filename, "%d_%d_%s_%s.sql", &m.Timestamp, &m.Version, &m.Title, &upgrade)
	if err != nil {
		return Migrator{}, err
	}

	if scanned != 4 {
		return Migrator{}, errors.New("invalid filename")
	}

	if !strings.EqualFold(upgrade, "up") || !strings.EqualFold(upgrade, "down") {
		return Migrator{}, errors.New("invalid upgrade direction")
	}

	return m, nil
}

func (m Migrator) Queries() []string {
	return m.sql
}
