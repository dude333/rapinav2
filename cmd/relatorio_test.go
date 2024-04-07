// SPDX-FileCopyrightText: 2023 Adriano Prado <dev@dude333.com>
//
// SPDX-License-Identifier: MIT

package main

import (
	"testing"
)

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

func Test_seq(t *testing.T) {
	type args struct {
		max     int
		n       int
		reverse bool
	}
	tests := []struct {
		name string
		args args
		want int
	}{
		{
			name: "1. should match",
			args: args{10, 5, false},
			want: 5,
		},
		{
			name: "2. should match",
			args: args{8, 2, true},
			want: 5,
		},
		{
			name: "3. should match",
			args: args{5, 0, false},
			want: 0,
		},
		{
			name: "4. should match",
			args: args{5, 0, true},
			want: 4,
		},
		{
			name: "5. should not match",
			args: args{5, 6, false},
			want: 0,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := seq(tt.args.max, tt.args.n, tt.args.reverse); got != tt.want {
				t.Errorf("seq() = %v, want %v", got, tt.want)
			}
		})
	}
}
