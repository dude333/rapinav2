package contabil

import (
	"bufio"
	"context"
	"errors"
	"os"
	"strconv"
	"strings"

	rapina "github.com/dude333/rapinav2"
	"github.com/dude333/rapinav2/pkg/contabil/dominio"
	"github.com/dude333/rapinav2/pkg/progress"
	"golang.org/x/text/encoding/charmap"
	"golang.org/x/text/transform"
)

type cvmFRE struct {
	infra infra
	cfg   *cfg
}

func NovaFRE(configs ...ConfigFn) (*cvmFRE, error) {
	var cvm cvmFRE
	cvm.cfg.loadConfigs(configs...)
	cvm.infra = &localInfra{dirDados: cvm.cfg.dirDados}

	return &cvm, nil
}

// Importar baixa o arquivo de FREs de todas as empresas de um determinado
// ano do site da CVM.
//
// Dentro do arquivo zip, extrair o arquivo com o formato fre_cia_capital_aberto_YYYY.csv
// que contém as colunas listadas abaixo:
/*
Campo: CNPJ_Companhia
-----------------------
   Descrição : CNPJ da companhia
   Domínio   : Alfanumérico
   Tipo Dados: varchar
   Tamanho   : 20

-----------------------
Campo: Data_Referencia
-----------------------
   Descrição : Data de referência do documento
   Domínio   : AAAA-MM-DD
   Tipo Dados: date
   Tamanho   : 10

-----------------------
Campo: Data_Ultima_Assembleia
-----------------------
   Descrição : Data da última assembléia geral de acionistas
   Domínio   : AAAA-MM-DD
   Tipo Dados: date
   Tamanho   : 10

-----------------------
Campo: ID_Documento
-----------------------
   Descrição : Identificador do documento
   Domínio   : Numérico
   Tipo Dados: int
   Precisão  : 10
   Scale     : 0

-----------------------
Campo: Nome_Companhia
-----------------------
   Descrição : Nome da Companhia
   Domínio   : Alfanumérico
   Tipo Dados: varchar
   Tamanho   : 100

-----------------------
Campo: Percentual_Acoes_Ordinarias_Circulacao
-----------------------
   Descrição : Percentual de ações ordinárias em circulação
   Domínio   : Numérico
   Tipo Dados: numeric
   Precisão  : 18
   Scale     : 6

-----------------------
Campo: Percentual_Acoes_Preferenciais_Circulacao
-----------------------
   Descrição : Percentual de ações preferenciais em circulação
   Domínio   : Numérico
   Tipo Dados: numeric
   Precisão  : 18
   Scale     : 6

-----------------------
Campo: Percentual_Total_Acoes_Circulacao
-----------------------
   Descrição : Percentual total de ações em circulação
   Domínio   : Numérico
   Tipo Dados: numeric
   Precisão  : 18
   Scale     : 6

-----------------------
Campo: Quantidade_Acionistas_Investidores_Institucionais
-----------------------
   Descrição : Quantidade de investidores institucionais
   Domínio   : Numérico
   Tipo Dados: numeric
   Precisão  : 18
   Scale     : 0

-----------------------
Campo: Quantidade_Acionistas_PF
-----------------------
   Descrição : Quantidade de acionistas pessoas físicas
   Domínio   : Numérico
   Tipo Dados: numeric
   Precisão  : 18
   Scale     : 0

-----------------------
Campo: Quantidade_Acionistas_PJ
-----------------------
   Descrição : Quantidade de acionistas pessoas jurídicas
   Domínio   : Numérico
   Tipo Dados: numeric
   Precisão  : 18
   Scale     : 0

-----------------------
Campo: Quantidade_Acoes_Ordinarias_Circulacao
-----------------------
   Descrição : Quantidade de ações ordinárias em circulação
   Domínio   : Numérico
   Tipo Dados: numeric
   Precisão  : 18
   Scale     : 0

-----------------------
Campo: Quantidade_Acoes_Preferenciais_Circulacao
-----------------------
   Descrição : Quantidade de ações preferenciais em circulação
   Domínio   : Numérico
   Tipo Dados: numeric
   Precisão  : 18
   Scale     : 0

-----------------------
Campo: Quantidade_Total_Acoes_Circulacao
-----------------------
   Descrição : Quantidade total de ações em circulação
   Domínio   : Numérico
   Tipo Dados: numeric
   Precisão  : 18
   Scale     : 0

-----------------------
Campo: Versao
-----------------------
   Descrição : Versão do documento
   Domínio   : Numérico
   Tipo Dados: smallint
   Precisão  : 5
   Scale     : 0
*/

func (c cvmFRE) existe(hash string) bool {
	if len(hash) == 0 || c.cfg.force {
		return false
	}
	for i := range c.cfg.arquivosJáProcessados {
		if c.cfg.arquivosJáProcessados[i] == hash {
			return true
		}
	}
	return false
}

func (c *cvmFRE) Importar(ctx context.Context, ano int, trimestral bool) <-chan dominio.Resultado {
	results := make(chan dominio.Resultado)

	go func() {
		defer close(results)

		url := urlArquivo(ano, trimestral)

		arquivos, zipHash, err := c.infra.DownloadAndUnzip(url, filtros())
		if err != nil {
			results <- dominio.Resultado{Error: err}
			return
		}

		defer c.infra.Cleanup(arquivos)

		if c.existe(zipHash) {
			progress.Warning("Este arquivo 'fre' já foi processado anteriormente")
			return
		}

		for _, arquivo := range arquivos {
			progress.Running(arquivo.path)

			// Processa o arquivo e envia o resultado para o canal 'results'
			err = processarArquivoFRE(ctx, arquivo, results)
			if err != nil {
				results <- dominio.Resultado{Hash: arquivo.hash}
			}

			progress.RunOK()
		}

		// Grava o hash do zip no banco de dados
		results <- dominio.Resultado{Hash: zipHash}
	}()

	return results
}

type csvFRE struct {
	sep           string // separador de campos
	cabeçalhoLido bool

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

// regFRE é usada para armazenar os dados (linhas) dos arquivos de FRE.
type regFRE struct {
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

func processarArquivoFRE(_ context.Context, arquivo Arquivo, results chan<- dominio.Resultado) error {
	fh, err := os.Open(arquivo.path)
	if err != nil {
		return err
	}
	defer fh.Close()

	csv := &csvFRE{sep: ";"}

	stream := transform.NewReader(fh, charmap.ISO8859_1.NewDecoder())
	scanner := bufio.NewScanner(stream)

	empresas := make(map[string][]*regFRE)

	for scanner.Scan() {
		linha := scanner.Text()

		fre, err := csv.carregaFRE(linha)
		if err != nil {
			continue
		}

		k := fre.CNPJ + ";" + fre.DataRef + ";" + strconv.Itoa(fre.Versao)
		empresas[k] = append(empresas[k], fre)
	}

	enviarFRE(empresas, results)

	return nil
}

func (c *csvFRE) lerCabeçalho(linha string) {
	c.cabeçalhoLido = true
	títulos := strings.Split(linha, c.sep)
	for i, t := range títulos {
		switch t {
		case "CNPJ_Companhia":
			c.posCnpj = i
		case "Nome_Companhia":
			c.posNome = i
		case "Data_Referencia":
			c.posDataRef = i
		case "Data_Ultima_Assembleia":
			c.posDataUltimaAssembleia = i
		case "ID_Documento":
			c.posIDDoc = i
		case "Percentual_Acoes_Ordinarias_Circulacao":
			c.posPctAcoesOrdCirculacao = i
		case "Percentual_Acoes_Preferenciais_Circulacao":
			c.posPctAcoesPrefCirculacao = i
		case "Percentual_Total_Acoes_Circulacao":
			c.posPctTotalAcoesCirculacao = i
		case "Quantidade_Acionistas_Investidores_Institucionais":
			c.posQtdAcionistasInstitucionais = i
		case "Quantidade_Acionistas_PF":
			c.posQtdAcionistasPF = i
		case "Quantidade_Acionistas_PJ":
			c.posQtdAcionistasPJ = i
		case "Quantidade_Acoes_Ordinarias_Circulacao":
			c.posQtdAcoesOrdCirculacao = i
		case "Quantidade_Acoes_Preferenciais_Circulacao":
			c.posQtdAcoesPrefCirculacao = i
		case "Quantidade_Total_Acoes_Circulacao":
			c.posQtdTotalAcoesCirculacao = i
		case "Versao":
			c.posVersao = i
		}
	}
}

func (c *csvFRE) carregaFRE(linha string) (*regFRE, error) {
	if !c.cabeçalhoLido {
		c.lerCabeçalho(linha)
		return nil, errors.New("cabeçalho")
	}

	items := strings.Split(linha, c.sep)

	fre := &regFRE{}

	// Converte os campos numéricos, ignorando erros
	fre.CNPJ = items[c.posCnpj]
	fre.Nome = items[c.posNome]
	fre.DataRef = items[c.posDataRef]
	fre.DataUltimaAssembleia = items[c.posDataUltimaAssembleia]

	// Converte campos numéricos, ignorando erros
	fre.IDDoc, _ = strconv.Atoi(items[c.posIDDoc])
	fre.PctAcoesOrdCirculacao, _ = strconv.ParseFloat(items[c.posPctAcoesOrdCirculacao], 64)
	fre.PctAcoesPrefCirculacao, _ = strconv.ParseFloat(items[c.posPctAcoesPrefCirculacao], 64)
	fre.PctTotalAcoesCirculacao, _ = strconv.ParseFloat(items[c.posPctTotalAcoesCirculacao], 64)
	fre.QtdAcionistasInstitucionais, _ = strconv.Atoi(items[c.posQtdAcionistasInstitucionais])
	fre.QtdAcionistasPF, _ = strconv.Atoi(items[c.posQtdAcionistasPF])
	fre.QtdAcionistasPJ, _ = strconv.Atoi(items[c.posQtdAcionistasPJ])
	fre.QtdAcoesOrdCirculacao, _ = strconv.ParseInt(items[c.posQtdAcoesOrdCirculacao], 10, 64)
	fre.QtdAcoesPrefCirculacao, _ = strconv.ParseInt(items[c.posQtdAcoesPrefCirculacao], 10, 64)
	fre.QtdTotalAcoesCirculacao, _ = strconv.ParseInt(items[c.posQtdTotalAcoesCirculacao], 10, 64)
	fre.Versao, _ = strconv.Atoi(items[c.posVersao])

	return fre, nil
}

// enviarFRE envia os dados de todas as empresas para o canal criado pelo método Importar.
func enviarFRE(empresas map[string][]*regFRE, results chan<- dominio.Resultado) {
	num := 0
	for k := range empresas {
		// Ignora se existir uma versão mais nova
		if _, ok := empresas[próxChaveFRE(k)]; ok {
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
				results <- dominio.Resultado{FRE: &freEmpresa}
				num++
			}
		}
	}

	progress.Debug("Registros processados: %d", num)
}

// próxChaveFRE retora a chave com a próxima versão:
// "CNPJ;DATA;1" => "CNPJ;DATA;2"
func próxChaveFRE(k string) string {
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
