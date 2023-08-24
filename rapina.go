package rapina

type InformeTrimestral struct {
	Codigo  string
	Descr   string
	Valores []ValoresTrimestrais
}

type ValoresTrimestrais struct {
	Ano   int
	T1    float64
	T2    float64
	T3    float64
	T4    float64
	Anual float64
}
