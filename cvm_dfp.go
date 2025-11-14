package rapina

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"strings"

	"github.com/dude333/rapinav2/pkg/progress"
	"github.com/jmoiron/sqlx"
)

const (
	_dfpInitLen    = 500
	_contasInitLen = 50
)

type DFP map[string]*DFPEmpresa

type DFPEmpresa struct {
	Empresa
	Ano          string
	DataIniExerc string
	Versão       string
	Contas       []Conta
}

// Conta com os dados das Demonstrações Financeiras Padronizadas (DFP) ou
// com as Informações Trimestrais (ITR).
type Conta struct {
	Código       string   // 1, 1.01, 1.02...
	Descr        string   // Descrição
	Grupo        string   // BPA, BPP, DRE, DFC...
	DataIniExerc string   // AAAA-MM-DD
	DataFimExerc string   // AAAA-MM-DD
	OrdemExerc   string   // ÚLTIMO ou PENÚLTIMO
	Total        Dinheiro // $
	Meses        int      // Meses acumulados desde o início do período
	Consolidado  bool     // Individual ou Consolidado
}

func (c *Conta) Válida() bool {
	if c.Meses%3 != 0 {
		progress.Trace("Ignorando registro não multiplo de 3: %v", c)
		return true
	}
	return len(c.Código) > 0 &&
		len(c.Descr) > 0 &&
		len(c.DataFimExerc) == len("AAAA-MM-DD") &&
		(c.OrdemExerc == "ÚLTIMO" ||
			(c.OrdemExerc == "PENÚLTIMO" && strings.HasPrefix(c.DataFimExerc, "2009")))
}

func NewDFP() *DFP {
	dfp := make(DFP, _dfpInitLen)
	return &dfp
}

func keyDFP(empresa *DFPEmpresa) (string, error) {
	key := empresa.CNPJ + empresa.Ano + ";" + empresa.Versão
	if len(key) < (14 + 4 + 1 + 1) {
		return "", fmt.Errorf("keyDFP: empresa inválida")
	}
	return key, nil
}

// // nextKeyDFP retora a chave com a próxima chave:
// // "CNPJANO;1" => "CNPJANO;2"
// func nextKeyDFP(k string) string {
// 	ks := strings.Split(k, ";")
// 	if len(ks) != 2 {
// 		return k
// 	}
// 	ver, err := strconv.Atoi(ks[1])
// 	if err != nil {
// 		return k
// 	}
// 	return ks[0] + strconv.Itoa(ver+1)
// }

// AppendConta adiciona uma conta ao DFP
func (dfp *DFP) AppendConta(e *DFPEmpresa) bool {
	progress.Debug("AppendConta: %#v", e)
	if len(e.Contas) != 1 {
		return false
	}
	conta := e.Contas[0]
	if !conta.Válida() {
		progress.Debug("Conta inválida: %#v", conta)
		return false
	}
	key, err := keyDFP(e)
	if err != nil {
		return false
	}

	progress.Debug("Adicionando conta: %s", key)

	_, exists := (*dfp)[key]
	if !exists {
		(*dfp)[key] = &DFPEmpresa{
			Empresa:      e.Empresa,
			Ano:          e.Ano,
			DataIniExerc: e.DataIniExerc,
			Versão:       e.Versão,
			Contas:       make([]Conta, 0, _contasInitLen),
		}
	}

	(*dfp)[key].Contas = append((*dfp)[key].Contas, conta)

	return true
}

// Salvar salva o DFP no banco de dados
func (dfp *DFP) Salvar(ctx context.Context, db *sqlx.DB) error {
	progress.Status("Salvar: %v", (*dfp))
	for i := range *dfp {
		if ctx.Err() != nil {
			return ctx.Err()
		}
		progress.Status("DFP: %#v", (*dfp)[i].Contas)
	}
	return nil
}

// csvDFP ---------------------------------------------------------------------

var (
	ErrCabeçalhoCsv    = errors.New("cabeçalho")
	ErrFaltaItemCsv    = errors.New("itens faltando")
	ErrDataInválidaCsv = errors.New("data inválida")
)

const (
	numItens int = 11 // número de itens (soma dos parâmetros pos___ da struct csv)
)

type csvDFP struct {
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

func (c *csvDFP) lerCabeçalho(linha string) {
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
//	-----------------------
//	Campo: CD_CONTA
//	-----------------------
//	Descrição : Código da conta
//	Domínio   : Numérico
//	Tipo Dados: varchar
//	Tamanho   : 18
//
//	-----------------------
//	Campo: CNPJ_CIA
//	-----------------------
//	Descrição : CNPJ da companhia
//	Domínio   : Alfanumérico
//	Tipo Dados: varchar
//	Tamanho   : 20
//
//	-----------------------
//	Campo: DENOM_CIA
//	-----------------------
//	Descrição : Nome empresarial da companhia
//	Domínio   : Alfanumérico
//	Tipo Dados: varchar
//	Tamanho   : 100
//
//	-----------------------
//	Campo: DS_CONTA
//	-----------------------
//	Descrição : Descrição da conta
//	Domínio   : Alfanumérico
//	Tipo Dados: varchar
//	Tamanho   : 100
//
//	-----------------------
//	Campo: DT_INI_EXERC
//	-----------------------
//	Descrição : Data início do exercício social
//	Domínio   : AAAA-MM-DD
//	Tipo Dados: date
//	Tamanho   : 10
//
//	-----------------------
//	Campo: DT_FIM_EXERC
//	-----------------------
//	Descrição : Data fim do exercício social
//	Domínio   : AAAA-MM-DD
//	Tipo Dados: date
//	Tamanho   : 10
//
//	-----------------------
//	Campo: ESCALA_MOEDA
//	-----------------------
//	Descrição : Escala monetária
//	Domínio   : Alfanumérico
//	Tipo Dados: varchar
//	Tamanho   : 100
//
//	-----------------------
//	Campo: GRUPO_DFP
//	-----------------------
//	Descrição : Nome e nível de agregação da demonstração
//	Domínio   : Alfanumérico
//	Tipo Dados: varchar
//	Tamanho   : 206
//
//	-----------------------
//	Campo: MOEDA
//	-----------------------
//	Descrição : Moeda
//	Domínio   : Alfanumérico
//	Tipo Dados: varchar
//	Tamanho   : 100
//
//	-----------------------
//	Campo: ORDEM_EXERC
//	-----------------------
//	Descrição : Ordem do exercício social
//	Domínio   : Alfanumérico
//	Tipo Dados: varchar
//	Tamanho   : 9
//
//	-----------------------
//	Campo: VERSAO
//	-----------------------
//	Descrição : Versão do documento
//	Domínio   : Numérico
//	Tipo Dados: smallint
//	Precisão  : 5
//	Scale     : 0
//
//	-----------------------
//	Campo: VL_CONTA
//	-----------------------
//	Descrição : Valor da conta
//	Domínio   : Numérico
//	Tipo Dados: decimal
//	Precisão  : 29
//	Scale     : 10
func (c *csvDFP) carregaDFP(linha string) (*DFPEmpresa, error) {
	if !c.cabeçalhoLido {
		c.lerCabeçalho(linha)
		return nil, ErrCabeçalhoCsv
	}

	itens := strings.Split(linha, c.sep)
	if len(itens) < numItens {
		return nil, ErrFaltaItemCsv
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

	return &DFPEmpresa{
		Empresa: Empresa{
			Nome: itens[c.posDenomCia],
			CNPJ: itens[c.posCnpj],
		},
		Ano:          itens[c.posDtFimExerc][:4],
		DataIniExerc: dtIni,
		Versão:       itens[c.posVersao],
		Contas: []Conta{{
			Código:       itens[c.posCdConta],
			Descr:        itens[c.posDsConta],
			Grupo:        itens[c.posGrupoDFP],
			DataIniExerc: dtIni,
			DataFimExerc: itens[c.posDtFimExerc],
			OrdemExerc:   itens[c.posOrdemExerc],
			Total: Dinheiro{
				Moeda:  moeda(itens[c.posMoeda]),
				Valor:  vl,
				Escala: escala(itens[c.posEscalaMoeda]),
			},
			Meses:       m,
			Consolidado: strings.Contains(itens[c.posGrupoDFP], "onsolidado"),
		}},
	}, nil
}

// meses retorna a diferença em meses entre ini e fim, com a data no formato
// AAAA-MM-DD.
func meses(ini, fim string) (int, error) {
	if len(fim) != 10 || (ini != "" && len(ini) != 10) {
		return 0, ErrDataInválidaCsv
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
		return 0, ErrDataInválidaCsv
	}

	meses := 0

	if anoF != anoI {
		meses = (anoF - anoI) * 12
		meses = meses - mesI + mesF + 1
	} else {
		meses = mesF - mesI + 1
	}

	if meses <= 0 {
		return 0, ErrDataInválidaCsv
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
