package api

import (
	"context"
	"fmt"
	"log/slog"
	"strings"

	"github.com/dtbead/moonpool/internal/log"
)

type SearchQuery []string

const (
	SEARCH_PREDICATE_NONE rune = 0
	SEARCH_PREDICATE_NOT  rune = '-'
	SEARCH_PREDICATE_OR   rune = '~'

	sqlSearchPreliminary = `SELECT archive.id FROM tags 
	INNER JOIN tag_map ON tag_map.tag_id = tags.tag_id
	INNER JOIN archive ON archive.id = tag_map.archive_id`
	sqlEpilogue = `SELECT id FROM predicate_none UNION SELECT archive FROM predicate_or WHERE id NOT IN predicate_not;`

	sqlSearchPredicateNONE_Prologue = `predicate_none AS (` + sqlSearchPreliminary + ` WHERE tags.text IS (`
	sqlSearchPredicateNOT_Prologue  = `predicate_not AS (` + sqlSearchPreliminary + ` WHERE tags.text IS (`
	sqlSearchPredicateOR_Prologue   = `predicat_or AS (` + sqlSearchPreliminary + ` WHERE tags.text IS (`

	sqlSearchPredicateNOT_Epilogue = `) AND archive.id IN predicate_none GROUP BY archive.id)`
	sqlSearchPredicateOR_Epilogue  = `) GROUP BY archive.id HAVING COUNT(archive.id) = %d)"`
)

// Add() adds a multiple tags together to search for with no special predicates
func (q *SearchQuery) Add(tag []string) {
	*q = append(*q, tag...)
}

func (a API) Query(ctx context.Context, q SearchQuery) ([]int64, error) {
	sqlStmt, sqlBindValues := buildQuery(q)
	a.log.LogAttrs(ctx, log.LogLevelVerbose, "built complex search query", slog.String("sql_query", sqlStmt), slog.Any("sql_paramaters", sqlBindValues))

	res, err := a.db.QueryContext(ctx, sqlStmt, stringSliceToInterface(sqlBindValues)...)
	if err != nil {
		a.log.LogAttrs(ctx, log.LogLevelError, "failed to execute complex search queryin", slog.String("sql_query", sqlStmt), slog.Any("error", err))
		return nil, err
	}
	defer res.Close()

	var archiveIDs []int64
	var archiveID int64 = -1
	for res.Next() {
		if err := res.Scan(&archiveID); err != nil {
			a.log.LogAttrs(ctx, log.LogLevelError, "failed to read result from db to memory", slog.String("sql_query", sqlStmt), slog.Any("error", err))
			return nil, err
		}

		if archiveID != -1 {
			archiveIDs = append(archiveIDs, archiveID)
			archiveID = -1
		}

	}
	a.log.LogAttrs(ctx, log.LogLevelVerbose, fmt.Sprintf("found '%v' archive_id(s) from custom SQL query", archiveIDs),
		slog.Any("archive_ids", archiveIDs))

	return archiveIDs, nil
}

/*
buildQuery() crafts an array of SQLite common table expression statements that corresponds to each SEARCH_PREDICATE_*.
buildQuery() will ALWAYS return a non-empty array of CTE's, even if said predicate is not included in the SearchQuery.

# The CTE's are in the following order:
  - 0. predicate_none AS (...)
  - 1. predicate not AS (...)
  - 2. predicate_or AS (...)

Each CTE does NOT contain the actual string content to query for, but rather uses bind values/parameters.
*/
func buildQuery(q SearchQuery) (string, []string) {
	var tagsGeneral []string
	var tagsNot []string
	var tagsOr []string
	for _, tag := range q {
		switch []rune(tag)[0] {
		default:
			tagsGeneral = append(tagsGeneral, tag)
		case SEARCH_PREDICATE_NOT:
			tagsNot = append(tagsNot, trimIndex(1, tag))
		case SEARCH_PREDICATE_OR:
			tagsOr = append(tagsOr, trimIndex(1, tag))
		}
	}

	var g, n, o string
	switch {
	case tagsGeneral != nil:
		g = buildPredicate(SEARCH_PREDICATE_NONE, tagsGeneral)
		fallthrough
	case tagsNot != nil:
		n = buildPredicate(SEARCH_PREDICATE_NOT, tagsNot)
		fallthrough
	case tagsOr != nil:
		o = buildPredicate(SEARCH_PREDICATE_OR, tagsOr)
	}

	var sqlBindValues = make([]string, 0, len(tagsGeneral)+len(tagsNot)+len(tagsOr))
	sqlBindValues = append(sqlBindValues, tagsGeneral...)
	sqlBindValues = append(sqlBindValues, tagsNot...)
	sqlBindValues = append(sqlBindValues, tagsOr...)

	sqlStmt := combineCTE([3]string{g, n, o})

	return deleteWhitespace(sqlStmt), sqlBindValues
}

func combineCTE(c [3]string) string {
	var s string = "WITH "
	var epilogue string = "SELECT id FROM predicate_none"

	totalPredicates := 0
	for _, v := range c {
		if v != "" {
			totalPredicates++
		}
	}

	for i, v := range c {
		if v != "" {
			switch i {
			case 0:
				if totalPredicates <= 1 {
					s += v + " "
				} else {
					s += v + ", "
				}
			case 1:
				s += v + " "
				epilogue += " WHERE id NOT IN predicate_not;"
			}
		}
	}

	return s + epilogue
}

func sqlSearchPredicateNONE_Epilogue(tags int) string {
	return fmt.Sprintf(") GROUP BY archive.id HAVING COUNT(archive.id) = %d)", tags)
}

// buildPredicate() takes a predicate and a slice of tags, and creates an SQL query
// with the same amount of SQL placeholders as tags
func buildPredicate(predicate rune, tags []string) string {
	if len(tags) == 0 {
		return ""
	}

	var str, end string
	switch predicate {
	case SEARCH_PREDICATE_NOT:
		str = sqlSearchPredicateNOT_Prologue
		end = sqlSearchPredicateNOT_Epilogue
	case SEARCH_PREDICATE_OR:
		str = sqlSearchPredicateOR_Prologue
		end = sqlSearchPredicateOR_Epilogue
	default:
		str = sqlSearchPredicateNONE_Prologue
		end = sqlSearchPredicateNONE_Epilogue(len(tags))

	}

	for i := range tags {
		if i == len(tags)-1 {
			str += "?" + end
		} else {
			str += "?, "
		}
	}

	return str
}

func (q SearchQuery) ToInterface() []interface{} {
	return stringSliceToInterface(q)
}

// NewSearchQuery() takes a string with comma separated tags and returns a new SearchQuery
func NewSearchQuery(s string) SearchQuery {
	return strings.Split(s, ",")
}

func stringSliceToInterface(s []string) []interface{} {
	intrf := make([]interface{}, len(s))
	for i, v := range s {
		intrf[i] = v
	}

	return intrf
}
