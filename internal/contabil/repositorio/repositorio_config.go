package repositório

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
