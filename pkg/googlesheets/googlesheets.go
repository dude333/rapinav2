// SPDX-FileCopyrightText: 2026 Adriano Prado <dev@dude333.com>
//
// SPDX-License-Identifier: MIT

// Package googlesheets fornece uma interface de alto nível para criação e
// manipulação de documentos do Google Sheets via API REST v4.
//
// A autenticação é feita com OAuth 2.0 usando uma credencial do tipo
// "Aplicativo de desktop": na primeira execução o usuário é direcionado a uma
// URL de consentimento no navegador; o token resultante é armazenado localmente
// para que execuções posteriores sejam não-interativas. O escopo OAuth é
// restrito a "drive.file", portanto a aplicação só enxerga os arquivos que ela
// mesma criou — nenhum outro conteúdo do Google Drive fica visível.
package googlesheets

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"math"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/dude333/rapinav2/pkg/progress"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/drive/v3"
	"google.golang.org/api/googleapi"
	"google.golang.org/api/option"
	"google.golang.org/api/sheets/v4"
)

// pxPerCharWidth é o fator de conversão de "character width units" (unidade
// do Excel/excelize) para pixels (unidade da API do Google Sheets).
// Baseado na fonte padrão Calibri 11pt a 96 DPI: 1 char unit ≈ 7 px.
const pxPerCharWidth = 7.0

// ---------------------------------------------------------------------------
// Configuração e construtores
// ---------------------------------------------------------------------------

// Config contém os caminhos de credencial e cache de token para o OAuth 2.0.
type Config struct {
	// CredentialsFile é o caminho para o arquivo JSON de credenciais OAuth2
	// baixado do Google Cloud Console (tipo: "Aplicativo de desktop").
	// Padrão: "credentials.json".
	CredentialsFile string

	// TokenFile é o caminho onde o token OAuth2 é armazenado após o primeiro
	// login interativo. Padrão: "token.json".
	TokenFile string

	// TokenPort é a porta para o redirect URI do OAuth. Se 0, usa porta automática.
	// Padrão: 0 (porta automática).
	TokenPort int

	// OAuthURL é a URL externa para o redirect URI do OAuth (ex: https://seu-dominio.com).
	// Se vazio, usa localhost. Padrão: "" (vazio, usa localhost).
	OAuthURL string
}

func (c *Config) applyDefaults() {
	if c.CredentialsFile == "" {
		c.CredentialsFile = "credentials.json"
	}
	if c.TokenFile == "" {
		c.TokenFile = "token.json"
	}
}

// New autentica o usuário, cria uma planilha Google Sheets em branco e retorna
// um *Sheet pronto para uso.
func New(ctx context.Context, cfg Config) (*Sheet, error) {
	cfg.applyDefaults()

	sheetsSrv, driveSrv, err := buildServices(ctx, cfg)
	if err != nil {
		return nil, err
	}

	ss, err := sheetsSrv.Spreadsheets.Create(&sheets.Spreadsheet{
		Properties: &sheets.SpreadsheetProperties{Title: "Sem título"},
		Sheets: []*sheets.Sheet{
			{Properties: &sheets.SheetProperties{Title: "Página1", SheetId: 0}},
		},
	}).Context(ctx).Do()
	if err != nil {
		return nil, fmt.Errorf("googlesheets: erro ao criar planilha: %w", err)
	}

	firstID := ss.Sheets[0].Properties.SheetId
	return &Sheet{
		ctx:           ctx,
		srv:           sheetsSrv,
		driveSrv:      driveSrv,
		spreadsheetID: ss.SpreadsheetId,
		sheetID:       firstID,
		sheetName:     ss.Sheets[0].Properties.Title,
		cellsBySheet:  make(map[int64][]pendingCell),
		colCount:      map[int64]int{firstID: defaultSheetCols},
		frozenCols:    make(map[int64]int),
		frozenRows:    make(map[int64]int),
	}, nil
}

// Open autentica o usuário e encapsula uma planilha existente identificada por
// spreadsheetID.
func Open(ctx context.Context, cfg Config, spreadsheetID string) (*Sheet, error) {
	cfg.applyDefaults()

	sheetsSrv, driveSrv, err := buildServices(ctx, cfg)
	if err != nil {
		return nil, err
	}

	ss, err := sheetsSrv.Spreadsheets.Get(spreadsheetID).Context(ctx).Do()
	if err != nil {
		return nil, fmt.Errorf("googlesheets: erro ao buscar planilha %q: %w", spreadsheetID, err)
	}

	out := &Sheet{
		ctx:           ctx,
		srv:           sheetsSrv,
		driveSrv:      driveSrv,
		spreadsheetID: spreadsheetID,
		cellsBySheet:  make(map[int64][]pendingCell),
		colCount:      make(map[int64]int),
		frozenCols:    make(map[int64]int),
		frozenRows:    make(map[int64]int),
	}
	for _, sh := range ss.Sheets {
		cols := int(sh.Properties.GridProperties.ColumnCount)
		if cols == 0 {
			cols = defaultSheetCols
		}
		out.colCount[sh.Properties.SheetId] = cols
		out.frozenCols[sh.Properties.SheetId] = int(sh.Properties.GridProperties.FrozenColumnCount)
		out.frozenRows[sh.Properties.SheetId] = int(sh.Properties.GridProperties.FrozenRowCount)
	}
	if len(ss.Sheets) > 0 {
		out.sheetID = ss.Sheets[0].Properties.SheetId
		out.sheetName = ss.Sheets[0].Properties.Title
	}
	return out, nil
}

// ListFilesInFolder autentica o usuário e retorna uma lista de arquivos na pasta
// especificada do Google Drive. folderPath é o caminho completo da pasta,
// como "/rapina" ou "root" para a pasta raiz.
func ListFilesInFolder(ctx context.Context, cfg Config, folderPath string) ([]*drive.File, error) {
	cfg.applyDefaults()

	_, driveSrv, err := buildServices(ctx, cfg)
	if err != nil {
		return nil, err
	}

	// Find the folder ID
	parentID := "root"
	if folderPath != "" && folderPath != "/" && folderPath != "root" {
		parts := strings.Split(strings.Trim(folderPath, "/"), "/")
		for _, part := range parts {
			if part == "" {
				continue
			}
			query := fmt.Sprintf("name='%s' and mimeType='application/vnd.google-apps.folder' and '%s' in parents and trashed=false", part, parentID)
			list, err := driveSrv.Files.List().Q(query).Fields("files(id)").Context(ctx).Do()
			if err != nil {
				return nil, fmt.Errorf("googlesheets: search folder %q: %w", part, err)
			}
			if len(list.Files) == 0 {
				return nil, fmt.Errorf("googlesheets: folder %q not found", part)
			}
			parentID = list.Files[0].Id
		}
	}

	// List files in the folder
	query := fmt.Sprintf("'%s' in parents and trashed=false", parentID)
	list, err := driveSrv.Files.List().Q(query).Fields("files(id,name,modifiedTime,size,mimeType)").Context(ctx).Do()
	if err != nil {
		return nil, fmt.Errorf("googlesheets: list files in folder: %w", err)
	}

	return list.Files, nil
}

// ---------------------------------------------------------------------------
// Tipos internos
// ---------------------------------------------------------------------------

type cellFormat struct {
	textFormat *sheets.TextFormat
	wrapText   bool
	numberFmt  string
}

type pendingCell struct {
	row, col int
	styleIdx int
	value    any
}

// Sheet é o tipo concreto retornado por New e Open.
type Sheet struct {
	ctx           context.Context
	srv           *sheets.Service
	driveSrv      *drive.Service
	spreadsheetID string
	sheetID       int64
	sheetName     string

	// formats é indexado base-1: styleIdx 0 = sem estilo, 1 = formats[0].
	formats      []cellFormat
	cellsBySheet map[int64][]pendingCell
	requests     []*sheets.Request
	colCount     map[int64]int
	frozenCols   map[int64]int
	frozenRows   map[int64]int
}

// SpreadsheetID retorna o ID do documento Google Sheets.
func (s *Sheet) SpreadsheetID() string { return s.spreadsheetID }

// SpreadsheetURL retorna a URL do navegador para a planilha.
func (s *Sheet) SpreadsheetURL() string {
	return "https://docs.google.com/spreadsheets/d/" + s.spreadsheetID
}

// ---------------------------------------------------------------------------
// Métodos de Sheet
// ---------------------------------------------------------------------------

// NewSheet cria uma nova aba com o nome informado e a torna ativa.
// Na primeira chamada (quando só existe a aba padrão "Página1"/"Sheet1"),
// renomeia-a em vez de criar uma segunda aba vazia.
//
// O nome é truncado a 31 caracteres (limite da API do Google Sheets).
func (s *Sheet) NewSheet(sheetName string) error {
	// Google Sheets limita nomes de aba a 31 caracteres (limite da API).
	// Usamos []rune para truncar corretamente strings UTF-8.
	if len([]rune(sheetName)) > 31 {
		sheetName = string([]rune(sheetName)[:31])
	}

	// Verifica se a aba já é a única existente (criada com New())
	ss, err := s.srv.Spreadsheets.Get(s.spreadsheetID).Context(s.ctx).Do()
	if err != nil {
		return fmt.Errorf("googlesheets: erro ao obter planilha: %w", err)
	}

	// Se só existe uma aba com o nome padrão, renomeia em vez de criar nova
	if len(ss.Sheets) == 1 &&
		(ss.Sheets[0].Properties.Title == "Página1" ||
			ss.Sheets[0].Properties.Title == "Sheet1") {
		_, err := s.srv.Spreadsheets.BatchUpdate(s.spreadsheetID,
			&sheets.BatchUpdateSpreadsheetRequest{
				Requests: []*sheets.Request{{
					UpdateSheetProperties: &sheets.UpdateSheetPropertiesRequest{
						Properties: &sheets.SheetProperties{
							SheetId: ss.Sheets[0].Properties.SheetId,
							Title:   sheetName,
						},
						Fields: "title",
					},
				}},
			}).Context(s.ctx).Do()
		if err != nil {
			return fmt.Errorf("googlesheets: erro ao renomear aba inicial %q: %w", sheetName, err)
		}
		s.sheetID = ss.Sheets[0].Properties.SheetId
		s.sheetName = sheetName
		if _, ok := s.colCount[s.sheetID]; !ok {
			s.colCount[s.sheetID] = defaultSheetCols
		}
		return nil
	}

	// Caso contrário, cria uma nova aba
	resp, err := s.srv.Spreadsheets.BatchUpdate(s.spreadsheetID,
		&sheets.BatchUpdateSpreadsheetRequest{
			Requests: []*sheets.Request{{
				AddSheet: &sheets.AddSheetRequest{
					Properties: &sheets.SheetProperties{Title: sheetName},
				},
			}},
		}).Context(s.ctx).Do()
	if err != nil {
		return fmt.Errorf("googlesheets: erro ao criar aba %q: %w", sheetName, err)
	}

	for _, r := range resp.Replies {
		if r.AddSheet != nil {
			s.sheetID = r.AddSheet.Properties.SheetId
			s.sheetName = r.AddSheet.Properties.Title
			s.colCount[s.sheetID] = defaultSheetCols
			s.frozenCols[s.sheetID] = 0
			s.frozenRows[s.sheetID] = 0
			return nil
		}
	}
	return errors.New("googlesheets: NewSheet: API não retornou resposta AddSheet")
}

// SetZoom é um no-op: a API REST do Google Sheets v4 não expõe zoom.
func (s *Sheet) SetZoom(_ float64) error { return nil }

// defaultSheetCols é o número de colunas criadas por padrão em uma nova aba.
const defaultSheetCols = 26

// SetColWidth define a largura das colunas a partir da coluna A.
// widths[0] → coluna A, widths[1] → coluna B, etc.
//
// As larguras devem estar em "character width units", o mesmo padrão usado
// pelo excelize. A conversão para pixels (exigida pela API do Google Sheets)
// é feita internamente (1 char unit ≈ 7 px), portanto o código chamador
// não precisa se preocupar com a diferença de unidades entre os dois backends.
//
// Exemplo — funciona igual para excelize e googlesheets:
//
//	x.SetColWidth(widths)     // excelize  — recebe char units
//	s.SetColWidth(widths)     // googlesheets — converte internamente para px
func (s *Sheet) SetColWidth(widths []float64) {
	if len(widths) == 0 {
		return
	}

	// Expande a grade se o número de colunas requeridas supera o atual
	current := s.colCount[s.sheetID]
	if len(widths) > current {
		extra := len(widths) - current
		s.requests = append(s.requests, &sheets.Request{
			AppendDimension: &sheets.AppendDimensionRequest{
				SheetId:   s.sheetID,
				Dimension: "COLUMNS",
				Length:    int64(extra),
			},
		})
		s.colCount[s.sheetID] += extra
	}

	for i, w := range widths {
		if w <= 0 {
			continue
		}
		// Converte char units → pixels internamente
		px := int64(math.Round(w * pxPerCharWidth))
		s.requests = append(s.requests, &sheets.Request{
			UpdateDimensionProperties: &sheets.UpdateDimensionPropertiesRequest{
				Range: &sheets.DimensionRange{
					SheetId:    s.sheetID,
					Dimension:  "COLUMNS",
					StartIndex: int64(i),
					EndIndex:   int64(i + 1),
				},
				Properties: &sheets.DimensionProperties{PixelSize: px},
				Fields:     "pixelSize",
			},
		})
	}
}

// FreezePane enfileira o congelamento de linhas/colunas até a célula informada
// em notação A1. Ex.: "B2" congela a primeira linha e a primeira coluna.
func (s *Sheet) FreezePane(cell string) error {
	col, row, err := a1ToIndices(cell)
	if err != nil {
		return fmt.Errorf("googlesheets: FreezePane: %w", err)
	}
	s.frozenCols[s.sheetID] = col
	s.frozenRows[s.sheetID] = row
	s.requests = append(s.requests, &sheets.Request{
		UpdateSheetProperties: &sheets.UpdateSheetPropertiesRequest{
			Fields: "gridProperties.frozenRowCount,gridProperties.frozenColumnCount",
			Properties: &sheets.SheetProperties{
				SheetId: s.sheetID,
				GridProperties: &sheets.GridProperties{
					FrozenRowCount:    int64(row),
					FrozenColumnCount: int64(col),
				},
			},
		},
	})
	return nil
}

// SetFont registra um estilo de texto e retorna o índice base-1.
func (s *Sheet) SetFont(size float64, bold, wrap bool, family ...string) (int, error) {
	tf := &sheets.TextFormat{
		FontSize: int64(size),
		Bold:     bold,
	}
	if len(family) > 0 && family[0] != "" {
		tf.FontFamily = family[0]
	}

	s.formats = append(s.formats, cellFormat{
		textFormat: tf,
		wrapText:   wrap,
	})
	return len(s.formats), nil
}

// SetNumber registra um estilo numérico e retorna o índice base-1.
func (s *Sheet) SetNumber(size float64, bold bool, format string, family ...string) (int, error) {
	tf := &sheets.TextFormat{
		FontSize: int64(size),
		Bold:     bold,
	}
	if len(family) > 0 && family[0] != "" {
		tf.FontFamily = family[0]
	}

	s.formats = append(s.formats, cellFormat{
		textFormat: tf,
		numberFmt:  format,
	})
	return len(s.formats), nil
}

// PrintCell acumula a escrita de value na célula (row, col) com base 1.
// style é o valor retornado por SetFont ou SetNumber (0 = sem estilo).
func (s *Sheet) PrintCell(row, col, style int, value any) {
	s.cellsBySheet[s.sheetID] = append(s.cellsBySheet[s.sheetID],
		pendingCell{row: row - 1, col: col - 1, styleIdx: style, value: value})
}

// RemoveRow enfileira a remoção da linha de índice base-1.
func (s *Sheet) RemoveRow(row int) error {
	s.requests = append(s.requests, &sheets.Request{
		DeleteDimension: &sheets.DeleteDimensionRequest{
			Range: &sheets.DimensionRange{
				SheetId:    s.sheetID,
				Dimension:  "ROWS",
				StartIndex: int64(row - 1),
				EndIndex:   int64(row),
			},
		},
	})
	return nil
}

// RemoveCol enfileira a remoção da coluna de índice base-1.
func (s *Sheet) RemoveCol(col int) error {
	totalAfter := s.colCount[s.sheetID] - 1
	frozen := s.frozenCols[s.sheetID]
	if frozen >= totalAfter {
		frozen = totalAfter - 1
		if frozen < 0 {
			frozen = 0
		}
	}
	if totalAfter-frozen < 1 {
		return fmt.Errorf("googlesheets: RemoveCol: não é possível remover todas as colunas não congeladas da aba %q", s.sheetName)
	}
	s.requests = append(s.requests, &sheets.Request{
		DeleteDimension: &sheets.DeleteDimensionRequest{
			Range: &sheets.DimensionRange{
				SheetId:    s.sheetID,
				Dimension:  "COLUMNS",
				StartIndex: int64(col - 1),
				EndIndex:   int64(col),
			},
		},
	})
	s.colCount[s.sheetID]--
	return nil
}

// findOrCreateFolder finds a folder by name in the given parent, or creates it if it doesn't exist.
func (s *Sheet) findOrCreateFolder(name, parentID string) (string, error) {
	// Search for existing folder
	query := fmt.Sprintf("name='%s' and mimeType='application/vnd.google-apps.folder' and '%s' in parents and trashed=false", name, parentID)
	list, err := s.driveSrv.Files.List().Q(query).Fields("files(id)").Context(s.ctx).Do()
	if err != nil {
		return "", fmt.Errorf("googlesheets: search folder %q: %w", name, err)
	}
	if len(list.Files) > 0 {
		return list.Files[0].Id, nil
	}

	// Create new folder
	folder := &drive.File{
		Name:     name,
		MimeType: "application/vnd.google-apps.folder",
		Parents:  []string{parentID},
	}
	created, err := s.driveSrv.Files.Create(folder).Context(s.ctx).Do()
	if err != nil {
		return "", fmt.Errorf("googlesheets: create folder %q: %w", name, err)
	}
	return created.Id, nil
}

// SaveAs envia todas as operações pendentes e renomeia a planilha no Drive.
// Se um arquivo com o mesmo nome já existir na pasta, adiciona um número de versão
// como "Report (1)", "Report (2)", etc.
func (s *Sheet) SaveAs(name string) error {
	if err := s.flush(); err != nil {
		return err
	}

	// Parse path for folders
	parts := strings.Split(name, "/")
	filename := parts[len(parts)-1]
	parentID := "root"
	for _, part := range parts[:len(parts)-1] {
		if part == "" {
			continue // skip empty parts
		}
		var err error
		parentID, err = s.findOrCreateFolder(part, parentID)
		if err != nil {
			return err
		}
	}

	// Check for existing files with similar names
	q := fmt.Sprintf("name contains '%s' and '%s' in parents and trashed = false", filename, parentID)
	fileList, err := s.driveSrv.Files.List().Q(q).Context(s.ctx).Do()
	if err != nil {
		return fmt.Errorf("googlesheets: SaveAs check existence: %w", err)
	}

	maxVersion := 0
	hasExact := false
	for _, f := range fileList.Files {
		if f.Name == filename {
			hasExact = true
		} else if strings.HasPrefix(f.Name, filename+" (") && strings.HasSuffix(f.Name, ")") {
			// Parse version
			versionStr := strings.TrimPrefix(f.Name, filename+" (")
			versionStr = strings.TrimSuffix(versionStr, ")")
			if v, err := strconv.Atoi(versionStr); err == nil {
				if v > maxVersion {
					maxVersion = v
				}
			}
		}
	}

	newName := filename
	if hasExact || maxVersion > 0 {
		newVersion := maxVersion + 1
		newName = fmt.Sprintf("%s (%d)", filename, newVersion)
	}

	// Update file with new name and parent
	if parentID == "root" {
		// No folders, just update name
		_, err := s.driveSrv.Files.Update(s.spreadsheetID, &drive.File{Name: newName}).
			Context(s.ctx).Do()
		if err != nil {
			return fmt.Errorf("googlesheets: SaveAs %q: %w", name, err)
		}
		return nil
	}

	// Handle folders
	_, err = s.driveSrv.Files.Update(s.spreadsheetID, &drive.File{
		Name: newName,
	}).
		AddParents(parentID).
		RemoveParents("root").
		Context(s.ctx).Do()
	if err != nil {
		return fmt.Errorf("googlesheets: SaveAs %q: %w", name, err)
	}
	return nil
}

// Close envia quaisquer operações pendentes.
func (s *Sheet) Close() error { return s.flush() }

// ---------------------------------------------------------------------------
// Helpers internos de batch
// ---------------------------------------------------------------------------

func (s *Sheet) buildBatchRequests() ([]*sheets.Request, error) {
	for sheetID, cells := range s.cellsBySheet {
		if len(cells) == 0 {
			continue
		}
		reqs, err := s.buildBatchCellRequest(sheetID, cells)
		if err != nil {
			return nil, err
		}
		s.requests = append(s.requests, reqs...)
	}
	s.cellsBySheet = make(map[int64][]pendingCell)

	if len(s.requests) == 0 {
		return nil, nil
	}

	// Reconcilia freeze: descarta requests antigos e re-emite com valores finais
	var filtered []*sheets.Request
	hasDelete := false
	for _, req := range s.requests {
		if req.DeleteDimension != nil {
			hasDelete = true
		}
		if req.UpdateSheetProperties != nil &&
			req.UpdateSheetProperties.Properties != nil &&
			req.UpdateSheetProperties.Properties.GridProperties != nil &&
			(req.UpdateSheetProperties.Properties.GridProperties.FrozenColumnCount != 0 ||
				req.UpdateSheetProperties.Properties.GridProperties.FrozenRowCount != 0) {
			continue
		}
		filtered = append(filtered, req)
	}
	s.requests = filtered

	for sheetID, frozenCols := range s.frozenCols {
		frozenRows := s.frozenRows[sheetID]
		if hasDelete && frozenCols >= s.colCount[sheetID] {
			if s.colCount[sheetID] <= 1 {
				frozenCols = 0
			} else {
				frozenCols = s.colCount[sheetID] - 1
			}
			s.frozenCols[sheetID] = frozenCols
		}
		if frozenCols == 0 && frozenRows == 0 {
			continue
		}
		s.requests = append(s.requests, &sheets.Request{
			UpdateSheetProperties: &sheets.UpdateSheetPropertiesRequest{
				Fields: "gridProperties.frozenRowCount,gridProperties.frozenColumnCount",
				Properties: &sheets.SheetProperties{
					SheetId: sheetID,
					GridProperties: &sheets.GridProperties{
						FrozenRowCount:    int64(frozenRows),
						FrozenColumnCount: int64(frozenCols),
					},
				},
			},
		})
	}

	var appends, writes, deletes []*sheets.Request
	for _, req := range s.requests {
		switch {
		case req.AppendDimension != nil:
			appends = append(appends, req)
		case req.DeleteDimension != nil:
			deletes = append(deletes, req)
		default:
			writes = append(writes, req)
		}
	}
	s.requests = nil

	ordered := make([]*sheets.Request, 0, len(appends)+len(writes)+len(deletes))
	ordered = append(ordered, appends...)
	ordered = append(ordered, writes...)
	ordered = append(ordered, deletes...)
	return ordered, nil
}

func isAllUnfrozenDeletionError(err error) bool {
	if err == nil {
		return false
	}
	msg := strings.ToLower(err.Error())
	if strings.Contains(msg, "não é possível excluir todas as colunas não congeladas") ||
		strings.Contains(msg, "cannot delete all the non") {
		return true
	}
	var gerr *googleapi.Error
	if errors.As(err, &gerr) && gerr.Code == http.StatusBadRequest {
		m := strings.ToLower(gerr.Message)
		if strings.Contains(m, "não é possível excluir todas as colunas não congeladas") ||
			strings.Contains(m, "cannot delete all the non") {
			return true
		}
	}
	return false
}

func (s *Sheet) flush() error {
	ordered, err := s.buildBatchRequests()
	if err != nil {
		return err
	}
	if len(ordered) == 0 {
		return nil
	}
	_, err = s.srv.Spreadsheets.BatchUpdate(s.spreadsheetID,
		&sheets.BatchUpdateSpreadsheetRequest{Requests: ordered}).
		Context(s.ctx).Do()
	if err != nil {
		if isAllUnfrozenDeletionError(err) {
			progress.Warning(fmt.Sprintf("googlesheets: ignorado erro deleteDimension: %v", err))
			return nil
		}
		return fmt.Errorf("googlesheets: flush: %w", err)
	}
	return nil
}

// buildBatchCellRequest empacota células de uma aba em requisições BatchUpdate.
func (s *Sheet) buildBatchCellRequest(sheetID int64, cells []pendingCell) ([]*sheets.Request, error) {
	minRow, minCol := cells[0].row, cells[0].col
	maxRow, maxCol := minRow, minCol
	for _, pc := range cells {
		if pc.row < minRow {
			minRow = pc.row
		}
		if pc.row > maxRow {
			maxRow = pc.row
		}
		if pc.col < minCol {
			minCol = pc.col
		}
		if pc.col > maxCol {
			maxCol = pc.col
		}
	}

	numRows := maxRow - minRow + 1
	numCols := maxCol - minCol + 1

	var reqs []*sheets.Request

	currentCols := s.colCount[sheetID]
	if maxCol+1 > currentCols {
		extra := int64(maxCol + 1 - currentCols)
		reqs = append(reqs, &sheets.Request{
			AppendDimension: &sheets.AppendDimensionRequest{
				SheetId:   sheetID,
				Dimension: "COLUMNS",
				Length:    extra,
			},
		})
		s.colCount[sheetID] = maxCol + 1
	}

	grid := make([][]*sheets.CellData, numRows)
	for i := range grid {
		row := make([]*sheets.CellData, numCols)
		for j := range row {
			row[j] = &sheets.CellData{}
		}
		grid[i] = row
	}

	for _, pc := range cells {
		cd, err := s.buildCellData(pc)
		if err != nil {
			return nil, err
		}
		grid[pc.row-minRow][pc.col-minCol] = cd
	}

	rowData := make([]*sheets.RowData, numRows)
	for i, row := range grid {
		rowData[i] = &sheets.RowData{Values: row}
	}

	reqs = append(reqs, &sheets.Request{
		UpdateCells: &sheets.UpdateCellsRequest{
			Start: &sheets.GridCoordinate{
				SheetId:     sheetID,
				RowIndex:    int64(minRow),
				ColumnIndex: int64(minCol),
			},
			Rows:   rowData,
			Fields: "userEnteredValue,userEnteredFormat",
		},
	})
	return reqs, nil
}

// buildCellData converte um pendingCell no tipo CellData da API.
// styleIdx é base-1: 0 = sem estilo, 1 = formats[0], etc.
func (s *Sheet) buildCellData(pc pendingCell) (*sheets.CellData, error) {
	cd := &sheets.CellData{}

	switch v := pc.value.(type) {
	case string:
		cd.UserEnteredValue = &sheets.ExtendedValue{StringValue: &v}
	case float64:
		cd.UserEnteredValue = &sheets.ExtendedValue{NumberValue: &v}
	case float32:
		f := float64(v)
		cd.UserEnteredValue = &sheets.ExtendedValue{NumberValue: &f}
	case int:
		f := float64(v)
		cd.UserEnteredValue = &sheets.ExtendedValue{NumberValue: &f}
	case int32:
		f := float64(v)
		cd.UserEnteredValue = &sheets.ExtendedValue{NumberValue: &f}
	case int64:
		f := float64(v)
		cd.UserEnteredValue = &sheets.ExtendedValue{NumberValue: &f}
	case bool:
		cd.UserEnteredValue = &sheets.ExtendedValue{BoolValue: &v}
	case nil:
		cd.UserEnteredValue = &sheets.ExtendedValue{}
	default:
		str := fmt.Sprintf("%v", v)
		cd.UserEnteredValue = &sheets.ExtendedValue{StringValue: &str}
	}

	if pc.styleIdx >= 1 && pc.styleIdx <= len(s.formats) {
		f := s.formats[pc.styleIdx-1]
		cf := &sheets.CellFormat{TextFormat: f.textFormat}
		if f.wrapText {
			cf.WrapStrategy = "WRAP"
		}
		if f.numberFmt != "" {
			cf.NumberFormat = &sheets.NumberFormat{
				Type:    "NUMBER",
				Pattern: f.numberFmt,
			}
		}
		cd.UserEnteredFormat = cf
	}
	return cd, nil
}

// ---------------------------------------------------------------------------
// Helper de notação A1
// ---------------------------------------------------------------------------

// a1ToIndices converte um endereço em notação A1 (ex.: "B2") para índices
// base-0 de (coluna, linha).
func a1ToIndices(cell string) (col, row int, err error) {
	cell = strings.ToUpper(strings.TrimSpace(cell))
	if cell == "" {
		return 0, 0, errors.New("endereço de célula vazio")
	}
	i := 0
	colNum := 0
	for i < len(cell) && cell[i] >= 'A' && cell[i] <= 'Z' {
		colNum = colNum*26 + int(cell[i]-'A') + 1
		i++
	}
	if i == 0 {
		return 0, 0, fmt.Errorf("nenhuma letra de coluna em %q", cell)
	}
	if i == len(cell) {
		return 0, 0, fmt.Errorf("nenhum número de linha em %q", cell)
	}
	rowNum, convErr := strconv.Atoi(cell[i:])
	if convErr != nil || rowNum < 1 {
		return 0, 0, fmt.Errorf("número de linha inválido em %q", cell)
	}
	return colNum - 1, rowNum - 1, nil
}

// ---------------------------------------------------------------------------
// Helpers OAuth 2.0
// ---------------------------------------------------------------------------

func buildServices(ctx context.Context, cfg Config) (*sheets.Service, *drive.Service, error) {
	credBytes, err := os.ReadFile(filepath.Clean(cfg.CredentialsFile))
	if err != nil {
		return nil, nil, fmt.Errorf("googlesheets: erro ao ler %q: %w", cfg.CredentialsFile, err)
	}

	oauthCfg, err := google.ConfigFromJSON(credBytes,
		sheets.SpreadsheetsScope,
		drive.DriveFileScope,
	)
	if err != nil {
		return nil, nil, fmt.Errorf("googlesheets: erro ao processar credenciais: %w", err)
	}

	tok, err := tokenFromFile(cfg.TokenFile)
	if err != nil {
		tok, err = tokenFromWeb(ctx, oauthCfg, cfg.TokenPort, cfg.OAuthURL)
		if err != nil {
			return nil, nil, err
		}
		if saveErr := saveToken(cfg.TokenFile, tok); saveErr != nil {
			fmt.Fprintf(os.Stderr,
				"googlesheets: aviso: não foi possível salvar o token em %q: %v\n",
				cfg.TokenFile, saveErr)
		}
	}

	httpClient := oauthCfg.Client(ctx, tok)

	sheetsSrv, err := sheets.NewService(ctx, option.WithHTTPClient(httpClient))
	if err != nil {
		return nil, nil, fmt.Errorf("googlesheets: erro ao criar serviço Sheets: %w", err)
	}

	driveSrv, err := drive.NewService(ctx, option.WithHTTPClient(httpClient))
	if err != nil {
		return nil, nil, fmt.Errorf("googlesheets: erro ao criar serviço Drive: %w", err)
	}

	return sheetsSrv, driveSrv, nil
}

func tokenFromFile(file string) (*oauth2.Token, error) {
	f, err := os.Open(filepath.Clean(file))
	if err != nil {
		return nil, err
	}
	defer f.Close()
	tok := &oauth2.Token{}
	return tok, json.NewDecoder(f).Decode(tok)
}

func saveToken(file string, tok *oauth2.Token) error {
	f, err := os.OpenFile(filepath.Clean(file),
		os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0600)
	if err != nil {
		return err
	}
	defer f.Close()
	return json.NewEncoder(f).Encode(tok)
}

func tokenFromWeb(ctx context.Context, cfg *oauth2.Config, tokenPort int, oauthURL string) (*oauth2.Token, error) {
	var addr string
	if tokenPort > 0 {
		addr = fmt.Sprintf("127.0.0.1:%d", tokenPort)
	} else {
		addr = "127.0.0.1:0"
	}
	listener, err := net.Listen("tcp", addr)
	if err != nil {
		return nil, fmt.Errorf("googlesheets: erro ao iniciar servidor local: %w", err)
	}
	port := listener.Addr().(*net.TCPAddr).Port

	// Determina a URL de redirecionamento para OAuth
	var redirectURL string
	if oauthURL != "" {
		redirectURL = oauthURL
	} else {
		redirectURL = fmt.Sprintf("http://localhost:%d", port)
	}

	cfg.RedirectURL = redirectURL

	state := "state-token"
	authURL := cfg.AuthCodeURL(state, oauth2.AccessTypeOffline)

	codeCh := make(chan string, 1)
	errCh := make(chan error, 1)

	server := &http.Server{
		Handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.URL.Query().Get("state") != state {
				http.Error(w, "estado inválido", http.StatusBadRequest)
				errCh <- errors.New("googlesheets: state OAuth não corresponde")
				return
			}
			code := r.URL.Query().Get("code")
			if code == "" {
				msg := r.URL.Query().Get("error")
				http.Error(w, "autorização negada", http.StatusForbidden)
				errCh <- fmt.Errorf("googlesheets: autorização negada: %s", msg)
				return
			}
			fmt.Fprintln(w, "<html><body><h2>Autorização concluída. Pode fechar esta aba.</h2></body></html>")
			codeCh <- code
		}),
	}

	go func() {
		if serveErr := server.Serve(listener); serveErr != nil && !errors.Is(serveErr, http.ErrServerClosed) {
			errCh <- fmt.Errorf("googlesheets: erro no servidor local: %w", serveErr)
		}
	}()

	fmt.Printf("\nAbra esta URL no seu navegador para autorizar o acesso:\n\n  %s\n\nAguardando autorização...\n", authURL)

	var code string
	select {
	case code = <-codeCh:
	case err = <-errCh:
		server.Shutdown(context.Background()) //nolint:errcheck
		return nil, err
	case <-ctx.Done():
		server.Shutdown(context.Background()) //nolint:errcheck
		return nil, errors.New("googlesheets: contexto cancelado enquanto aguardava autorização")
	}

	server.Shutdown(context.Background()) //nolint:errcheck

	tok, err := cfg.Exchange(ctx, code)
	if err != nil {
		return nil, fmt.Errorf("googlesheets: erro ao trocar código de autorização: %w", err)
	}

	fmt.Println("Autorização bem-sucedida.")
	return tok, nil
}
