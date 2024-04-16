package excel

import (
	"github.com/xuri/excelize/v2"
)

type Excel struct {
	file       *excelize.File
	sheetName  string
	sheetIndex int
}

func New() *Excel {
	file := excelize.NewFile()
	_ = file.SetDocProps(&excelize.DocProperties{
		Creator:     "Rapina",
		Description: "https://github.com/dude333/rapinav2",
		Version:     "2",
	})
	return &Excel{
		file:       file,
		sheetName:  "rapina",
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

func (x *Excel) SetFont(size float64, bold, wrap bool) (int, error) {
	return x.file.NewStyle(&excelize.Style{
		Font: &excelize.Font{
			Size: size,
			Bold: bold,
		},
		Alignment: &excelize.Alignment{
			WrapText: wrap,
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

func (x *Excel) PrintCell(row, col, style int, value interface{}) {
	_ = x.file.SetCellValue(x.sheetName, cell(row, col), value)
	_ = x.file.SetCellStyle(x.sheetName, cell(row, col), cell(row, col), style)
}

// RemoveRow removes rows starting with 1
func (x *Excel) RemoveRow(row int) error {
	return x.file.RemoveRow(x.sheetName, row)
}

// RemoveCol removes columns starting with 1 (=column"A")
func (x *Excel) RemoveCol(col int) error {
	return x.file.RemoveCol(x.sheetName, num2name(col))
}

func (x *Excel) SaveAs(name string) error {
	return x.file.SaveAs(name)
}

func (x *Excel) Close() error {
	return x.file.Close()
}
