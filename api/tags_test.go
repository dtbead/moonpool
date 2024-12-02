package api

import (
	"context"
	"fmt"
	"reflect"
	"slices"
	"testing"

	"github.com/dtbead/moonpool/entry"
)

func TestAPI_QueryTags(t *testing.T) {
	mockAPI, err := newMockAPI(Config{ArchiveLocation: ":memory:", ThumbnailLocation: ":memory:"}, t)
	if err != nil {
		t.Fatalf("failed to create mock API. %v", err)
	}
	archive_ids, err := GenerateMockData(mockAPI, 2, false)
	if err != nil {
		t.Fatalf("failed to generate mock data, %v", err)
	}

	err = mockAPI.SetTags(context.Background(), archive_ids[0], []string{"foo"})
	if err != nil {
		t.Fatalf("failed to set tag, %v", err)
	}

	err = mockAPI.SetTags(context.Background(), archive_ids[1], []string{"foo", "bar"})
	if err != nil {
		t.Fatalf("failed to set tag, %v", err)
	}

	err = mockAPI.NewTagAlias(context.Background(), "foo", "foo_alias")
	if err != nil {
		t.Fatalf("failed to set tag, %v", err)
	}

	err = mockAPI.NewTagAlias(context.Background(), "bar", "bar_alias")
	if err != nil {
		t.Fatalf("failed to set tag, %v", err)
	}

	/*
		foo -> foo_alias
		bar -> bar_alias

		'foo' tag IN archive_id[1, 2]
		'foo', 'bar' tags IN archive_id[2]
	*/

	type args struct {
		ctx  context.Context
		sort string
		q    QueryTags
	}
	tests := []struct {
		name    string
		a       *API
		args    args
		want    []int64
		wantErr bool
	}{
		{"tagInclude only", mockAPI, args{context.Background(), "imported", QueryTags{[]string{"foo"}, nil}}, archive_ids, false},
		{"tagInclude + tagExclude", mockAPI, args{context.Background(), "imported", QueryTags{[]string{"foo"}, []string{"bar"}}}, []int64{1}, false},
		{"resolve tagInclude alias", mockAPI, args{context.Background(), "imported", QueryTags{[]string{"foo_alias"}, []string{}}}, archive_ids, false},
		{"resolve tagExclude alias", mockAPI, args{context.Background(), "imported", QueryTags{[]string{"foo"}, []string{"bar_alias"}}}, []int64{1}, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := tt.a.QueryTags(tt.args.ctx, tt.args.sort, tt.args.q)
			if (err != nil) != tt.wantErr {
				t.Errorf("API.QueryTags() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !slices.Equal(got, tt.want) {
				t.Errorf("API.QueryTags() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestAPI_GetTagsByRange(t *testing.T) {
	mockAPI, err := newMockAPI(Config{ArchiveLocation: ":memory:", ThumbnailLocation: ":memory:"}, t)
	if err != nil {
		t.Fatalf("failed to create mock API. %v", err)
	}

	if _, err := GenerateMockData(mockAPI, 50, true); err != nil {
		t.Fatalf("failed to generate mock data. %v", err)
	}

	type args struct {
		ctx    context.Context
		start  int64
		end    int64
		offset int64
	}
	tests := []struct {
		name    string
		a       *API
		args    args
		wantErr bool
	}{
		{"generic", mockAPI, args{context.Background(), 0, 50, 0}, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := tt.a.GetTagsByRange(tt.args.ctx, tt.args.start, tt.args.end, tt.args.offset)
			if (err != nil) != tt.wantErr {
				t.Errorf("API.GetTagsByRange() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
		})
	}
}

func TestAPI_GetTagCount(t *testing.T) {
	mockAPI, err := newMockAPI(Config{ArchiveLocation: ":memory:", ThumbnailLocation: ":memory:"}, t)
	if err != nil {
		t.Fatalf("failed to create mock API. %v", err)
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

func TestAPI_SetTags(t *testing.T) {
	mockAPI, err := newMockAPI(Config{ArchiveLocation: t.TempDir() + "/moonpool_SetTags.sqlite3", ThumbnailLocation: ":memory:"}, t)
	if err != nil {
		t.Fatalf("failed to create mock API. %v", err)
	}
	defer mockAPI.Close(context.Background())

	archive_ids, err := GenerateMockData(mockAPI, 3, false)
	if err != nil {
		t.Fatalf("failed to generate mock data. %v", err)
	}

	multipleTags := make([]string, 0, 10)
	for i := 0; i < 10; i++ {
		multipleTags = append(multipleTags, randomString(6))
	}

	type args struct {
		ctx        context.Context
		archive_id int64
		tags       []string
	}
	tests := []struct {
		test    bool
		name    string
		a       *API
		args    args
		wantErr bool
	}{
		{true, "single tag", mockAPI, args{context.Background(), archive_ids[0], []string{randomString(6)}}, false},
		{true, "multiple tags", mockAPI, args{context.Background(), archive_ids[1], multipleTags}, false},
		{true, "ignore duplicate tags error", mockAPI, args{context.Background(), archive_ids[2], []string{"foo", "foo"}}, false},
	}
	for _, tt := range tests {
		if tt.test {
			t.Run(tt.name, func(t *testing.T) {
				fmt.Printf("Database path: %s\n", tt.a.Config.ArchiveLocation)
				if err := tt.a.SetTags(tt.args.ctx, tt.args.archive_id, tt.args.tags); (err != nil) != tt.wantErr {
					t.Fatalf("API.SetTags() error = %v, wantErr %v", err, tt.wantErr)
				}

				got, err := tt.a.GetTags(tt.args.ctx, tt.args.archive_id)
				if err != nil {
					t.Errorf("API.SetTags()/API.GetTags() error = %v, wantErr %v", err, tt.wantErr)
				}

				if !reflect.DeepEqual(got, removeDuplicateStr(tt.args.tags)) {
					t.Errorf("API.SetTags() got %v, want %v", got, tt.args.tags)
				}

			})
		} else {
			fmt.Printf("!!!! SKIPPED TEST '%s' !!!!\n", tt.name)
		}
	}
}

func TestAPI_SetTagsWithAlias(t *testing.T) {
	mockAPI, err := newMockAPI(Config{ArchiveLocation: ":memory:", ThumbnailLocation: ":memory:"}, t)
	if err != nil {
		t.Fatalf("failed to create mock API. %v", err)
	}
	defer mockAPI.Close(context.Background())

	_, err = GenerateMockData(mockAPI, 3, false)
	if err != nil {
		t.Fatalf("failed to generate mock data. %v", err)
	}

	type args struct {
		ctx        context.Context
		archive_id int64
		aliasTag   string
	}
	tests := []struct {
		name    string
		a       *API
		args    args
		wantErr bool
		wantTag string
	}{
		{"single tag", mockAPI, args{context.Background(), 1, "bar"}, false, "foo"},
		{"no alias tag", mockAPI, args{context.Background(), 2, "zzz"}, true, "zzz"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.wantTag != "" {
				_, err = tt.a.archive.NewTag(tt.args.ctx, tt.wantTag)
				if err != nil {
					t.Fatalf("API.SetTags() error = %v, wantErr %v", err, tt.wantErr)
				}
			}

			if err = tt.a.NewTagAlias(tt.args.ctx, tt.wantTag, tt.args.aliasTag); (err != nil) != tt.wantErr {
				t.Fatalf("API.SetTags()/API.NewTagAlias() error = %v, wantErr %v", err, tt.wantErr)
			}

			if err := tt.a.SetTags(tt.args.ctx, tt.args.archive_id, []string{tt.args.aliasTag}); err != nil {
				t.Fatalf("API.SetTags() error = %v", err)
			}

			got, err := tt.a.GetTags(tt.args.ctx, tt.args.archive_id)
			if err != nil {
				t.Errorf("API.SetTags()/API.GetTags() error = %v", err)
			}

			if got[0] != tt.wantTag {
				t.Errorf("API.SetTags() got = %v, wantTag %v", tt.args.aliasTag, tt.wantTag)
			}
		})
	}
}

func TestAPI_GetTagsByList(t *testing.T) {
	mockAPI, err := newMockAPI(Config{ArchiveLocation: ":memory:", ThumbnailLocation: ":memory:"}, t)
	if err != nil {
		t.Fatalf("failed to create mock API. %v", err)
	}
	archive_ids, err := GenerateMockData(mockAPI, 2, false)
	if err != nil {
		t.Fatalf("failed to generate mock data, %v", err)
	}

	err = mockAPI.SetTags(context.Background(), archive_ids[0], []string{"foo"})
	if err != nil {
		t.Fatalf("failed to set tag, %v", err)
	}

	err = mockAPI.SetTags(context.Background(), archive_ids[1], []string{"foo", "bar"})
	if err != nil {
		t.Fatalf("failed to set tag, %v", err)
	}

	type args struct {
		ctx         context.Context
		archive_ids []int64
	}
	tests := []struct {
		name    string
		a       *API
		args    args
		want    []entry.TagCount
		wantErr bool
	}{
		{"exists", mockAPI, args{context.Background(), []int64{1}}, []entry.TagCount{
			{Text: "foo", Count: 1},
		}, false},
		{"multiple exists", mockAPI, args{context.Background(), []int64{1, 2}}, []entry.TagCount{
			{Text: "foo", Count: 2},
			{Text: "bar", Count: 1},
		}, false},
		{"not exists", mockAPI, args{context.Background(), []int64{3}}, []entry.TagCount{}, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := tt.a.GetTagsByList(tt.args.ctx, tt.args.archive_ids)
			if (err != nil) != tt.wantErr {
				t.Errorf("API.GetTagsByList() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("API.GetTagsByList() = %v, want %v", got, tt.want)
			}

			cmp := func(a entry.TagCount, b entry.TagCount) int {
				switch {
				default:
					return 0
				case a.Count < b.Count:
					return -1
				case a.Count > b.Count:
					return 1
				}
			}

			if !slices.IsSortedFunc(got, cmp) {
				return
			}
		})
	}
}
