// SPDX-FileCopyrightText: 2022 Adriano Prado <dev@dude333.com>
//
// SPDX-License-Identifier: MIT

package rapina

type config struct {
	Grupos []string
}

var Config = config{}

func init() {
	Config.Grupos = []string{
		"DF Individual - Balanço Patrimonial Ativo",
		"DF Consolidado - Balanço Patrimonial Ativo",
		"DF Individual - Balanço Patrimonial Passivo",
		"DF Consolidado - Balanço Patrimonial Passivo",
		"DF Individual - Demonstração do Fluxo de Caixa (Método Direto)",
		"DF Consolidado - Demonstração do Fluxo de Caixa (Método Direto)",
		"DF Individual - Demonstração do Fluxo de Caixa (Método Indireto)",
		"DF Consolidado - Demonstração do Fluxo de Caixa (Método Indireto)",
		"DF Individual - Demonstração do Resultado",
		"DF Consolidado - Demonstração do Resultado",
		"DF Individual - Demonstração de Valor Adicionado",
		"DF Consolidado - Demonstração de Valor Adicionado",
	}
}
