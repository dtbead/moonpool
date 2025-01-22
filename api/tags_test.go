package api

import (
	"context"
	"fmt"
	"reflect"
	"slices"
	"testing"
	"time"

	"github.com/dtbead/moonpool/entry"
)

func TestAPI_QueryTags(t *testing.T) {
	mockAPI, err := newMockAPI(Config{ArchiveLocation: ":memory:", ThumbnailLocation: ":memory:"}, t)
	if err != nil {
		t.Fatalf("failed to create mock API. %v", err)
	}

	archive_ids, err := GenerateMockData(mockAPI, 2, false, true)
	if err != nil {
		t.Fatalf("failed to generate mock data, %v", err)
	}

	var inc = 0
	for _, v := range archive_ids {
		err := mockAPI.SetTimestamps(context.Background(), v, entry.Timestamp{
			DateImported: time.Now().Add(time.Duration(inc) * time.Hour)})
		if err != nil {
			t.Fatalf("failed to set import timestamp, %v", err)
		}

		inc += 5
	}

	err = mockAPI.AssignTags(context.Background(), archive_ids[0], []string{"foo"})
	if err != nil {
		t.Fatalf("failed to set tag, %v", err)
	}

	err = mockAPI.AssignTags(context.Background(), archive_ids[1], []string{"foo", "bar"})
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
		ctx         context.Context
		sort, order string
		q           QueryTags
	}
	tests := []struct {
		name    string
		a       *API
		args    args
		want    []int64
		wantErr bool
	}{
		{"tagInclude only", mockAPI, args{context.Background(), "imported", "descending", QueryTags{[]string{"foo"}, nil}}, []int64{2, 1}, false},
		{"tagInclude + tagExclude", mockAPI, args{context.Background(), "imported", "descending", QueryTags{[]string{"foo"}, []string{"bar"}}}, []int64{1}, false},
		{"resolve tagInclude alias", mockAPI, args{context.Background(), "imported", "descending", QueryTags{[]string{"foo_alias"}, []string{}}}, []int64{2, 1}, false},
		{"resolve tagExclude alias", mockAPI, args{context.Background(), "imported", "descending", QueryTags{[]string{"foo"}, []string{"bar_alias"}}}, []int64{1}, false},
		{"tagInclude only ascending order", mockAPI, args{context.Background(), "imported", "ascending", QueryTags{[]string{"foo"}, nil}}, []int64{1, 2}, false},
		{"resolve tagInclude alias ascending order", mockAPI, args{context.Background(), "imported", "ascending", QueryTags{[]string{"foo_alias"}, []string{}}}, []int64{1, 2}, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := tt.a.QueryTags(tt.args.ctx, tt.args.sort, tt.args.order, tt.args.q)
			if (err != nil) != tt.wantErr {
				t.Errorf("API.QueryTags() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !slices.Equal(got, tt.want) {
				for _, archive_id := range got {
					ts, err := tt.a.GetTimestamps(tt.args.ctx, archive_id)
					if err != nil {
						t.Errorf("failed to get timestamp for archive_id %d, %v\n", archive_id, err)
					}

					fmt.Printf("archive_id: %d\t %+v\n", archive_id, ts)
				}
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

	if _, err := GenerateMockData(mockAPI, 50, true, true); err != nil {
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

	if _, err := GenerateMockData(mockAPI, 1, false, false); err != nil {
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

			if err := mockAPI.AssignTags(tt.args.ctx, tt.want, []string{tag}); err != nil {
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

func TestAPI_AssignTags(t *testing.T) {
	mockAPI, err := newMockAPI(Config{ArchiveLocation: t.TempDir() + "/moonpool_AssignTags.sqlite3", ThumbnailLocation: ":memory:"}, t)
	if err != nil {
		t.Fatalf("failed to create mock API. %v", err)
	}
	defer mockAPI.Close(context.Background())

	archive_ids, err := GenerateMockData(mockAPI, 4, false, false)
	if err != nil {
		t.Fatalf("failed to generate mock data. %v", err)
	}

	type args struct {
		ctx        context.Context
		archive_id int64
		tags       []string
	}
	tests := []struct {
		test     bool
		name     string
		a        *API
		args     args
		wantTags []string
		wantErr  bool
	}{
		{false, "single tag", mockAPI, args{context.Background(), archive_ids[0], []string{"foo"}}, []string{"foo"}, false},
		{false, "multiple tags", mockAPI, args{context.Background(), archive_ids[1], []string{"foo", "bar"}}, []string{"foo", "bar"}, false},
		{false, "ignore duplicate tags", mockAPI, args{context.Background(), archive_ids[2], []string{"foo", "foo"}}, []string{"foo"}, false},
		{true, "ignore invalid tag", mockAPI, args{context.Background(), archive_ids[3], []string{"\n", "\r", "\t", " "}}, nil, false},
	}
	for _, tt := range tests {
		if tt.test {
			t.Run(tt.name, func(t *testing.T) {
				fmt.Printf("Database path: %s\n", tt.a.Config.ArchiveLocation)
				if err := tt.a.AssignTags(tt.args.ctx, tt.args.archive_id, tt.args.tags); (err != nil) != tt.wantErr {
					t.Fatalf("API.AssignTags() error = %v, wantErr %v", err, tt.wantErr)
				}

				got, err := tt.a.GetTags(tt.args.ctx, tt.args.archive_id)
				if err != nil {
					t.Errorf("API.AssignTags()/API.GetTags() error = %v, wantErr %v", err, tt.wantErr)
				}

				if !reflect.DeepEqual(got, tt.wantTags) {
					t.Errorf("API.AssignTags() got %v, want %v", got, tt.wantTags)
				}

			})
		} else {
			fmt.Printf("!!!! SKIPPED TEST '%s' !!!!\n", tt.name)
		}
	}
}

func TestAPI_AssignTagsWithAlias(t *testing.T) {
	mockAPI, err := newMockAPI(Config{ArchiveLocation: ":memory:", ThumbnailLocation: ":memory:"}, t)
	if err != nil {
		t.Fatalf("failed to create mock API. %v", err)
	}
	defer mockAPI.Close(context.Background())

	_, err = GenerateMockData(mockAPI, 3, false, false)
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
					t.Fatalf("API.AssignTags() error = %v, wantErr %v", err, tt.wantErr)
				}
			}

			if err = tt.a.NewTagAlias(tt.args.ctx, tt.wantTag, tt.args.aliasTag); (err != nil) != tt.wantErr {
				t.Fatalf("API.AssignTags()/API.NewTagAlias() error = %v, wantErr %v", err, tt.wantErr)
			}

			if err := tt.a.AssignTags(tt.args.ctx, tt.args.archive_id, []string{tt.args.aliasTag}); err != nil {
				t.Fatalf("API.AssignTags() error = %v", err)
			}

			got, err := tt.a.GetTags(tt.args.ctx, tt.args.archive_id)
			if err != nil {
				t.Errorf("API.AssignTags()/API.GetTags() error = %v", err)
			}

			if got[0] != tt.wantTag {
				t.Errorf("API.AssignTags() got = %v, wantTag %v", tt.args.aliasTag, tt.wantTag)
			}
		})
	}
}

func TestAPI_GetTagsByList(t *testing.T) {
	mockAPI, err := newMockAPI(Config{ArchiveLocation: ":memory:", ThumbnailLocation: ":memory:"}, t)
	if err != nil {
		t.Fatalf("failed to create mock API. %v", err)
	}
	archive_ids, err := GenerateMockData(mockAPI, 2, false, false)
	if err != nil {
		t.Fatalf("failed to generate mock data, %v", err)
	}

	err = mockAPI.AssignTags(context.Background(), archive_ids[0], []string{"foo"})
	if err != nil {
		t.Fatalf("failed to set tag, %v", err)
	}

	err = mockAPI.AssignTags(context.Background(), archive_ids[1], []string{"foo", "bar"})
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

func TestAPI_ReplaceTags(t *testing.T) {
	mockAPI, err := newMockAPI(Config{ArchiveLocation: ":memory:", ThumbnailLocation: ":memory:"}, t)
	if err != nil {
		t.Fatalf("failed to create mock API. %v", err)
	}

	_, err = GenerateMockData(mockAPI, 1, true, false)
	if err != nil {
		t.Fatalf("failed to generate mock data, %v", err)
	}

	type args struct {
		ctx        context.Context
		archive_id int64
		tags       []string
	}
	tests := []struct {
		name    string
		a       *API
		args    args
		wantErr bool
	}{
		{"generic", mockAPI, args{context.Background(), 1, []string{"foo", "bar"}}, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := tt.a.ReplaceTags(tt.args.ctx, tt.args.archive_id, tt.args.tags); (err != nil) != tt.wantErr {
				t.Errorf("API.ReplaceTags() error = %v, wantErr %v", err, tt.wantErr)
			}
			got, err := tt.a.GetTags(tt.args.ctx, tt.args.archive_id)
			if err != nil {
				t.Errorf("API.ReplaceTags()/API.GetTags() error = %v, wantErr %v", err, tt.wantErr)
			}

			if !slices.Equal(tt.args.tags, got) {
				t.Errorf("API.ReplaceTags() got %v, want %v", got, tt.args.tags)
			}
		})
	}
}

func TestAPI_RemoveTags(t *testing.T) {
	mockAPI, err := newMockAPI(Config{ArchiveLocation: ":memory:", ThumbnailLocation: ":memory:"}, t)
	if err != nil {
		t.Fatalf("failed to create mock API. %v", err)
	}

	_, err = GenerateMockData(mockAPI, 5, false, false)
	if err != nil {
		t.Fatalf("failed to create mock data. %v", err)
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
		wantInArchive int64 // skips removing a tag mapped to an archive_id. use 0 to ignore
	}{
		{"remove entire tag (tag is no longer mapped to any archive)", mockAPI, args{context.Background(), []string{"foo"}, []int64{1, 2}}, false, 0},
		{"remove tag for single archive (tag is still mapped to an archive)", mockAPI, args{context.Background(), []string{"bar"}, []int64{3, 4}}, false, 4},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			for _, v := range tt.args.archive_id {
				if err := tt.a.AssignTags(tt.args.ctx, v, tt.args.tags); err != nil {
					t.Errorf("API.RemoveTags()/API.AssignTags() error = %v, wantErr %v", err, tt.wantErr)
				}
			}

			for _, v := range tt.args.archive_id {
				if tt.wantInArchive != v {
					if err := tt.a.RemoveTags(tt.args.ctx, v, tt.args.tags); (err != nil) != tt.wantErr {
						t.Errorf("API.RemoveTags() error = %v, wantErr %v", err, tt.wantErr)
					}
				}
			}

			for _, tag := range tt.args.tags {
				searchRes, err := tt.a.SearchTag(tt.args.ctx, tag)
				if (err != nil) != tt.wantErr {
					t.Errorf("API.RemoveTags() error = %v, wantErr %v", err, tt.wantErr)
				}
				t.Logf("found %v archive(s) for tag %s", searchRes, tag)

				for _, got := range searchRes {
					for _, notWant := range tt.args.archive_id {
						if got == notWant && got != tt.wantInArchive {
							t.Errorf("API.RemoveTags() got archive_id %d in tag search '%s'", got, tag)
						}
					}
				}
			}
		})
	}
}

func TestAPI_RemoveTagsWithAlias(t *testing.T) {
	mockAPI, err := newMockAPI(Config{ArchiveLocation: ":memory:", ThumbnailLocation: ":memory:"}, t)
	if err != nil {
		t.Fatalf("failed to create mock API. %v", err)
	}

	_, err = GenerateMockData(mockAPI, 2, false, false)
	if err != nil {
		t.Fatalf("failed to create mock data. %v", err)
	}

	// assigning to archive_id 1. will be removed from archive_id 2 when test is ran.
	err = mockAPI.archive.AssignTag(context.Background(), 1, "zzz")
	if err != nil {
		t.Fatalf("failed to assign tag. %v", err)
	}

	type args struct {
		ctx        context.Context
		tag        entry.TagAlias
		archive_id int64
	}
	tests := []struct {
		name                string
		a                   *API
		args                args
		wantErr             bool
		wantDeletedTagAlias bool
	}{
		{"full tag deletion", mockAPI, args{context.Background(), entry.TagAlias{BaseTag: "foo", AliasTag: "bar"}, 1}, false, true},
		{"multiple tag_alias references", mockAPI, args{context.Background(), entry.TagAlias{BaseTag: "zzz", AliasTag: "fff"}, 2}, false, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			err := tt.a.AssignTags(tt.args.ctx, tt.args.archive_id, []string{tt.args.tag.BaseTag})
			if err != nil {
				t.Errorf("API.RemoveTags()/API.AssignTags() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			err = tt.a.NewTagAlias(tt.args.ctx, tt.args.tag.BaseTag, tt.args.tag.AliasTag)
			if err != nil {
				t.Errorf("API.RemoveTags()/API.NewTagAlias() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			err = tt.a.RemoveTags(tt.args.ctx, tt.args.archive_id, []string{tt.args.tag.BaseTag})
			if err != nil {
				t.Errorf("API.RemoveTags() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			tag_aliases, err := tt.a.ResolveTagAlias(tt.args.ctx, []string{tt.args.tag.AliasTag})
			if err != nil {
				t.Errorf("API.RemoveTags()/API.ResolveTagAlias() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if len(tag_aliases) != 0 && tt.wantDeletedTagAlias {
				t.Errorf("wanted nil tag_aliases, got %v", tag_aliases)
			}
		})
	}
}
