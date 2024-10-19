package www

import "testing"

func Test_getProjectDirectory(t *testing.T) {
	tests := []struct {
		name string
		want string
	}{
		{name: "go mod", want: `e:\Programming\go\src\github.com\dtbead\moonpool\server\www`},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := getProjectDirectory(); got != tt.want {
				t.Errorf("getProjectDirectory() = %v, want %v", got, tt.want)
			}
		})
	}
}
