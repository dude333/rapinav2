// SPDX-FileCopyrightText: 2021 Adriano Prado <dev@dude333.com>
//
// SPDX-License-Identifier: MIT

package contabil

import "os"

// cfg contém as configurações usadas nos construtores deste repositório.
type cfg struct {
	dirDados              string   // Diretório de dados temporários
	arquivosJáProcessados []string // Hashes dos arquivos já processados
	force                 bool     // Forçar atualização, mesmo que o dado já exista
}

func (c *cfg) loadConfigs(configs ...ConfigFn) error {
	for _, config := range configs {
		config(c)
	}

	if c.dirDados == "" {
		c.dirDados = os.TempDir()
	} else {
		err := os.MkdirAll(c.dirDados, os.ModePerm)
		if err != nil {
			return err
		}
	}
	return nil
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

func CfgForce(force bool) ConfigFn {
	return func(c *cfg) {
		c.force = force
	}
}
