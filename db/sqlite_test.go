package db

import (
	"database/sql"
	"testing"

	_ "modernc.org/sqlite"
)

var testdb = "testdata/db.sqlite3"
var testdb2 = "testdata/prefilled.sqlite3"
var db, _ = sql.Open("sqlite", testdb)
var pdb, _ = sql.Open("sqlite", testdb2)

func TestCreateArchiveDatabaseScheme(t *testing.T) {
	type args struct {
		TableName string
		db        *sql.DB
	}

	type tables struct {
		name string
		sql  string
	}

	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{"general name", args{"foobar", db}, false},
		{"gibberish", args{"juy76FD*&YUIOP{ERPLK ,'-p2/},", db}, true},
	}

	statement, err := db.Prepare(`SELECT name, sql FROM "main".sqlite_master WHERE name == ?;`)
	if err != nil {
		t.Fatalf("CreateArchiveDatabaseScheme() unable to create SQL query. %v", err)
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := CreateArchiveDatabaseScheme(tt.args.TableName, tt.args.db); (err != nil) != tt.wantErr {
				t.Errorf("CreateArchiveDatabaseScheme() error = %v, wantErr %v", err, tt.wantErr)
			}
		})

		rows, err := statement.Query(tt.args.TableName)
		if err != nil {
			t.Errorf("CreateArchiveDatabaseScheme() failed to query database for table names: %v", err)
		}
		for rows.Next() {
			var res tables
			if err := rows.Scan(
				&res.name, &res.sql,
			); err != nil {
				t.Log(err.Error())
			}

			if res.name == "sqlite_sequence" {
				return
			}

			if res.name == tt.args.TableName {
				return
			} else {
				t.Logf("CreateArchiveDatabaseScheme() SQL query = '%v'", res.sql)
				t.Errorf("CreateArchiveDatabaseScheme() expected %v, got %v", tt.args.TableName, res.name)

			}
		}
		if rows.Err() != nil {
			t.Log(rows.Close())
		}
	}
}

func TestSingleQuery(t *testing.T) {
	type args struct {
		table string
		row   string
		value string
		db    *sql.DB
	}
	tests := []struct {
		name    string
		args    args
		want    string
		wantErr bool
	}{
		{"valid", args{"foo", "bar", "test", pdb}, "test", false},
		{"missing table", args{"blah", "bar", "test", pdb}, "", true},
		{"empty results", args{"foo", "bar", "123", pdb}, "", false},
		{"missing row", args{"foo", "123", "test", pdb}, "", true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := SingleQuery(tt.args.table, tt.args.row, tt.args.value, tt.args.db)
			if (err != nil) != tt.wantErr {
				t.Errorf("SingleQuery() error on test '%v' = %v, wantErr %v", tt.name, err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("SingleQuery() on test '%v', got = %v, want %v", tt.name, got, tt.want)
			}
		})
	}
}

func Test_doesTableExist(t *testing.T) {
	type args struct {
		table string
		db    *sql.DB
	}
	tests := []struct {
		name string
		args args
		want bool
	}{
		{"exists", args{"foo", pdb}, true},
		{"non-existent", args{"123", pdb}, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := doesTableExist(tt.args.table, tt.args.db); got != tt.want {
				t.Errorf("doesTableExist() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestAddTag(t *testing.T) {
	type args struct {
		table string
		tag   string
		db    *sql.DB
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{"general", args{"tags", "bar", pdb}, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := AddTag(tt.args.table, tt.args.tag, tt.args.db); (err != nil) != tt.wantErr {
				t.Errorf("AddTag() error = %v, wantErr %v", err, tt.wantErr)
			}
			if res, err := SingleQuery(tt.args.table, "text", tt.args.tag, tt.args.db); err != nil || res == "" {
				t.Errorf("AddTag() error = %v, wantErr %v, got result '%v'", err, tt.wantErr, res)
			}
		})
	}
}

func TestAddTags(t *testing.T) {
	type args struct {
		tags []string
		db   *sql.DB
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{"general", args{[]string{"foo"}, pdb}, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := AddTags(tt.args.tags, tt.args.db); (err != nil) != tt.wantErr {
				t.Errorf("AddTags() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func Test_doesColumnExist(t *testing.T) {
	type args struct {
		table string
		row   string
		db    *sql.DB
	}
	tests := []struct {
		name string
		args args
		want bool
	}{
		{"exists", args{"foo", "bar", pdb}, true},
		{"not exist", args{"foo", "123", pdb}, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := doesColumnExist(tt.args.table, tt.args.row, tt.args.db); got != tt.want {
				t.Errorf("doesColumnExist() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestFindTagID(t *testing.T) {
	type args struct {
		tag string
		db  *sql.DB
	}
	tests := []struct {
		name string
		args args
		want uint
	}{
		{"exists", args{"foo", pdb}, 1},
		{"not exists", args{"testsetset", pdb}, 0},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := FindTagID(tt.args.tag, tt.args.db); got != tt.want {
				t.Errorf("FindTagID() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestMapTags(t *testing.T) {
	type args struct {
		archiveID uint
		tag       []string
		db        *sql.DB
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{"add", args{1, []string{"foo"}, pdb}, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := MapTags(tt.args.archiveID, tt.args.tag, tt.args.db); (err != nil) != tt.wantErr {
				t.Errorf("MapTags() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestSearchTag(t *testing.T) {
	type args struct {
		tag string
		db  *sql.DB
	}
	tests := []struct {
		name string
		args args
		want uint
	}{
		{"exists", args{"foo", pdb}, 1},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := SearchTag(tt.args.tag, tt.args.db); got != tt.want {
				t.Errorf("SearchTag() = %v, want %v", got, tt.want)
			}
		})
	}
}
