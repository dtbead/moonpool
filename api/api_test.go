package api

import (
	"context"
	"database/sql"
	"fmt"
	"io"
	"math/rand"
	"os"
	"reflect"
	"testing"
	"time"

	"github.com/dtbead/moonpool/archive"
	"github.com/dtbead/moonpool/file"
	"github.com/dtbead/moonpool/log"
	"github.com/go-test/deep"
)

var mockAPI *API

func TestMain(m *testing.M) {
	l := log.NewSlogLogger(context.Background())

	sql, err := sql.Open("sqlite", "mock.sqlite3?_journal_mode=WAL")
	if err != nil {
		l.Error(err.Error())
		os.Exit(1)
	}

	defer sql.Close()
	archive.InitializeSQLite3(sql)
	mockAPI = New(l, sql)

	code := m.Run()
	os.Exit(code)
}

func generateMockData(amount int) error {
	for i := 0; i < amount; i++ {
		e := NewMockEntry()
		e.Entry.Metadata.Extension = ".png"

		_, err := mockAPI.Import(context.Background(), e, []string{randomString(6)})
		if err != nil {
			return err
		}
	}

	return nil
}

func randomString(length int) string {
	b := make([]byte, length+2)
	rand.Read(b)
	return fmt.Sprintf("%x", b)[2 : length+2]
}

func BenchmarkImport(b *testing.B) {
	if err := generateMockData(b.N); err != nil {
		b.Errorf("BenchmarkImport() error = %v", err)
	}

}

func TestAPI_Import(t *testing.T) {
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

func TestAPI_GetHashes(t *testing.T) {
	hash1 := archive.Hashes{
		MD5:    randomBytes(16),
		SHA1:   randomBytes(20),
		SHA256: randomBytes(32),
	}

	if err := mockAPI.SetHashes(context.Background(), 1, hash1); err != nil {
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
		{"generic", mockAPI, args{context.Background(), 1}, hash1, false},
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
		{"generic", *mockAPI, args{context.Background(), 100, archive.Hashes{
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
	ts1 := archive.Timestamp{
		DateModified: time.Now().Add(time.Hour * -300),
		DateImported: time.Now(),
	}

	type args struct {
		ctx            context.Context
		archive_id     int64
		localTimeStamp archive.Timestamp
	}
	tests := []struct {
		name             string
		a                *API
		args             args
		wantUTCTimeStamp archive.Timestamp
		wantErr          bool
	}{
		{"generic import/local to utc conversion", mockAPI, args{context.Background(), 2, ts1},
			archive.Timestamp{
				DateModified: cleanTimestamp(ts1.DateModified),
				DateImported: cleanTimestamp(ts1.DateImported),
			},
			false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := mockAPI.SetTimestamps(context.Background(), tt.args.archive_id, tt.args.localTimeStamp); err != nil {
				t.Fatalf("API.GetTimestamps()/API.SetTimestamps() error = %v, wantErr %v", err, tt.wantErr)
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
		{"golang compatible format", mockAPI, args{context.Background(), 2, archive.Timestamp{
			DateModified: time.Now().Add(time.Hour * -300),
			DateImported: time.Now()},
		}, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := tt.a.SetTimestamps(tt.args.ctx, tt.args.archive_id, tt.args.t); (err != nil) != tt.wantErr {
				t.Errorf("API.SetTimestamps() error = %v, wantErr %v", err, tt.wantErr)
			}

			ts, err := tt.a.GetTimestamps(tt.args.ctx, tt.args.archive_id)
			if (err != nil) != tt.wantErr {
				t.Errorf("API.GetTimestamps() error = %v, wantErr %v", err, tt.wantErr)
			}

			if !reflect.DeepEqual(archive.Timestamp{
				DateModified: cleanTimestamp(tt.args.t.DateModified),
				DateImported: cleanTimestamp(tt.args.t.DateImported),
			}, ts) {
				t.Errorf("API.SetTimestamps() got = %v, want %v", ts, tt.args.t)
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

func TestAPI_Get(t *testing.T) {
	generic := NewMockEntry()
	generic.Entry.Metadata.PathRelative = file.BuildPath(generic.Hash().MD5, ".png")
	type args struct {
		ctx   context.Context
		entry Importer
	}
	tests := []struct {
		name    string
		a       *API
		args    args
		want    archive.Entry
		wantErr bool
	}{
		{"generic", mockAPI, args{context.Background(), generic}, generic.Entry, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			archive_id, err := tt.a.Import(tt.args.ctx, tt.args.entry, nil)
			if err != nil {
				t.Errorf("API.Get()/API.Import() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			got, err := tt.a.Get(tt.args.ctx, archive_id)
			if (err != nil) != tt.wantErr {
				t.Errorf("API.Get() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			// prevent deep.Equal from crying about blank timestamps
			// though we should properly test this in the future eventually...
			got.Metadata.Timestamp.DateImported = tt.want.Metadata.Timestamp.DateImported

			if diff := deep.Equal(got, tt.want); diff != nil {
				t.Errorf("API.Get() = %v", diff)
			}
		})
	}
}

func TestAPI_GetFile(t *testing.T) {
	f, err := os.Open("testdata/82d233bf13e0ebe6636db4d405d846c357d73c3cc491a97b85b9b235b4efdc80.png")
	if err != nil {
		t.Fatalf("GetFile() failed to open test data, %v", err)
	}

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
			// create new entry
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

			// import entry
			archive_id, err := tt.a.Import(tt.args.ctx, entry, nil)
			if err != nil {
				t.Fatalf("API.GetFile()/API.Import() unable to import entry. %v", err)
			}

			got, err := tt.a.GetFile(tt.args.ctx, archive_id)
			if (err != nil) != tt.wantErr {
				t.Errorf("API.GetFile() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			defer got.Close()

			w, err := io.Copy(io.Discard, got)
			if err != nil {
				t.Errorf("API.GetFile() error = %v, wantErr %v", err, tt.wantErr)
			}

			if w <= 0 {
				t.Error("API.GetFile() read 0 bytes from file")
			}
		})
	}
}
