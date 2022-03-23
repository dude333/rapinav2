// SPDX-FileCopyrightText: 2021 Adriano Prado <dev@dude333.com>
//
// SPDX-License-Identifier: MIT

package repositorio

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	cotação "github.com/dude333/rapinav2/internal/cotacao/dominio"
	"github.com/dude333/rapinav2/pkg/progress"
	"golang.org/x/text/encoding/charmap"
	"golang.org/x/text/transform"
	"os"
	"strconv"
	"strings"
)

// b3 implementa RepositórioImportaçãoAtivo. Busca a cotação de ativos
// no site da B3.
type b3 struct {
	infra
}

func B3(dirDados string) cotação.RepositórioImportaçãoAtivo {
	return &b3{
		infra: &localInfra{dirDados: dirDados},
	}
}

// Importar baixa o arquivo de cotações de todas as empresas de um determinado
// dia do site da B3.
func (b *b3) Importar(ctx context.Context, dia cotação.Data) <-chan cotação.ResultadoImportaçãoDFP {
	results := make(chan cotação.ResultadoImportaçãoDFP)

	go func() {
		defer close(results)

		url, zip, err := arquivoCotação(dia)
		if err != nil {
			results <- cotação.ResultadoImportaçãoDFP{Error: err}
			return
		}

		arquivos, err := b.infra.DownloadAndUnzip(url, zip, []string{})
		if err != nil {
			results <- cotação.ResultadoImportaçãoDFP{Error: err}
			return
		}
		defer func() {
			_ = b.infra.Cleanup(arquivos)
		}()

		for _, arquivo := range arquivos {
			progress.Running(arquivo)
			b.processarSériesHistóricas(ctx, arquivo, results)
			progress.RunOK()
			select {
			case <-ctx.Done():
				results <- cotação.ResultadoImportaçãoDFP{Error: ctx.Err()}
				return
			default:
			}
		}
	}()

	return results
}

func arquivoCotação(dia cotação.Data) (url, zip string, err error) {
	data := dia.String()
	if len(data) != len("2021-05-03") {
		return "", "", ErrDataInválidaFn(data)
	}
	conv := data[8:10] + data[5:7] + data[0:4] // DDMMAAAA

	zip = fmt.Sprintf(`COTAHIST_D%s.ZIP`, conv)
	url = `http://bvmf.bmfbovespa.com.br/InstDados/SerHist/` + zip

	return url, zip, nil
}

// processarSériesHistóricas lê o arquivo de séries históricas baixado da B3
// e envia os valores do ativo para o canal "result".
func (b *b3) processarSériesHistóricas(ctx context.Context, arquivo string, result chan<- cotação.ResultadoImportaçãoDFP) {
	fh, err := os.Open(arquivo)
	if err != nil {
		result <- cotação.ResultadoImportaçãoDFP{Error: err}
		return
	}
	defer fh.Close()

	stream := transform.NewReader(fh, charmap.ISO8859_1.NewDecoder())
	scanner := bufio.NewScanner(stream)
	for scanner.Scan() {
		l := scanner.Text()
		atv, err := analisarLinha(l)
		if err != nil {
			continue
		}
		select {
		case <-ctx.Done():
			result <- cotação.ResultadoImportaçãoDFP{Error: ctx.Err()}
			return
		default:
			result <- cotação.ResultadoImportaçãoDFP{Ativo: atv}
		}
	}
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
func analisarLinha(linha string) (*cotação.Ativo, error) {
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
	data, err := cotação.NovaData(linha[2:6] + "-" + linha[6:8] + "-" + linha[8:10])
	if err != nil {
		return &cotação.Ativo{}, err
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
			return &cotação.Ativo{}, err
		}
		vals[i] = float64(num) / 100
	}

	const r = "R$"

	return &cotação.Ativo{
		Código:       código,
		Data:         data,
		Abertura:     cotação.Dinheiro{Valor: vals[0], Moeda: r},
		Máxima:       cotação.Dinheiro{Valor: vals[1], Moeda: r},
		Mínima:       cotação.Dinheiro{Valor: vals[2], Moeda: r},
		Encerramento: cotação.Dinheiro{Valor: vals[3], Moeda: r},
		Volume:       vals[4],
	}, nil
}
