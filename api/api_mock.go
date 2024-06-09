package api

import (
	"io"
	"math/rand/v2"
	"time"

	"github.com/dtbead/moonpool/archive"
	"github.com/dtbead/moonpool/file"
)

type MockEntry struct {
	Entry archive.Entry
}

func NewMockEntry() MockEntry {
	return MockEntry{
		Entry: archive.Entry{
			Metadata: archive.Metadata{
				Timestamp: archive.Timestamp{DateModified: cleanTimestamp(time.Now().Add(-time.Hour * 300))}, // TODO: randomize DateModified time
				Hash: archive.Hashes{
					MD5:    randomBytes(16),
					SHA1:   randomBytes(20),
					SHA256: randomBytes(32),
				},
				Extension: ".png",
			},
		},
	}
}

func (m MockEntry) Hash() archive.Hashes {
	return m.Entry.Metadata.Hash
}

func (m MockEntry) Path() string {
	if m.Entry.Metadata.Hash.MD5 != nil {
		m.Hash()
	}
	m.Entry.Metadata.PathRelative = file.BuildPath(m.Entry.Metadata.Hash.MD5, m.Entry.Extension())
	return m.Entry.Metadata.PathRelative
}

func (m MockEntry) Extension() string {
	return m.Entry.Metadata.Extension
}

func (m MockEntry) Timestamp() archive.Timestamp {
	return m.Entry.Metadata.Timestamp
}
func (m MockEntry) File() io.Reader {
	return nil // empty method
}

func (m MockEntry) Store(baseDirectory string) error {
	return nil // empty method
}

func (m MockEntry) DeleteTemp() error {
	return nil // empty method
}

func randomBytes(length int) []byte {
	data := make([]byte, length)
	for i := 0; i < length; i++ {
		data[i] = byte(rand.Int())
	}

	return data
}

func trimString(s string, maxLen int) string {
	runes := []rune(s)
	if len(runes) <= maxLen {
		return s
	}
	return string(runes[0:maxLen])
}
