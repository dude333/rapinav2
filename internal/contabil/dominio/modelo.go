// SPDX-FileCopyrightText: 2022 Adriano Prado <dev@dude333.com>
//
// SPDX-License-Identifier: MIT

package dominio

import (
	"fmt"
	"strings"
)

type Conta struct {
	Código       string
	Descr        string
	Consolidado  bool // Individual ou Consolidado
	Grupo        string
	DataFimExerc string // AAAA-MM-DD
	OrdemExerc   string
	Total        Dinheiro
}

// Válida retorna verdadeiro se os dados da conta são válidos. Ignora os registros
// do penúltimo ano, com exceção de 2009, uma vez que a CVM só disponibliza (pelo
// menos em 2021) dados até 2010.
func (c Conta) Válida() bool {
	return len(c.Código) > 0 &&
		len(c.Descr) > 0 &&
		len(c.DataFimExerc) == len("AAAA-MM-DD") &&
		(c.OrdemExerc == "ÚLTIMO" ||
			(c.OrdemExerc == "PENÚLTIMO" && strings.HasPrefix(c.DataFimExerc, "2009")))
}

type Dinheiro struct {
	Valor  float64
	Escala int
	Moeda  string
}

func (d Dinheiro) String() string {
	return fmt.Sprintf(`%s %.2f`, d.Moeda, d.Valor*float64(d.Escala))
}

type config struct {
	GruposDFP []string
	GruposITR []string
}

var Config = config{}
