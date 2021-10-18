// SPDX-FileCopyrightText: 2021 Adriano Prado <dev@dude333.com>
//
// SPDX-License-Identifier: MIT

package repositório_test

import (
	"context"
	"fmt"
	contábil "github.com/dude333/rapinav2/internal/contabil/dominio"
	repositório "github.com/dude333/rapinav2/internal/contabil/repositorio"
	"os"
	"testing"
)

func Test_cvm_Importar(t *testing.T) {
	type args struct {
		ctx context.Context
		ano int
	}
	tests := []struct {
		name    string
		args    args
		want    <-chan contábil.ResultadoImportação
		wantErr bool
	}{
		{
			name: "deveria funcionar",
			args: args{
				ctx: context.Background(),
				ano: 2020,
			},
			want:    make(<-chan contábil.ResultadoImportação),
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := repositório.NovoCVM(os.TempDir())
			for result := range c.Importar(tt.args.ctx, tt.args.ano) {
				if (result.Error != nil) != tt.wantErr {
					t.Errorf("RepositórioImportaçãoDFP.Importar() error = %v, wantErr %v", result.Error, tt.wantErr)
					return
				}
				if result.Error != nil {
					fmt.Printf("=> %+v\n", result.Error)
				}
			}
		})
	}
}
