// A helper interface to handle the importing of new entries from filesystem to database
package importer

import (
	"fmt"
	"io"
	"os"

	"github.com/dtbead/moonpool/entry"
	"github.com/dtbead/moonpool/internal/file"
)

type Importer struct {
	file io.Reader
	e    entry.Entry
}

func (i Importer) Path() string {
	return i.e.Metadata.Paths.FileRelative
}

func (i Importer) Extension() string {
	return i.e.Metadata.Paths.FileExtension
}

// Store copies a file into baseDirectory with its filename as its MD5 hash + its extension
//
// for example: if baseDirectory is "media", then "media/78/78f7f3b074f759b5dbc2ba0224457b15.png"
func (i Importer) Store(baseDirectory string) error {
	return file.Copy(fmt.Sprintf("%s/%s", baseDirectory, i.e.Metadata.Paths.FileRelative), i.file)
}

func (i Importer) Timestamp() entry.Timestamp {
	return i.e.Metadata.Timestamp
}

func (i Importer) Hash() entry.Hashes {
	return i.e.Metadata.Hash
}

func New(r io.Reader, extension string) (Importer, error) {
	f, isFile := r.(*os.File)
	if isFile {
		defer f.Seek(0, io.SeekStart)
	}

	hashes, err := file.GetHash(r)
	if err != nil {
		return Importer{}, err
	}

	i := Importer{
		e: entry.Entry{
			Metadata: entry.Metadata{
				Hash: entry.Hashes(hashes),
				Paths: entry.Path{
					FileRelative:  file.BuildPath(hashes.MD5, extension),
					FileExtension: extension,
				},
			},
		},
	}

	if isFile {
		i.file = f

		dateModified, err := file.DateModified(f)
		if err != nil {
			return Importer{}, err
		}

		dateCreated, err := file.DateCreated(f)
		if err != nil {
			return Importer{}, err
		}

		i.e.Metadata.Timestamp = entry.Timestamp{
			DateCreated:  dateCreated,
			DateModified: dateModified,
		}
	}

	return i, nil
}
