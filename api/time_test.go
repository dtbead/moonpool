package api

import (
	"reflect"
	"testing"
	"time"
)

func Test_timeToUnixEpoch(t *testing.T) {
	type args struct {
		t time.Time
	}
	tests := []struct {
		name string
		args args
		want time.Time
	}{
		{"generic", args{time.Date(2024, 5, 29, 21, 44, 20, 50010, time.Local)}, time.Date(2024, 5, 29, 21, 44, 20, 50010, time.Local).Round(time.Second * 1).UTC()},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := timeToUnixEpoch(tt.args.t); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("timeToUnixEpoch() = %v, want %v", got, tt.want)
			}
		})
	}
}
