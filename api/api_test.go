package api

import (
	"context"
	"database/sql"
	"fmt"
	"io"
	"os"
	"reflect"
	"sort"
	"strings"
	"testing"
	"time"

	archive "github.com/dtbead/moonpool/db"
	"github.com/dtbead/moonpool/log"
)

// newMockAPI returns a disposable Moonpool API used for testing purposes. if useFile is set to true,
// newMockAPI will create a temporary database on filesystem and return a string pointing to its filepath.
func newMockAPI(c Config, t *testing.T, useFile bool) (*API, string, error) {
	var src, path string
	if useFile {
		path = t.TempDir() + "\\moonpool.sqlite3?cache=shared&mode=rwc&journal_mode=WAL"
		src = path
	} else {
		src = ":memory:?_journal_mode=WAL"
	}

	sql, err := sql.Open("sqlite", src)
	if err != nil {
		return nil, "", err
	}

	if err = archive.InitializeSQLite3(sql); err != nil {
		return nil, "", err
	}

	logger := log.NewSlogger(context.Background(), log.LogLevelVerbose, "api")

	return New(sql, logger, c), strings.ReplaceAll(path, "?cache=shared&mode=rwc&journal_mode=WAL", ""), nil
}

func BenchmarkImport(b *testing.B) {
	a, _, _ := newMockAPI(Config{}, nil, false)
	if _, err := GenerateMockData(a, b.N, true); err != nil {
		b.Errorf("BenchmarkImport() error = %v", err)
	}
}

func TestAPI_Import(t *testing.T) {
	mockAPI, _, err := newMockAPI(Config{}, nil, false)
	if err != nil {
		t.Fatal(err)
	}

	type args struct {
		ctx  context.Context
		i    Importer
		tags []string
	}
	tests := []struct {
		name    string
		a       *API
		args    args
		wantErr bool
	}{
		{"generic", mockAPI, args{context.Background(), NewMockEntry(), nil}, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := tt.a.Import(tt.args.ctx, tt.args.i, tt.args.tags)
			if (err != nil) != tt.wantErr {
				t.Errorf("API.Import() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got < 1 {
				t.Errorf("API.Import() = %v, want archive_id >= 1", got)
			}

			hashes, err := tt.a.service.GetHashes(tt.args.ctx, got)
			if err != nil {
				t.Errorf("API.Import() error on getting hash. %v", err)
			}

			entry, err := tt.a.service.GetEntry(tt.args.ctx, got)
			if err != nil {
				t.Errorf("API.Import() error on getting entry. %v", err)
			}

			validPath := fmt.Sprintf("%s/%s%s", byteToHex(hashes.Md5[:1]), byteToHex(hashes.Md5), tt.args.i.Extension())
			if validPath != entry.Path {
				t.Errorf("API.Import() path = %v, want %v", entry.Path, validPath)
			}
		})
	}
}

func TestAPI_Import_Multiple(t *testing.T) {
	mockAPI, _, err := newMockAPI(Config{}, nil, false)
	if err != nil {
		t.Fatal(err)
	}

	type args struct {
		ctx    context.Context
		amount int
		tags   []string
	}
	tests := []struct {
		name    string
		a       *API
		args    args
		wantErr bool
	}{
		{"generic", mockAPI, args{context.Background(), 100, nil}, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			for i := 0; i < tt.args.amount; i++ {
				mockEntry := NewMockEntry()
				got, err := tt.a.Import(tt.args.ctx, mockEntry, tt.args.tags)
				if (err != nil) != tt.wantErr {
					t.Errorf("API.Import() error = %v, wantErr %v", err, tt.wantErr)
					return
				}
				if got < 1 {
					t.Errorf("API.Import() = %v, want archive_id >= 1", got)
				}

				hashes, err := tt.a.service.GetHashes(tt.args.ctx, got)
				if err != nil {
					t.Errorf("API.Import() error on getting hash. %v", err)
				}

				entry, err := tt.a.service.GetEntry(tt.args.ctx, got)
				if err != nil {
					t.Errorf("API.Import() error on getting entry. %v", err)
				}

				validPath := fmt.Sprintf("%s/%s%s", byteToHex(hashes.Md5[:1]), byteToHex(hashes.Md5), mockEntry.Extension())
				if validPath != entry.Path {
					t.Errorf("API.Import() path = %v, want %v", entry.Path, validPath)
				}
			}

		})
	}
}

func TestAPI_GetHashes(t *testing.T) {
	mockAPI, _, _ := newMockAPI(Config{}, nil, false)
	archive_id, err := mockAPI.Import(context.Background(), NewMockEntry(), nil)
	if err != nil {
		t.Fatalf("failed to import mock entry. %v", err)
	}

	hash := archive.Hashes{
		MD5:    randomBytes(16),
		SHA1:   randomBytes(20),
		SHA256: randomBytes(32),
	}

	if err := mockAPI.SetHashes(context.Background(), archive_id, hash); err != nil {
		t.Fatalf("failed to set hash. %v", err)
	}

	type args struct {
		ctx        context.Context
		archive_id int64
	}
	tests := []struct {
		name    string
		a       *API
		args    args
		want    archive.Hashes
		wantErr bool
	}{
		{"generic", mockAPI, args{context.Background(), 1}, hash, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := tt.a.GetHashes(tt.args.ctx, tt.args.archive_id)
			if (err != nil) != tt.wantErr {
				t.Errorf("API.GetHashes() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("API.GetHashes() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestAPI_SetHashes(t *testing.T) {
	mockAPI, _, _ := newMockAPI(Config{}, nil, false)
	archive_id, err := mockAPI.Import(context.Background(), NewMockEntry(), nil)
	if err != nil {
		t.Fatalf("failed to import mock entry. %v", err)
	}

	type args struct {
		ctx        context.Context
		archive_id int64
		h          archive.Hashes
	}
	tests := []struct {
		name    string
		a       API
		args    args
		wantErr bool
	}{
		{"generic", *mockAPI, args{context.Background(), archive_id, archive.Hashes{
			MD5:    randomBytes(16),
			SHA1:   randomBytes(20),
			SHA256: randomBytes(32),
		}}, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := tt.a.SetHashes(tt.args.ctx, tt.args.archive_id, tt.args.h); (err != nil) != tt.wantErr {
				t.Errorf("API.SetHashes() error = %v, wantErr %v", err, tt.wantErr)
			}

			hashes, err := tt.a.GetHashes(tt.args.ctx, tt.args.archive_id)
			if err != nil {
				t.Errorf("API.GetHashes() error = %v, wantErr %v", err, tt.wantErr)
			}

			if !reflect.DeepEqual(hashes, tt.args.h) {
				t.Errorf("API.SetHashes() got %v, want %v", hashes, tt.args.h)
			}
		})
	}
}

func TestAPI_GetTimestamps(t *testing.T) {
	mockAPI, _, err := newMockAPI(Config{}, nil, false)
	if err != nil {
		t.Fatalf("failed to create mock API. %v", err)
	}
	archive_id, err := mockAPI.Import(context.Background(), NewMockEntry(), nil)
	if err != nil {
		t.Fatalf("failed to import mock entry. %v", err)
	}

	ts1 := newTimestamp()

	type args struct {
		ctx        context.Context
		archive_id int64
		Timestamp  archive.Timestamp
	}
	tests := []struct {
		name             string
		a                *API
		args             args
		wantUTCTimeStamp archive.Timestamp
		wantErr          bool
	}{
		{"non-UTC import", mockAPI, args{context.Background(), archive_id, ts1}, ts1.UTC(), false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := mockAPI.SetTimestamps(context.Background(), tt.args.archive_id, tt.args.Timestamp); err != nil {
				t.Fatalf("failed to set timestamp. %v\n", err)
				return
			}

			got, err := tt.a.GetTimestamps(tt.args.ctx, tt.args.archive_id)
			if (err != nil) != tt.wantErr {
				t.Errorf("API.GetTimestamps() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !reflect.DeepEqual(got, tt.wantUTCTimeStamp) {
				t.Errorf("API.GetTimestamps() = %v, want %v", got, tt.wantUTCTimeStamp)
			}

		})
	}
}

func TestAPI_SetTimestamps(t *testing.T) {
	mockAPI, _, _ := newMockAPI(Config{}, nil, false)
	archive_id, err := mockAPI.Import(context.Background(), NewMockEntry(), nil)
	if err != nil {
		t.Fatalf("failed to import mock entry. %v", err)
	}

	ts1 := newTimestamp()

	type args struct {
		ctx        context.Context
		archive_id int64
		t          archive.Timestamp
	}
	tests := []struct {
		name    string
		a       *API
		args    args
		wantErr bool
	}{
		{"golang compatible format", mockAPI, args{context.Background(), archive_id, ts1}, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := tt.a.SetTimestamps(tt.args.ctx, tt.args.archive_id, tt.args.t); (err != nil) != tt.wantErr {
				t.Errorf("API.SetTimestamps() error = %v, wantErr %v", err, tt.wantErr)
			}

			got, err := tt.a.GetTimestamps(tt.args.ctx, tt.args.archive_id)
			if (err != nil) != tt.wantErr {
				t.Errorf("API.GetTimestamps() error = %v, wantErr %v", err, tt.wantErr)
			}

			want := archive.Timestamp{
				DateCreated:  tt.args.t.DateCreated.Round(time.Second * 1).UTC(),
				DateModified: tt.args.t.DateModified.Round(time.Second * 1).UTC(),
				DateImported: tt.args.t.DateImported.Round(time.Second * 1).UTC(),
			}

			if !reflect.DeepEqual(got, want) {
				t.Errorf("API.SetTimestamps() got = %v, want %v", got, want)
			}
		})
	}
}

func Test_isValidHash(t *testing.T) {
	type args struct {
		b      []byte
		length int
	}
	tests := []struct {
		name string
		args args
		want bool
	}{
		{"valid md5", args{b: randomBytes(16), length: 16}, true},
		{"valid sha1", args{b: randomBytes(20), length: 20}, true},
		{"valid sha256", args{b: randomBytes(32), length: 32}, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := isValidHash(tt.args.b, tt.args.length); got != tt.want {
				t.Errorf("isValidHash() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestAPI_GetFile(t *testing.T) {
	mockAPI, _, _ := newMockAPI(Config{}, nil, false)

	f, err := os.Open("testdata/82d233bf13e0ebe6636db4d405d846c357d73c3cc491a97b85b9b235b4efdc80.png")
	if err != nil {
		t.Fatalf("failed to open test file. %v\n", err)
	}
	defer f.Close()

	type args struct {
		ctx  context.Context
		file io.Reader
	}
	tests := []struct {
		name    string
		a       *API
		args    args
		wantErr bool
	}{
		{"generic", mockAPI, args{ctx: context.Background(), file: f}, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			entry, err := archive.New(tt.args.file, ".png")
			if err != nil {
				t.Fatalf("API.GetFile() unable to create new entry. %v", err)
			}

			defer func() error {
				if err := entry.DeleteTemp(); err != nil {
					t.Fatalf("API.GetFile() unable to delete temporary file. %v", err)
					return err
				}
				return nil
			}()

			archive_id, err := tt.a.Import(tt.args.ctx, entry, nil)
			if err != nil {
				t.Fatalf("API.GetFile()/API.Import() unable to import entry. %v", err)
			}

			path, err := tt.a.GetPath(tt.args.ctx, archive_id)
			if err != nil {
				t.Fatalf("API.GetFile()/API.GetPath() failed to fetch filepath. %v", err)
			}

			t.Logf("imported media to %s/%s\n", tt.a.Conf.MediaLocation, path.Filepath)

			got, err := tt.a.GetFile(tt.args.ctx, archive_id)
			if (err != nil) != tt.wantErr {
				t.Fatalf("API.GetFile() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			defer got.Close()

			w, err := io.Copy(io.Discard, got)
			if err != nil {
				t.Fatalf("API.GetFile() error = %v, wantErr %v", err, tt.wantErr)
			}

			if stat, _ := f.Stat(); stat.Size() != w {
				t.Errorf("API.GetFile() expected %d copied bytes from file, got %d", stat, w)
			}
		})
	}
}

func TestAPI_SetTags(t *testing.T) {
	mockAPI, _, _ := newMockAPI(Config{}, nil, false)
	archive_ids, err := GenerateMockData(mockAPI, 2, false)
	if err != nil {
		t.Fatalf("failed to generate mock data. %v", err)
	}

	type args struct {
		ctx        context.Context
		archive_id int64
		tags       []string
	}
	tests := []struct {
		name        string
		a           *API
		args        args
		wantErr     bool
		wantSetOnly bool
	}{
		{"generic", mockAPI, args{context.Background(), archive_ids[0], []string{randomString(6)}}, false, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := tt.a.SetTags(tt.args.ctx, tt.args.archive_id, tt.args.tags); (err != nil) != tt.wantErr {
				t.Errorf("API.SetTags() error = %v, wantErr %v", err, tt.wantErr)
			}

			got, err := tt.a.GetTags(tt.args.ctx, tt.args.archive_id)
			if err != nil {
				t.Errorf("API.SetTags()/API.GetTags()e rror = %v, wantErr %v", err, tt.wantErr)
			}

			if !reflect.DeepEqual(got, tt.args.tags) {
				t.Errorf("API.SetTags() got %v, want %v", got, tt.args.tags)
			}

		})
	}
}

func TestAPI_RemoveTags(t *testing.T) {
	mockAPI, _, err := newMockAPI(Config{}, nil, false)
	if err != nil {
		t.Fatalf("failed to create mock API. %v\n", err)
	}

	for i := 0; i < 4; i++ {
		_, err = mockAPI.Import(context.Background(), NewMockEntry(), nil)
		if err != nil {
			t.Fatalf("failed to create new mock entry. %v\n", err)
		}
	}

	type args struct {
		ctx        context.Context
		tags       []string
		archive_id []int64
	}
	tests := []struct {
		name          string
		a             *API
		args          args
		wantErr       bool
		wantInArchive int64 // skips removing a tag mapped to an archive_id
	}{
		{"remove entire tag (tag is no longer mapped to any archive)", mockAPI, args{context.Background(), []string{"foo"}, []int64{1, 2}}, false, 0},
		{"remove tag for single archive (tag is still mapped to an archive)", mockAPI, args{context.Background(), []string{"bar"}, []int64{3, 4}}, false, 4},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			for _, v := range tt.args.archive_id {
				if err := tt.a.SetTags(context.Background(), v, tt.args.tags); err != nil {
					t.Errorf("API.RemoveTags()/API.SetTags() error = %v, wantErr %v", err, tt.wantErr)
				}
			}

			for _, v := range tt.args.archive_id {
				if tt.wantInArchive == v {
					break
				}
				if err := tt.a.RemoveTags(tt.args.ctx, v, tt.args.tags); (err != nil) != tt.wantErr {
					t.Errorf("API.RemoveTags() error = %v, wantErr %v", err, tt.wantErr)
				}
			}

			// iterate through the tags we previously inserted
			for _, tag := range tt.args.tags {
				searchTags, err := tt.a.SearchTag(tt.args.ctx, tag)
				if (err != nil) != tt.wantErr {
					t.Errorf("API.RemoveTags() error = %v, wantErr %v", err, tt.wantErr)
				}
				t.Logf("found tags %v", searchTags)

				// iterate through the archive_id's we assigned tags to
				for _, archive_id := range tt.args.archive_id {
					// iterate through the list of tags we've just searched for after removing our tags.
					// ideally, this loop would be skipped entirely, assuming a tag is no longer mapped to any
					// entry and has been deleted entirely already, but never say never...
					for _, searchTag := range searchTags {
						if searchTag.ArchiveID == archive_id {
							t.Logf("found matching archive_id %d in tag search", archive_id)
							// make a new slice containing only the text of tags. makes it easier to work with instead
							// of dealing with a struct of archive.EntryTags
							tagText := make([]string, len(searchTag.Tags))
							for i, v := range searchTag.Tags {
								tagText[i] = v.Text
							}

							if !inSlice(tagText, tt.args.tags) && searchTag.ArchiveID == tt.wantInArchive {
								t.Errorf("API.RemoveTags() found tag = %v for archive_id %d, want none", tag, archive_id)
							}
						}
					}

				}
			}
		})
	}
}

func TestAPI_NewSavepoint(t *testing.T) {
	mockAPI, _, err := newMockAPI(Config{}, nil, false)
	if err != nil {
		t.Fatal(err)
	}

	type args struct {
		ctx  context.Context
		name string
	}
	tests := []struct {
		name    string
		a       API
		args    args
		wantErr bool
	}{
		{"generic", *mockAPI, args{context.Background(), "meow"}, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := tt.a.NewSavepoint(tt.args.ctx, tt.args.name); (err != nil) != tt.wantErr {
				t.Errorf("API.NewSavepoint() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestAPI_DoesEntryExist(t *testing.T) {
	mockAPI, _, err := newMockAPI(Config{}, nil, false)
	if err != nil {
		t.Fatalf("failed to create mock API. %v", err)
	}
	GenerateMockData(mockAPI, 1, true)

	type args struct {
		ctx context.Context
		id  int64
	}
	tests := []struct {
		name string
		a    *API
		args args
		want bool
	}{
		{"exists", mockAPI, args{context.Background(), 1}, true},
		{"not exist", mockAPI, args{context.Background(), 2}, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.a.DoesEntryExist(tt.args.ctx, tt.args.id); got != tt.want {
				t.Errorf("API.DoesEntryExist() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestAPI_GetTagCount(t *testing.T) {
	mockAPI, _, err := newMockAPI(Config{}, nil, false)
	if err != nil {
		t.Fatalf("failed to create mock api. %v", err)
	}

	if _, err := GenerateMockData(mockAPI, 1, false); err != nil {
		t.Fatalf("failed to generate mock data. %v", err)
	}

	type args struct {
		ctx context.Context
	}
	tests := []struct {
		name    string
		a       *API
		args    args
		want    int64
		wantErr bool
	}{
		{"new tag", mockAPI, args{context.Background()}, 1, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var tagCount int64
			var err error
			tag := randomString(6)

			if err := mockAPI.SetTags(tt.args.ctx, tt.want, []string{tag}); err != nil {
				t.Fatalf("failed to set tag '%s', %v", tag, err)
			}

			tagCount, err = tt.a.GetTagCount(tt.args.ctx, tag)
			if (err != nil) != tt.wantErr {
				t.Errorf("API.GetTagCount() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if tagCount != tt.want {
				t.Errorf("API.GetTagCount() = %v, want %v", tagCount, tt.want)
			}
		})
	}
}

// inSlice compares two slices of any type against each other and returns
// true whether or not they're equivalent. inSlice assumes each slice is of the same
// length and is sorted.
func inSlice(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}

	sort.Strings(a)
	sort.Strings(b)

	for _, str1 := range a {
		for _, str2 := range b {
			if str1 != str2 {
				return false
			}
		}
	}

	return true
}

func ParseString(s string) (time.Time, error) {
	location, err := time.LoadLocation("America/Chicago")
	if err != nil {
		return time.Time{}, err
	}
	const layout = "2006-01-02 15:04:05 -0700"

	date, err := time.ParseInLocation(layout, s, location)
	if err != nil {
		return time.Time{}, err
	}

	return date, nil
}
