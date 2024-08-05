package api

import (
	"testing"
)

func Test_buildSearchQuery(t *testing.T) {
	type args struct {
		queries []search
	}
	tests := []struct {
		name    string
		args    args
		want    string
		wantErr bool
	}{
		{"simple OR predicate", args{queries: []search{{"foo", SEARCH_PREDICATE_OR}}}, preliminarySearchStatement + "tags.text OR (?)", false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := buildSearchQuery(tt.args.queries)
			if (err != nil) != tt.wantErr {
				t.Errorf("buildSearchQuery() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("buildSearchQuery() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_convertSystemTagsToSQLKeyword(t *testing.T) {
	type args struct {
		s string
	}
	tests := []struct {
		name    string
		args    args
		want    string
		wantErr bool
	}{
		{"valid 'limit:' tag", args{"limit:5"}, "5", false},
		{"empty 'limit:' tag", args{"limit:"}, "", false},
		{"valid 'range:' tag", args{"range:1-100"}, "1-100", false},
		{"empty 'range:' tag", args{"range:"}, "", false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := convertSystemTagsToSQLKeyword(tt.args.s)
			if (err != nil) != tt.wantErr {
				t.Errorf("convertSystemTagsToSQLKeyword() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("convertSystemTagsToSQLKeyword() = %v, want %v", got, tt.want)
			}
		})
	}
}
