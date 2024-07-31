// SPDX-FileCopyrightText: 2024 Adriano Prado <dev@dude333.com>
//
// SPDX-License-Identifier: MIT

// cvm.go:
//  1. Preparação de arquivos:
//     input: zip
//     output: csv
//  2. Parsing:
//     input: csv
//     output: struct
//  3. Storage:
//     input: struct
//     output: tabela
//  4. Load:
//     input: tabela
//     output: struct
//
// Uso:
// - Importar (preparação de arquivos, parsing, storage)
//   . importar por tipo, anual ou trimestral
//   . tipos diferentes podem ser armazenados na mesma tabela
// - Carregar (load)

package rapina

import (
	"bufio"
	"context"
	"fmt"
	"net/url"
	"os"
	"path"
	"time"

	"github.com/dude333/rapinav2/pkg/infra"
	"github.com/dude333/rapinav2/pkg/progress"
	"github.com/jmoiron/sqlx"
	"golang.org/x/text/encoding/charmap"
	"golang.org/x/text/transform"
)

type CVM struct {
	db       *sqlx.DB
	dirDados string   // Diretório de dados temporários
	hashes   []string // Hashes dos arquivos já processados
	force    bool     // Forçar atualização, mesmo que o dado já exista
}

func NewCVM(db *sqlx.DB, tempDir string, force bool) (*CVM, error) {
	c := &CVM{
		db:       db,
		dirDados: tempDir,
		force:    force,
	}
	if tempDir == "" {
		c.dirDados = os.TempDir()
	} else {
		err := os.MkdirAll(c.dirDados, os.ModePerm)
		if err != nil {
			panic(err)
		}
	}
	return c, nil
}

type CvmType int

const (
	CvmDfp CvmType = iota
	CvmItr
	CvmFre
)

type CvmDataSource interface {
	Salvar(ctx context.Context, db *sqlx.DB) error
}

// ImportarCVMimporta dados da CVM para a base de dados local
func (c *CVM) Importar(ctx context.Context, ano int) error {
	for _, tipo := range []CvmType{CvmDfp} {
		dados, err := c.importarTipo(ctx, tipo, ano)
		if err != nil {
			return err
		}
		err = dados.Salvar(ctx, c.db)
		if err != nil {
			progress.Error(err)
		}
	}
	return nil
}

// importarTipo prepara os dados e os carrega na struct correspondente.
func (c *CVM) importarTipo(ctx context.Context, tipo CvmType, ano int) (CvmDataSource, error) {
	switch tipo {
	case CvmDfp:
		dfp, err := c.importarDFP(ctx, ano, false)
		return dfp, err
	case CvmItr:
		return c.importarDFP(ctx, ano, true)
		// case CvmFre:
		// 	return c.importarFRE(ctx, ano)
	}
	return nil, fmt.Errorf("tipo %d inválido", tipo)
}

func (c CVM) existe(hash string) bool {
	if len(hash) == 0 || c.force {
		return false
	}
	for i := range c.hashes {
		if c.hashes[i] == hash {
			return true
		}
	}
	return false
}

func (c *CVM) addHash(hash string) {
	for i := range c.hashes {
		if c.hashes[i] == hash {
			return
		}
	}
	c.hashes = append(c.hashes, hash)
}

// importarDFP prepara os dados e os carrega na struct DFP
func (c *CVM) importarDFP(ctx context.Context, ano int, trimestre bool) (*DFP, error) {
	// url := urlArquivo(CvmDfp, ano, trimestre)
	// arquivos, zipHash, err := DownloadAndUnzip(url, c.dirDados, filtros())
	// if err != nil {
	// 	return nil, err
	// }
	// if c.existe(zipHash) {
	// 	return nil, fmt.Errorf("este arquivo 'dfp/itr' já foi processado anteriormente")
	// }

	arquivos := []Arquivo{
		{
			path: path.Join(c.dirDados, "dfp.zip"),
			hash: "dfp",
		},
		{
			path: path.Join(c.dirDados, "itr.zip"),
			hash: "itr",
		},
	}

	defer Cleanup(arquivos)

	dfp := NewDFP()
	for _, arq := range arquivos {
		if ctx.Err() != nil {
			return dfp, ctx.Err()
		}
		progress.Running(arq.path)
		err := ProcessarArquivoDFP(ctx, arq, dfp)
		if err != nil {
			progress.RunFailMsg(err.Error())
			continue
		}
		c.addHash(arq.hash)
		progress.RunOK()
	}
	return dfp, nil
}

func ProcessarArquivoDFP(ctx context.Context, arq Arquivo, dfp *DFP) error {
	fh, err := os.Open(arq.path)
	if err != nil {
		return err
	}
	defer fh.Close()

	csv := &csvDFP{sep: ";"}

	stream := transform.NewReader(fh, charmap.ISO8859_1.NewDecoder())
	scanner := bufio.NewScanner(stream)

	for scanner.Scan() {
		if ctx.Err() != nil {
			return ctx.Err()
		}

		linha := scanner.Text()

		conta, err := csv.carregaDFP(linha)
		if err != nil {
			continue
		}

		dfp.AppendConta(conta)
	}

	return nil
}

// FRE ------------------------------------------------------------------------

func (c *CVM) importarFRE(ctx context.Context, ano int) (<-chan CvmDataSource, error) {
	ch := make(chan CvmDataSource)
	go func() {
		defer close(ch)
		for i := 1; i <= 10; i++ {
			if ctx.Err() != nil {
				return
			}
			ch <- &FRE{}
			time.Sleep(time.Second)
		}
	}()
	return ch, nil
}

type FRE struct{}

func (fre *FRE) Salvar(ctx context.Context, db *sqlx.DB) error {
	// Salvar FRE no bando de dados
	progress.Status("FRE")
	return nil
}

// --------------------------------------------------------------------------A-

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
		filtros = append(filtros,
			"dfp_cia_aberta_"+t+"_con",
			"dfp_cia_aberta_"+t+"_ind",
			"itr_cia_aberta_"+t+"_con",
			"itr_cia_aberta_"+t+"_ind",
		)
	}

	return filtros
}

func urlArquivo(tipoDado CvmType, ano int, trimestral bool) string {
	var tipo string
	switch tipoDado {
	case CvmDfp:
		tipo = "DFP"
		if trimestral {
			tipo = "ITR"
		}
	case CvmFre:
		tipo = "FRE"
	}

	zip := fmt.Sprintf(`%s_cia_aberta_%d.zip`, tipo, ano)
	return `http://dados.cvm.gov.br/dados/CIA_ABERTA/DOC/` + tipo + `/DADOS/` + zip
}

// --------------------------------------------------------------------------A-

type Arquivo struct {
	path string
	hash string
}

// DownloadAndUnzip baixa e descompacta o arquivo e retorna a lista de arquivos
// descompactados, com o hash de cada arquivo; retorna também o hash do .zip e
// o erro, se houver.
func DownloadAndUnzip(urlString string, tempDir string, filtros []string) ([]Arquivo, string, error) {
	u, err := url.Parse(urlString)
	if err != nil {
		return []Arquivo{}, "", err
	}
	arquivo := path.Base(u.Path)
	zip := path.Join(tempDir, arquivo)
	arqs, err := infra.DownloadAndUnzip(urlString, zip, filtros)
	if err != nil {
		return []Arquivo{}, "", err
	}

	arquivos := make([]Arquivo, len(arqs))
	for i := range arqs {
		h, _ := infra.FileHash(arqs[i])
		arquivos[i] = Arquivo{
			path: arqs[i],
			hash: h,
		}
	}

	zipHash, err := infra.FileHash(zip)
	os.Remove(zip)

	return arquivos, zipHash, err
}

// Cleanup remove os arquivos
func Cleanup(arqs []Arquivo) []string {
	files := make([]string, len(arqs))
	for i := range arqs {
		files[i] = arqs[i].path
	}
	return infra.Cleanup(files)
}
