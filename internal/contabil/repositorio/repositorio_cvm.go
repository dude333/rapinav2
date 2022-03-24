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
	"strconv"
	"strings"

	"github.com/dude333/rapinav2/internal/contabil/dominio"
	"github.com/dude333/rapinav2/pkg/progress"
	"golang.org/x/text/encoding/charmap"
	"golang.org/x/text/transform"
)

type cvmDFP struct {
	CNPJ        string
	Nome        string // Nome da empresa
	Ano         string
	Consolidado bool
	Versão      string

	Código       string
	Descr        string
	GrupoDFP     string
	DataFimExerc string // AAAA-MM-DD
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
		Consolidado:  strings.Contains(c.GrupoDFP, "onsolidado"),
		Grupo:        grp,
		DataFimExerc: c.DataFimExerc,
		OrdemExerc:   c.OrdemExerc,
		Total: dominio.Dinheiro{
			Valor:  c.Valor,
			Escala: c.Escala,
			Moeda:  c.Moeda,
		},
	}

	return conta
}

// cvm implementa RepositórioImportação. Busca demonstrações financeiras
// no site da CVM.
type cvm struct {
	infra
}

func NovoCVM(dirDados string) dominio.RepositórioImportação {
	return &cvm{
		infra: &localInfra{dirDados: dirDados},
	}
}

// Importar baixa o arquivo de DFPs de todas as empresas de um determinado
// ano do site da CVM.
func (c *cvm) Importar(ctx context.Context, ano int) <-chan dominio.ResultadoImportação {
	results := make(chan dominio.ResultadoImportação)

	go func() {
		defer close(results)

		url, zip, err := arquivoDFP(ano)
		if err != nil {
			results <- dominio.ResultadoImportação{Error: err}
			return
		}

		arquivos, err := c.infra.DownloadAndUnzip(url, zip, Config.Filtros)
		if err != nil {
			results <- dominio.ResultadoImportação{Error: err}
			return
		}
		defer func() {
			_ = c.infra.Cleanup(arquivos)
		}()

		for _, arquivo := range arquivos {
			progress.Running(arquivo)
			// Processa o arquivo e envia o resultado para o canal 'results'
			_ = processarArquivoDFP(ctx, arquivo, results)
			progress.RunOK()
		}

	}()
	return results
}

func arquivoDFP(ano int) (url, zip string, err error) {
	if ano < 2000 || ano > 3000 {
		return "", "", ErrAnoInválidoFn(ano)
	}
	zip = fmt.Sprintf(`dfp_cia_aberta_%d.zip`, ano)
	url = `http://dados.cvm.gov.br/dados/CIA_ABERTA/DOC/DFP/DADOS/` + zip

	return url, zip, nil
}

func processarArquivoDFP(ctx context.Context, arquivo string, results chan<- dominio.ResultadoImportação) error {
	fh, err := os.Open(arquivo)
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
		if err != nil {
			continue
		}

		k := dfp.CNPJ + dfp.Ano + ";" + dfp.Versão
		empresas[k] = append(empresas[k], dfp)
	}

	enviarDFP(empresas, results)

	return nil
}

// enviarDFP envia os dados de todas as empresas de todos os anos do arquivo
// lido, com base no o mapa empresas[ano]*cvmDFP. Os dados são enviados pelo
// canal criado pelo método Importar.
func enviarDFP(empresas map[string][]*cvmDFP, results chan<- dominio.ResultadoImportação) {
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

			empresa := dominio.Empresa{
				CNPJ:         registros[0].CNPJ,
				Nome:         registros[0].Nome,
				Ano:          a,
				ContasAnuais: contas[ano],
			}

			if empresa.Válida() {
				results <- dominio.ResultadoImportação{Empresa: &empresa}
			} else {
				results <- dominio.ResultadoImportação{Error: ErrDFPInválida}
			}
		}
	} // next k
	progress.Status("Linhas processadas: %d", num)
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
	c.cabeçalhoLido = true
	títulos := strings.Split(linha, c.sep)
	for i, t := range títulos {
		switch t {
		case "CNPJ_CIA":
			c.posCnpj = i
		case "DENOM_CIA":
			c.posDenomCia = i
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

	// A validação da Conta é feita pelo modelo, mas aqui é feita para
	// evitar erro ao pegar o ano
	if len(itens[c.posDtFimExerc]) != len("AAAA-MM-DD") {
		return nil, ErrDataInválida
	}

	vl, err := strconv.ParseFloat(itens[c.posVlConta], 64)
	if err != nil {
		return nil, err
	}

	return &cvmDFP{
		CNPJ:         itens[c.posCnpj],
		Nome:         itens[c.posDenomCia],
		Ano:          itens[c.posDtFimExerc][:4],
		Versão:       itens[c.posVersao],
		Código:       itens[c.posCdConta],
		Descr:        itens[c.posDsConta],
		GrupoDFP:     itens[c.posGrupoDFP],
		DataFimExerc: itens[c.posDtFimExerc],
		OrdemExerc:   itens[c.posOrdemExerc],
		Valor:        vl,
		Escala:       escala(itens[c.posEscalaMoeda]),
		Moeda:        moeda(itens[c.posMoeda]),
	}, nil
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
