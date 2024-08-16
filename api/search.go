package api

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/dtbead/moonpool/archive"
	"github.com/dtbead/moonpool/log"
)

const (
	sqlSearchPreliminary = `SELECT archive.id, tags.tag_id, tags.text FROM tags 
	INNER JOIN tagmap ON tagmap.tag_id = tags.tag_id
	INNER JOIN archive ON archive.id = tagmap.archive_id`
)

type SearchQuery []string
type queryResult struct {
	ArchiveID, TagID int64
	Tag              string
}

func getPredicate(r rune) string {
	switch r {
	default:
		return ""
	case '-':
		return "NOT"
	case '~':
		return "OR"
	}
}

func (a API) Query(ctx context.Context, q SearchQuery) ([]archive.EntryTags, error) {
	sqlStmt, err := buildQuery(q)
	if err != nil {
		return nil, err
	}

	a.log.LogAttrs(context.Background(), log.LogLevelVerbose, "built custom SQL query", slog.String("sql_query", sqlStmt))

	res, err := a.db.QueryContext(ctx, sqlStmt)
	if err != nil {
		a.log.LogAttrs(context.Background(), log.LogLevelError, "failed to execute custom SQL query", slog.String("sql_query", sqlStmt), slog.Any("error", err))
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
	var v SearchQuery
	var generalTags SearchQuery

	sqlQuery := sqlSearchPreliminary
	for _, query := range q {
		switch getPredicate([]rune(query)[0]) {
		default:
			generalTags = append(generalTags, query)
		case "OR":
		case "NOT":
			v = append(v, sqlQuery+fmt.Sprintf("\nWHERE (tags.text NOT = '%s')", trimStringIndex(query, 1)))
		}

	}

	return buildGeneralTagQuery(generalTags), nil
}

// buildGeneralSearch builds a search query with tags that do not have
// any special modifiers or predicates
func buildGeneralTagQuery(s []string) string {
	generalSearch := sqlSearchPreliminary + " WHERE tags.text IN ("
	for i, v := range s {
		if i >= len(s)-1 {
			generalSearch += fmt.Sprintf("'%s');", v)
		} else {
			generalSearch += fmt.Sprintf("'%s', ", v)
		}
	}

	return deleteWhitespace(generalSearch)
}

func trimStringIndex(s string, index int) string {
	return string([]rune(s)[index:])
}
