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

	scanned, err := fmt.Sscanf(filename, "%d_%d_%s_%b.sql", m.Timestamp, m.Version, m.Title, m.Upgrade)
	if err != nil {
		return Migrator{}, err
	}

	if scanned != 4 {
		return Migrator{}, errors.New("invalid filename")
	}

	return m, nil
}

func (m Migrator) Queries() []string {
	return m.sql
}
