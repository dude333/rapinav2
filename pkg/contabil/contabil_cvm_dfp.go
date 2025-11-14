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

func NewDFPImporter(configs ...Option) (CVM, error) {
	var cvm CVMImporter
	cvm.cfg = &cfg{}
	if err := cvm.cfg.apply(configs...); err != nil {
		return nil, err
	}
	cvm.nome = "dfp/itr"
	cvm.infra = &LocalInfra{dirDados: cvm.cfg.tempDir}
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

func processarArquivoDFP(_ context.Context, arquivo Arquivo, results chan<- dominio.ImportResult) error {
	fh, err := os.Open(arquivo.path)
	if err != nil {
		return err
	}
	defer func() { _ = fh.Close() }()

	leitorCSV := csv.NewReader(transform.NewReader(fh, charmap.ISO8859_1.NewDecoder()))
	leitorCSV.Comma = ';'
	leitorCSV.LazyQuotes = true

	// Pula o cabeçalho
	header, err := leitorCSV.Read()
	if err != nil {
		return err
	}

	parser := novoParserDFP(header)

	empresas := make(map[string][]*DFPRecord)

	for {
		linha, err := leitorCSV.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			continue
		}

		dfp, err := parser.parseDFP(linha)

		if err != nil || dfp.registroInválido() {
			continue
		}

		k := dfp.CNPJ + dfp.Ano + ";" + dfp.Versão
		empresas[k] = append(empresas[k], dfp)
	}

	sendDFP(empresas, results)

	return nil
}

// DFPRecord é usada para armazenar os dados (linhas) dos arquivos de DFP.
type DFPRecord struct {
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

func (dfp *DFPRecord) toConta() dominio.Conta {
	contém := func(str string) bool {
		return strings.Contains(dfp.GrupoDFP, str)
	}

	grp := dfp.GrupoDFP
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
		Código:       dfp.Código,
		Descr:        dfp.Descr,
		Consolidado:  dfp.Consolidado,
		Grupo:        grp,
		DataIniExerc: dfp.DataIniExerc,
		DataFimExerc: dfp.DataFimExerc,
		Meses:        dfp.Meses,
		OrdemExerc:   dfp.OrdemExerc,
		Total: rapina.Dinheiro{
			Valor:  dfp.Valor,
			Escala: dfp.Escala,
			Moeda:  dfp.Moeda,
		},
	}

	return conta
}

func (dfp *DFPRecord) registroInválido() bool {
	if dfp.Meses%3 != 0 {
		progress.Trace("Ignorando registro não multiplo de 3: %v", dfp)
		return true
	}
	return false
}

// sendDFP envia os dados de todas as empresas de todos os anos do arquivo
// lido, com base no o mapa empresas[ano]*cvmDFP. Os dados são enviados pelo
// canal criado pelo método Importar.
func sendDFP(empresas map[string][]*DFPRecord, results chan<- dominio.ImportResult) {
	num := 0
	for k := range empresas {
		// Ignora se existir uma versão mais nova
		if _, ok := empresas[nextDFPKey(k)]; ok {
			continue
		}

		registros := empresas[k]
		if len(registros) == 0 {
			continue
		}

		contas := make(map[string][]dominio.Conta, len(registros))

		for _, reg := range registros {
			c := reg.toConta()
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

			dfpEmpresa := dominio.DemonstracaoFinanceira{
				Empresa: rapina.Empresa{
					CNPJ: registros[0].CNPJ,
					Nome: registros[0].Nome,
				},
				Ano:    a,
				Contas: contas[ano],
			}

			if dfpEmpresa.Válida() {
				results <- dominio.ImportResult{DFP: &dfpEmpresa}
			} else {
				results <- dominio.ImportResult{Error: ErrDFPInválida}
			}
		}
	} // next k

	// results <- dominio.Resultado{Hash: hash}
	progress.Debug("Linhas processadas: %d", num)
}

// nextDFPKey retora a chave com a próxima chave:
// "CNPJANO;1" => "CNPJANO;2"
func nextDFPKey(k string) string {
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

type DFPParser struct {
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

func novoParserDFP(cabeçalho []string) *DFPParser {
	p := &DFPParser{
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

func (p *DFPParser) parseDFP(itens []string) (*DFPRecord, error) {
	if len(itens) < numItens {
		return nil, ErrFaltaItem
	}

	dtIni := "" // dado não aparece no BP
	if p.posDtIniExerc >= 0 {
		dtIni = itens[p.posDtIniExerc]
	}
	m, err := monthsDiff(dtIni, itens[p.posDtFimExerc])
	if err != nil {
		return nil, err
	}

	vl, err := strconv.ParseFloat(itens[p.posVlConta], 64)
	if err != nil {
		return nil, err
	}

	return &DFPRecord{
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

// monthsDiff retorna a diferença em meses entre ini e fim, com a data no formato
// AAAA-MM-DD.
func monthsDiff(ini, fim string) (int, error) {
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
