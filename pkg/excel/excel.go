package excel

import (
	"fmt"

	"github.com/xuri/excelize/v2"

	"github.com/dude333/rapinav2/pkg/progress"
)

type Excel struct {
	file       *excelize.File
	sheetName  string
	sheetIndex int
}

func New() *Excel {
	return &Excel{
		file:       excelize.NewFile(),
		sheetName:  "Informe Trimestral",
		sheetIndex: 0,
	}
}

func (x *Excel) NewSheet(sheetName string) error {
	n := len(x.file.GetSheetList())
	if n == 1 && x.sheetIndex == 0 {
		err := x.file.SetSheetName(x.file.GetSheetList()[x.sheetIndex], sheetName)
		if err != nil {
			return err
		}
	} else {
		_, err := x.file.NewSheet(sheetName)
		if err != nil {
			return err
		}
	}
	x.sheetIndex++
	x.sheetName = sheetName
	return nil
}

func (x *Excel) SetZoom(zoomScale float64) error {
	return x.file.SetSheetView(x.sheetName, 0, &excelize.ViewOptions{ZoomScale: &zoomScale})
}

func (x *Excel) SetColWidth(widths []float64) {
	for i := range widths {
		col := num2name(i + 1)
		_ = x.file.SetColWidth(x.sheetName, col, col, widths[i])
	}
}

func (x *Excel) FreezePane(cell string) error {
	row, col, err := excelize.CellNameToCoordinates(cell)
	if err != nil {
		return err
	}
	return x.file.SetPanes(x.sheetName, &excelize.Panes{
		Freeze:      true,
		Split:       false,
		XSplit:      row - 1,
		YSplit:      col - 1,
		TopLeftCell: cell,
		ActivePane:  "bottomRight",
		Panes: []excelize.PaneOptions{
			{SQRef: cell, ActiveCell: cell, Pane: "bottomRight"},
		},
	})
}

func (x *Excel) SetFont(size float64, bold bool) (int, error) {
	return x.file.NewStyle(&excelize.Style{
		Font: &excelize.Font{
			Size: size,
			Bold: bold,
		},
	})
}

func (x *Excel) SetNumber(size float64, bold bool, format string) (int, error) {
	return x.file.NewStyle(&excelize.Style{
		CustomNumFmt: &format,
		Font: &excelize.Font{
			Size: size,
			Bold: bold,
		},
	})
}

func (x *Excel) PrintCell(row, col int, style int, value interface{}) {
	_ = x.file.SetCellValue(x.sheetName, cell(row, col), value)
	_ = x.file.SetCellStyle(x.sheetName, cell(row, col), cell(row, col), style)
}

// RemoveCol removes columns starting with 1 (=column"A")
func (x *Excel) RemoveCol(col int) error {
	return x.file.RemoveCol(x.sheetName, num2name(col))
}

func cell(row, col int) string {
	return fmt.Sprintf("%s%d", num2name(col), row)
}

func num2name(col int) string {
	n, _ := excelize.ColumnNumberToName(col)
	return n
}

func (x *Excel) SaveAs(name string) error {
	return x.file.SaveAs(name)
}

func (x *Excel) Close() error {
	return x.file.Close()
}

func StringWidth(str string) float64 {
	var width float64 = 0.0
	for _, ch := range str {
		width += charWidth(ch)
	}
	return width
}

// charWidth foi criado com este script em Python:
// from PIL import ImageFont
// font = ImageFont.truetype("calibri.ttf", 11)
// for char in "0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyzÀÁÂÃÇÉÊÍÓÔÕÚàáâãçéêíóôõú.-_ /(),":
//
//	width = font.getlength(char)
//	print(f"\'{char}\': {width},")
func charWidth(ch rune) float64 {
	keys := map[rune]float64{
		'0': 5.0,
		'1': 5.0,
		'2': 5.0,
		'3': 5.0,
		'4': 5.0,
		'5': 5.0,
		'6': 5.0,
		'7': 5.0,
		'8': 5.0,
		'9': 5.0,
		'A': 6.0,
		'B': 5.0,
		'C': 5.0,
		'D': 6.0,
		'E': 5.0,
		'F': 5.0,
		'G': 6.0,
		'H': 6.0,
		'I': 3.0,
		'J': 3.0,
		'K': 5.0,
		'L': 4.0,
		'M': 9.0,
		'N': 6.0,
		'O': 7.0,
		'P': 5.0,
		'Q': 7.0,
		'R': 5.0,
		'S': 5.0,
		'T': 5.0,
		'U': 6.0,
		'V': 6.0,
		'W': 9.0,
		'X': 5.0,
		'Y': 5.0,
		'Z': 5.0,
		'a': 5.0,
		'b': 5.0,
		'c': 4.0,
		'd': 5.0,
		'e': 5.0,
		'f': 3.0,
		'g': 5.0,
		'h': 5.0,
		'i': 2.0,
		'j': 2.0,
		'k': 5.0,
		'l': 2.0,
		'm': 8.0,
		'n': 5.0,
		'o': 5.0,
		'p': 5.0,
		'q': 5.0,
		'r': 3.0,
		's': 4.0,
		't': 3.0,
		'u': 5.0,
		'v': 5.0,
		'w': 7.0,
		'x': 4.0,
		'y': 5.0,
		'z': 4.0,
		'À': 6.0,
		'Á': 6.0,
		'Â': 6.0,
		'Ã': 6.0,
		'Ç': 5.0,
		'É': 5.0,
		'Ê': 5.0,
		'Í': 3.0,
		'Ó': 7.0,
		'Ô': 7.0,
		'Õ': 7.0,
		'Ú': 6.0,
		'à': 5.0,
		'á': 5.0,
		'â': 5.0,
		'ã': 5.0,
		'ç': 4.0,
		'é': 5.0,
		'ê': 5.0,
		'í': 2.0,
		'ó': 5.0,
		'ô': 5.0,
		'õ': 5.0,
		'ú': 5.0,
		'.': 3.0,
		'-': 3.0,
		'_': 5.0,
		' ': 2.0,
		'/': 4.0,
		'(': 3.0,
		')': 3.0,
		',': 3.0,
	}
	width, ok := keys[ch]
	if !ok {
		progress.Trace("charWidth: Caractere '%c' não encontrado na tabela", ch)
		return 1.4
	}
	return width / 5.2
}
