package api

import (
	"context"
	"log/slog"
	"strconv"
	"strings"

	"github.com/dtbead/moonpool/internal/log"
)

type QueryTags struct {
	TagsInclude, TagsExclude []string
}

// Valid sort options are "imported", "created", and "modified".
// Valid order options are "descending", "ascending".
func (a *API) QueryTags(ctx context.Context, sort, order string, q QueryTags) ([]int64, error) {
	res, err := a.archive.SearchTagByList(ctx, sort, order, q.TagsInclude, q.TagsExclude)
	if err != nil {
		return nil, err
	}

	a.log.LogAttrs(ctx, log.LogLevelVerbose,
		"found "+strconv.Itoa(len(res))+" archive_id's",
		slog.Any("query_tags", q))

	return res, nil
}

// BuildQuery takes a string and returns a QueryTag to be used in API.QueryTags.
// Each tag can be separated by a comma (,) and have a prefixed special character to
// assign specific tag modifiers to it.
//
// tag modifiers:
// dash (-) = exclude tag from search
func BuildQuery(s string) QueryTags {
	var tagsInclude, tagsExclude []string

	tags := strings.Split(s, ",")
	for _, tag := range tags {
		if tag == "" {
			break
		}

		tag = strings.TrimSpace(tag)

		if strings.HasPrefix(tag, "-") {
			tagsExclude = append(tagsExclude, tag[len("-"):])
		} else {
			tagsInclude = append(tagsInclude, tag)
		}
	}

	return QueryTags{tagsInclude, tagsExclude}
}

// SearchHash takes a hexadecimal string of either md5, sha1, or sha256, and returns an archive_id.
// hash can be upper or lowercase.
func (a API) SearchHash(ctx context.Context, hash string) (archive_id int64, err error) {
	return a.archive.SearchHash(ctx, hash)
}
