WITH 
acumulado AS (
	SELECT codigo, descr, data_ini_exerc, data_fim_exerc, SUBSTR(c.data_fim_exerc, 1, 4) ano, meses, 
		SUM(CASE 
			WHEN meses = 3 THEN valor 
			WHEN meses = 12 AND SUBSTR(c.data_fim_exerc, 6, 2) = '03' THEN valor 
			ELSE NULL END) AS q1,
		SUM(CASE 
			WHEN meses = 6 THEN valor 
			WHEN meses = 12 AND SUBSTR(c.data_fim_exerc, 6, 2) = '06' THEN valor 
			ELSE NULL END) AS q2,
		SUM(CASE 
			WHEN meses = 9 THEN valor 
			WHEN meses = 12 AND SUBSTR(c.data_fim_exerc, 6, 2) = '09' THEN valor 
			ELSE NULL END) AS q3,
		SUM(CASE WHEN data_ini_exerc <> '' AND meses = 12 THEN valor ELSE NULL END) AS q4,
		SUM(CASE WHEN meses = 12 AND SUBSTR(c.data_fim_exerc, 6, 2) = '12' THEN valor ELSE NULL END) AS q4_anual
	FROM
	    empresas e
	JOIN contas c ON e.id = c.id_empresa
	WHERE c.id_empresa IN (%s)
	    AND c.consolidado = %d
		AND (c.data_ini_exerc = '' OR SUBSTR(c.data_ini_exerc, 6, 2) = "01") -- APENAS data_ini_exec DE JANEIRO
	GROUP BY ano, codigo, descr
	ORDER BY data_fim_exerc
),
calculado AS (
	SELECT
		ano,
		codigo,
		descr,
		COALESCE(q1, 0) AS t1,
		CASE WHEN data_ini_exerc <> '' AND q1 IS NOT NULL AND q2 IS NOT NULL THEN q2-q1 ELSE COALESCE(q2, 0) END AS t2,
		CASE WHEN data_ini_exerc <> '' AND q2 IS NOT NULL AND q3 IS NOT NULL THEN q3-q2 ELSE COALESCE(q3, 0) END AS t3,
		CASE WHEN data_ini_exerc <> '' AND q4 IS NOT NULL THEN q4-COALESCE(q3, 0) ELSE COALESCE(q4_anual, 0) END AS t4
		FROM acumulado
),
agrupado AS (
	SELECT
	    codigo,
	    descr,
	    '[' || GROUP_CONCAT(
	        '{"ano":' || ano ||
	        ',"t1":' || COALESCE(t1, 0) ||
	        ',"t2":' || COALESCE(t2, 0) ||
	        ',"t3":' || COALESCE(t3, 0) ||
	        ',"t4":' || COALESCE(t4, 0) || '}'
	    ) || ']' AS valores
	FROM calculado
	WHERE t1 <> 0 OR t2 <> 0 OR t3 <> 0 OR t4 <> 0 -- FILTRA LINHAS VAZIAS
	GROUP BY codigo, descr
)
SELECT * from agrupado
