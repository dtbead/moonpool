package db

import (
	_ "embed"
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
