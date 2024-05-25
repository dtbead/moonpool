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
	"time"
)

type Hashes struct {
	MD5    []byte
	SHA1   []byte
	SHA256 []byte
}

func CopyAndHash(baseDirectory, extension string, r io.Reader) (Hashes, error) {
	h, err := GetHash(r)
	if err != nil {
		return Hashes{}, err
	}

	path := BuildPath(h.MD5, extension)
	destination := baseDirectory + "/" + path

	if !doesPathExist(destination) {
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

func byteToHexString(h []byte) string {
	return hex.EncodeToString(h)
}

// BuildPath builds a path to store media. md5 gets encoded to a hexidecimal string
// to create a storage path such as "f1/f15f38b5cfdbfd56aeb6da48b65d3d6f.png".
// BuildPath expects an extension to have a period prefix already added by caller
func BuildPath(md5 []byte, extension string) string {
	return fmt.Sprintf("%s/%s%s", string(byteToHexString(md5[:1])), string(byteToHexString(md5[:])), extension)
}

func GetDateModified(f *os.File) (time.Time, error) {
	fi, err := f.Stat()
	if err != nil {
		return time.Time{}, err
	}

	return fi.ModTime(), nil
}

// NewStorage creates a directory to store media
func NewStorage(rootPath string) error {
	if err := os.MkdirAll(path.Clean(fmt.Sprintf("%s/db/media/storage", rootPath)), os.ModePerm); err != nil {
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
