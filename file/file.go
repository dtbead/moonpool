package file

import (
	"bufio"
	"crypto/md5"
	"crypto/sha1"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"os"
	"path"
	"path/filepath"
	"sync"
	"time"

	"github.com/dtbead/moonpool/media"
)

func CopyAndHash(destination, extension string, r io.Reader) (media.Entry, error) {
	var e media.Entry
	var wg sync.WaitGroup

	type res struct {
		hash       media.Hashes
		hashString string
		err        error
	}

	dataChan := make(chan res)

	wg.Add(1)
	go func() {
		defer wg.Done()
		h, err := GetHashes(r)

		dataChan <- res{
			hash:       h,
			hashString: ByteToString(h.MD5),
			err:        err,
		}
	}()

	tmp := <-dataChan
	if tmp.err != nil {
		return media.Entry{}, tmp.err
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
		err := copy(destination, r)
		dataChan <- res{
			err: err,
		}
	}()

	tmp = <-dataChan
	if tmp.err != nil {
		return media.Entry{}, tmp.err
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

	_, err = io.Copy(file, r)
	if err != nil {
		return err
	}

	return err
}

func GetHashes(r io.Reader) (media.Hashes, error) {
	md5 := md5.New()
	sha1 := sha1.New()
	sha256 := sha256.New()

	pagesize := os.Getpagesize()

	reader := bufio.NewReaderSize(r, pagesize)
	multiWriter := io.MultiWriter(md5, sha1, sha256)

	_, err := io.Copy(multiWriter, reader)
	if err != nil {
		return media.Hashes{}, err
	}

	return media.Hashes{
		MD5:    md5.Sum(nil),
		SHA1:   sha1.Sum(nil),
		SHA256: sha256.Sum(nil),
	}, nil
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

func GetTimestamp(f *os.File) (media.Timestamp, error) {
	fi, err := f.Stat()
	if err != nil {
		return media.Timestamp{}, err
	}

	return media.Timestamp{DateCreated: fi.ModTime(), DateImported: time.Now()}, nil
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
	_, err := os.Stat(path)
	if err == nil {
		return true
	}

	if os.IsNotExist(err) {
		return false
	}

	return false
}

func trimTrailingSlashes(s string) string {
	if s[len(s)-1] == '\\' || s[len(s)-1] == '/' {
		return s[:len(s)-1]
	}
	return s
}
