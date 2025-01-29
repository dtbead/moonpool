package api

import (
	"context"
	"fmt"
	"io"
	"os"
	"reflect"
	"testing"
	"time"

	"github.com/dtbead/moonpool/entry"
	"github.com/dtbead/moonpool/importer"
	"github.com/dtbead/moonpool/internal/db"
	"github.com/dtbead/moonpool/internal/log"
	"github.com/go-test/deep"
)

// newMockAPI returns a disposable Moonpool API used for testing purposes.
func newMockAPI(c Config, t *testing.T) (*API, error) {
	logger := log.New(log.LogLevelVerbose)
	api, err := New(c, logger)
	if err != nil {
		return nil, err
	}

	if t != nil {
		t.Cleanup(func() { api.Close(context.Background()) })
	}

	return api, nil
}

func BenchmarkImport(b *testing.B) {
	a, _ := newMockAPI(Config{ArchiveLocation: ":memory:", ThumbnailLocation: ":memory:"}, nil)
	if _, err := GenerateMockData(a, b.N, true, true); err != nil {
		b.Errorf("BenchmarkImport() error = %v", err)
	}
}

func TestAPI_Import(t *testing.T) {
	mockAPI, err := newMockAPI(Config{ArchiveLocation: ":memory:", ThumbnailLocation: ":memory:"}, t)
	if err != nil {
		t.Fatal(err)
	}

	mockEntry := newMockEntry()
	err = mockEntry.LoadFile("testdata/6ba11adbdb35ee10f9353608a7b97ef248733a72.jpg")
	if err != nil {
		t.Fatal(err)
	}
	defer mockEntry.CloseFile()

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
		{"generic", mockAPI, args{context.Background(), mockEntry, nil}, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := tt.a.Import(tt.args.ctx, tt.args.i)
			if (err != nil) != tt.wantErr {
				t.Errorf("API.Import() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got < 1 {
				t.Errorf("API.Import() = %v, want archive_id >= 1", got)
			}

			hashes, err := tt.a.archive.GetHashes(tt.args.ctx, got)
			if err != nil {
				t.Errorf("API.Import() error on getting hash. %v", err)
			}

			entry, err := tt.a.archive.GetEntry(tt.args.ctx, got)
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
	mockAPI, err := newMockAPI(Config{ArchiveLocation: ":memory:", ThumbnailLocation: ":memory:"}, t)
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
			var mockEntry mockEntry
			for i := 0; i < tt.args.amount; i++ {
				mockEntry = newMockEntry()

				got, err := tt.a.Import(tt.args.ctx, mockEntry)
				if (err != nil) != tt.wantErr {
					t.Errorf("API.Import() error = %v, wantErr %v", err, tt.wantErr)
					return
				}
				if got < 1 {
					t.Errorf("API.Import() = %v, want archive_id >= 1", got)
				}

				hashes, err := tt.a.archive.GetHashes(tt.args.ctx, got)
				if err != nil {
					t.Errorf("API.Import() error on getting hash. %v", err)
				}

				entry, err := tt.a.archive.GetEntry(tt.args.ctx, got)
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
	mockAPI, err := newMockAPI(Config{ArchiveLocation: ":memory:", ThumbnailLocation: ":memory:"}, t)
	if err != nil {
		t.Fatalf("failed to create mock api, %v", err)
	}
	archive_id, err := mockAPI.Import(context.Background(), newMockEntry())
	if err != nil {
		t.Fatalf("failed to import mock entry. %v", err)
	}

	hash := entry.Hashes{
		MD5:    hexToByte("2ec268313d4d0bbc765144b6334df68b"),
		SHA1:   hexToByte("6ba11adbdb35ee10f9353608a7b97ef248733a72"),
		SHA256: hexToByte("7aaa7471fed00d0bcb416f123d364ec28a9080708601bd308cc4301d3fadb0e1"),
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
		want    entry.Hashes
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
				t.Errorf("API.GetHashes() got\nMD5 = %s\nSHA1 = %s\nSHA256 = %s\n", byteToHex(got.MD5), byteToHex(got.SHA1), byteToHex(got.SHA256))
				t.Errorf("want\nMD5 = %s\nSHA1 = %s\nSHA256 = %s\n", byteToHex(tt.want.MD5), byteToHex(tt.want.SHA1), byteToHex(tt.want.SHA256))
			}
		})
	}
}

func TestAPI_SetHashes(t *testing.T) {
	mockAPI, err := newMockAPI(Config{ArchiveLocation: ":memory:", ThumbnailLocation: ":memory:"}, t)
	if err != nil {
		t.Fatalf("failed to create mock api, %v", err)
	}
	archive_id, err := mockAPI.Import(context.Background(), newMockEntry())
	if err != nil {
		t.Fatalf("failed to import mock entry. %v", err)
	}

	type args struct {
		ctx        context.Context
		archive_id int64
		h          entry.Hashes
	}
	tests := []struct {
		name    string
		a       API
		args    args
		wantErr bool
	}{
		{"generic", *mockAPI, args{context.Background(), archive_id, randomHashes()}, false},
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
	mockAPI, err := newMockAPI(Config{ArchiveLocation: ":memory:", ThumbnailLocation: ":memory:"}, t)
	if err != nil {
		t.Fatalf("failed to create mock api, %v", err)
	}
	archive_id, _ := GenerateMockData(mockAPI, 1, false, false)

	type args struct {
		ctx        context.Context
		archive_id int64
	}
	tests := []struct {
		name          string
		a             *API
		args          args
		wantTimestamp entry.Timestamp
		wantErr       bool
	}{
		{"generic", mockAPI, args{context.Background(), archive_id[0]}, randomTimestamp(), false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err = mockAPI.archive.SetTimestamps(context.Background(), tt.args.archive_id, db.Timestamp(tt.wantTimestamp))
			if err != nil {
				t.Errorf("failed to set mock timestamp, %v", err)
			}

			got, err := tt.a.GetTimestamps(tt.args.ctx, tt.args.archive_id)
			if (err != nil) != tt.wantErr {
				t.Errorf("API.GetTimestamps() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			wantUTC := entry.Timestamp{
				DateCreated:  tt.wantTimestamp.DateCreated.UTC(),
				DateModified: tt.wantTimestamp.DateModified.UTC(),
				DateImported: tt.wantTimestamp.DateImported.UTC(),
			}

			if diff := deep.Equal(got, wantUTC); diff != nil {
				t.Errorf("API.GetTimestamps() got difference: %v", diff)
			}
		})
	}
}

func TestAPI_SetTimestamps(t *testing.T) {
	mockAPI, err := newMockAPI(Config{ArchiveLocation: ":memory:", ThumbnailLocation: ":memory:"}, t)
	if err != nil {
		t.Fatalf("failed to create mock API. %v", err)
	}
	archive_id, err := mockAPI.Import(context.Background(), newMockEntry())
	if err != nil {
		t.Fatalf("failed to import mock entry. %v", err)
	}

	ts1 := randomTimestamp()

	type args struct {
		ctx        context.Context
		archive_id int64
		t          entry.Timestamp
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

			want := entry.Timestamp{
				DateCreated:  tt.args.t.DateCreated.Round(time.Second * 1).UTC(),
				DateModified: tt.args.t.DateModified.Round(time.Second * 1).UTC(),
				DateImported: tt.args.t.DateImported.Round(time.Second * 1).UTC(),
			}

			if !reflect.DeepEqual(got, want) {
				const msg = `API.GetTimestamps()
				got
				DateCreated = %s
				DateModified = %s
				DateImported = %s
				
				want
				DateCreated = %s
				DateModified = %s
				DateImported = %s
				`
				t.Errorf(msg, got.DateCreated, got.DateModified, got.DateImported, want.DateCreated, want.DateModified, want.DateImported)
			}
		})
	}
}

func TestAPI_SetTimestampsPartial(t *testing.T) {
	mockAPI, err := newMockAPI(Config{ArchiveLocation: ":memory:", ThumbnailLocation: ":memory:"}, t)
	if err != nil {
		t.Fatalf("failed to create mock API. %v", err)
	}

	archive_ids, err := GenerateMockData(mockAPI, 2, false, true)
	if err != nil {
		t.Fatalf("failed to generate mock data. %v", err)
	}

	rndTimestampExclude := func(field string) entry.Timestamp {
		t := randomTimestamp()
		switch field {
		default:
			return entry.Timestamp{}
		case "imported":
			t.DateImported = time.Time{}
		case "modified":
			t.DateModified = time.Time{}
		case "created":
			t.DateCreated = time.Time{}
		}

		return t
	}
	type args struct {
		ctx        context.Context
		archive_id int64
		t          entry.Timestamp
	}
	tests := []struct {
		name    string
		a       *API
		args    args
		wantErr bool
	}{
		{"missing imported", mockAPI, args{context.Background(), archive_ids[0], rndTimestampExclude("imported")}, false},
		{"missing created", mockAPI, args{context.Background(), archive_ids[0], rndTimestampExclude("created")}, false},
		{"missing modified", mockAPI, args{context.Background(), archive_ids[0], rndTimestampExclude("modified")}, false},
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

			switch {
			case got.DateCreated == time.Time{}:
				t.Errorf("got emptry DateCreated, expected %v", tt.args.t.DateCreated)
			case got.DateImported == time.Time{}:
				t.Errorf("got emptry DateImported, expected %v", tt.args.t.DateImported)
			case got.DateModified == time.Time{}:
				t.Errorf("got emptry DateModified, expected %v", tt.args.t.DateModified)
			}
		})
	}
}

func TestAPI_GetFile(t *testing.T) {
	mockAPI, err := newMockAPI(Config{ArchiveLocation: ":memory:", ThumbnailLocation: ":memory:", MediaLocation: t.TempDir()}, t)
	if err != nil {
		t.Fatalf("failed to create mock API. %v", err)
	}

	f, err := os.Open("testdata/6ba11adbdb35ee10f9353608a7b97ef248733a72.jpg")
	if err != nil {
		t.Fatalf("failed to open test file. %v\n", err)
	}
	defer f.Close()

	type args struct {
		ctx  context.Context
		file *os.File
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
			entry, err := importer.New(tt.args.file, ".png")
			if err != nil {
				t.Fatalf("importer.New() failed to create new entry. %v", err)
			}

			archive_id, err := tt.a.Import(tt.args.ctx, entry)
			if err != nil {
				t.Fatalf("API.Import() failed to import entry. %v", err)
			}

			path, err := tt.a.GetPath(tt.args.ctx, archive_id)
			if err != nil {
				t.Fatalf("API.GetPath() failed to fetch filepath. %v", err)
			}

			t.Logf("imported media to %s/%s\n", tt.a.Config.MediaLocation, path.FileRelative)

			got, err := tt.a.GetFile(tt.args.ctx, archive_id)
			if (err != nil) != tt.wantErr {
				t.Fatalf("API.GetFile() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			defer got.Close()

			written, err := io.Copy(io.Discard, got)
			if err != nil {
				t.Fatalf("io.Copy error = %v, wantErr %v", err, tt.wantErr)
			}

			want, err := tt.args.file.Stat()
			if err != nil {
				t.Fatalf("failed to get file info, %v", err)
			}

			if want.Size() != written {
				t.Errorf("got %d bytes from file, copied %d", want.Size(), written)
			}
		})
	}
}

func TestAPI_NewSavepoint(t *testing.T) {
	mockAPI, err := newMockAPI(Config{ArchiveLocation: ":memory:", ThumbnailLocation: ":memory:"}, t)
	if err != nil {
		t.Fatalf("failed to create mock API. %v", err)
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
	mockAPI, err := newMockAPI(Config{ArchiveLocation: ":memory:", ThumbnailLocation: ":memory:"}, t)
	if err != nil {
		t.Fatalf("failed to create mock API. %v", err)
	}

	GenerateMockData(mockAPI, 1, true, true)

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

func randomHashes() entry.Hashes {
	return entry.Hashes{
		MD5:    randomBytes(16),
		SHA1:   randomBytes(20),
		SHA256: randomBytes(32),
	}
}

func TestAPI_GenerateFileMetadata(t *testing.T) {
	mockAPI, err := newMockAPI(Config{ArchiveLocation: ":memory:", ThumbnailLocation: ":memory:", MediaLocation: t.TempDir()}, t)
	if err != nil {
		t.Fatalf("failed to create mock API. %v", err)
	}

	f, err := os.Open("testdata/6ba11adbdb35ee10f9353608a7b97ef248733a72.jpg")
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
		want    entry.FileMetadata
		wantErr bool
	}{
		{"generic", mockAPI, args{context.Background(), f}, entry.FileMetadata{
			MediaOrientation: "landscape",
			MediaHeight:      1993,
			MediaWidth:       3000,
			FileMimetype:     "image/jpeg",
			FileSize:         258934,
		}, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			entry, err := importer.New(tt.args.file, ".jpg")
			if err != nil {
				t.Fatalf("importer.New() failed to create new entry. %v", err)
			}

			archive_id, err := mockAPI.Import(context.Background(), entry)
			if err != nil {
				t.Fatalf("failed to create mock data. %v", err)
			}

			if err := tt.a.GenerateFileMetadata(tt.args.ctx, archive_id); (err != nil) != tt.wantErr {
				t.Errorf("API.GenerateFileMetadata() error = %v, wantErr %v", err, tt.wantErr)
			}

			got, err := tt.a.GetMetadata(tt.args.ctx, archive_id)
			if err != nil {
				t.Errorf("API.GenerateFileMetadata()/API.GetMetadata error = %v, wantErr %v", err, tt.wantErr)
			}

			res := deep.Equal(got, tt.want)
			for _, v := range res {
				t.Errorf("got %v", v)
			}
		})
	}
}
