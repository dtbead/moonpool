package api

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"path"
	"strconv"
	"strings"
	"time"
)

func byteToHex(b []byte) string {
	return hex.EncodeToString(b)
}

func hexToByte(s string) []byte {
	h, _ := hex.DecodeString(s)
	return h
}

func randomString(length int) string {
	b := make([]byte, length+2)
	rand.Read(b)
	return fmt.Sprintf("%x", b)[2 : length+2]
}

// cleanPath cleans a filepath by replacing all instances of '\' with '/'
// and calling func path.Clean
func cleanPath(s string) string {
	p := path.Clean(strings.ReplaceAll(s, `\`, `/`))
	if p == "." {
		return ""
	}
	return p
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

// timeToRFC3339_UTC returns a RFC3339 string-formatted timestamp in UTC timezone.
func timeToRFC3339_UTC(t time.Time) string {
	return t.UTC().Format(time.RFC3339)
}

func int64ToString(i int64) string {
	return strconv.FormatInt(i, 10)
}
