package file

import (
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"io"
	"os"
	"path"
	"path/filepath"
	"time"

	"github.com/dtbead/moonpool/tags"
)

type Entry struct {
	File     os.File
	Metadata EntryMetadata
	Tags     []tags.Tag
}

type EntryMetadata struct {
	MD5Hash   []byte
	Timestamp Timestamp
	Path      string
}

type Timestamp struct {
	DateCreated  time.Time
	DateModified time.Time
	DateImported time.Time
}

func CopyFile(destination string, f *os.File) error {
	if !doesPathExist(filepath.Dir(destination)) {
		if err := os.MkdirAll(filepath.Dir(destination), 0664); err != nil {
			return err
		}
	}

	dst, err := os.Create(destination)
	if err != nil {
		return err
	}

	if _, err := io.Copy(dst, f); err != nil {
		return err
	}

	return nil
}

// TODO: this func shouldn't load the entire file into memory
func GetMD5Hash(f *os.File) []byte {
	buf := []byte{}
	f.Read(buf)

	h := md5.New()
	sum := h.Sum(buf)

	return sum
}

func ByteToString(h []byte) string {
	return hex.EncodeToString(h)
}

// BuildFilePath builds a path to store imported media. The first 2 characters of HashString
// will be used to create a string such as "/f1/f15f38b5cfdbfd56aeb6da48b65d3d6f.png"
// to organize media for quicker lookup
func BuildFilePath(RootPath, HashString, extension string) string {
	r := []rune(HashString)
	return filepath.Clean(fmt.Sprintf("%s/%s/%s%s", RootPath, string(r[0:2]), HashString, extension))
}

func GetTimestamp(f *os.File) (Timestamp, error) {
	fi, err := f.Stat()
	if err != nil {
		return Timestamp{}, err
	}

	return Timestamp{DateCreated: fi.ModTime(), DateImported: time.Now()}, nil
}

// NewStorage creates a base directory to store imported files
func NewStorage(rootPath string) error {
	p := fmt.Sprintf("%s/db/media/storage", rootPath)
	p = path.Clean(p)
	if err := os.MkdirAll(path.Clean(p), os.ModePerm); err != nil {
		return err
	}
	return nil
}

func doesPathExist(path string) bool {
	if stat, err := os.Stat(path); err == nil && stat.IsDir() {
		return true
	}
	return false
}
