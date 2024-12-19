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
		{"disallow spaces", args{"foo bar"}, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := IsClean(tt.args.s); got != tt.want {
				t.Errorf("isClean() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_DeleteWhitespace(t *testing.T) {
	type args struct {
		s string
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{"generic", args{"foo \r\n bar \t 123"}, "foo bar 123"},
		{"multiple spaces", args{"foo  bar"}, "foo bar"},
		{"single space", args{"foo bar"}, "foo bar"},
		{"newlines", args{"foo\n\nbar"}, "foo bar"},
		{"trailing space", args{" foo bar "}, "foo bar"},
		{"tabs", args{"foo\tbar"}, "foo bar"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := DeleteWhitespace(tt.args.s); got != tt.want {
				t.Errorf("DeleteWhitespace() = %v, want %v", got, tt.want)
			}
		})
	}
}
