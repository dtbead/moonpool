package api

import (
	"context"
	"io"
	rand "math/rand/v2"
	"os"
	"time"

	"github.com/dtbead/moonpool/entry"
	"github.com/dtbead/moonpool/internal/file"
)

type mockEntry struct {
	media                       io.Reader
	PathRelative, PathExtension string
	Hashes                      entry.Hashes
	Timestamps                  entry.Timestamp
}

// newMockEntry creates a new entry in archive which populates the following fields with valid
// but random data:
/*
	media
	Hashes
	Extension
	PathRelative
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

func (m mockEntry) FileSize() int {
	return rand.IntN(999999)
}

func (m mockEntry) FileData() io.Reader {
	resetFileSeek(m.media)
	return m.media
}

func (m *mockEntry) LoadFile(path string) error {
	f, err := os.Open(path)
	if err != nil {
		return err
	}

	m.media = f
	return nil
}

func (m *mockEntry) CloseFile() {
	f, ok := m.media.(*os.File)
	if ok {
		f.Close()
	}
}

// empty method
func (m mockEntry) Store(baseDirectory string) (path string, err error) {
	return "", nil
}

// empty method
func (m mockEntry) DeleteTemp() error {
	return nil
}

// GenerateMockData creates an x amount of new entries with a random tag and .png extension as its metadata.
// GenerateMockData may return an error, along with partial ArchiveIDs if only some imports are successful.
func GenerateMockData(a *API, amount int, mockTags, mockTimestamps bool) ([]int64, error) {
	var ArchiveIDs = make([]int64, 0, amount)
	ctx := context.Background()

	err := a.NewSavepoint(ctx, "mockgen")
	if err != nil {
		return nil, err
	}

	for i := 0; i < amount; i++ {
		archive_id, err := a.Import(ctx, newMockEntry())
		if err != nil {
			return ArchiveIDs, err
		}

		if mockTags {
			if err := a.AssignTags(ctx, archive_id, []string{randomString(6)}); err != nil {
				return ArchiveIDs, err
			}
		}

		if mockTimestamps {
			if err := a.SetTimestamps(ctx, archive_id, randomTimestamp()); err != nil {
				return ArchiveIDs, err
			}

		}

		ArchiveIDs = append(ArchiveIDs, archive_id)
	}

	err = a.ReleaseSavepoint(ctx, "mockgen")
	if err != nil {
		return nil, err
	}

	return ArchiveIDs, nil
}

// randomTimestamp generates a random Timestamp between the start of 2020 and the beginning of 2024,
// rounded by the nearest second. DateImported will ALWAYS be of time.Now()
func randomTimestamp() entry.Timestamp {
	min := time.Date(2020, 1, 0, 0, 0, 0, 0, time.UTC).Unix()
	max := time.Date(2024, 1, 0, 0, 0, 0, 0, time.UTC).Unix()
	delta := max - min

	randomTime := time.Unix(rand.Int64N(delta)+min, 0)

	return entry.Timestamp{
		DateCreated:  randomTime.Round(time.Second * 1),
		DateModified: randomTime.Add(-200 * time.Hour).Round(time.Second * 1),
		DateImported: time.Now().Round(time.Second * 1),
	}
}

func randomBytes(length int) []byte {
	data := make([]byte, length)
	for i := 0; i < length; i++ {
		data[i] = byte(rand.Int())
	}

	return data
}
