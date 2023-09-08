package repositorio

import (
	_ "embed"
	"encoding/json"
	"fmt"
	"strings"

	rapina "github.com/dude333/rapinav2"
)

type resultadoTrimestral struct {
	Codigo  string `db:"codigo"`
	Descr   string `db:"descr"`
	Valores string `db:"valores"`
}

type jsonTrimestral []struct {
	Ano   int     `json:"year"`
	T1    float64 `json:"q1"`
	T2    float64 `json:"q2"`
	T3    float64 `json:"q3"`
	T4    float64 `json:"q4"`
	Anual float64 `json:"yearly"`
}

func converterResultadosTrimestrais(resultados []resultadoTrimestral) ([]rapina.InformeTrimestral, error) {
	itr := make([]rapina.InformeTrimestral, len(resultados))

	for i, resultado := range resultados {
		var valoresJSON jsonTrimestral
		err := json.Unmarshal([]byte(resultado.Valores), &valoresJSON)
		if err != nil {
			return nil, err
		}

		valoresTrimestrais := make([]rapina.ValoresTrimestrais, len(valoresJSON))
		for j, valorJSON := range valoresJSON {
			valoresTrimestrais[j].Ano = valorJSON.Ano
			valoresTrimestrais[j].T1 = valorJSON.T1
			valoresTrimestrais[j].T2 = valorJSON.T2
			valoresTrimestrais[j].T3 = valorJSON.T3
			valoresTrimestrais[j].T4 = valorJSON.T4
			valoresTrimestrais[j].Anual = valorJSON.Anual
		}

		itr[i] = rapina.InformeTrimestral{
			Codigo:  resultado.Codigo,
			Descr:   resultado.Descr,
			Valores: valoresTrimestrais,
		}
	}

	return itr, nil
}

//go:embed repositorio_sqlite_trimestral.sql
var sqlQueryTrimestral string

func sqlTrimestral(ids []int, consolidado bool) string {
	strIds := strings.Trim(strings.Join(strings.Fields(fmt.Sprint(ids)), ","), "[]")
	intConsolidado := 0
	if consolidado {
		intConsolidado = 1
	}
	return fmt.Sprintf(sqlQueryTrimestral, strIds, intConsolidado)
}
