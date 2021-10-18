// SPDX-FileCopyrightText: 2021 Adriano Prado <dev@dude333.com>
//
// SPDX-License-Identifier: MIT

package repositório

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	contábil "github.com/dude333/rapinav2/internal/contabil/dominio"
	"golang.org/x/text/encoding/charmap"
	"golang.org/x/text/transform"
	"os"
	"strconv"
	"strings"
)

type cvmDFP struct {
	CNPJ   string
	Nome   string // Nome da empresa
	Ano    string
	Versão string

	Código       string
	Descr        string
	GrupoDFP     string
	DataFimExerc string // AAAA-MM-DD
	OrdemExerc   string // ÚLTIMO ou PENÚLTIMO
	Valor        float64
	Escala       int
	Moeda        string
}

// cvm implemente RepositórioImportaçãoDFP. Busca demonstrações financeiras
// no site da CVM.
type cvm struct {
	infra
}

func NovoCVM(dirDados string) contábil.RepositórioImportaçãoDFP {
	return &cvm{
		infra: &localInfra{dirDados: dirDados},
	}
}

// Importar baixa o arquivo de DFPs de todas as empresas de um determinado
// ano do site da CVM.
func (c *cvm) Importar(ctx context.Context, ano int) <-chan contábil.ResultadoImportação {
	results := make(chan contábil.ResultadoImportação)

	go func() {
		defer close(results)

		url, zip, err := arquivoDFP(ano)
		if err != nil {
			results <- contábil.ResultadoImportação{Error: err}
			return
		}

		arquivos, err := c.infra.DownloadAndUnzip(url, zip)
		if err != nil {
			results <- contábil.ResultadoImportação{Error: err}
			return
		}
		defer func() {
			_ = c.infra.Cleanup(arquivos)
		}()

		for _, arquivo := range arquivos {
			fmt.Println("-", arquivo)
			_ = c.processarArquivoDFP(ctx, arquivo, results)
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

func (c *cvm) processarArquivoDFP(ctx context.Context, arquivo string, results chan<- contábil.ResultadoImportação) error {
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

		dfp, err := csv.unmarshalDFP(linha)
		if err != nil {
			fmt.Println("* ", err)
			continue
		}

		k := dfp.CNPJ + dfp.Ano + ";" + dfp.Versão
		empresas[k] = append(empresas[k], dfp)
	}

	paraDFP(empresas, results)

	return nil
}

func paraDFP(empresas map[string][]*cvmDFP, results chan<- contábil.ResultadoImportação) {
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

		contas := make(map[string][]contábil.Conta)

		for _, reg := range registros {
			c := contábil.Conta{
				Código:       reg.Código,
				Descr:        reg.Descr,
				GrupoDFP:     reg.GrupoDFP,
				DataFimExerc: reg.DataFimExerc,
				OrdemExerc:   reg.OrdemExerc,
				Total: contábil.Dinheiro{
					Valor:  reg.Valor,
					Escala: reg.Escala,
					Moeda:  reg.Moeda,
				},
			}
			if c.Válida() {
				contas[reg.Ano] = append(contas[reg.Ano], c)
			}
			num++
		}

		for ano := range contas {
			a, err := strconv.Atoi(ano)
			if err != nil {
				continue
			}

			dfp := contábil.DFP{
				CNPJ:   registros[0].CNPJ,
				Nome:   registros[0].Nome,
				Ano:    a,
				Contas: contas[ano],
			}

			if dfp.Válida() {
				fmt.Println(">", dfp)
				results <- contábil.ResultadoImportação{DFP: &dfp}
			} else {
				results <- contábil.ResultadoImportação{Error: ErrDFPInválida}
			}
		}
	} // next k
	fmt.Println("- Linhas processadas:", num)
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

type csv struct {
	sep     string // separador de campos
	títulos map[string]int
}

func (c *csv) lerCabeçalho(linha string) {
	títulos := strings.Split(linha, c.sep)
	c.títulos = make(map[string]int, len(títulos))
	for i, t := range títulos {
		c.títulos[t] = i
	}
}

// unmarshalDFP transforma uma linha do arquivo DFP em uma estrutura DFP.
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
func (c *csv) unmarshalDFP(linha string) (*cvmDFP, error) {
	if len(c.títulos) == 0 {
		c.lerCabeçalho(linha)
		return nil, ErrCabeçalho
	}

	itens := strings.Split(linha, c.sep)
	if len(itens) < len(c.títulos) {
		return nil, ErrFaltaItem
	}

	// A validação da Conta é feita pelo modelo, mas aqui é feita para
	// evitar erro ao pegar o ano
	if len(itens[c.títulos["DT_FIM_EXERC"]]) != len("AAAA-MM-DD") {
		return nil, ErrDataInválida
	}

	vl, err := strconv.ParseFloat(itens[c.títulos["VL_CONTA"]], 64)
	if err != nil {
		return nil, err
	}

	return &cvmDFP{
		CNPJ:         itens[c.títulos["CNPJ"]],
		Nome:         itens[c.títulos["DENOM_CIA"]],
		Ano:          itens[c.títulos["DT_FIM_EXERC"]][:4],
		Versão:       itens[c.títulos["VERSAO"]],
		Código:       itens[c.títulos["CD_CONTA"]],
		Descr:        itens[c.títulos["DS_CONTA"]],
		GrupoDFP:     itens[c.títulos["GRUPO_DFP"]],
		DataFimExerc: itens[c.títulos["DT_FIM_EXERC"]],
		OrdemExerc:   itens[c.títulos["ORDEM_EXERC"]],
		Valor:        vl,
		Escala:       escala(itens[c.títulos["ESCALA_MOEDA"]]),
		Moeda:        moeda(itens[c.títulos["MOEDA"]]),
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
