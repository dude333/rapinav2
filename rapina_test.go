// SPDX-FileCopyrightText: 2023 Adriano Prado <dev@dude333.com>
//
// SPDX-License-Identifier: MIT

package rapina

import (
	"testing"
)

func Test_Zerado(t *testing.T) {
	type args struct {
		valores []ValoresTrimestrais
	}
	tests := []struct {
		name string
		args args
		want bool
	}{
		{
			name: "testar zerado",
			args: args{
				valores: []ValoresTrimestrais{
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
				valores: []ValoresTrimestrais{
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
				valores: []ValoresTrimestrais{
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
			if got := Zerado(tt.args.valores); got != tt.want {
				t.Errorf("zerado() = %v, want %v", got, tt.want)
			}
		})
	}
}
