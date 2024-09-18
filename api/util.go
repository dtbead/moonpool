package api

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"path"
	"regexp"
	"strings"
)

func byteToHex(b []byte) string {
	return hex.EncodeToString(b)
}

func deleteWhitespace(s string) string {
	whitespace, _ := regexp.Compile(`[\t\r\n]+`)
	excessSpaces, _ := regexp.Compile(`[ ]{2,}`)

	s = whitespace.ReplaceAllString(s, " ")
	s = excessSpaces.ReplaceAllString(s, " ")
	return s
}

func randomString(length int) string {
	b := make([]byte, length+2)
	rand.Read(b)
	return fmt.Sprintf("%x", b)[2 : length+2]
}

func trimIndex(i int, s string) string {
	return string([]rune(s)[i:])
}

// cleanPath() cleans a filepath by replacing all instances of '\' with '/'
// and calling func path.Clean()
func cleanPath(s string) string {
	return path.Clean(strings.ReplaceAll(s, `\`, `/`))
}

func removeDuplicateStr(strSlice []string) []string {
	allKeys := make(map[string]bool)
	list := []string{}
	for _, item := range strSlice {
		if _, value := allKeys[item]; !value {
			allKeys[item] = true
			list = append(list, item)
		}
	}
	return list
}

// cleanTag() removes any excess whitespace, blacklisted characters, etc from a given string.
func cleanTag(s string) string {
	r := []rune(s)
	NewString := make([]rune, len(r))

	for _, v := range r {
		if v != ' ' {
			NewString = append(NewString, v)
		}
	}

	return strings.TrimSpace(string(NewString))
}
