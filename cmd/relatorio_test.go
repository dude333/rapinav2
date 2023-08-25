package main

import (
	"testing"

	rapina "github.com/dude333/rapinav2"
)

func Test_zerado(t *testing.T) {
	type args struct {
		valores []rapina.ValoresTrimestrais
	}
	tests := []struct {
		name string
		args args
		want bool
	}{
		{
			name: "testar zerado",
			args: args{
				valores: []rapina.ValoresTrimestrais{
					{Ano: 2020, T1: 0, T2: 0, T3: 0, T4: 0, Anual: 0},
					{Ano: 2021, T1: 0, T2: 0, T3: 0, T4: 0, Anual: 0},
					{Ano: 2022, T1: 0, T2: 0, T3: 0, T4: 0, Anual: 0},
				},
			},
			want: true,
		},
		{
			name: "testar não zerado 1",
			args: args{
				valores: []rapina.ValoresTrimestrais{
					{Ano: 2020, T1: 0, T2: 0, T3: 0, T4: 0, Anual: 0},
					{Ano: 2021, T1: 0, T2: 0, T3: 0, T4: 0, Anual: 0},
					{Ano: 2022, T1: 0, T2: 1, T3: 0, T4: 0, Anual: 0},
				},
			},
			want: false,
		},
		{
			name: "testar não zerado 2",
			args: args{
				valores: []rapina.ValoresTrimestrais{
					{Ano: 2020, T1: 0, T2: 0, T3: 0, T4: 0, Anual: 0},
					{Ano: 2021, T1: 0, T2: 0, T3: 0, T4: 0, Anual: 0},
					{Ano: 2022, T1: 0, T2: 0, T3: 0, T4: 0, Anual: 0},
					{Ano: 2023, T1: 10.1, T2: 0, T3: 0, T4: 0, Anual: 0},
				},
			},
			want: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := zerado(tt.args.valores); got != tt.want {
				t.Errorf("zerado() = %v, want %v", got, tt.want)
			}
		})
	}
}
