package api

import (
	"context"
	_ "embed"
	"fmt"
	"reflect"
	"strings"
	"testing"
)

const (
	sqlSearchPredicateNONE_Template = sqlSearchPredicateNONE_Prologue + "?BINDVALUES?" + `) GROUP BY archive.id HAVING COUNT(archive.id) = ?TAGAMOUNT?)`
	sqlSearchPredicateNOT_Template  = sqlSearchPredicateNOT_Prologue + "?BINDVALUES?" + sqlSearchPredicateNOT_Epilogue
	sqlSearchPredicateOR_Template   = sqlSearchPredicateOR_Prologue + "?BINDVALUES?" + sqlSearchPredicateOR_Epilogue
)

func TestAPI_Query(t *testing.T) {
	mockAPI, dbPath, err := newMockAPI(Config{}, t, true)
	if err != nil {
		t.Fatalf("failed to create mock API. %v\n", err)
	}
	defer mockAPI.Close()

	archiveIDs, err := GenerateMockData(mockAPI, 6, true)
	if err != nil {
		t.Fatalf("failed to generate mock datmockAPI. %v\n", err)
	}

	if err := mockAPI.SetTags(context.Background(), archiveIDs[0], []string{"foo"}); err != nil {
		t.Fatalf("failed to set tag for archive_id %d. %v\n", archiveIDs[0], err)
	}

	if err := mockAPI.SetTags(context.Background(), archiveIDs[1], []string{"foo", "bar"}); err != nil {
		t.Fatalf("failed to set tag for archive_id %d. %v\n", archiveIDs[1], err)
	}

	if err := mockAPI.SetTags(context.Background(), archiveIDs[2], []string{"bar", "zap"}); err != nil {
		t.Fatalf("failed to set tag for archive_id %d. %v\n", archiveIDs[2], err)
	}

	if err := mockAPI.SetTags(context.Background(), archiveIDs[3], []string{"zap"}); err != nil {
		t.Fatalf("failed to set tag for archive_id %d. %v\n", archiveIDs[3], err)
	}

	if err := mockAPI.SetTags(context.Background(), archiveIDs[4], []string{"bee"}); err != nil {
		t.Fatalf("failed to set tag for archive_id %d. %v\n", archiveIDs[4], err)
	}

	if err := mockAPI.SetTags(context.Background(), archiveIDs[5], []string{"bar"}); err != nil {
		t.Fatalf("failed to set tag for archive_id %d. %v\n", archiveIDs[5], err)
	}

	type args struct {
		ctx context.Context
		q   SearchQuery
	}
	tests := []struct {
		test    bool
		name    string
		a       API
		args    args
		want    []int64
		wantErr bool
	}{
		{true, "search single with no predicates", *mockAPI, args{context.Background(), []string{"bee"}}, []int64{archiveIDs[4]}, false},
		{false, "search multiple with no predicates", *mockAPI, args{context.Background(), []string{"foo", "bar"}}, []int64{archiveIDs[1]}, false},
		{false, "search with NOT predicate", *mockAPI, args{context.Background(), []string{"-foo", "bar"}}, []int64{archiveIDs[2], archiveIDs[5]}, false},
		{false, "search with multiple NOT predicate", *mockAPI, args{context.Background(), []string{"-foo", "-zap", "bar"}}, []int64{archiveIDs[5]}, false},
	}
	for _, tt := range tests {
		if tt.test {
			t.Run(tt.name, func(t *testing.T) {
				got, err := tt.a.Query(tt.args.ctx, tt.args.q)
				if (err != nil) != tt.wantErr {
					t.Fatalf("API.Query() error = %v, wantErr %v", err, tt.wantErr)
				}
				if !reflect.DeepEqual(got, tt.want) {
					t.Logf("Database path: %s", dbPath)
					t.Errorf("API.Query() archive_id = %v, want %v", got, tt.want)
				}
			})
		} else {
			fmt.Printf("!!!! SKIPPED TEST '%s' !!!!\n", tt.name)
		}
	}
}

func Test_buildPredicate(t *testing.T) {
	type args struct {
		predicate rune
		s         []string
	}
	tests := []struct {
		test bool
		name string
		args args
		want string
	}{
		{true, "empty/no input", args{0, nil}, ""},
		{true, "predicate NONE string formatting", args{SEARCH_PREDICATE_NONE, []string{"foo", "bar"}}, fmt.Sprintf("%s?, ?) GROUP BY archive.id HAVING COUNT(archive.id) = 2)", sqlSearchPredicateNONE_Prologue)},
		{true, "predicate NOT string formatting", args{SEARCH_PREDICATE_NOT, []string{"foo", "bar"}}, fmt.Sprintf("%s?, ?) AND archive.id IN predicate_none GROUP BY archive.id)", sqlSearchPredicateNOT_Prologue)},
	}
	for _, tt := range tests {
		if tt.test {
			t.Run(tt.name, func(t *testing.T) {
				got := buildPredicate(tt.args.predicate, tt.args.s)
				got = deleteWhitespace(got)
				tt.want = deleteWhitespace(tt.want)

				if got != tt.want {
					t.Errorf("got\n%v\nwant\n%v\n", got, tt.want)
				}
			})
		} else {
			fmt.Printf("!!!! SKIPPED TEST '%s' !!!!\n", tt.name)
		}
	}
}

func Test_buildQuery(t *testing.T) {
	test2 := "WITH " + strings.ReplaceAll(sqlSearchPredicateNONE_Template, "?BINDVALUES?", "?")
	test2 = strings.ReplaceAll(test2, "?TAGAMOUNT?", "1")
	test2 = test2 + ", " + strings.ReplaceAll(sqlSearchPredicateNOT_Template, "?BINDVALUES?", "?") + " SELECT id FROM predicate_none WHERE id NOT IN predicate_not;"

	type args struct {
		q SearchQuery
	}
	tests := []struct {
		test  bool
		name  string
		args  args
		want  string
		want1 []string
	}{
		{true, "NONE predicates", args{q: []string{"foo", "bar"}}, fmt.Sprintf("WITH %s?, ?) GROUP BY archive.id HAVING COUNT(archive.id) = %d) SELECT id FROM predicate_none", sqlSearchPredicateNONE_Prologue, 2), []string{"foo", "bar"}},
		{true, "NONE and NOT predicates", args{q: []string{"foo", "-bar"}}, test2, []string{"foo", "bar"}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, got1 := buildQuery(tt.args.q)
			got = deleteWhitespace(got)
			tt.want = deleteWhitespace(tt.want)

			if got != tt.want {
				t.Errorf("buildQuery()\ngot\n%v\n\nwant\n%v\n", got, tt.want)
			}
			if !reflect.DeepEqual(got1, tt.want1) {
				t.Errorf("buildQuery() got1 = %v, want %v", got1, tt.want1)
			}
		})
	}
}
