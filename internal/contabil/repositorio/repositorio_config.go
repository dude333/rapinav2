// SPDX-FileCopyrightText: 2021 Adriano Prado <dev@dude333.com>
//
// SPDX-License-Identifier: MIT

package repositorio

type config struct {
	Filtros []string // Parte do nome dos arquivos que serão usados
}

var Config config

func init() {
	tipo := []string{
		"BPA",
		"BPP",
		"DFC_MD",
		"DFC_MI",
		"DRE",
		"DVA",
	}

	for _, t := range tipo {
		// Por hora serão usados apenas os dados consolidados
		Config.Filtros = append(Config.Filtros, "dfp_cia_aberta_"+t+"_con")
	}
}

// cfg contém as configurações usadas nos construtores deste repositório.
type cfg struct {
	dirBD            string // Diretório onde o banco de dados será armazenado
	dirDados         string // Diretório com dados baixados da CMV (temporários)
	limiteAnual      int    // Limite de anos de DFP a serem baixados da CVM (0 = todos)
	limiteTrimestral int    // Limite de anos de ITR a serem baixados da CVM (0 = todos)
}

type ConfigFn func(*cfg)

func RodarBDNaMemória() ConfigFn {
	return func(c *cfg) {
		c.dirBD = ":memory:"
	}
}
func DirBD(dir string) ConfigFn {
	return func(c *cfg) {
		c.dirBD = dir
	}
}
func DirDados(dir string) ConfigFn {
	return func(c *cfg) {
		c.dirDados = dir
	}
}
func ComLimiteAnual(anos int) ConfigFn {
	return func(c *cfg) {
		c.limiteAnual = anos
	}
}
func ComLimiteTrimestral(anos int) ConfigFn {
	return func(c *cfg) {
		c.limiteTrimestral = anos
	}
}
