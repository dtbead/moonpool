package db

import (
	"reflect"
	"testing"
	"time"
)

func TestParseTimestamp(t *testing.T) {
	s := time.Date(2018, 1, 12, 13, 12, 49, (18 * 1000000), time.UTC) // 2018-01-12 13:12:49.018
	type args struct {
		s string
	}
	tests := []struct {
		name string
		args args
		want time.Time
	}{
		{"SQLite timestamp", args{"2018-01-12 13:12:49.018"}, s},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := ParseTimestamp(tt.args.s); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("ParseTimestamp() = %v, want %v", got, tt.want)
			}
		})
	}
}
