package archive

import (
	_ "embed"
	"reflect"
	"testing"

	_ "modernc.org/sqlite"
)

func Test_isClean(t *testing.T) {
	type args struct {
		s string
	}
	tests := []struct {
		name string
		args args
		want bool
	}{
		{"typical input", args{"Foobar123"}, true},
		{"over length limit", args{"foobar1233453456dfgh45h6w"}, false},
		{"invalid characters", args{"foob^![]r123"}, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := isClean(tt.args.s); got != tt.want {
				t.Errorf("isClean() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_decodeHexString(t *testing.T) {
	type args struct {
		s string
	}
	tests := []struct {
		name string
		args args
		want []byte
	}{
		{"generic", args{"1998a30583dd5112bbefc59fd5e8dbbd"}, []byte{25, 152, 163, 5, 131, 221, 81, 18, 187, 239, 197, 159, 213, 232, 219, 189}},
		{"invalid", args{""}, []byte{}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := decodeHexString(tt.args.s); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("decodeHexString() = %v, want %v", got, tt.want)
			}
		})
	}
}
