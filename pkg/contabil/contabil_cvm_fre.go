// SPDX-FileCopyrightText: 2021 Adriano Prado <dev@dude333.com>
//
// SPDX-License-Identifier: MIT

package contabil

import (
	"context"
	"encoding/csv"
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"

	rapina "github.com/dude333/rapinav2"
	"github.com/dude333/rapinav2/pkg/contabil/dominio"
	"github.com/dude333/rapinav2/pkg/progress"
	"golang.org/x/text/encoding/charmap"
	"golang.org/x/text/transform"
)

func NovaFRE(configs ...Option) (CVM, error) {
	var cvm CVMImporter
	cvm.cfg = &cfg{}
	if err := cvm.cfg.apply(configs...); err != nil {
		return nil, err
	}
	cvm.nome = "fre"
	cvm.infra = &LocalInfra{dirDados: cvm.cfg.tempDir}
	cvm.url = urlArquivoFRE
	cvm.filtros = filtrosFRE
	cvm.processar = processarArquivoFRE

	return &cvm, nil
}

func filtrosFRE() []string {
	return []string{"fre_cia_aberta_distribuicao_capital_social"}
}

func urlArquivoFRE(ano int, _ bool) string {
	zip := fmt.Sprintf(`fre_cia_aberta_%d.zip`, ano)
	return `http://dados.cvm.gov.br/dados/CIA_ABERTA/DOC/FRE/DADOS/` + zip
}

// FRERecord é usada para armazenar os dados (linhas) dos arquivos de FRE.
type FRERecord struct {
	CNPJ                        string
	Nome                        string // Nome da empresa
	DataRef                     string
	DataUltimaAssembleia        string
	IDDoc                       int
	PctAcoesOrdCirculacao       float64
	PctAcoesPrefCirculacao      float64
	PctTotalAcoesCirculacao     float64
	QtdAcionistasInstitucionais int
	QtdAcionistasPF             int
	QtdAcionistasPJ             int
	QtdAcoesOrdCirculacao       int64
	QtdAcoesPrefCirculacao      int64
	QtdTotalAcoesCirculacao     int64
	Versao                      int
}

func processarArquivoFRE(_ context.Context, arquivo Arquivo, results chan<- dominio.ImportResult) error {
	fh, err := os.Open(arquivo.path)
	if err != nil {
		return err
	}
	defer func() { _ = fh.Close() }()

	leitorCSV := csv.NewReader(transform.NewReader(fh, charmap.ISO8859_1.NewDecoder()))
	leitorCSV.Comma = ';'
	leitorCSV.LazyQuotes = true

	cabeçalho, err := leitorCSV.Read()
	if err != nil {
		return err
	}

	parser := NewFREParser(cabeçalho)
	empresas := make(map[string][]*FRERecord)

	for {
		linha, err := leitorCSV.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			continue
		}

		fre, err := parser.parseFRE(linha)
		if err != nil {
			continue
		}

		k := fre.CNPJ + ";" + fre.DataRef + ";" + strconv.Itoa(fre.Versao)
		empresas[k] = append(empresas[k], fre)
	}

	sendFRE(empresas, results)

	return nil
}

type parserFRE struct {
	posCnpj                        int
	posNome                        int
	posDataRef                     int
	posDataUltimaAssembleia        int
	posIDDoc                       int
	posPctAcoesOrdCirculacao       int
	posPctAcoesPrefCirculacao      int
	posPctTotalAcoesCirculacao     int
	posQtdAcionistasInstitucionais int
	posQtdAcionistasPF             int
	posQtdAcionistasPJ             int
	posQtdAcoesOrdCirculacao       int
	posQtdAcoesPrefCirculacao      int
	posQtdTotalAcoesCirculacao     int
	posVersao                      int
}

func NewFREParser(cabeçalho []string) *parserFRE {
	p := &parserFRE{}
	for i, t := range cabeçalho {
		switch t {
		case "CNPJ_Companhia":
			p.posCnpj = i
		case "Nome_Companhia":
			p.posNome = i
		case "Data_Referencia":
			p.posDataRef = i
		case "Data_Ultima_Assembleia":
			p.posDataUltimaAssembleia = i
		case "ID_Documento":
			p.posIDDoc = i
		case "Percentual_Acoes_Ordinarias_Circulacao":
			p.posPctAcoesOrdCirculacao = i
		case "Percentual_Acoes_Preferenciais_Circulacao":
			p.posPctAcoesPrefCirculacao = i
		case "Percentual_Total_Acoes_Circulacao":
			p.posPctTotalAcoesCirculacao = i
		case "Quantidade_Acionistas_Investidores_Institucionais":
			p.posQtdAcionistasInstitucionais = i
		case "Quantidade_Acionistas_PF":
			p.posQtdAcionistasPF = i
		case "Quantidade_Acionistas_PJ":
			p.posQtdAcionistasPJ = i
		case "Quantidade_Acoes_Ordinarias_Circulacao":
			p.posQtdAcoesOrdCirculacao = i
		case "Quantidade_Acoes_Preferenciais_Circulacao":
			p.posQtdAcoesPrefCirculacao = i
		case "Quantidade_Total_Acoes_Circulacao":
			p.posQtdTotalAcoesCirculacao = i
		case "Versao":
			p.posVersao = i
		}
	}
	return p
}

func (p *parserFRE) parseFRE(itens []string) (*FRERecord, error) {
	var err error
	fre := &FRERecord{}

	fre.CNPJ = itens[p.posCnpj]
	fre.Nome = itens[p.posNome]
	fre.DataRef = itens[p.posDataRef]
	fre.DataUltimaAssembleia = itens[p.posDataUltimaAssembleia]

	fre.IDDoc, err = strconv.Atoi(itens[p.posIDDoc])
	if err != nil {
		progress.ErrorMsg("erro ao converter IDDoc: %v", err)
	}
	fre.PctAcoesOrdCirculacao, err = strconv.ParseFloat(itens[p.posPctAcoesOrdCirculacao], 64)
	if err != nil {
		progress.ErrorMsg("erro ao converter PctAcoesOrdCirculacao: %v", err)
	}
	fre.PctAcoesPrefCirculacao, err = strconv.ParseFloat(itens[p.posPctAcoesPrefCirculacao], 64)
	if err != nil {
		progress.ErrorMsg("erro ao converter PctAcoesPrefCirculacao: %v", err)
	}
	fre.PctTotalAcoesCirculacao, err = strconv.ParseFloat(itens[p.posPctTotalAcoesCirculacao], 64)
	if err != nil {
		progress.ErrorMsg("erro ao converter PctTotalAcoesCirculacao: %v", err)
	}
	fre.QtdAcionistasInstitucionais, err = strconv.Atoi(itens[p.posQtdAcionistasInstitucionais])
	if err != nil {
		progress.ErrorMsg("erro ao converter QtdAcionistasInstitucionais: %v", err)
	}
	fre.QtdAcionistasPF, err = strconv.Atoi(itens[p.posQtdAcionistasPF])
	if err != nil {
		progress.ErrorMsg("erro ao converter QtdAcionistasPF: %v", err)
	}
	fre.QtdAcionistasPJ, err = strconv.Atoi(itens[p.posQtdAcionistasPJ])
	if err != nil {
		progress.ErrorMsg("erro ao converter QtdAcionistasPJ: %v", err)
	}
	fre.QtdAcoesOrdCirculacao, err = strconv.ParseInt(itens[p.posQtdAcoesOrdCirculacao], 10, 64)
	if err != nil {
		progress.ErrorMsg("erro ao converter QtdAcoesOrdCirculacao: %v", err)
	}
	fre.QtdAcoesPrefCirculacao, err = strconv.ParseInt(itens[p.posQtdAcoesPrefCirculacao], 10, 64)
	if err != nil {
		progress.ErrorMsg("erro ao converter QtdAcoesPrefCirculacao: %v", err)
	}
	fre.QtdTotalAcoesCirculacao, err = strconv.ParseInt(itens[p.posQtdTotalAcoesCirculacao], 10, 64)
	if err != nil {
		progress.ErrorMsg("erro ao converter QtdTotalAcoesCirculacao: %v", err)
	}
	fre.Versao, err = strconv.Atoi(itens[p.posVersao])
	if err != nil {
		progress.ErrorMsg("erro ao converter Versao: %v", err)
	}

	return fre, nil
}

// sendFRE envia os dados de todas as empresas para o canal criado pelo método Importar.
func sendFRE(empresas map[string][]*FRERecord, results chan<- dominio.ImportResult) {
	num := 0
	for k := range empresas {
		// Ignora se existir uma versão mais nova
		if _, ok := empresas[nextFREkey(k)]; ok {
			continue
		}

		registros := empresas[k]
		if len(registros) == 0 {
			continue
		}

		for _, reg := range registros {
			freEmpresa := dominio.FreDistribCapital{
				Empresa: rapina.Empresa{
					CNPJ: reg.CNPJ,
					Nome: reg.Nome,
				},
				DataRef:                 reg.DataRef,
				DataUltimaAssembleia:    reg.DataUltimaAssembleia,
				IDDoc:                   reg.IDDoc,
				PctAcoesOrdCirculacao:   reg.PctAcoesOrdCirculacao,
				PctAcoesPrefCirculacao:  reg.PctAcoesPrefCirculacao,
				PctTotalAcoesCirculacao: reg.PctTotalAcoesCirculacao,
				QtdAcionistasInst:       reg.QtdAcionistasInstitucionais,
				QtdAcionistasPF:         reg.QtdAcionistasPF,
				QtdAcionistasPJ:         reg.QtdAcionistasPJ,
				QtdAcoesOrdCirculacao:   reg.QtdAcoesOrdCirculacao,
				QtdAcoesPrefCirculacao:  reg.QtdAcoesPrefCirculacao,
				QtdTotalAcoesCirculacao: reg.QtdTotalAcoesCirculacao,
			}

			if freEmpresa.Válida() {
				results <- dominio.ImportResult{FRE: &freEmpresa}
				num++
			}
		}
	}

	progress.Debug("Registros processados: %d", num)
}

// nextFREkey retora a chave com a próxima versão:
// "CNPJ;DATA;1" => "CNPJ;DATA;2"
func nextFREkey(k string) string {
	ks := strings.Split(k, ";")
	if len(ks) != 3 {
		return k
	}
	ver, err := strconv.Atoi(ks[2])
	if err != nil {
		return k
	}
	return ks[0] + ";" + ks[1] + ";" + strconv.Itoa(ver+1)
}
