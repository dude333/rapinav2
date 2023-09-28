// SPDX-FileCopyrightText: 2023 Adriano Prado <dev@dude333.com>
//
// SPDX-License-Identifier: MIT

package main

import "testing"

func Test_acctCode(t *testing.T) {
	type args struct {
		cod   string
		descr string
	}
	tests := []struct {
		name string
		args args
		want accountType
	}{
		{
			name: "should match",
			args: args{
				cod:   "7.1",
				descr: "Depreciaçao Amortização e Exaustao",
			},
			want: Deprec,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := acctCode(tt.args.cod, tt.args.descr); got != tt.want {
				t.Errorf("acctCode() = %v, want %v", got, tt.want)
			}
		})
	}
}
