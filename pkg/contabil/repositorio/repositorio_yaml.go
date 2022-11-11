// SPDX-FileCopyrightText: 2022 Adriano Prado <dev@dude333.com>
//
// SPDX-License-Identifier: MIT

package repositorio

import (
	"log"

	"github.com/dude333/rapinav2/pkg/progress"
	"gopkg.in/yaml.v2"
)

// Este arquivo provê dados com base no banco de dados usando as definições do
// arquivo yaml com as definições de qual é o código e a descrição de cada
// conta, pois em alguns casos há a variação dessas definições entre empresas
// e períodos.

// ContaYaml armazena os dados importados do arquivo yaml (deve ser convertido
// em contábil.ConfigConta).
// Formato: [["1", "Descrição 1"], ["1.1", "Descrição 1.1"]]
type Modelo struct {
	AtivoTotal        [][]string `yaml:"AtivoTotal"`
	AtivoCirc         [][]string `yaml:"AtivoCirc"`
	AtivoNCirc        [][]string `yaml:"AtivoNCirc"`
	Caixa             [][]string `yaml:"Caixa"`
	AplicFinanceiras  [][]string `yaml:"AplicFinanceiras"`
	Estoque           [][]string `yaml:"Estoque"`
	ContasARecebCirc  [][]string `yaml:"ContasARecebCirc"`
	ContasARecebNCirc [][]string `yaml:"ContasARecebNCirc"`
	PassivoTotal      [][]string `yaml:"PassivoTotal"`
	PassivoCirc       [][]string `yaml:"PassivoCirc"`
	PassivoNCirc      [][]string `yaml:"PassivoNCirc"`
	Equity            [][]string `yaml:"Equity"`
	DividaCirc        [][]string `yaml:"DividaCirc"`
	DividaNCirc       [][]string `yaml:"DividaNCirc"`
	DividendosJCP     [][]string `yaml:"DividendosJCP"`
	DividendosMin     [][]string `yaml:"DividendosMin"`
	Vendas            [][]string `yaml:"Vendas"`
	CustoVendas       [][]string `yaml:"CustoVendas"`
	DespesasOp        [][]string `yaml:"DespesasOp"`
	EBIT              [][]string `yaml:"EBIT"`
	ResulFinanc       [][]string `yaml:"ResulFinanc"`
	ResulOpDescont    [][]string `yaml:"ResulOpDescont"`
	LucLiq            [][]string `yaml:"LucLiq"`
	FCO               [][]string `yaml:"FCO"`
	FCI               [][]string `yaml:"FCI"`
	FCF               [][]string `yaml:"FCF"`
	Deprec            [][]string `yaml:"Deprec"`
	JurosCapProp      [][]string `yaml:"JurosCapProp"`
	Dividendos        [][]string `yaml:"Dividendos"`
}

type Contas struct {
	Modelos map[string]Modelo
}

func Unmarshal(data string) {
	var c Contas
	err := yaml.Unmarshal([]byte(data), &c)
	if err != nil {
		log.Fatalf("error: %v", err)
	}

	for k := range c.Modelos {
		progress.Status("%s", k)
		progress.Status("%+v", c.Modelos[k])
	}
}
