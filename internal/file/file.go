package file

import (
	"bufio"
	"crypto/md5"
	"crypto/sha1"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"image"
	"io"
	"os"
	"path"
	"path/filepath"
	"strings"
	"time"

	"github.com/corona10/goimagehash"
)

type Hashes struct {
	MD5    []byte
	SHA1   []byte
	SHA256 []byte
}

type PerceptualHashes struct {
	Type string
	Hash uint64
}

func CopyAndHash(baseDirectory, extension string, r io.Reader) (Hashes, error) {
	h, err := GetHash(r)
	if err != nil {
		return Hashes{}, err
	}

	path := BuildPath(h.MD5, extension)
	destination := baseDirectory + "/" + path

	if !DoesPathExist(destination) {
		if err := os.MkdirAll(filepath.Dir(destination), 0664); err != nil {
			return Hashes{}, err
		}
	}

	dest, err := os.Create(destination)
	if err != nil {
		return Hashes{}, err
	}
	defer dest.Close()

	_, err = io.Copy(dest, r)
	if err != nil {
		return Hashes{}, err
	}

	return h, nil
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

func GetHash(r io.Reader) (Hashes, error) {
	reader := bufio.NewReader(r)

	md5 := md5.New()
	sha1 := sha1.New()
	sha256 := sha256.New()

	mw := io.MultiWriter(md5, sha1, sha256)
	_, err := io.Copy(mw, reader)
	if err != nil {
		return Hashes{}, err
	}

	return Hashes{
		MD5:    md5.Sum(nil),
		SHA1:   sha1.Sum(nil),
		SHA256: sha256.Sum(nil),
	}, nil
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

func unixTimeToWindowsTicks(unix uint64) uint64 {
	return (unix * 10000000) + 116444736000000000
}
