// SPDX-FileCopyrightText: 2021 Adriano Prado <dev@dude333.com>
//
// SPDX-License-Identifier: MIT

package repositório

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	domínio "github.com/dude333/rapinav2/internal/cotacao/dominio"
	"golang.org/x/text/encoding/charmap"
	"golang.org/x/text/transform"
	"os"
	"strconv"
	"strings"
)

type b3 struct {
	bd domínio.RepositórioEscritaAtivo
	infra
}

func B3(
	bd domínio.RepositórioEscritaAtivo,
	dirDados string) domínio.RepositórioLeituraAtivo {
	return &b3{
		bd:    bd,
		infra: &localInfra{dirDados: dirDados},
	}
}

//
// Cotação baixa o arquivo de cotações de todas as empresas de um determinado
// dia do site da B3.
//
func (b *b3) Cotação(ctx context.Context, código string, dia domínio.Data) (*domínio.Ativo, error) {
	url, zip, err := arquivoCotação(dia)
	if err != nil {
		return &domínio.Ativo{}, err
	}

	arquivos, err := b.infra.DownloadAndUnzip(url, zip)
	if err != nil {
		return &domínio.Ativo{}, err
	}
	defer func() {
		// _ = b.infra.Cleanup(arquivos)
	}()

	var ativo domínio.Ativo

	for _, arquivo := range arquivos {
		fmt.Println("-", arquivo)
		atv, err := b.processarSériesHistóricas(arquivo, código)
		if err == nil && atv.Código == código {
			ativo = *atv
		}
	}

	if ativo == (domínio.Ativo{}) {
		return &domínio.Ativo{}, ErrAtivoNãoEncontrado
	}

	return &ativo, nil
}

func arquivoCotação(dia domínio.Data) (url, zip string, err error) {
	data := dia.String()
	if len(data) != len("2021-05-03") {
		return "", "", ErrDataInválida(data)
	}
	conv := data[8:10] + data[5:7] + data[0:4] // DDMMAAAA

	zip = fmt.Sprintf(`COTAHIST_D%s.ZIP`, conv)
	url = `http://bvmf.bmfbovespa.com.br/InstDados/SerHist/` + zip

	return url, zip, nil
}

// processarSériesHistóricas lê o arquivo de séries históricas baixado da B3,
// retorna com os valores do ativo selecionado pelo "código" e armazena todos os
// ativos no banco de dados.
func (b *b3) processarSériesHistóricas(arquivo, código string) (*domínio.Ativo, error) {
	fh, err := os.Open(arquivo)
	if err != nil {
		return &domínio.Ativo{}, err
	}
	defer fh.Close()

	var ativo domínio.Ativo

	stream := transform.NewReader(fh, charmap.ISO8859_1.NewDecoder())
	scanner := bufio.NewScanner(stream)
	for scanner.Scan() {
		l := scanner.Text()
		atv, err := analisarLinha(l)
		if err != nil {
			continue
		}

		if atv.Código == código {
			ativo = *atv
		}

		if b.bd != nil {
			_ = b.bd.Salvar(context.Background(), atv)
		}
	}

	return &ativo, nil
}

// parseB3Quote parses the line based on this layout:
// http://www.b3.com.br/data/files/33/67/B9/50/D84057102C784E47AC094EA8/SeriesHistoricas_Layout.pdf
//
//   CAMPO/CONTEÚDO  TIPO E TAMANHO  POS. INIC.	 POS. FINAL
//   TIPREG “01”     N(02)           01          02
//   DATA “AAAAMMDD” N(08)           03          10
//   CODBDI          X(02)           11          12
//   CODNEG          X(12)           13          24
//   TPMERC          N(03)           25          27
//   PREABE          (11)V99         57          69
//   PREMAX          (11)V99         70          82
//   PREMIN          (11)V99         83          95
//   PREULT          (11)V99         109         121
//   QUATOT          N18             153         170
//   VOLTOT          (16)V99         171         188
//
// CODBDI:
//   02 LOTE PADRÃO
//   12 FUNDO IMOBILIÁRIO
//
// TPMERC:
//   010 VISTA
//   020 FRACIONÁRIO
func analisarLinha(linha string) (*domínio.Ativo, error) {
	if len(linha) != 245 {
		return nil, errors.New("linha deve conter 245 bytes")
	}

	tipReg := linha[0:2]
	if tipReg != "01" {
		return nil, fmt.Errorf("registro %s ignorado", tipReg)
	}

	codBDI := linha[10:12]
	if codBDI != "02" && codBDI != "12" {
		return nil, fmt.Errorf("BDI %s ignorado", codBDI)
	}

	tpMerc := linha[24:27]
	if tpMerc != "010" && tpMerc != "020" {
		return nil, fmt.Errorf("tipo de mercado %s ignorado", tpMerc)
	}

	código := strings.TrimSpace(linha[12:24])
	data, err := domínio.NovaData(linha[2:6] + "-" + linha[6:8] + "-" + linha[8:10])
	if err != nil {
		return &domínio.Ativo{}, err
	}

	numRanges := [5]struct {
		i, f int
	}{
		{56, 69},   // PREABE = Abertura
		{69, 82},   // PREMAX = Màxima
		{82, 95},   // PREMIN = Mínima
		{108, 121}, // PREULT = Encerramento
		{170, 188}, // VOLTOT = Volume
	}
	var vals [5]float64
	for i, r := range numRanges {
		num, err := strconv.Atoi(linha[r.i:r.f])
		if err != nil {
			return &domínio.Ativo{}, err
		}
		vals[i] = float64(num) / 100
	}

	const r = "R$"

	return &domínio.Ativo{
		Código:       código,
		Data:         data,
		Abertura:     domínio.Dinheiro{Valor: vals[0], Moeda: r},
		Máxima:       domínio.Dinheiro{Valor: vals[1], Moeda: r},
		Mínima:       domínio.Dinheiro{Valor: vals[2], Moeda: r},
		Encerramento: domínio.Dinheiro{Valor: vals[3], Moeda: r},
		Volume:       vals[4],
	}, nil
}
