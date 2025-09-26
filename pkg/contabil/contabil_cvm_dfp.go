// SPDX-FileCopyrightText: 2021 Adriano Prado <dev@dude333.com>
//
// SPDX-License-Identifier: MIT

package contabil

import (
	"context"
	"encoding/csv"
	"errors"
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"

	"golang.org/x/text/encoding/charmap"
	"golang.org/x/text/transform"

	rapina "github.com/dude333/rapinav2"
	"github.com/dude333/rapinav2/pkg/contabil/dominio"
	"github.com/dude333/rapinav2/pkg/progress"
)

func NovaDFP(configs ...ConfigFn) (CVM, error) {
	var cvm cvmImporter
	cvm.cfg = &cfg{}
	cvm.cfg.loadConfigs(configs...)
	cvm.nome = "dfp/itr"
	cvm.infra = &localInfra{dirDados: cvm.cfg.dirDados}
	cvm.url = urlArquivoDFP
	cvm.filtros = filtrosDFP
	cvm.processar = processarArquivoDFP
	cvm.reportarAviso = true

	return &cvm, nil
}

func filtrosDFP() []string {
	var filtros []string // Parte do nome dos arquivos que serão usados

	tipo := []string{
		"BPA",
		"BPP",
		"DFC_MD",
		"DFC_MI",
		"DRE",
		"DVA",
	}

	for _, t := range tipo {
		filtros = append(filtros,
			"dfp_cia_aberta_"+t+"_con",
			"dfp_cia_aberta_"+t+"_ind",
			"itr_cia_aberta_"+t+"_con",
			"itr_cia_aberta_"+t+"_ind",
		)
	}

	return filtros
}

func urlArquivoDFP(ano int, trimestral bool) string {
	tipo := "DFP"
	if trimestral {
		tipo = "ITR"
	}
	zip := fmt.Sprintf(`%s_cia_aberta_%d.zip`, tipo, ano)
	return `http://dados.cvm.gov.br/dados/CIA_ABERTA/DOC/` + tipo + `/DADOS/` + zip
}

func processarArquivoDFP(_ context.Context, arquivo Arquivo, results chan<- dominio.Resultado) error {
	fh, err := os.Open(arquivo.path)
	if err != nil {
		return err
	}
	defer fh.Close()

	leitorCSV := csv.NewReader(transform.NewReader(fh, charmap.ISO8859_1.NewDecoder()))
	leitorCSV.Comma = ';'
	leitorCSV.LazyQuotes = true

	// Pula o cabeçalho
	header, err := leitorCSV.Read()
	if err != nil {
		return err
	}

	parser := novoParserDFP(header)

	empresas := make(map[string][]*regDFP)

	for {
		linha, err := leitorCSV.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			continue
		}

		dfp, err := parser.carregaDFP(linha)

		if err != nil || dfp.registroInválido() {
			continue
		}

		k := dfp.CNPJ + dfp.Ano + ";" + dfp.Versão
		empresas[k] = append(empresas[k], dfp)
	}

	enviarDFP(empresas, results)

	return nil
}

// regDFP é usada para armazenar os dados (linhas) dos arquivos de DFP.
type regDFP struct {
	CNPJ   string
	Nome   string // Nome da empresa
	Ano    string
	Versão string

	Código       string
	Descr        string
	GrupoDFP     string
	DataIniExerc string // AAAA-MM-DD
	DataFimExerc string // AAAA-MM-DD
	OrdemExerc   string // ÚLTIMO ou PENÚLTIMO
	Moeda        string
	Meses        int // Número de meses acumulados desde o início do exercício
	Valor        float64
	Escala       int
	Consolidado  bool
}

func (reg *regDFP) converteConta() dominio.Conta {
	contém := func(str string) bool {
		return strings.Contains(reg.GrupoDFP, str)
	}

	grp := reg.GrupoDFP
	switch {
	case contém("Balanço Patrimonial Passivo"):
		grp = "BPP"
	case contém("Balanço Patrimonial Ativo"):
		grp = "BPA"
	case contém("Demonstração do Fluxo de Caixa"):
		grp = "DFC"
	case contém("Demonstração do Resultado"):
		grp = "DRE"
	case contém("Demonstração de Valor Adicionado"):
		grp = "DVA"
	}

	conta := dominio.Conta{
		Código:       reg.Código,
		Descr:        reg.Descr,
		Consolidado:  reg.Consolidado,
		Grupo:        grp,
		DataIniExerc: reg.DataIniExerc,
		DataFimExerc: reg.DataFimExerc,
		Meses:        reg.Meses,
		OrdemExerc:   reg.OrdemExerc,
		Total: rapina.Dinheiro{
			Valor:  reg.Valor,
			Escala: reg.Escala,
			Moeda:  reg.Moeda,
		},
	}

	return conta
}

func (dfp *regDFP) registroInválido() bool {
	if dfp.Meses%3 != 0 {
		progress.Trace("Ignorando registro não multiplo de 3: %v", dfp)
		return true
	}
	return false
}

// enviarDFP envia os dados de todas as empresas de todos os anos do arquivo
// lido, com base no o mapa empresas[ano]*cvmDFP. Os dados são enviados pelo
// canal criado pelo método Importar.
func enviarDFP(empresas map[string][]*regDFP, results chan<- dominio.Resultado) {
	num := 0
	for k := range empresas {
		// Ignora se existir uma versão mais nova
		if _, ok := empresas[próxChave(k)]; ok {
			continue
		}

		registros := empresas[k]
		if len(registros) == 0 {
			continue
		}

		contas := make(map[string][]dominio.Conta, len(registros))

		for _, reg := range registros {
			c := reg.converteConta()
			if c.Válida() {
				contas[reg.Ano] = append(contas[reg.Ano], c)
				num++
			}
		}

		for ano := range contas {
			a, err := strconv.Atoi(ano)
			if err != nil {
				continue
			}

			dfpEmpresa := dominio.DemonstraçãoFinanceira{
				Empresa: rapina.Empresa{
					CNPJ: registros[0].CNPJ,
					Nome: registros[0].Nome,
				},
				Ano:    a,
				Contas: contas[ano],
			}

			if dfpEmpresa.Válida() {
				results <- dominio.Resultado{DFP: &dfpEmpresa}
			} else {
				results <- dominio.Resultado{Error: ErrDFPInválida}
			}
		}
	} // next k

	// results <- dominio.Resultado{Hash: hash}
	progress.Debug("Linhas processadas: %d", num)
}

// próxChave retora a chave com a próxima chave:
// "CNPJANO;1" => "CNPJANO;2"
func próxChave(k string) string {
	ks := strings.Split(k, ";")
	if len(ks) != 2 {
		return k
	}
	ver, err := strconv.Atoi(ks[1])
	if err != nil {
		return k
	}
	return ks[0] + strconv.Itoa(ver+1)
}

var (
	ErrFaltaItem    = errors.New("itens faltando")
	ErrDataInválida = errors.New("data inválida")
)

const (
	numItens int = 11 // número de itens (soma dos parâmetros pos___ da struct csv)
)

type parserDFP struct {
	posCnpj        int
	posDenomCia    int
	posDtIniExerc  int
	posDtFimExerc  int
	posVersao      int
	posCdConta     int
	posDsConta     int
	posGrupoDFP    int
	posOrdemExerc  int
	posVlConta     int
	posEscalaMoeda int
	posMoeda       int
}

func novoParserDFP(cabeçalho []string) *parserDFP {
	p := &parserDFP{
		posDtIniExerc: -1, // Este campo não aparece nos dados do balanço patrimonial
	}

	for i, t := range cabeçalho {
		switch t {
		case "CNPJ_CIA":
			p.posCnpj = i
		case "DENOM_CIA":
			p.posDenomCia = i
		case "DT_INI_EXERC":
			p.posDtIniExerc = i
		case "DT_FIM_EXERC":
			p.posDtFimExerc = i
		case "VERSAO":
			p.posVersao = i
		case "CD_CONTA":
			p.posCdConta = i
		case "DS_CONTA":
			p.posDsConta = i
		case "GRUPO_DFP":
			p.posGrupoDFP = i
		case "ORDEM_EXERC":
			p.posOrdemExerc = i
		case "VL_CONTA":
			p.posVlConta = i
		case "ESCALA_MOEDA":
			p.posEscalaMoeda = i
		case "MOEDA":
			p.posMoeda = i
		}
	}

	return p
}

func (p *parserDFP) carregaDFP(itens []string) (*regDFP, error) {
	if len(itens) < numItens {
		return nil, ErrFaltaItem
	}

	dtIni := "" // dado não aparece no BP
	if p.posDtIniExerc >= 0 {
		dtIni = itens[p.posDtIniExerc]
	}
	m, err := meses(dtIni, itens[p.posDtFimExerc])
	if err != nil {
		return nil, err
	}

	vl, err := strconv.ParseFloat(itens[p.posVlConta], 64)
	if err != nil {
		return nil, err
	}

	return &regDFP{
		CNPJ:         itens[p.posCnpj],
		Nome:         itens[p.posDenomCia],
		Ano:          itens[p.posDtFimExerc][:4],
		Consolidado:  strings.Contains(itens[p.posGrupoDFP], "onsolidado"),
		Versão:       itens[p.posVersao],
		Código:       itens[p.posCdConta],
		Descr:        itens[p.posDsConta],
		GrupoDFP:     itens[p.posGrupoDFP],
		DataIniExerc: dtIni,
		DataFimExerc: itens[p.posDtFimExerc],
		Meses:        m,
		OrdemExerc:   itens[p.posOrdemExerc],
		Valor:        vl,
		Escala:       escala(itens[p.posEscalaMoeda]),
		Moeda:        moeda(itens[p.posMoeda]),
	}, nil
}

// meses retorna a diferença em meses entre ini e fim, com a data no formato
// AAAA-MM-DD.
func meses(ini, fim string) (int, error) {
	if len(fim) != 10 || (ini != "" && len(ini) != 10) {
		return 0, ErrDataInválida
	}

	// Quando ini == "" significa que o dado veio do balanço patrimonial
	// e só o data do fim do exercício é fornecido
	if ini == "" {
		return 12, nil
	}

	anoI, _ := strconv.Atoi(ini[0:4])
	mesI, _ := strconv.Atoi(ini[5:7])
	anoF, _ := strconv.Atoi(fim[0:4])
	mesF, _ := strconv.Atoi(fim[5:7])

	if anoI == 0 || mesI == 0 || anoF == 0 || mesF == 0 {
		return 0, ErrDataInválida
	}

	meses := 0

	if anoF != anoI {
		meses = (anoF - anoI) * 12
		meses = meses - mesI + mesF + 1
	} else {
		meses = mesF - mesI + 1
	}

	if meses <= 0 {
		return 0, ErrDataInválida
	}

	return meses, nil
}

func escala(s string) int {
	switch strings.ToUpper(s) {
	case "UNIDADE":
		return 1
	case "MIL":
		return 1000
	case "MILHÃO", "MILHAO":
		return 1e6
	}
	return 1
}

func moeda(s string) string {
	switch strings.ToUpper(s) {
	case "REAL":
		return "R$"
	}
	return s
}
