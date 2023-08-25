// SPDX-FileCopyrightText: 2021 Adriano Prado <dev@dude333.com>
//
// SPDX-License-Identifier: MIT

package repositorio

// cfg contém as configurações usadas nos construtores deste repositório.
type cfg struct {
	dirDados              string   // Diretório de dados temporários
	arquivosJáProcessados []string // Hashes dos arquivos já processados
}

type ConfigFn func(*cfg)

func CfgDirDados(dir string) ConfigFn {
	return func(c *cfg) {
		if len(dir) > 0 {
			c.dirDados = dir
		}
	}
}

func CfgArquivosJáProcessados(hashes []string) ConfigFn {
	return func(c *cfg) {
		if len(hashes) > 0 {
			c.arquivosJáProcessados = hashes
		}
	}
}
