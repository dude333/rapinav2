// SPDX-FileCopyrightText: 2023 Adriano Prado <dev@dude333.com>
//
// SPDX-License-Identifier: MIT

package relatorios

import (
	"fmt"
	"os"
	"time"
)

/*
func TestDfp(t *testing.T) {
	type args struct {
		filepath string
		anoi     int
		anof     int
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "salvar output.txt",
			args: args{
				filepath: "/tmp/output.txt",
				anoi:     0,
				anof:     0,
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// if err := Dfp(tt.args.filepath, tt.args.anoi, tt.args.anof); (err != nil) != tt.wantErr {
			// 	t.Errorf("Dfp() error = %v, wantErr %v", err, tt.wantErr)
			// } else {
			if err := verificaArquivo(tt.args.filepath); (err != nil) != tt.wantErr {
				t.Errorf("Arquivo criado por Dfp() %v", err)
			}
			// }
		})
	}
}
*/

func verificaArquivo(filepath string) error {
	fileInfo, err := os.Stat(filepath)
	if err != nil {
		return err
	}

	if fileInfo.Size() == 0 {
		return fmt.Errorf("está vazio")
	}

	tenSecondsAgo := time.Now().Add(-10 * time.Second)
	if !fileInfo.ModTime().After(tenSecondsAgo) {
		return fmt.Errorf("foi criado há mais de 10 segundos")
	}

	return nil
}
