package api

import "testing"

func Test_deleteWhitespace(t *testing.T) {
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
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := deleteWhitespace(tt.args.s); got != tt.want {
				t.Errorf("deleteWhitespace() = %v, want %v", got, tt.want)
			}
		})
	}
}
