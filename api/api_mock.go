package api

import (
	"context"
	"io"
	rand "math/rand/v2"
	"time"

	"github.com/dtbead/moonpool/entry"
	"github.com/dtbead/moonpool/internal/file"
)

type mockEntry struct {
	PathRelative, PathExtension string
	Hashes                      entry.Hashes
	Timestamps                  entry.Timestamp
}

// newMockEntry() creates a new entry in archive which populates the following fields with valid
// but random data:
/*
	Metadata.Timestamp{}
	Metadata.Hash{}
	Metadata.Extension = ".png"
*/
func newMockEntry() mockEntry {
	h := entry.Hashes{
		MD5:    randomBytes(16),
		SHA1:   randomBytes(20),
		SHA256: randomBytes(32),
	}

	return mockEntry{
		PathRelative:  file.BuildPath(h.MD5[:], ".png"),
		PathExtension: ".png",
		Hashes:        h,
	}
}

func (m mockEntry) Hash() entry.Hashes {
	return m.Hashes
}

func (m mockEntry) Path() string {
	return m.PathRelative
}

func (m mockEntry) Extension() string {
	return m.PathExtension
}

func (m mockEntry) Timestamp() entry.Timestamp {
	return m.Timestamps
}

// empty method
func (m mockEntry) File() io.Reader {
	return nil
}

// empty method
func (m mockEntry) Store(string) error {
	return nil
}

// empty method
func (m mockEntry) DeleteTemp() error {
	return nil
}

// GenerateMockData() creates an x amount of new entries with a random tag and .png extension as its metadata.
// GenerateMockData() may return partial ArchiveIDs if only some imports are successful.
//
// You should ALWAYS check if "len(ArchiveID) <= 0 && err != nil"
func GenerateMockData(a *API, amount int, mockTags bool) ([]int64, error) {
	var ArchiveIDs = make([]int64, 0, amount)
	var mock mockEntry
	switch mockTags {
	case true:
		for i := 0; i < amount; i++ {
			mock = newMockEntry()
			archive_id, err := a.Import(context.Background(), mock, []string{randomString(6)})
			if err != nil {
				return ArchiveIDs, err
			}
			ArchiveIDs = append(ArchiveIDs, archive_id)
		}
	case false:
		for i := 0; i < amount; i++ {
			mock = newMockEntry()
			archive_id, err := a.Import(context.Background(), mock, nil)
			if err != nil {
				return ArchiveIDs, err
			}
			ArchiveIDs = append(ArchiveIDs, archive_id)
		}
	}
	return ArchiveIDs, nil
}

// randomTimestamp() generates a random Timestamp between the start of 2020 and the beginning of 2024
func randomTimestamp() Timestamp {
	min := time.Date(2020, 1, 0, 0, 0, 0, 0, time.UTC).Unix()
	max := time.Date(2024, 1, 0, 0, 0, 0, 0, time.UTC).Unix()
	delta := max - min

	randomTime := time.Unix(rand.Int64N(delta)+min, 0)

	return Timestamp{
		DateCreated:  randomTime,
		DateModified: randomTime.Add(-200 * time.Hour),
		DateImported: time.Now().Add(-100 * time.Hour),
	}
}

func randomBytes(length int) []byte {
	data := make([]byte, length)
	for i := 0; i < length; i++ {
		data[i] = byte(rand.Int())
	}

	return data
}
