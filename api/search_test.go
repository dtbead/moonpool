package api

import (
	"context"
	_ "embed"
	"fmt"
	"testing"

	archive "github.com/dtbead/moonpool/db"
)

func Test_buildGeneralTagQuery(t *testing.T) {
	type args struct {
		s []string
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{"generic", args{[]string{"foo", "bar"}}, " WHERE tags.text IN ('foo', 'bar')"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := buildGeneralTagQuery(tt.args.s); got != tt.want {
				t.Errorf("buildGeneralTagQuery() = %v, want %v", got, tt.want)
			}
		})
	}
}

//go:embed testdata/Test_buildQuery_general_tags.txt
var generalTagsWant string

/*
func TestAPI_Query(t *testing.T) {
	mockAPI, err := newMockAPI()
	if err != nil {
		t.Fatalf("failed to create new mock API. %v", err)
	}

	bsearch, err := populateQuery(10, mockAPI)
	if err != nil {
		t.Fatalf("failed to populate entries with random tags. %v", err)
	}

	type args struct {
		ctx context.Context
		q   SearchQuery
	}
	tests := []struct {
		name    string
		a       API
		args    args
		want    []int64
		wantErr bool
	}{
		{"basic searching (no predicates)", *mockAPI,
			args{context.Background(), getTagSlice(bsearch[0])[:3]},
			[]int64{bsearch[0].ArchiveID}, false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := tt.a.Query(tt.args.ctx, tt.args.q)
			if (err != nil) != tt.wantErr {
				t.Errorf("API.Query() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			var archiveIDs []int64
			for _, v := range got {
				archiveIDs = append(archiveIDs, v.ArchiveID)
			}

			if res := deep.Equal(archiveIDs, tt.want); res != nil {
				for i := range res {
					res[i] = res[i] + "\n"
				}
				t.Errorf("API.Query() = \n%v", res)
			}
		})
	}
}
*/

func getTagSlice(e archive.EntryTags) []string {
	t := make([]string, len(e.Tags))

	for i, v := range e.Tags {
		t[i] = v.Text
	}

	return t
}

func populateQuery(amount int, a *API) ([]archive.EntryTags, error) {
	const totalTags = 10
	e := make([]archive.EntryTags, amount)

	archiveIDs, err := GenerateMockData(a, amount, true)
	if err != nil {
		return nil, err
	}

	// generate random tags that will be shared across every entry
	mockTags := make([]string, totalTags)
	for i := 0; i < totalTags; i++ {
		mockTags[i] = randomString(16)
	}

	// do the same thing as above, but for a archive.Tag{} struct to later fit
	// into our archive.EntryTags{} slice struct
	archiveTags := make([]archive.Tag, totalTags)
	for i := 0; i < totalTags; i++ {
		archiveTags[i] = archive.Tag{Text: mockTags[i]}
	}

	// set the random tags that will be shared across every entry
	for i := 0; i < amount; i++ {
		if err := a.SetTags(context.Background(), archiveIDs[i], mockTags); err != nil {
			return nil, err
		}
		e[i] = archive.EntryTags{ArchiveID: archiveIDs[i], Tags: archiveTags}
	}

	// generate random tags
	var randomTags = make([]string, totalTags)
	for i := 0; i < amount; i++ {
		for i := 0; i < totalTags; i++ {
			randomTags[i] = randomString(12)
		}

		if err := a.SetTags(context.Background(), archiveIDs[i], randomTags); err != nil {
			return nil, err
		}

		for _, v := range randomTags {
			e[i].Tags = append(e[i].Tags, archive.Tag{Text: v})
		}
	}

	return e, nil
}

func Test_buildNotTagQuery(t *testing.T) {
	type args struct {
		s []string
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{"generic", args{[]string{"foo", "bar"}}, " WHERE tags.text NOT IN ('foo', 'bar')"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := buildNotTagQuery(tt.args.s); got != tt.want {
				t.Errorf("buildNotTagQuery() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_buildQuery(t *testing.T) {
	type args struct {
		q SearchQuery
	}
	tests := []struct {
		name    string
		args    args
		want    string
		wantErr bool
	}{
		{"general with NOT predicate", args{SearchQuery{
			[]query{
				{
					Predicate: SEARCH_PREDICATE_NONE,
					Tag:       []string{"foo"},
				},
				{
					Predicate: SEARCH_PREDICATE_NOT,
					Tag:       []string{"bar"},
				},
			},
		}}, deleteWhitespace(fmt.Sprintf("%s %s AND %s;", sqlSearchPreliminary, "WHERE tags.text IN ('foo')", "WHERE tags.text NOT IN ('bar')")), false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := buildQuery(tt.args.q)
			if (err != nil) != tt.wantErr {
				t.Errorf("buildQuery() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("buildQuery() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_buildOrTagQuery(t *testing.T) {
	type args struct {
		s []string
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{"generic", args{[]string{"foo", "bar"}}, " WHERE tags.text IN ('foo', 'bar') GROUP BY tags.tag_id;"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := buildOrTagQuery(tt.args.s); got != tt.want {
				t.Errorf("buildOrTagQuery() = %v, want %v", got, tt.want)
			}
		})
	}
}
