package db

import (
	"errors"
	"os"
	"strings"
)

func DoesFileExist(s string) bool {
	if _, err := os.Stat(s); err == nil {
		return true
	} else if errors.Is(err, os.ErrNotExist) {
		return false
	} else {
		return false
	}
}

// sanitizes a string for protection against SQL injections
func sanitizeString(s string) string {
	return strings.ReplaceAll(s, "'", "")
}

func columnsToMap(c []string) map[string]string {
	m := make(map[string]string, len(c))
	for _, v := range c {
		m[v] = ""
	}

	return m
}
