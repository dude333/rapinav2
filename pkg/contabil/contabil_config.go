// SPDX-FileCopyrightText: 2021 Adriano Prado <dev@dude333.com>
//
// SPDX-License-Identifier: MIT

package contabil

import "os"

// cfg contém as configurações usadas nos construtores deste repositório.
type cfg struct {
	tempDir         string   // Diretório de dados temporários
	processedHashes []string // Hashes dos arquivos já processados
	force           bool     // Forçar atualização, mesmo que o dado já exista
}

func (c *cfg) apply(options ...Option) error {
	for _, option := range options {
		option(c)
	}

	if c.tempDir == "" {
		c.tempDir = os.TempDir()
	} else {
		err := os.MkdirAll(c.tempDir, os.ModePerm)
		if err != nil {
			return err
		}
	}
	return nil
}

type Option func(*cfg)

func WithDataDir(dir string) Option {
	return func(c *cfg) {
		if len(dir) > 0 {
			c.tempDir = dir
		}
	}
}

func WithProcessedHashes(hashes []string) Option {
	return func(c *cfg) {
		if len(hashes) > 0 {
			c.processedHashes = hashes
		}
	}
}

func WithForce(force bool) Option {
	return func(c *cfg) {
		c.force = force
	}
}

