package db

import "testing"

func TestSanitizeStringForDatabase(t *testing.T) {
	type args struct {
		s string
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{"single quotes", args{"'foobar'"}, "foobar"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := sanitizeString(tt.args.s); got != tt.want {
				t.Errorf("sanitizeString() = %v, want %v", got, tt.want)
			}
		})
	}
}
