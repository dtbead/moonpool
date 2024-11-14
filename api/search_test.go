package api

import (
	"reflect"
	"testing"
)

func TestBuildQuery(t *testing.T) {
	type args struct {
		s string
	}
	tests := []struct {
		name string
		args args
		want QueryTags
	}{
		{"include only", args{"foobar"}, QueryTags{[]string{"foobar"}, nil}},
		{"include + exclude", args{"foobar,-bar"}, QueryTags{[]string{"foobar"}, []string{"bar"}}},
		{"include + exclude with excess whitespace", args{"foobar, -bar"}, QueryTags{[]string{"foobar"}, []string{"bar"}}},
		{"empty args", args{""}, QueryTags{nil, nil}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := BuildQuery(tt.args.s); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("BuildQuery() = %v, want %v", got, tt.want)
			}
		})
	}
}
