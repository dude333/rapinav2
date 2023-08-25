package repositorio

import (
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

func sqlTrimestral(ids []int) string {
	strIds := strings.Trim(strings.Join(strings.Fields(fmt.Sprint(ids)), ","), "[]")
	return fmt.Sprintf(`WITH CalculatedValues AS (
	SELECT
	    year,
	    codigo,
	    descr,
	    COALESCE(q1_value, 0) q1_value,
	    COALESCE(q2_value, 0) q2_value,
	    COALESCE(q3_value, 0) q3_value,
	    CASE WHEN q1_value IS NOT NULL AND q2_value IS NOT NULL AND q3_value IS NOT NULL
	         THEN yearly_value - (q1_value + q2_value + q3_value)
	         ELSE 0
	    END AS q4_calculated,
		CASE WHEN q1_value IS NOT NULL AND q2_value IS NOT NULL AND q3_value IS NOT NULL
			THEN COALESCE(yearly_value, 0)
			ELSE 0
		END AS yearly_value
	FROM (
	    SELECT
	        CASE WHEN c.data_ini_exerc <> '' THEN SUBSTR(c.data_ini_exerc, 1, 4)
		         ELSE SUBSTR(c.data_fim_exerc, 1, 4)
		    END AS year,
		    c.data_fim_exerc,
	        c.codigo,
	        c.descr,
	        SUM(CASE 
		        WHEN meses =  3 AND SUBSTR(c.data_ini_exerc, 6, 2) IN ('01', '02', '03') THEN valor 
		        WHEN meses = 12 AND SUBSTR(c.data_fim_exerc, 6, 2) IN ('01', '02', '03') THEN valor
		        ELSE NULL END) AS q1_value,
	        SUM(CASE 
		        WHEN meses =  3 AND SUBSTR(c.data_ini_exerc, 6, 2) IN ('04', '05', '06') THEN valor 
		        WHEN meses = 12 AND SUBSTR(c.data_fim_exerc, 6, 2) IN ('04', '05', '06') THEN valor
		        ELSE NULL END) AS q2_value,
	        SUM(CASE 
		        WHEN meses =  3 AND SUBSTR(c.data_ini_exerc, 6, 2) IN ('07', '08', '09') THEN valor 
		        WHEN meses = 12 AND SUBSTR(c.data_fim_exerc, 6, 2) IN ('07', '08', '09') THEN valor
		        ELSE NULL END) AS q3_value,
	        (CASE WHEN meses = 12 AND c.data_fim_exerc = MAX(c.data_fim_exerc) THEN valor ELSE NULL END) AS yearly_value
	    FROM
	        empresas e
	    JOIN contas c ON e.id = c.id_empresa
	    WHERE c.id_empresa IN (%s)
	    GROUP BY
			CASE WHEN c.data_ini_exerc <> '' THEN SUBSTR(c.data_ini_exerc, 1, 4)
		         ELSE SUBSTR(c.data_fim_exerc, 1, 4)
		    END,
	        c.codigo,
	        c.descr
		ORDER BY data_fim_exerc
	)
	ORDER BY year, codigo
)
SELECT 	
	codigo,
    descr,
    '[' || GROUP_CONCAT(
        '{"year":' || year || 
        ',"q1":' || COALESCE(q1_value, 0) ||
        ',"q2":' || COALESCE(q2_value, 0) ||
        ',"q3":' || COALESCE(q3_value, 0) ||
        ',"q4":' || COALESCE(q4_calculated, 0) ||
        ',"yearly":' || COALESCE(yearly_value, 0) || '}'
    ) || ']' AS valores
FROM CalculatedValues
GROUP BY codigo, descr`, strIds)
}
