package file

import (
	"bufio"
	"crypto/md5"
	"crypto/sha1"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"hash"
	"image"
	"io"
	"mime"
	"os"
	"path"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/corona10/goimagehash"
)

var hashPool sync.Pool

type hasher struct {
	buf    bufio.Reader
	md5    hash.Hash
	sha1   hash.Hash
	sha256 hash.Hash
}

func (h *hasher) Reset() {
	h.md5.Reset()
	h.sha1.Reset()
	h.sha256.Reset()
}

func init() {
	hashPool = sync.Pool{
		New: func() any {
			h := hasher{}
			h.md5 = md5.New()
			h.sha1 = sha1.New()
			h.sha256 = sha256.New()
			return &h
		},
	}
}

func getHashpool() *hasher {
	h := hashPool.Get()
	if h != nil {
		return h.(*hasher)
	}
	return nil
}

func putHashPool(h hasher) {
	h.Reset()
	hashPool.Put(&h)
}

type Hashes struct {
	MD5    []byte
	SHA1   []byte
	SHA256 []byte
}

func (h Hashes) string() hashesString {
	return hashesString{
		MD5:    hex.EncodeToString(h.MD5),
		SHA1:   hex.EncodeToString(h.SHA1),
		SHA256: hex.EncodeToString(h.SHA256),
	}
}

type hashesString struct {
	MD5, SHA1, SHA256 string
}

type PerceptualHashes struct {
	Type string
	Hash uint64
}

func Copy(destination string, r io.Reader) error {
	baseDirectory := filepath.Dir(destination)
	if !DoesPathExist(baseDirectory) {
		if err := os.MkdirAll(baseDirectory, os.ModePerm); err != nil {
			return err
		}
	}

	file, err := os.Create(destination)
	if err != nil {
		return err
	}
	defer file.Close()

	buf := bufio.NewReader(r)

	w, err := buf.WriteTo(file)
	if err != nil {
		return err
	}

	if w <= 0 {
		return errors.New("copied 0 bytes")
	}

	return nil
}

// GetHash returns the MD5, SHA1, SHA256 hash and total bytes read of a given io.Reader.
func GetHash(r io.Reader) (hash Hashes, read int64, err error) {
	hashPool := getHashpool()
	hashPool.buf.Reset(r)

	mw := io.MultiWriter(hashPool.md5, hashPool.sha1, hashPool.sha256)
	bytesRead, err := io.Copy(mw, &hashPool.buf)
	if err != nil {
		return Hashes{}, -1, err
	}

	h := Hashes{
		MD5:    hashPool.md5.Sum(nil),
		SHA1:   hashPool.sha1.Sum(nil),
		SHA256: hashPool.sha256.Sum(nil),
	}

	putHashPool(*hashPool)
	return h, bytesRead, nil
}

func GetPerceptualHash(i image.Image) (PerceptualHashes, error) {
	hash, err := goimagehash.PerceptionHash(i)
	if err != nil {
		return PerceptualHashes{}, err
	}

	ph := PerceptualHashes{}

	switch hash.GetKind() {
	case 0:
		ph.Type = "unknown"
	case 1:
		ph.Type = "AHash"
	case 2:
		ph.Type = "PHash"
	case 3:
		ph.Type = "DHash"
	case 4:
		ph.Type = "WHash"
	}

	ph.Hash = hash.GetHash()
	return ph, nil

}

func ByteToHexString(h []byte) string {
	return hex.EncodeToString(h)
}

// BuildPath builds a path to store media. md5 gets encoded to a hexidecimal string
// to create a storage path such as "f1/f15f38b5cfdbfd56aeb6da48b65d3d6f.png".
// BuildPath expects an extension to have a period prefix already added by caller
func BuildPath(md5 []byte, extension string) string {
	return fmt.Sprintf("%s/%s%s", string(ByteToHexString(md5[:1])), string(ByteToHexString(md5[:])), extension)
}

// DateModified returns the UTC time of the date modified on a file
func DateModified(f *os.File) (time.Time, error) {
	fi, err := f.Stat()
	if err != nil {
		return time.Time{}, err
	}

	return fi.ModTime().UTC(), nil
}

// NewStorage creates a new directory to store media
func NewStorage(rootPath string) error {
	if err := os.MkdirAll(path.Clean(fmt.Sprintf("%s/db/media/storage", rootPath)), os.ModePerm); err != nil {
		return err
	}
	return nil
}

func DoesPathExist(path string) bool {
	_, err := os.Stat(path)
	if err == nil {
		return true
	}

	if os.IsNotExist(err) {
		return false
	}

	return false
}

func IsDirectoryEmpty(name string) bool {
	f, err := os.Open(name)
	if err != nil {
		return false
	}
	defer f.Close()

	_, err = f.Readdirnames(1)
	return err == io.EOF
}

// CleanPath cleans a filepath by replacing all instances of '\' with '/'
// and calling func path.Clean
func CleanPath(s string) string {
	return path.Clean(strings.ReplaceAll(s, `\`, `/`))
}

// GetMimeTypeByExtension returns the mime type according to a given extension. It returns
// a blank string if no type is found.
func GetMimeTypeByExtension(ext string) string {
	return mime.TypeByExtension(ext)
}

func unixTimeToWindowsTicks(unix uint64) uint64 {
	return (unix * 10000000) + 116444736000000000
}

// GetSize returns the size of a given io.Reader.
func GetSize(r io.Reader) (bytes int64, err error) {
	f, ok := r.(*os.File)
	if ok {
		st, err := f.Stat()
		if err != nil {
			return -1, err
		}

		return st.Size(), nil
	}

	n, err := io.Copy(io.Discard, r)
	if err != nil {
		return -1, err
	}

	return n, nil
}
