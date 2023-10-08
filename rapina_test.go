// SPDX-FileCopyrightText: 2023 Adriano Prado <dev@dude333.com>
//
// SPDX-License-Identifier: MIT

package rapina

import (
	"reflect"
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
					{Ano: 2020, T1: 0, T2: 0, T3: 0, T4: 0},
					{Ano: 2021, T1: 0, T2: 0, T3: 0, T4: 0},
					{Ano: 2022, T1: 0, T2: 0, T3: 0, T4: 0},
				},
			},
			want: true,
		},
		{
			name: "testar não zerado 1",
			args: args{
				valores: []ValoresTrimestrais{
					{Ano: 2020, T1: 0, T2: 0, T3: 0, T4: 0},
					{Ano: 2021, T1: 0, T2: 0, T3: 0, T4: 0},
					{Ano: 2022, T1: 0, T2: 1, T3: 0, T4: 0},
				},
			},
			want: false,
		},
		{
			name: "testar não zerado 2",
			args: args{
				valores: []ValoresTrimestrais{
					{Ano: 2020, T1: 0, T2: 0, T3: 0, T4: 0},
					{Ano: 2021, T1: 0, T2: 0, T3: 0, T4: 0},
					{Ano: 2022, T1: 0, T2: 0, T3: 0, T4: 0},
					{Ano: 2023, T1: 10.1, T2: 0, T3: 0, T4: 0},
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

func Test_codPai(t *testing.T) {
	tests := []string{"1", "2", "1.02", "2.03.04", "3.01.02.04", "6.01", "7.02.01"}
	expected := []string{"1", "2", "1.02", "2.03.04", "3.01.02", "6.01", "7.02.01"}

	for i := range tests {
		if got := codPai(tests[i]); got != expected[i] {
			t.Errorf("codPai(%s) = %s, want %s", tests[i], got, expected[i])
		}
	}
}

func TestAddVTs(t *testing.T) {
	type args struct {
		v1 []ValoresTrimestrais
		v2 []ValoresTrimestrais
	}
	tests := []struct {
		name string
		args args
		want []ValoresTrimestrais
	}{
		{
			name: "deveria somar anos balanceados",
			args: args{
				v1: []ValoresTrimestrais{
					{Ano: 2023, T1: 10.0, T2: 20.0, T3: 30.0, T4: 40.0},
					{Ano: 2024, T1: 15.0, T2: 25.0, T3: 35.0, T4: 45.0},
				},
				v2: []ValoresTrimestrais{
					{Ano: 2023, T1: 5.0, T2: 10.0, T3: 15.0, T4: 20.0},
					{Ano: 2024, T1: 7.0, T2: 12.0, T3: 17.0, T4: 22.0},
				},
			},
			want: []ValoresTrimestrais{{2023, 15, 30, 45, 60}, {2024, 22, 37, 52, 67}},
		},
		{
			name: "deveria somar anos desbalanceados (final)",
			args: args{
				v1: []ValoresTrimestrais{
					{Ano: 2022, T1: 10.0, T2: 20.0, T3: 30.0, T4: 40.0},
					{Ano: 2024, T1: 15.0, T2: 25.0, T3: 35.0, T4: 45.0},
				},
				v2: []ValoresTrimestrais{
					{Ano: 2023, T1: 5.0, T2: 10.0, T3: 15.0, T4: 20.0},
					{Ano: 2024, T1: 7.0, T2: 12.0, T3: 17.0, T4: 22.0},
				},
			},
			want: []ValoresTrimestrais{{2022, 10, 20, 30, 40}, {2023, 5, 10, 15, 20}, {2024, 22, 37, 52, 67}},
		},
		{
			name: "deveria somar anos desbalanceados (início')",
			args: args{
				v1: []ValoresTrimestrais{{2010, 1, 1, 1, 1}, {2012, 10, 10, 10, 10}},
				v2: []ValoresTrimestrais{{2011, 2, 2, 2, 2}, {2012, 5, 5, 5, 5}, {2023, 100, 100, 100, 100}},
			},
			want: []ValoresTrimestrais{{2010, 1, 1, 1, 1}, {2011, 2, 2, 2, 2}, {2012, 15, 15, 15, 15}, {2023, 100, 100, 100, 100}},
		},
		{
			name: "deveria somar anos desbalanceados (meio)",
			args: args{
				v1: []ValoresTrimestrais{{2011, 2, 2, 2, 2}, {2012, 5, 5, 5, 5}, {2023, 100, 100, 100, 100}},
				v2: []ValoresTrimestrais{{2011, 1, 1, 1, 1}, {2023, 10, 10, 10, 10}},
			},
			want: []ValoresTrimestrais{{2011, 3, 3, 3, 3}, {2012, 5, 5, 5, 5}, {2023, 110, 110, 110, 110}},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := AddVTs(tt.args.v1, tt.args.v2); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("AddVTs() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestSubVTs(t *testing.T) {
	type args struct {
		v1 []ValoresTrimestrais
		v2 []ValoresTrimestrais
	}
	tests := []struct {
		name string
		args args
		want []ValoresTrimestrais
	}{
		{
			name: "deveria subtrair anos balanceados",
			args: args{
				v1: []ValoresTrimestrais{{2010, 10, 10, 10, 10}, {2011, 20, 20, 20, 20}, {2020, 30, 30, 30, 3}},
				v2: []ValoresTrimestrais{{2010, 2, 2, 2, 2}, {2011, 10, 2, 2, 2}, {2020, 2, 2, 2, 2}},
			},
			want: []ValoresTrimestrais{{2010, 8, 8, 8, 8}, {2011, 10, 18, 18, 18}, {2020, 28, 28, 28, 1}},
		},
		{
			name: "deveria subtrair anos desbalanceados",
			args: args{
				v1: []ValoresTrimestrais{{2011, 2, 2, 2, 2}, {2012, 5, 5, 5, 5}, {2023, 100, 100, 100, 100}},
				v2: []ValoresTrimestrais{{2011, 1, 1, 1, 1}, {2023, 110, 110, 110, 110}},
			},
			want: []ValoresTrimestrais{{2011, 1, 1, 1, 1}, {2012, 5, 5, 5, 5}, {2023, -10, -10, -10, -10}},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := SubVTs(tt.args.v1, tt.args.v2); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("AddVTs() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestDivVTs(t *testing.T) {
	type args struct {
		v1 []ValoresTrimestrais
		v2 []ValoresTrimestrais
	}
	tests := []struct {
		name string
		args args
		want []ValoresTrimestrais
	}{
		{
			name: "deveria dividir anos balanceados",
			args: args{
				v1: []ValoresTrimestrais{{2010, 10, 10, 10, 10}, {2011, 20, 20, 20, 20}, {2020, 30, 30, 30, 3}},
				v2: []ValoresTrimestrais{{2010, 2, 2, 2, 2}, {2011, 10, 2, 2, 2}, {2020, 2, 2, 2, 2}},
			},
			want: []ValoresTrimestrais{{2010, 5, 5, 5, 5}, {2011, 2, 10, 10, 10}, {2020, 15, 15, 15, 1.5}},
		},
		{
			name: "deveria dividir anos desbalanceados",
			args: args{
				v1: []ValoresTrimestrais{{2011, 2, 2, 2, 2}, {2012, 5, 5, 5, 5}, {2023, 100, 100, 100, 100}},
				v2: []ValoresTrimestrais{{2011, 1, 1, 1, 1}, {2023, 10, 10, 10, 10}},
			},
			want: []ValoresTrimestrais{{2011, 2, 2, 2, 2}, {2012, 0, 0, 0, 0}, {2023, 10, 10, 10, 10}},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := DivVTs(tt.args.v1, tt.args.v2); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("AddVTs() = %v, want %v", got, tt.want)
			}
		})
	}
}
