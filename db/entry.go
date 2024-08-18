package db

import (
	"errors"
	"io"
	"os"
	"strings"
	"time"

	"github.com/dtbead/moonpool/file"
)

type Entry struct {
	file     *os.File
	Metadata Metadata
	Tags     []string
}

type EntryTags struct {
	ArchiveID int64
	Tags      []Tag
}

type Metadata struct {
	Hash         Hashes
	Timestamp    Timestamp
	PathRelative string
	Extension    string
}

type Hashes struct {
	MD5    []byte
	SHA1   []byte
	SHA256 []byte
}

type Timestamp struct {
	DateCreated  time.Time
	DateModified time.Time
	DateImported time.Time
}

type Tag struct {
	Text  string
	TagID int
}

// CleanTag removes any excess whitespace, blacklisted characters, etc from a given string.
func CleanTag(s string) string {
	r := []rune(s)
	NewString := make([]rune, len(r))

	for _, v := range r {
		if v != ' ' {
			NewString = append(NewString, v)
		}
	}

	return strings.TrimSpace(string(NewString))
}

type Importer interface {
	Timestamp() Timestamp
	Store(baseDirectory string) error
	Path() string
	Extension() string
	Hash() Hashes
}

func (e Entry) Path() string {
	return e.Metadata.PathRelative
}

func (e Entry) Extension() string {
	return e.Metadata.Extension
}

func (e Entry) Hash() Hashes {
	return e.Metadata.Hash
}

func (e Entry) Store(baseDirectory string) error {
	return file.Copy(baseDirectory, e.Metadata.PathRelative, e.file)
}

func (e Entry) Timestamp() Timestamp {
	return e.Metadata.Timestamp
}

// DeleteTemp closes and deletes the temporary file. There is no need to call Close on file
// when calling DeleteTemp
func (e Entry) DeleteTemp() error {
	name := e.file.Name()
	if err := e.file.Close(); !errors.Is(err, os.ErrClosed) && err != nil {
		return err
	}

	return os.Remove(name)
}

// New takes an io.Reader and a file extension, and returns a new Entry.
func New(r io.Reader, extension string) (Entry, error) {
	var e Entry

	f, err := os.CreateTemp(os.TempDir(), "moonpool_*")
	if err != nil {
		return Entry{}, err
	}

	if _, err := io.Copy(f, r); err != nil {
		return Entry{}, errors.New("failed to copy r to temporary file")
	}

	f.Seek(0, io.SeekStart)
	h, err := file.GetHash(f)
	if err != nil {
		return Entry{}, err
	}

	e.Metadata.Hash.MD5 = h.MD5
	e.Metadata.Hash.SHA1 = h.SHA1
	e.Metadata.Hash.SHA256 = h.SHA256
	e.Metadata.PathRelative = file.BuildPath(h.MD5, extension)

	f.Seek(0, io.SeekStart)
	return Entry{
		file: f,
		Metadata: Metadata{
			PathRelative: e.Metadata.PathRelative,
			Extension:    extension,
			Hash: Hashes{
				MD5:    h.MD5,
				SHA1:   h.SHA1,
				SHA256: h.SHA256,
			},
		},
	}, nil
}
