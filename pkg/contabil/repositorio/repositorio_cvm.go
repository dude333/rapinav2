// SPDX-FileCopyrightText: 2021 Adriano Prado <dev@dude333.com>
//
// SPDX-License-Identifier: MIT

package repositorio

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	rapina "github.com/dude333/rapinav2"
	"github.com/dude333/rapinav2/pkg/contabil/dominio"
	"github.com/dude333/rapinav2/pkg/progress"
	"golang.org/x/text/encoding/charmap"
	"golang.org/x/text/transform"
)

// cvmDFP é usada para armazenar os dados (linhas) dos arquivos de DFP para
// ser posteriormente transformada no modelo Conta, do domínio rapina.
type cvmDFP struct {
	CNPJ        string
	Nome        string // Nome da empresa
	Ano         string
	Consolidado bool
	Versão      string

	Código       string
	Descr        string
	GrupoDFP     string
	DataIniExerc string // AAAA-MM-DD
	DataFimExerc string // AAAA-MM-DD
	Meses        int    // Número de meses acumulados desde o início do exercício
	OrdemExerc   string // ÚLTIMO ou PENÚLTIMO
	Valor        float64
	Escala       int
	Moeda        string
}

func (c *cvmDFP) converteConta() dominio.Conta {

	contém := func(str string) bool {
		return strings.Contains(c.GrupoDFP, str)
	}

	grp := c.GrupoDFP
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
		Código:       c.Código,
		Descr:        c.Descr,
		Consolidado:  c.Consolidado,
		Grupo:        grp,
		DataIniExerc: c.DataIniExerc,
		DataFimExerc: c.DataFimExerc,
		Meses:        c.Meses,
		OrdemExerc:   c.OrdemExerc,
		Total: rapina.Dinheiro{
			Valor:  c.Valor,
			Escala: c.Escala,
			Moeda:  c.Moeda,
		},
	}

	return conta
}

// CVM implementa RepositórioImportação. Busca demonstrações financeiras
// no site da CVM.
type CVM struct {
	infra
	cfg
}

func NovoCVM(configs ...ConfigFn) (*CVM, error) {
	var cvm CVM
	for _, cfg := range configs {
		cfg(&cvm.cfg)
	}

	if cvm.dirDados == "" {
		cvm.dirDados = os.TempDir()
	} else {
		err := os.MkdirAll(cvm.dirDados, os.ModePerm)
		if err != nil {
			return nil, err
		}
	}

	cvm.infra = &localInfra{dirDados: cvm.dirDados}

	return &cvm, nil
}

// Importar baixa o arquivo de DFPs de todas as empresas de um determinado
// ano do site da CVM. Com trimestral == true, é baixado o arquivo de ITRs.
func (c *CVM) Importar(ctx context.Context, ano int, trimestral bool) <-chan dominio.Resultado {
	results := make(chan dominio.Resultado)

	go func() {
		defer close(results)

		url := urlArquivo(ano, trimestral)

		arquivos, err := c.infra.DownloadAndUnzip(url, filtros())
		if err != nil {
			results <- dominio.Resultado{Error: err}
			return
		}
		defer func() {
			_ = c.infra.Cleanup(arquivos)
		}()

		for _, arquivo := range arquivos {
			progress.Running(arquivo.path)

			// Ignora arquivos já processados
			if c.existe(arquivo.hash) {
				progress.RunFail()
				progress.Warning("Arquivo %s já foi processado anteriormente", filepath.Base(arquivo.path))
				continue
			}
			// Processa o arquivo e envia o resultado para o canal 'results'
			_ = processarArquivoDFP(ctx, arquivo, results)
			progress.RunOK()
		}
	}()

	return results
}

func (c CVM) existe(hash string) bool {
	if len(hash) == 0 {
		return false
	}
	for i := range c.arquivosJáProcessados {
		if c.arquivosJáProcessados[i] == hash {
			return true
		}
	}
	return false
}

func filtros() []string {
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
		filtros = append(filtros, "dfp_cia_aberta_"+t+"_con")
		filtros = append(filtros, "dfp_cia_aberta_"+t+"_ind")
		filtros = append(filtros, "itr_cia_aberta_"+t+"_con")
		filtros = append(filtros, "itr_cia_aberta_"+t+"_ind")
	}

	return filtros
}

func urlArquivo(ano int, trimestral bool) string {
	tipo := "DFP"
	if trimestral {
		tipo = "ITR"
	}
	zip := fmt.Sprintf(`%s_cia_aberta_%d.zip`, tipo, ano)
	return `http://dados.cvm.gov.br/dados/CIA_ABERTA/DOC/` + tipo + `/DADOS/` + zip
}

func processarArquivoDFP(ctx context.Context, arquivo Arquivo, results chan<- dominio.Resultado) error {
	fh, err := os.Open(arquivo.path)
	if err != nil {
		return err
	}
	defer fh.Close()

	csv := &csv{sep: ";"}

	stream := transform.NewReader(fh, charmap.ISO8859_1.NewDecoder())
	scanner := bufio.NewScanner(stream)

	empresas := make(map[string][]*cvmDFP)

	for scanner.Scan() {
		linha := scanner.Text()

		dfp, err := csv.carregaDFP(linha)

		if err != nil || ignorarRegistro(dfp) {
			continue
		}

		k := dfp.CNPJ + dfp.Ano + ";" + dfp.Versão
		empresas[k] = append(empresas[k], dfp)
	}

	enviarDFP(empresas, arquivo.hash, results)

	return nil
}

func ignorarRegistro(dfp *cvmDFP) bool {
	if dfp.Meses != 3 && dfp.Meses != 12 {
		progress.Trace("Ignorando registro não trimestral ou anual: %v", dfp)
		return true
	}
	return false
}

// enviarDFP envia os dados de todas as empresas de todos os anos do arquivo
// lido, com base no o mapa empresas[ano]*cvmDFP. Os dados são enviados pelo
// canal criado pelo método Importar.
func enviarDFP(empresas map[string][]*cvmDFP, hash string, results chan<- dominio.Resultado) {
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

		contas := make(map[string][]dominio.Conta)

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

			empresa := dominio.DemonstraçãoFinanceira{
				Empresa: rapina.Empresa{
					CNPJ: registros[0].CNPJ,
					Nome: registros[0].Nome,
				},
				Ano:    a,
				Contas: contas[ano],
			}

			if empresa.Válida() {
				results <- dominio.Resultado{Empresa: &empresa}
			} else {
				results <- dominio.Resultado{Error: ErrDFPInválida}
			}
		}
	} // next k

	results <- dominio.Resultado{Hash: hash}
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
	ErrCabeçalho    = errors.New("cabeçalho")
	ErrFaltaItem    = errors.New("itens faltando")
	ErrDataInválida = errors.New("data inválida")
)

const (
	numItens int = 11 // número de itens (soma dos parâmetros pos___ da struct csv)
)

type csv struct {
	sep           string // separador de campos
	cabeçalhoLido bool

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

func (c *csv) lerCabeçalho(linha string) {
	c.posDtIniExerc = -1 // Este campo não aparece nos dados do balanço patrimonial
	c.cabeçalhoLido = true
	títulos := strings.Split(linha, c.sep)
	for i, t := range títulos {
		switch t {
		case "CNPJ_CIA":
			c.posCnpj = i
		case "DENOM_CIA":
			c.posDenomCia = i
		case "DT_INI_EXERC":
			c.posDtIniExerc = i
		case "DT_FIM_EXERC":
			c.posDtFimExerc = i
		case "VERSAO":
			c.posVersao = i
		case "CD_CONTA":
			c.posCdConta = i
		case "DS_CONTA":
			c.posDsConta = i
		case "GRUPO_DFP":
			c.posGrupoDFP = i
		case "ORDEM_EXERC":
			c.posOrdemExerc = i
		case "VL_CONTA":
			c.posVlConta = i
		case "ESCALA_MOEDA":
			c.posEscalaMoeda = i
		case "MOEDA":
			c.posMoeda = i
		}
	}
}

// carregaDFP transforma uma linha do arquivo DFP em uma estrutura DFP.
//
//		-----------------------
//		Campo: CD_CONTA
//		-----------------------
//			Descrição : Código da conta
//			Domínio   : Numérico
//			Tipo Dados: varchar
//			Tamanho   : 18
//
//		-----------------------
//		Campo: CNPJ_CIA
//		-----------------------
//			Descrição : CNPJ da companhia
//			Domínio   : Alfanumérico
//			Tipo Dados: varchar
//			Tamanho   : 20
//
//		-----------------------
//		Campo: DENOM_CIA
//		-----------------------
//			Descrição : Nome empresarial da companhia
//			Domínio   : Alfanumérico
//			Tipo Dados: varchar
//			Tamanho   : 100
//
//		-----------------------
//		Campo: DS_CONTA
//		-----------------------
//			Descrição : Descrição da conta
//			Domínio   : Alfanumérico
//			Tipo Dados: varchar
//			Tamanho   : 100
//
//		-----------------------
//		Campo: DT_INI_EXERC
//		-----------------------
//			Descrição : Data início do exercício social
//			Domínio   : AAAA-MM-DD
//			Tipo Dados: date
//			Tamanho   : 10
//
//		-----------------------
//		Campo: DT_FIM_EXERC
//		-----------------------
//			Descrição : Data fim do exercício social
//			Domínio   : AAAA-MM-DD
//			Tipo Dados: date
//			Tamanho   : 10
//
//		-----------------------
//		Campo: ESCALA_MOEDA
//		-----------------------
//			Descrição : Escala monetária
//			Domínio   : Alfanumérico
//			Tipo Dados: varchar
//			Tamanho   : 100
//
//		-----------------------
//		Campo: GRUPO_DFP
//		-----------------------
//			Descrição : Nome e nível de agregação da demonstração
//			Domínio   : Alfanumérico
//			Tipo Dados: varchar
//			Tamanho   : 206
//
//		-----------------------
//		Campo: MOEDA
//		-----------------------
//			Descrição : Moeda
//			Domínio   : Alfanumérico
//			Tipo Dados: varchar
//			Tamanho   : 100
//
//		-----------------------
//		Campo: ORDEM_EXERC
//		-----------------------
//			Descrição : Ordem do exercício social
//			Domínio   : Alfanumérico
//			Tipo Dados: varchar
//			Tamanho   : 9
//
//		-----------------------
//		Campo: VERSAO
//		-----------------------
//			Descrição : Versão do documento
//			Domínio   : Numérico
//			Tipo Dados: smallint
//			Precisão  : 5
//			Scale     : 0
//
//		-----------------------
//		Campo: VL_CONTA
//		-----------------------
//			Descrição : Valor da conta
//			Domínio   : Numérico
//			Tipo Dados: decimal
//			Precisão  : 29
//			Scale     : 10
//
func (c *csv) carregaDFP(linha string) (*cvmDFP, error) {
	if !c.cabeçalhoLido {
		c.lerCabeçalho(linha)
		return nil, ErrCabeçalho
	}

	itens := strings.Split(linha, c.sep)
	if len(itens) < numItens {
		return nil, ErrFaltaItem
	}

	dtIni := "" // dado não aparece no BP
	if c.posDtIniExerc >= 0 {
		dtIni = itens[c.posDtIniExerc]
	}
	m, err := meses(dtIni, itens[c.posDtFimExerc])
	if err != nil {
		return nil, err
	}

	vl, err := strconv.ParseFloat(itens[c.posVlConta], 64)
	if err != nil {
		return nil, err
	}

	return &cvmDFP{
		CNPJ:         itens[c.posCnpj],
		Nome:         itens[c.posDenomCia],
		Ano:          itens[c.posDtFimExerc][:4],
		Consolidado:  strings.Contains(itens[c.posGrupoDFP], "onsolidado"),
		Versão:       itens[c.posVersao],
		Código:       itens[c.posCdConta],
		Descr:        itens[c.posDsConta],
		GrupoDFP:     itens[c.posGrupoDFP],
		DataIniExerc: dtIni,
		DataFimExerc: itens[c.posDtFimExerc],
		Meses:        m,
		OrdemExerc:   itens[c.posOrdemExerc],
		Valor:        vl,
		Escala:       escala(itens[c.posEscalaMoeda]),
		Moeda:        moeda(itens[c.posMoeda]),
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
