package file

import (
	"fmt"
	"os"
	"reflect"
	"testing"
	"time"
)

func TestGetDateModified(t *testing.T) {
	testFileTime, _ := time.Parse(time.RFC3339, "2018-01-12T13:12:49Z")
	testFile, err := newFile(t, t.TempDir(), "blank.txt")
	if err != nil {
		t.Fatalf("GetDateModified() error = unable to create test file. %v", err)
	}

	type args struct {
		f *os.File
	}
	tests := []struct {
		name    string
		args    args
		want    time.Time
		wantErr bool
	}{
		{"generic", args{testFile}, testFileTime.UTC(), false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := GetDateModified(tt.args.f)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetDateModified() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got.UTC(), tt.want) {
				t.Errorf("GetDateModified() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_doesPathExist(t *testing.T) {
	f, err := os.Create("testdata/hawk.png")
	if err != nil {
		t.Fatalf("doesPathExist() error. unable to create temporary file. %v", err)
	}
	defer f.Close()

	type args struct {
		path string
	}
	tests := []struct {
		name string
		args args
		want bool
	}{
		{"exists directory", args{path: "testdata"}, true},
		{"exists file", args{path: "testdata/hawk.png"}, true},
		{"not exists directory", args{path: "thispathdoesntexist"}, false},
		{"not exists file", args{path: "testdata/thisfiledoesnotexist.png"}, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := doesPathExist(tt.args.path); got != tt.want {
				t.Errorf("doesPathExist() = %v, want %v", got, tt.want)
			}
		})
	}

	t.Cleanup(func() {
		if err := os.Remove("testdata/hawk.png"); err != nil {
			t.Fatalf("doesPathExist() error = unable to delete temporary testdata, %v", err)
		}
	})
}
func TestNewStorage(t *testing.T) {
	type args struct {
		rootPath string
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{"valid", args{"testdata/tmp/testpath"}, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := NewStorage(tt.args.rootPath); (err != nil) != tt.wantErr {
				t.Errorf("NewStorage() error = %v, wantErr %v", err, tt.wantErr)
			}

			p := fmt.Sprintf("%s/db/media/storage", tt.args.rootPath)
			if got, err := exists(t, p); (got == false) && tt.wantErr == true || (err != nil) != tt.wantErr {
				t.Errorf("NewStorage() error = %v, wantErr %v. path = %v", err, tt.wantErr, got)
			}

			t.Cleanup(func() {
				err := os.Remove(p)
				if err != nil {
					t.Fatal(err)
				}
			})
		})
	}

}

func TestBuildPath(t *testing.T) {
	type args struct {
		md5       []byte
		extension string
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{"generic", args{md5: []byte{91, 115, 3, 1, 18, 87, 5, 166, 60, 160, 100, 218, 24, 159, 125, 80}, extension: ".png"}, "5b/5b730301125705a63ca064da189f7d50.png"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := BuildPath(tt.args.md5, tt.args.extension); got != tt.want {
				t.Errorf("BuildPath() = %v, want %v", got, tt.want)
			}
		})
	}
}

func exists(t *testing.T, path string) (bool, error) {
	t.Helper()
	_, err := os.Stat(path)
	if err == nil {
		return true, nil
	}
	if os.IsNotExist(err) {
		return false, nil
	}
	return false, err
}

func newFile(t *testing.T, path, filename string) (*os.File, error) {
	t.Helper()
	f, err := os.Create(path + "/" + filename)
	if err != nil {
		return nil, err
	}

	return f, nil
}
