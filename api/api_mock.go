package api

import (
	"context"
	"io"
	rand "math/rand/v2"
	"time"

	archive "github.com/dtbead/moonpool/db"
	"github.com/dtbead/moonpool/file"
)

type MockEntry struct {
	Entry archive.Entry
}

// NewMockEntry creates a new entry in archive which populates the following fields with valid
// but random data: Metadata.Timestamp{*}, Metadata.Hash{*}, Metadata.Extension = ".png"
func NewMockEntry() MockEntry {
	return MockEntry{
		Entry: archive.Entry{
			Metadata: archive.Metadata{
				Timestamp: archive.Timestamp{
					DateCreated:  timeToUnixEpoch(time.Now().Add(-time.Hour * time.Duration(rand.IntN(300)))),
					DateModified: timeToUnixEpoch(time.Now().Add(-time.Hour * time.Duration(rand.IntN(300)))),
					DateImported: timeToUnixEpoch(time.Now().Add(-time.Hour * time.Duration(rand.IntN(300))))},
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

// empty method
func (m MockEntry) File() io.Reader {
	return nil // empty method
}

// empty method
func (m MockEntry) Store(baseDirectory string) error {
	return nil // empty method
}

// empty method
func (m MockEntry) DeleteTemp() error {
	return nil
}

// GenerateMockData creates an x amount of new entries with a random tag and .png extension as its
// metadata
func GenerateMockData(a *API, amount int, addTags bool) ([]int64, error) {
	var ArchiveIDs = make([]int64, amount)

	switch addTags {
	case true:
		for i := 0; i < amount; i++ {
			e := NewMockEntry()

			archiveID, err := a.Import(context.Background(), e, []string{randomString(6)})
			if err != nil {
				return nil, err
			}

			ArchiveIDs[i] = archiveID
		}
	case false:
		for i := 0; i < amount; i++ {
			e := NewMockEntry()

			archiveID, err := a.Import(context.Background(), e, nil)
			if err != nil {
				return nil, err
			}

			ArchiveIDs[i] = archiveID
		}
	}
	return ArchiveIDs, nil
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
