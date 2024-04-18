package media

import (
	"os"
	"time"
)

type Entry struct {
	ArchiveID int
	File      os.File
	Metadata  Metadata
	Tags      []string
}

type Metadata struct {
	MD5Hash      string
	Hash         Hashes
	Timestamp    Timestamp
	PathDirect   string
	PathRelative string
	Extension    string
}

type Hashes struct {
	MD5    []byte
	SHA1   []byte
	SHA256 []byte
}

type Timestamp struct {
	DateModifiedUTC time.Time
	DateImportedUTC time.Time
}

type Tag struct {
	ID   int
	Text string
}
