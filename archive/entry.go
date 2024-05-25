package archive

import (
	"io"
	"time"

	"github.com/dtbead/moonpool/file"
)

type Entry struct {
	file     io.Reader
	Metadata Metadata
	Tags     []string `json:"tags"`
}

type Metadata struct {
	Hash         Hashes
	Timestamp    Timestamp
	PathRelative string
	Extension    string `json:"file_extension"`
}

type Hashes struct {
	MD5    []byte `json:"md5"`
	SHA1   []byte `json:"sha1"`
	SHA256 []byte `json:"sha256"`
}

type Timestamp struct {
	DateModified time.Time `json:"timestamp_modified"`
	DateImported time.Time `json:"timestamp_imported"`
}

type Importer interface {
	Store() error
	Path() string
	Extension() string
	Hash() Hashes // TODO: should this return an error?
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

func (e Entry) Store() error {
	return file.Copy(e.Metadata.PathRelative, e.file)
}

// New takes an io.Reader, and file extension and returns a new entry. New will automatically hash the io.Reader
// and build a PathRelative to be used for importing. Entry.Timestamp and Entry.Tags is not modified
// and is up to the caller to populate
func New(r io.Reader, extension string) (Entry, error) {
	var e Entry

	h, err := file.GetHash(r)
	if err != nil {
		return Entry{}, err
	}
	e.Metadata.Hash.MD5 = h.MD5
	e.Metadata.Hash.SHA1 = h.SHA1
	e.Metadata.Hash.SHA256 = h.SHA256

	e.Metadata.PathRelative = file.BuildPath(h.MD5, extension)

	return e, nil
}
