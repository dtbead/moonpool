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
	"runtime"
	"syscall"
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

func Copy(baseDirectory, destination string, r io.Reader) error {
	dest := fmt.Sprintf("%s/%s", baseDirectory, destination)
	if !doesPathExist(baseDirectory + "/" + filepath.Dir(destination)) {
		if err := os.MkdirAll(baseDirectory+"/"+filepath.Dir(destination), 0664); err != nil {
			return err
		}
	}

	file, err := os.Create(dest)
	if err != nil {
		return err
	}
	defer file.Close()

	buf := bufio.NewReader(r)

	_, err = buf.WriteTo(file)
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

func GetPerceptualHash(i image.Image) (PerceptualHashes, error) {
	hash, err := goimagehash.DifferenceHash(i)
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
		ph.Type = "Whash"
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

func DateModified(f *os.File) (time.Time, error) {
	fi, err := f.Stat()
	if err != nil {
		return time.Time{}, err
	}

	return fi.ModTime(), nil
}

// DateCreated returns the Time of when a file was created for Windows OS only.
// If not on Windows, DateCreated will return the date modified instead
func DateCreated(f *os.File) (time.Time, error) {
	if runtime.GOOS != "windows" {
		return DateModified(f)
	}

	fi, err := f.Stat()
	if err != nil {
		return time.Time{}, err
	}

	d := fi.Sys().(*syscall.Win32FileAttributeData)
	if d == nil {
		return time.Time{}, errors.New("failed to get Win32FileAttributeData")
	}

	// Windows uses Jan 1st 1601 as epoch, Unix as Jan 1st 1970
	unixEpoch := d.CreationTime.Nanoseconds() / int64(time.Second)
	return time.Unix(0, unixEpoch).Add(-604854 * time.Hour), nil // 69 years
}

// NewStorage creates a new directory to store media
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
