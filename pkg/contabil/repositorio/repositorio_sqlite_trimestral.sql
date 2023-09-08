WITH CalculatedValues AS (
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
            ELSE CASE WHEN SUBSTR(data_fim_exerc, 6, 2) IN ('10', '11', '12')
                THEN COALESCE(yearly_value, 0)
                ELSE 0
                END
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
            (CASE 
                WHEN meses = 12 AND c.data_fim_exerc = MAX(c.data_fim_exerc) THEN valor 
                ELSE NULL END) AS yearly_value

	    FROM
	        empresas e
	    JOIN contas c ON e.id = c.id_empresa
	    WHERE c.id_empresa IN (%s)
		AND c.consolidado = %d
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
GROUP BY codigo, descr
