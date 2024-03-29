package file

import (
	"bytes"
	"crypto/md5"
	"crypto/sha1"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"os"
	"path"
	"path/filepath"
	"sync"
	"time"
)

type Entry struct {
	File     os.File
	Metadata Metadata
	Tags     []string
}

type Metadata struct {
	MD5Hash      string
	Hash         Hashes
	Timestamp    Timestamp
	PathDirect   string
	PathRelative string
	Extension    string
}

type Hashes struct {
	MD5    []byte
	SHA1   []byte
	SHA256 []byte
}

type Timestamp struct {
	DateCreated  time.Time
	DateModified time.Time
	DateImported time.Time
}

func CopyAndHash(destination, extension string, r io.Reader) (Entry, error) {
	var e Entry
	var buf bytes.Buffer
	var wg sync.WaitGroup

	type res struct {
		hash       Hashes
		hashString string
		err        error
	}

	tee := io.TeeReader(r, &buf)

	dataChan := make(chan res)

	wg.Add(1)
	go func() {
		defer wg.Done()
		h, err := GetHashes(tee)

		dataChan <- res{
			hash:       h,
			hashString: ByteToString(h.MD5),
			err:        err,
		}
	}()
	tmp := <-dataChan
	if tmp.err != nil {
		return Entry{}, tmp.err
	}

	path := BuildPath(tmp.hashString, extension)
	destination = trimTrailingSlashes(destination) + "/" + path

	e.Metadata.PathDirect = destination
	e.Metadata.PathRelative = path
	e.Metadata.Extension = extension
	e.Metadata.MD5Hash = string(tmp.hashString)
	e.Metadata.Hash.MD5 = tmp.hash.MD5
	e.Metadata.Hash.SHA1 = tmp.hash.SHA1
	e.Metadata.Hash.SHA256 = tmp.hash.SHA256
	e.Metadata.Timestamp.DateImported = time.Now()

	wg.Add(1)
	go func() {
		defer wg.Done()
		err := copy(destination, tee)
		dataChan <- res{
			err: err,
		}
	}()

	tmp = <-dataChan
	if tmp.err != nil {
		return Entry{}, tmp.err
	}

	return e, nil
}

func copy(destination string, r io.Reader) error {
	if !doesPathExist(filepath.Dir(destination)) {
		if err := os.MkdirAll(filepath.Dir(destination), 0664); err != nil {
			return err
		}
	}

	file, err := os.Create(destination)
	if err != nil {
		return err
	}
	defer file.Close()

	written, err := io.Copy(file, r)
	if err != nil {
		return err
	}

	fmt.Println(written)
	return err
}

func GetHashes(r io.Reader) (Hashes, error) {
	var h Hashes
	h.MD5 = Hash("md5", r)
	h.SHA1 = Hash("sha1", r)
	h.SHA256 = Hash("sha256", r)

	if h.MD5 == nil || h.SHA1 == nil || h.SHA256 == nil {
		return Hashes{}, errors.New("unable to calculate hash digest")
	}

	return h, nil
}

// valid methods are "md5", "sha1", and "sha256". returns nil if given invalid HashType.
func Hash(hashType string, r io.Reader) []byte {
	switch hashType {
	case "md5":
		buf := []byte{}
		r.Read(buf)

		h := md5.New()
		sum := h.Sum(buf)

		return sum
	case "sha1":
		buf := []byte{}
		r.Read(buf)

		h := sha1.New()
		sum := h.Sum(buf)

		return sum
	case "sha256":
		buf := []byte{}
		r.Read(buf)

		h := sha256.New()
		sum := h.Sum(buf)

		return sum
	}

	return nil
}

func ByteToString(h []byte) string {
	return hex.EncodeToString(h)
}

// BuildPath builds a path to store imported media. The first 2 characters of HashString
// will be used to create a string such as "f1/f15f38b5cfdbfd56aeb6da48b65d3d6f.png"
// for quicker file lookup. BuildPath expects the caller to include a period for its extension variable
func BuildPath(HashString, extension string) string {
	h := []rune(HashString)
	return fmt.Sprintf("%s/%s%s", string(h[0:2]), HashString, extension)
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

func trimTrailingSlashes(s string) string {
	if s[len(s)-1] == '\\' || s[len(s)-1] == '/' {
		return s[:len(s)-1]
	}
	return s
}
