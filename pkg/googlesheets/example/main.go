// Command example demonstrates the googlesheets package by creating a simple
// expense-report spreadsheet in the authenticated user's Google Drive.
package main

import (
	"context"
	"fmt"
	"log"

	"github.com/dude333/rapinav2/pkg/googlesheets"
)

func main() {
	ctx := context.Background()

	// credentials.json must be in the current directory (see README.md).
	ss, err := googlesheets.New(ctx, googlesheets.Config{
		CredentialsFile: "credentials.json",
		TokenFile:       "token.json",
	})
	if err != nil {
		log.Fatalf("googlesheets.New: %v", err)
	}
	defer ss.Close()

	// --- Style definitions ---
	headerStyle, _ := ss.SetFont(12, true, false)
	bodyStyle, _ := ss.SetFont(11, false, false)
	moneyStyle, _ := ss.SetNumber(11, false, `"$"#,##0.00`)

	// --- Column widths (pixels) ---
	ss.SetColWidth([]float64{200, 120, 120, 160})

	// --- Freeze the header row ---
	if err := ss.FreezePane("A2"); err != nil {
		log.Fatalf("FreezePane: %v", err)
	}

	// --- Header row (row 0) ---
	ss.PrintCell(0, 0, headerStyle, "Description")
	ss.PrintCell(0, 1, headerStyle, "Date")
	ss.PrintCell(0, 2, headerStyle, "Category")
	ss.PrintCell(0, 3, headerStyle, "Amount")

	// --- Data rows ---
	rows := []struct {
		desc, date, cat string
		amount          float64
	}{
		{"Office supplies", "2025-01-10", "Admin", 45.99},
		{"Team lunch", "2025-01-12", "Entertainment", 128.50},
		{"Cloud hosting", "2025-01-15", "IT", 299.00},
	}

	for i, r := range rows {
		row := i + 1
		ss.PrintCell(row, 0, bodyStyle, r.desc)
		ss.PrintCell(row, 1, bodyStyle, r.date)
		ss.PrintCell(row, 2, bodyStyle, r.cat)
		ss.PrintCell(row, 3, moneyStyle, r.amount)
	}

	// --- Save ---
	if err := ss.SaveAs("Expense Report – January 2025"); err != nil {
		log.Fatalf("SaveAs: %v", err)
	}

	fmt.Println("Spreadsheet created:", ss.SpreadsheetURL())
}
