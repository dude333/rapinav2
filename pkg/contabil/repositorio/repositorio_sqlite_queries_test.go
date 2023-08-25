package repositorio

import (
	"strings"
	"testing"
)

func Test_sqlTrimestral(t *testing.T) {
	type args struct {
		ids []int
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{
			name: "teste 1",
			args: args{
				ids: []int{1, 2, 3, 4, 5},
			},
			want: "(1,2,3,4,5)",
		},
		{
			name: "teste 2",
			args: args{
				ids: []int{0, 99, 1234},
			},
			want: "(0,99,1234)",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := sqlTrimestral(tt.args.ids, true); !strings.Contains(got, tt.want) {
				t.Errorf("sqlTrimestral() = %v, want %v", got, tt.want)
			}
		})
	}
}
