package db

import (
	"database/sql"
	"fmt"
	"log/slog"
	"os"
	"reflect"
	"testing"

	"github.com/dtbead/moonpool/media"
	_ "modernc.org/sqlite"
)

const mockDBPath = "testdata/mock/db.sqlite3"

var mockDB *SQLite3 // database with pre-existing data
var tempDB *SQLite3 // database with tables but no data

func TestMain(m *testing.M) {
	memDB, _ := sql.Open("sqlite", ":memory:")
	tempDB = &SQLite3{db: memDB}
	if err := initializeTempDB(); err != nil {
		slog.Error(fmt.Sprintf("failed to initialize in-memory database, %v", err))
		os.Exit(1)
	}

	var err error
	mockDB, err = OpenSQLite3(mockDBPath)
	if err != nil {
		slog.Error("failed to open '%s'. %v", mockDBPath, err)
		os.Exit(1)
	}

	code := m.Run()
	os.Exit(code)
}

func initializeTempDB() error {
	var err error

	err = tempDB.Initialize()
	if err != nil {
		return err
	}

	err = tempDB.AddTag("foo")
	if err != nil {
		return err
	}

	var fakeHash = media.Hashes{
		MD5:    generateRandomHash(16),
		SHA1:   generateRandomHash(20),
		SHA256: generateRandomHash(32),
	}

	_, err = tempDB.InsertEntry(fakeHash, "cmd/moonpool/hawky.png", "png")
	if err != nil {
		return err
	}

	return nil
}

func TestSQLite3_SearchTag(t *testing.T) {
	var exists = media.Entry{
		ArchiveID: 1,
		Metadata: media.Metadata{
			PathRelative: "d4/d41d8cd98f00b204e9800998ecf8427e.png",
			MD5Hash:      "d41d8cd98f00b204e9800998ecf8427e",
			Hash: media.Hashes{
				MD5: []byte{212, 29, 140, 217, 143, 0, 178, 4, 233, 128, 9, 152, 236, 248, 66, 126},
			},
		},
	}

	type args struct {
		tag string
	}
	tests := []struct {
		name    string
		s       *SQLite3
		args    args
		want    []media.Entry
		wantErr bool
	}{
		{"not exist", mockDB, args{"foo"}, nil, false},
		{"exists", mockDB, args{"hawkfrost"}, []media.Entry{exists}, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := tt.s.SearchTag(tt.args.tag)
			if (err != nil) != tt.wantErr {
				t.Errorf("SQLite3.SearchTag() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("SQLite3.SearchTag() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestSQLite3_MapTags(t *testing.T) {
	type args struct {
		a    int
		tags []string
	}
	tests := []struct {
		name    string
		s       *SQLite3
		args    args
		wantErr bool
	}{
		{"add", tempDB, args{1, []string{"foo"}}, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := tt.s.MapTags(tt.args.a, tt.args.tags); (err != nil) != tt.wantErr {
				t.Errorf("SQLite3.MapTags() error = %v, wantErr %v", err, tt.wantErr)
			}

			res, err := tempDB.searchTagMaps(1)
			if err != nil {
				t.Errorf("SQLite3.MapTags()/SQLite3.SingleQuery() error = %v, wantErr %v", err, tt.wantErr)
				t.FailNow()
			}

			if len(res) == 0 {
				t.Errorf("SQLite3.MapTags()/SQLite3.SingleQuery() found no result for tag mapping")
			}

			if res[0] != int(tt.args.a) {
				t.Errorf("SQLite3.MapTags()/SQLite3.SingleQuery() expected archive_id %d, got '%v'", tt.args.a, res[0])
			}
		})
	}
}

func TestSQLite3_searchTagID(t *testing.T) {
	type args struct {
		tag string
	}
	tests := []struct {
		name string
		s    *SQLite3
		args args
		want int
	}{
		{"exists", mockDB, args{"foo"}, 1},
		{"not exist", mockDB, args{"testsetset"}, -1},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got, err := tt.s.searchTagID(tt.args.tag); got != tt.want || err != nil {
				t.Errorf("SQLite3.FindTagID() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestSQLite3_AddTag(t *testing.T) {
	type args struct {
		table string
		tag   string
	}
	tests := []struct {
		name    string
		s       *SQLite3
		args    args
		wantErr bool
	}{
		{"general", mockDB, args{"foo", "bar"}, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := tt.s.AddTag(tt.args.tag); (err != nil) != tt.wantErr {
				t.Errorf("SQLite3.AddTag() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestSQLite3_doesTableExist(t *testing.T) {
	type args struct {
		table string
	}
	tests := []struct {
		name string
		s    *SQLite3
		args args
		want bool
	}{
		{"exists", mockDB, args{"archive"}, true},
		{"non-existent", mockDB, args{"123"}, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.s.doesTableExist(tt.args.table); got != tt.want {
				t.Errorf("SQLite3.doesTableExist() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestSQLite3_doesColumnExist(t *testing.T) {
	type args struct {
		table string
		row   string
	}
	tests := []struct {
		name string
		s    *SQLite3
		args args
		want bool
	}{
		{"exists", mockDB, args{"tags", "text"}, true},
		{"not exist", mockDB, args{"tags", "123"}, false},
		{"table not exists", mockDB, args{"foo", "bar"}, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.s.doesColumnExist(tt.args.table, tt.args.row); got != tt.want {
				t.Errorf("SQLite3.doesColumnExist() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestSQLite3_getTotalResults(t *testing.T) {
	type args struct {
		table string
		row   string
		value string
	}
	tests := []struct {
		name string
		s    *SQLite3
		args args
		want int
	}{
		{"single results", mockDB, args{"tags", "text", "foo"}, 1},
		// {"multiple results", mockDB, args{"tags", "text", "foo"}, 4}, // todo: add more data to mockDB for test
		{"no results", mockDB, args{"tags", "text", "thistagshouldntexist"}, 0},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.s.getTotalResults(tt.args.table, tt.args.row, tt.args.value); got != tt.want {
				t.Errorf("SQLite3.getTotalResults() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestSQLite3_AddTags(t *testing.T) {
	multipleTags := []media.Tag{
		{Text: "meow", ID: 3},
		{Text: "wolf", ID: 4},
	}

	type args struct {
		t []string
	}
	tests := []struct {
		name    string
		s       *SQLite3
		args    args
		want    []media.Tag
		wantErr bool
	}{
		{"single", tempDB, args{[]string{"what"}}, []media.Tag{{Text: "what", ID: 2}}, false},
		{"multiple", tempDB, args{[]string{"meow", "wolf"}}, multipleTags, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := tt.s.AddTags(tt.args.t)
			if (err != nil) != tt.wantErr {
				t.Errorf("SQLite3.AddTags() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("SQLite3.AddTags() = %v, want %v", got, tt.want)
			}

			for i := 0; i < len(tt.args.t); i++ {

				if id, err := tt.s.searchTagID(tt.want[i].Text); id != tt.want[i].ID || err != nil {
					t.Errorf("SQLite3.AddTags()/SQLite3.searchTagID = %v, want %v", id, tt.want[i].ID)

				}
			}
		})
	}
}