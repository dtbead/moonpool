package api

import (
	"strings"
)

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
