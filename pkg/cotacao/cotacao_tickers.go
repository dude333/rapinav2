package cotacao

import (
	"archive/zip"
	"encoding/base64"
	"encoding/csv"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"

	ext "github.com/dude333/rapinav2/pkg/infra"
	"github.com/dude333/rapinav2/pkg/progress"
	"github.com/jmoiron/sqlx"
)

type EmissorData struct {
	Ticker string
	Nome   string
	CNPJ   string
}

func AtualizarTickers(db *sqlx.DB) error {
	progress.Status("Iniciando a atualização dos tickers")

	err := ext.CriarTabelas(db, tabelas)
	if err != nil {
		return err
	}

	url, err := getURL()
	if err != nil {
		return err
	}
	progress.Debug("URL: %s", url)
	payload, err := downloadBinaryPayload(url)
	if err != nil {
		return err
	}

	h := ext.Hash(&payload)
	ok, _ := ext.HasHash(db, h)
	progress.Debug("hash: %s, ok: %t", h, ok)
	if ok {
		progress.Warning("Este arquivo de 'tickers' já foi processado anteriormente")
		return nil
	}
	_ = ext.SaveHash(db, h)

	// Extrai o EMISSOR.TXT do zip
	emissorData, err := extractEmissorData(&payload)
	if err != nil {
		return err
	}

	err = insertDataIntoDatabase(db, emissorData)
	if err != nil {
		return err
	}

	progress.Status("Dados inseridos com sucesso!")
	return nil
}

func downloadBinaryPayload(url string) ([]byte, error) {
	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("falha ao baixar arquivo, status code: %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	return body, nil
}

func extractEmissorData(zipContent *[]byte) ([]EmissorData, error) {
	reader, err := zip.NewReader(strings.NewReader(string(*zipContent)), int64(len(*zipContent)))
	if err != nil {
		return nil, err
	}

	var emissorData []EmissorData

	progress.Running("Extraindo arquivo EMISSOR.TXT")
	for _, file := range reader.File {
		if file.Name == "EMISSOR.TXT" {
			rc, err := file.Open()
			if err != nil {
				return nil, err
			}
			defer rc.Close()

			csvReader := csv.NewReader(rc)
			// csvReader.Comma = ';'
			csvData, err := csvReader.ReadAll()
			if err != nil {
				return nil, err
			}

			for _, row := range csvData {
				progress.Spinner()
				if len(row) >= 3 {
					emissorData = append(emissorData, EmissorData{
						Ticker: row[0],
						Nome:   row[1],
						CNPJ:   row[2],
					})
				}
			}
			break
		}
	}
	progress.RunOK()

	return emissorData, nil
}

func insertDataIntoDatabase(db *sqlx.DB, data []EmissorData) (err error) {
	if len(data) < 1000 {
		return fmt.Errorf("quantidade de dados insuficiente (%d)", len(data))
	}
	tx, err := db.Begin()
	if err != nil {
		return err
	}
	defer func() {
		if err != nil {
			err2 := tx.Rollback()
			err = errors.Join(err, err2)
		} else {
			err = tx.Commit()
		}
	}()

	// Limpa tabela antes de inserir novos dados
	if _, err = tx.Exec("DELETE FROM isin"); err != nil {
		return err
	}

	stmt, err := tx.Prepare("INSERT OR IGNORE INTO isin(key, ticker, cnpj, nome) VALUES(?, ?, ?, ?)")
	if err != nil {
		return err
	}
	defer stmt.Close()

	progress.Running("Inserindo dados no banco de dados")
	for _, d := range data {
		progress.Debug("Inserindo: %+v", d)
		_, err = stmt.Exec(d.Ticker, d.Ticker, d.CNPJ, d.Nome)
		if err != nil {
			progress.RunFail()
			return err
		}
		progress.Spinner()
	}
	progress.RunOK()

	return nil
}

func getURL() (string, error) {
	// Step 1: Fetch the JSON object from the specified URL
	url := "https://sistemaswebb3-listados.b3.com.br/isinProxy/IsinCall/GetTextDownload/"
	resp, err := http.Get(url)
	if err != nil {
		fmt.Println("Error fetching data:", err)
		return "", err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read response body: %w", err)
	}

	var obj struct {
		GeralPt struct {
			ID int `json:"id"`
		} `json:"geralPt"`
	}

	if err := json.Unmarshal(body, &obj); err != nil {
		return "", fmt.Errorf("failed to unmarshal JSON: %w", err)
	}

	// Step 2: Transform the ID into base64
	idJSON, err := json.Marshal(obj.GeralPt.ID)
	if err != nil {
		return "", fmt.Errorf("failed to marshal ID to JSON: %w", err)
	}
	base64ID := base64.StdEncoding.EncodeToString(idJSON)

	// Step 3: Use the base64 ID to request the file download
	downloadURL := fmt.Sprintf("https://sistemaswebb3-listados.b3.com.br/isinProxy/IsinCall/GetFileDownload/%s", base64ID)

	return downloadURL, nil
}

// GetTickerPrefix retorna o prefixo do código das ações a partir do CNPJ
func GetTickerPrefix(db *sqlx.DB, cnpj string) (string, error) {
	return getTickerPrefix(db, onlyNums(cnpj))
}

func getTickerPrefix(db *sqlx.DB, cnpj string) (string, error) {
	var ticker string
	err := db.Get(&ticker, "SELECT ticker FROM isin WHERE cnpj = ?", cnpj)
	if err == nil {
		return ticker, nil
	}

	// err = Update(db)
	// if err != nil {
	// 	return "", err
	// }

	err = db.Get(&ticker, "SELECT ticker FROM isin WHERE cnpj = ?", cnpj)
	if err != nil {
		return "", err
	}
	return ticker, nil
}

func onlyNums(str string) string {
	result := make([]rune, 0, len(str))
	for _, r := range str {
		if r >= '0' && r <= '9' {
			result = append(result, r)
		}
	}
	return string(result)
}

func init() {
	tabelas = append(tabelas, []ext.Tabela{
		{
			Nome:   "isin",
			Versão: _ver_,
			Up: `CREATE TABLE IF NOT EXISTS isin (
			key TEXT PRIMARY KEY,
			ticker TEXT,
			cnpj TEXT,	
			nome TEXT
		);`,
			Down: `DROP TABLE IF EXISTS isin;`,
		},
	}...)
}
