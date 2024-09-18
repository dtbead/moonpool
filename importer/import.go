// A helper interface to handle the importing of new entries from filesystem to database
package importer

import (
	"errors"
	"io"
	"os"

	"github.com/dtbead/moonpool/entry"
	"github.com/dtbead/moonpool/internal/file"
)

// an Entry defines a struct that contains info relating to the importing/exporting of media in Moonpool.
type Entry struct {
	file     *os.File
	Metadata Metadata
	Tags     Tags
}

type Metadata struct {
	Hash      entry.Hashes
	Timestamp entry.Timestamp
	Path      entry.Path
}

type Tags struct {
	ArchiveID int64
	Tags      []Tag
}

type Tag struct {
	Text  string
	TagID int
}

func (e Entry) Path() string {
	return e.Metadata.Path.FileRelative
}

func (e Entry) Extension() string {
	return e.Metadata.Path.FilExtension
}

func (e Entry) Store(baseDirectory string) error {
	return file.Copy(baseDirectory, e.Metadata.Path.FileRelative, e.file)
}

func (e Entry) Timestamp() entry.Timestamp {
	return e.Metadata.Timestamp
}

func (e Entry) Hash() entry.Hashes {
	return e.Metadata.Hash
}

// DeleteTemp() closes and deletes the temporary file. There is no need to call Close on file
// when calling DeleteTemp
func (e *Entry) DeleteTemp() error {
	name := e.file.Name()
	if err := e.file.Close(); !errors.Is(err, os.ErrClosed) && err != nil {
		return err
	}

	return os.Remove(name)
}

func New(r io.Reader, extension string) (Entry, error) {
	f, err := os.CreateTemp(os.TempDir(), "moonpool_*")
	if err != nil {
		return Entry{}, err
	}

	if _, err := io.Copy(f, r); err != nil {
		return Entry{}, errors.New("failed to copy r to temporary file")
	}

	dateMod, err := file.DateModified(f)
	if err != nil {
		return Entry{}, err
	}

	dateCreated, err := file.DateCreated(f)
	if err != nil {
		return Entry{}, err
	}

	hashes, err := file.GetHash(r)
	if err != nil {
		return Entry{}, err
	}

	return Entry{
		file: f,
		Metadata: Metadata{
			Path: entry.Path{
				FileRelative: file.BuildPath(hashes.MD5, extension),
				FilExtension: extension,
			},
			Hash: entry.Hashes{
				MD5:    hashes.MD5,
				SHA1:   hashes.SHA1,
				SHA256: hashes.SHA256,
			},
			Timestamp: entry.Timestamp{
				DateModified: dateMod,
				DateCreated:  dateCreated,
			},
		},
	}, nil
}
