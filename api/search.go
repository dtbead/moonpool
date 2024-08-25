package api

import (
	"context"
	"fmt"
	"log/slog"

	archive "github.com/dtbead/moonpool/db"
	"github.com/dtbead/moonpool/log"
)

const (
	SEARCH_PREDICATE_NONE = 0
	SEARCH_PREDICATE_NOT  = 1
	SEARCH_PREDICATE_OR   = 2

	sqlSearchPreliminary = `SELECT archive.id, tags.tag_id, tags.text FROM tags 
	INNER JOIN tagmap ON tagmap.tag_id = tags.tag_id
	INNER JOIN archive ON archive.id = tagmap.archive_id`
)

type SearchQuery struct {
	s []query
}

type query struct {
	Tag       []string
	Predicate int
}

type queryResult struct {
	ArchiveID, TagID int64
	Tag              string
}

// Add adds a multiple tags together to search for with no special predicates
func (q *SearchQuery) Add(tag []string) {
	q.s = append(q.s, query{Tag: tag, Predicate: SEARCH_PREDICATE_NONE})
}

func (q *SearchQuery) AddWithPredicate(tag string, predicate int) {
	q.s = append(q.s, query{Tag: []string{tag}, Predicate: predicate})
}

func (a API) Query(ctx context.Context, q SearchQuery) ([]archive.EntryTags, error) {
	sqlStmt, err := buildQuery(q)
	if err != nil {
		return nil, err
	}

	a.log.LogAttrs(context.Background(), log.LogLevelVerbose, "built complex search query", slog.String("sql_query", sqlStmt))

	res, err := a.db.QueryContext(ctx, sqlStmt)
	if err != nil {
		a.log.LogAttrs(context.Background(), log.LogLevelError, "failed to execute complex search queryin", slog.String("sql_query", sqlStmt), slog.Any("error", err))
		return nil, err
	}
	defer res.Close()

	var searchRes []archive.EntryTags
	var archiveID, tagID int64
	var tagText string
	for res.Next() {
		if err := res.Scan(&archiveID, &tagID, &tagText); err != nil {
			a.log.LogAttrs(context.Background(), log.LogLevelError, "failed to read result from db to memory", slog.String("sql_query", sqlStmt), slog.Any("error", err))
			return nil, err
		}
		searchRes = append(searchRes, archive.EntryTags{
			ArchiveID: archiveID,
			Tags: []archive.Tag{
				{
					TagID: int(tagID),
					Text:  tagText,
				},
			},
		})
		a.log.LogAttrs(context.Background(), log.LogLevelVerbose, "found result '%s' from question custom SQL query",
			slog.Any("result", searchRes),
			slog.String("sql_query", sqlStmt))

	}

	return searchRes, nil
}

func buildQuery(q SearchQuery) (string, error) {
	var tagsGeneral []string
	var tagsNot []string
	for _, v := range q.s {
		switch v.Predicate {
		default:
			tagsGeneral = append(tagsGeneral, v.Tag...)
		case SEARCH_PREDICATE_NOT:
			tagsNot = append(tagsNot, v.Tag...)
		}
	}

	query := fmt.Sprintf("%s WHERE %s AND %s;", sqlSearchPreliminary, buildGeneralTagQuery(tagsGeneral), buildNotTagQuery(tagsNot))
	return query, nil
}

func buildNotTagQuery(s []string) string {
	generalSearch := ` tags.text NOT IN (`
	for i, v := range s {
		if i >= len(s)-1 {
			generalSearch += fmt.Sprintf("'%s')", v)
		} else {
			generalSearch += fmt.Sprintf("'%s', ", v)
		}
	}

	return deleteWhitespace(generalSearch)
}

func buildOrTagQuery(s []string) string {
	generalSearch := ` tags.text IN (`
	for i, v := range s {
		if i >= len(s)-1 {
			generalSearch += fmt.Sprintf("'%s') GROUP BY tags.tag_id;", v)
		} else {
			generalSearch += fmt.Sprintf("'%s', ", v)
		}
	}

	return deleteWhitespace(generalSearch)
}

// buildGeneralSearch builds a search query with tags that do not have
// any special modifiers or predicates
func buildGeneralTagQuery(s []string) string {
	generalSearch := " tags.text IN ("
	for i, v := range s {
		if i+1 == len(s) {
			generalSearch += fmt.Sprintf("'%s')", v)
		} else {
			generalSearch += fmt.Sprintf("'%s', ", v)
		}
	}

	return deleteWhitespace(generalSearch)
}

func trimStringIndex(s string, index int) string {
	return string([]rune(s)[index:])
}

func getPredicate(r rune) int {
	switch r {
	default:
		return SEARCH_PREDICATE_NONE
	case '-':
		return SEARCH_PREDICATE_NOT
	case '~':
		return SEARCH_PREDICATE_OR
	}
}
