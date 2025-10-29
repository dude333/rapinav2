package contabil

import (
	_ "embed"
	"encoding/json"
	"fmt"
	"math"
	"strings"

	rapina "github.com/dude333/rapinav2"
)

type resultadoTrimestral struct {
	Codigo  string `db:"codigo"`
	Descr   string `db:"descr"`
	Valores string `db:"valores"`
}

type TrimestralItem struct {
	Ano int      `json:"ano"`
	T1  *float64 `json:"t1,omitempty"`
	T2  *float64 `json:"t2,omitempty"`
	T3  *float64 `json:"t3,omitempty"`
	T4  *float64 `json:"t4,omitempty"`
}

type JSONTrimestral []TrimestralItem

func converterResultadosTrimestrais(resultados []resultadoTrimestral) ([]rapina.InformeTrimestral, error) {
	itr := make([]rapina.InformeTrimestral, len(resultados))

	for i, resultado := range resultados {
		var valoresJSON JSONTrimestral
		err := json.Unmarshal([]byte(resultado.Valores), &valoresJSON)
		if err != nil {
			fmt.Printf("Error parsing JSON: %s\n- %v\n", err, valoresJSON)
			return nil, err
		}

		valoresTrimestrais := make([]rapina.ValoresTrimestrais, len(valoresJSON))
		for j, valorJSON := range valoresJSON {
			valoresTrimestrais[j].Ano = valorJSON.Ano
			valoresTrimestrais[j].T1 = getValue(valorJSON.T1)
			valoresTrimestrais[j].T2 = getValue(valorJSON.T2)
			valoresTrimestrais[j].T3 = getValue(valorJSON.T3)
			valoresTrimestrais[j].T4 = getValue(valorJSON.T4)
		}

		itr[i] = rapina.InformeTrimestral{
			Codigo:  resultado.Codigo,
			Descr:   resultado.Descr,
			Valores: valoresTrimestrais,
		}
	}

	return itr, nil
}

func getValue(val *float64) float64 {
	if val != nil {
		return *val
	}
	return math.NaN()
}

//go:embed contabil_sqlite_trimestral.sql
var sqlQueryTrimestral string

func sqlTrimestral(ids []int, consolidado bool) string {
	strIds := strings.Trim(strings.Join(strings.Fields(fmt.Sprint(ids)), ","), "[]")
	intConsolidado := 0
	if consolidado {
		intConsolidado = 1
	}
	return fmt.Sprintf(sqlQueryTrimestral, strIds, intConsolidado)
}
