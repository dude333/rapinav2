package repositório

import "strings"

type config struct {
	Prefixos []string
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
		Config.Prefixos = append(Config.Prefixos, "dfp_cia_aberta_"+t+"_con")
	}
}

func prefixoVálido(arquivo string) bool {
	for i := range Config.Prefixos {
		if strings.Contains(arquivo, Config.Prefixos[i]) {
			return true
		}
	}
	return false
}
