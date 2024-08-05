package api

import (
	"errors"
	"fmt"
	"regexp"
	"strings"
)

const (
	SEARCH_PREDICATE_OR    = '~'
	SEARCH_PREDICATE_NOT   = '-'
	SEARCH_PREDICATE_FUZZY = '*'

	SEARCH_SYSTEM_RANGE = "range:"
	SEARCH_SYSTEM_LIMIT = "limit:"
)

var regexSystemLimit = regexp.MustCompile(`^limit: *(\d+)$`)
var regexSystemRange = regexp.MustCompile(`^range: *(\d+-\d+)$`)

const preliminarySearchStatement = `
	SELECT archive.id, tags.tag_id, tags.text FROM tags 
	INNER JOIN tagmap ON tagmap.tag_id = tags.tag_id
	INNER JOIN archive ON archive.id = tagmap.archive_id WHERE `

type search struct {
	query     string
	predicate rune
}

func convertPredicateToSQLKeyword(r rune) (string, error) {
	switch r {
	default:
		return "", errors.New("invalid search predicate")
	case SEARCH_PREDICATE_OR:
		return "OR", nil
	case SEARCH_PREDICATE_NOT:
		return "NOT", nil
	case SEARCH_PREDICATE_FUZZY:
		return "LIKE", nil
	}
}

func convertSystemTagsToSQLKeyword(s string) (string, error) {
	switch {
	case strings.Contains(s, SEARCH_SYSTEM_LIMIT):
		if regexSystemLimit.FindAllStringSubmatch(s, 1) != nil {
			return regexSystemLimit.FindAllStringSubmatch(s, 1)[0][1], nil
		}

	case strings.Contains(s, SEARCH_SYSTEM_RANGE):
		if regexSystemRange.FindAllStringSubmatch(s, 1) != nil {
			return regexSystemRange.FindAllStringSubmatch(s, 1)[0][1], nil
		}
	}

	return "", nil
}

func parsePredicate(s []string) []search {
	var q []search
	var sq search
	for _, str := range s {
		if str == "" {
			continue
		}

		switch []rune(str)[0] {
		default:
			sq.predicate = 0
		case SEARCH_PREDICATE_OR:
			sq.predicate = SEARCH_PREDICATE_OR

		case SEARCH_PREDICATE_NOT:
			sq.predicate = SEARCH_PREDICATE_NOT

		case SEARCH_PREDICATE_FUZZY:
			sq.predicate = SEARCH_PREDICATE_FUZZY
		}
		q = append(q, sq)

		sq.predicate = 0
		sq.query = ""
	}

	return q
}

func buildSearchQuery(queries []search) (string, error) {
	stmt := preliminarySearchStatement
	for _, query := range queries {
		p, err := convertPredicateToSQLKeyword(query.predicate)
		if err != nil {
			return "", err
		}
		stmt += fmt.Sprintf("tags.text %s (?)", p)
	}

	return stmt, nil
}
