package db

import (
	"database/sql"
	"fmt"
	"log/slog"
	"os"
	"reflect"
	"testing"

	"github.com/dtbead/moonpool/file"
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

	err = tempDB.createArchiveSchema()
	if err != nil {
		return err
	}

	err = tempDB.AddTag("foo")
	if err != nil {
		return err
	}

	fakeHash, err := generateRandomHash(128)
	if err != nil {
		return err
	}

	err = tempDB.InsertEntry(string(fakeHash), "cmd/moonpool/hawky.png", "png")
	if err != nil {
		return err
	}

	return nil
}

func TestSQLite3_SearchTag(t *testing.T) {
	type args struct {
		tag string
	}
	tests := []struct {
		name    string
		s       *SQLite3
		args    args
		want    []file.Entry
		wantErr bool
	}{
		{"exists", mockDB, args{"foo"}, []file.Entry{}, false}, // todo: add mock files for entry
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

func TestSQLite3_mapTags(t *testing.T) {
	type args struct {
		a    ArchiveID
		tags []Tag
	}
	tests := []struct {
		name    string
		s       *SQLite3
		args    args
		wantErr bool
	}{
		{"add", tempDB, args{1, []Tag{{ID: 1, Text: "foo"}}}, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := tt.s.mapTags(tt.args.a, tt.args.tags); (err != nil) != tt.wantErr {
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
		want TagID
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

func TestSQLite3_AddTags(t *testing.T) {
	tag1 := Tag{3, "meow"}
	tag2 := Tag{4, "woof"}
	multiTag := []Tag{tag1, tag2}
	type args struct {
		tags []Tag
	}
	tests := []struct {
		name    string
		s       *SQLite3
		args    args
		wantErr bool
	}{
		{"single", tempDB, args{[]Tag{Tag{2, "what"}}}, false},
		{"multiple", tempDB, args{multiTag}, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := tt.s.AddTags(tt.args.tags); (err != nil) != tt.wantErr {
				t.Errorf("SQLite3.AddTags() error = %v, wantErr %v", err, tt.wantErr)
			}

			for i := 0; i < len(tt.args.tags); i++ {
				// check if tag_id matches when searching for text
				if resID, err := tt.s.searchTagID(tt.args.tags[i].Text); resID != TagID(tt.args.tags[i].ID) || err != nil {
					t.Errorf("SQLite3.AddTags() expected {ID: %v}, got {ID: %v}", tt.args.tags[i].ID, resID)
				}
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
