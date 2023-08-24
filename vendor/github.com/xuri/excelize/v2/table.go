// Copyright 2016 - 2023 The excelize Authors. All rights reserved. Use of
// this source code is governed by a BSD-style license that can be found in
// the LICENSE file.
//
// Package excelize providing a set of functions that allow you to write to and
// read from XLAM / XLSM / XLSX / XLTM / XLTX files. Supports reading and
// writing spreadsheet documents generated by Microsoft Excel™ 2007 and later.
// Supports complex components by high compatibility, and provided streaming
// API for generating or reading data from a worksheet with huge amounts of
// data. This library needs Go version 1.16 or later.

package excelize

import (
	"encoding/xml"
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"unicode"
	"unicode/utf8"
)

// parseTableOptions provides a function to parse the format settings of the
// table with default value.
func parseTableOptions(opts *Table) (*Table, error) {
	var err error
	if opts == nil {
		return &Table{ShowRowStripes: boolPtr(true)}, err
	}
	if opts.ShowRowStripes == nil {
		opts.ShowRowStripes = boolPtr(true)
	}
	if err = checkTableName(opts.Name); err != nil {
		return opts, err
	}
	return opts, err
}

// AddTable provides the method to add table in a worksheet by given worksheet
// name, range reference and format set. For example, create a table of A1:D5
// on Sheet1:
//
//	err := f.AddTable("Sheet1", &excelize.Table{Range: "A1:D5"})
//
// Create a table of F2:H6 on Sheet2 with format set:
//
//	disable := false
//	err := f.AddTable("Sheet2", &excelize.Table{
//	    Range:             "F2:H6",
//	    Name:              "table",
//	    StyleName:         "TableStyleMedium2",
//	    ShowFirstColumn:   true,
//	    ShowLastColumn:    true,
//	    ShowRowStripes:    &disable,
//	    ShowColumnStripes: true,
//	})
//
// Note that the table must be at least two lines including the header. The
// header cells must contain strings and must be unique, and must set the
// header row data of the table before calling the AddTable function. Multiple
// tables range reference that can't have an intersection.
//
// Name: The name of the table, in the same worksheet name of the table should
// be unique, starts with a letter or underscore (_), doesn't include a
// space or character, and should be no more than 255 characters
//
// StyleName: The built-in table style names
//
//	TableStyleLight1 - TableStyleLight21
//	TableStyleMedium1 - TableStyleMedium28
//	TableStyleDark1 - TableStyleDark11
func (f *File) AddTable(sheet string, table *Table) error {
	options, err := parseTableOptions(table)
	if err != nil {
		return err
	}
	// Coordinate conversion, convert C1:B3 to 2,0,1,2.
	coordinates, err := rangeRefToCoordinates(options.Range)
	if err != nil {
		return err
	}
	// Correct table reference range, such correct C1:B3 to B1:C3.
	_ = sortCoordinates(coordinates)
	tableID := f.countTables() + 1
	sheetRelationshipsTableXML := "../tables/table" + strconv.Itoa(tableID) + ".xml"
	tableXML := strings.ReplaceAll(sheetRelationshipsTableXML, "..", "xl")
	// Add first table for given sheet.
	sheetXMLPath, _ := f.getSheetXMLPath(sheet)
	sheetRels := "xl/worksheets/_rels/" + strings.TrimPrefix(sheetXMLPath, "xl/worksheets/") + ".rels"
	rID := f.addRels(sheetRels, SourceRelationshipTable, sheetRelationshipsTableXML, "")
	if err = f.addSheetTable(sheet, rID); err != nil {
		return err
	}
	f.addSheetNameSpace(sheet, SourceRelationship)
	if err = f.addTable(sheet, tableXML, coordinates[0], coordinates[1], coordinates[2], coordinates[3], tableID, options); err != nil {
		return err
	}
	return f.addContentTypePart(tableID, "table")
}

// countTables provides a function to get table files count storage in the
// folder xl/tables.
func (f *File) countTables() int {
	count := 0
	f.Pkg.Range(func(k, v interface{}) bool {
		if strings.Contains(k.(string), "xl/tables/table") {
			count++
		}
		return true
	})
	return count
}

// addSheetTable provides a function to add tablePart element to
// xl/worksheets/sheet%d.xml by given worksheet name and relationship index.
func (f *File) addSheetTable(sheet string, rID int) error {
	ws, err := f.workSheetReader(sheet)
	if err != nil {
		return err
	}
	table := &xlsxTablePart{
		RID: "rId" + strconv.Itoa(rID),
	}
	if ws.TableParts == nil {
		ws.TableParts = &xlsxTableParts{}
	}
	ws.TableParts.Count++
	ws.TableParts.TableParts = append(ws.TableParts.TableParts, table)
	return err
}

// setTableHeader provides a function to set cells value in header row for the
// table.
func (f *File) setTableHeader(sheet string, showHeaderRow bool, x1, y1, x2 int) ([]*xlsxTableColumn, error) {
	var (
		tableColumns []*xlsxTableColumn
		idx          int
	)
	for i := x1; i <= x2; i++ {
		idx++
		cell, err := CoordinatesToCellName(i, y1)
		if err != nil {
			return tableColumns, err
		}
		name, _ := f.GetCellValue(sheet, cell)
		if _, err := strconv.Atoi(name); err == nil {
			if showHeaderRow {
				_ = f.SetCellStr(sheet, cell, name)
			}
		}
		if name == "" {
			name = "Column" + strconv.Itoa(idx)
			if showHeaderRow {
				_ = f.SetCellStr(sheet, cell, name)
			}
		}
		tableColumns = append(tableColumns, &xlsxTableColumn{
			ID:   idx,
			Name: name,
		})
	}
	return tableColumns, nil
}

// checkSheetName check whether there are illegal characters in the table name.
// Verify that the name:
// 1. Starts with a letter or underscore (_)
// 2. Doesn't include a space or character that isn't allowed
func checkTableName(name string) error {
	if utf8.RuneCountInString(name) > MaxFieldLength {
		return ErrTableNameLength
	}
	for i, c := range name {
		if string(c) == "_" {
			continue
		}
		if unicode.IsLetter(c) {
			continue
		}
		if i > 0 && unicode.IsDigit(c) {
			continue
		}
		return newInvalidTableNameError(name)
	}
	return nil
}

// addTable provides a function to add table by given worksheet name,
// range reference and format set.
func (f *File) addTable(sheet, tableXML string, x1, y1, x2, y2, i int, opts *Table) error {
	// Correct the minimum number of rows, the table at least two lines.
	if y1 == y2 {
		y2++
	}
	hideHeaderRow := opts != nil && opts.ShowHeaderRow != nil && !*opts.ShowHeaderRow
	if hideHeaderRow {
		y1++
	}
	// Correct table range reference, such correct C1:B3 to B1:C3.
	ref, err := f.coordinatesToRangeRef([]int{x1, y1, x2, y2})
	if err != nil {
		return err
	}
	tableColumns, _ := f.setTableHeader(sheet, !hideHeaderRow, x1, y1, x2)
	name := opts.Name
	if name == "" {
		name = "Table" + strconv.Itoa(i)
	}
	t := xlsxTable{
		XMLNS:       NameSpaceSpreadSheet.Value,
		ID:          i,
		Name:        name,
		DisplayName: name,
		Ref:         ref,
		AutoFilter: &xlsxAutoFilter{
			Ref: ref,
		},
		TableColumns: &xlsxTableColumns{
			Count:       len(tableColumns),
			TableColumn: tableColumns,
		},
		TableStyleInfo: &xlsxTableStyleInfo{
			Name:              opts.StyleName,
			ShowFirstColumn:   opts.ShowFirstColumn,
			ShowLastColumn:    opts.ShowLastColumn,
			ShowRowStripes:    *opts.ShowRowStripes,
			ShowColumnStripes: opts.ShowColumnStripes,
		},
	}
	if hideHeaderRow {
		t.AutoFilter = nil
		t.HeaderRowCount = intPtr(0)
	}
	table, _ := xml.Marshal(t)
	f.saveFileList(tableXML, table)
	return nil
}

// AutoFilter provides the method to add auto filter in a worksheet by given
// worksheet name, range reference and settings. An auto filter in Excel is a
// way of filtering a 2D range of data based on some simple criteria. For
// example applying an auto filter to a cell range A1:D4 in the Sheet1:
//
//	err := f.AutoFilter("Sheet1", "A1:D4", []excelize.AutoFilterOptions{})
//
// Filter data in an auto filter:
//
//	err := f.AutoFilter("Sheet1", "A1:D4", []excelize.AutoFilterOptions{
//	    {Column: "B", Expression: "x != blanks"},
//	})
//
// Column defines the filter columns in an auto filter range based on simple
// criteria
//
// It isn't sufficient to just specify the filter condition. You must also
// hide any rows that don't match the filter condition. Rows are hidden using
// the SetRowVisible function. Excelize can't filter rows automatically since
// this isn't part of the file format.
//
// Setting a filter criteria for a column:
//
// Expression defines the conditions, the following operators are available
// for setting the filter criteria:
//
//	==
//	!=
//	>
//	<
//	>=
//	<=
//	and
//	or
//
// An expression can comprise a single statement or two statements separated
// by the 'and' and 'or' operators. For example:
//
//	x <  2000
//	x >  2000
//	x == 2000
//	x >  2000 and x <  5000
//	x == 2000 or  x == 5000
//
// Filtering of blank or non-blank data can be achieved by using a value of
// Blanks or NonBlanks in the expression:
//
//	x == Blanks
//	x == NonBlanks
//
// Excel also allows some simple string matching operations:
//
//	x == b*      // begins with b
//	x != b*      // doesn't begin with b
//	x == *b      // ends with b
//	x != *b      // doesn't end with b
//	x == *b*     // contains b
//	x != *b*     // doesn't contains b
//
// You can also use '*' to match any character or number and '?' to match any
// single character or number. No other regular expression quantifier is
// supported by Excel's filters. Excel's regular expression characters can be
// escaped using '~'.
//
// The placeholder variable x in the above examples can be replaced by any
// simple string. The actual placeholder name is ignored internally so the
// following are all equivalent:
//
//	x     < 2000
//	col   < 2000
//	Price < 2000
func (f *File) AutoFilter(sheet, rangeRef string, opts []AutoFilterOptions) error {
	coordinates, err := rangeRefToCoordinates(rangeRef)
	if err != nil {
		return err
	}
	_ = sortCoordinates(coordinates)
	// Correct reference range, such correct C1:B3 to B1:C3.
	ref, _ := f.coordinatesToRangeRef(coordinates, true)
	filterDB := "_xlnm._FilterDatabase"
	wb, err := f.workbookReader()
	if err != nil {
		return err
	}
	sheetID, err := f.GetSheetIndex(sheet)
	if err != nil {
		return err
	}
	filterRange := fmt.Sprintf("'%s'!%s", sheet, ref)
	d := xlsxDefinedName{
		Name:         filterDB,
		Hidden:       true,
		LocalSheetID: intPtr(sheetID),
		Data:         filterRange,
	}
	if wb.DefinedNames == nil {
		wb.DefinedNames = &xlsxDefinedNames{
			DefinedName: []xlsxDefinedName{d},
		}
	} else {
		var definedNameExists bool
		for idx := range wb.DefinedNames.DefinedName {
			definedName := wb.DefinedNames.DefinedName[idx]
			if definedName.Name == filterDB && *definedName.LocalSheetID == sheetID && definedName.Hidden {
				wb.DefinedNames.DefinedName[idx].Data = filterRange
				definedNameExists = true
			}
		}
		if !definedNameExists {
			wb.DefinedNames.DefinedName = append(wb.DefinedNames.DefinedName, d)
		}
	}
	columns := coordinates[2] - coordinates[0]
	return f.autoFilter(sheet, ref, columns, coordinates[0], opts)
}

// autoFilter provides a function to extract the tokens from the filter
// expression. The tokens are mainly non-whitespace groups.
func (f *File) autoFilter(sheet, ref string, columns, col int, opts []AutoFilterOptions) error {
	ws, err := f.workSheetReader(sheet)
	if err != nil {
		return err
	}
	if ws.SheetPr != nil {
		ws.SheetPr.FilterMode = true
	}
	ws.SheetPr = &xlsxSheetPr{FilterMode: true}
	filter := &xlsxAutoFilter{
		Ref: ref,
	}
	ws.AutoFilter = filter
	for _, opt := range opts {
		if opt.Column == "" || opt.Expression == "" {
			continue
		}
		fsCol, err := ColumnNameToNumber(opt.Column)
		if err != nil {
			return err
		}
		offset := fsCol - col
		if offset < 0 || offset > columns {
			return fmt.Errorf("incorrect index of column '%s'", opt.Column)
		}
		fc := &xlsxFilterColumn{ColID: offset}
		re := regexp.MustCompile(`"(?:[^"]|"")*"|\S+`)
		token := re.FindAllString(opt.Expression, -1)
		if len(token) != 3 && len(token) != 7 {
			return fmt.Errorf("incorrect number of tokens in criteria '%s'", opt.Expression)
		}
		expressions, tokens, err := f.parseFilterExpression(opt.Expression, token)
		if err != nil {
			return err
		}
		f.writeAutoFilter(fc, expressions, tokens)
		filter.FilterColumn = append(filter.FilterColumn, fc)
	}
	ws.AutoFilter = filter
	return nil
}

// writeAutoFilter provides a function to check for single or double custom
// filters as default filters and handle them accordingly.
func (f *File) writeAutoFilter(fc *xlsxFilterColumn, exp []int, tokens []string) {
	if len(exp) == 1 && exp[0] == 2 {
		// Single equality.
		var filters []*xlsxFilter
		filters = append(filters, &xlsxFilter{Val: tokens[0]})
		fc.Filters = &xlsxFilters{Filter: filters}
	} else if len(exp) == 3 && exp[0] == 2 && exp[1] == 1 && exp[2] == 2 {
		// Double equality with "or" operator.
		var filters []*xlsxFilter
		for _, v := range tokens {
			filters = append(filters, &xlsxFilter{Val: v})
		}
		fc.Filters = &xlsxFilters{Filter: filters}
	} else {
		// Non default custom filter.
		expRel := map[int]int{0: 0, 1: 2}
		andRel := map[int]bool{0: true, 1: false}
		for k, v := range tokens {
			f.writeCustomFilter(fc, exp[expRel[k]], v)
			if k == 1 {
				fc.CustomFilters.And = andRel[exp[k]]
			}
		}
	}
}

// writeCustomFilter provides a function to write the <customFilter> element.
func (f *File) writeCustomFilter(fc *xlsxFilterColumn, operator int, val string) {
	operators := map[int]string{
		1:  "lessThan",
		2:  "equal",
		3:  "lessThanOrEqual",
		4:  "greaterThan",
		5:  "notEqual",
		6:  "greaterThanOrEqual",
		22: "equal",
	}
	customFilter := xlsxCustomFilter{
		Operator: operators[operator],
		Val:      val,
	}
	if fc.CustomFilters != nil {
		fc.CustomFilters.CustomFilter = append(fc.CustomFilters.CustomFilter, &customFilter)
	} else {
		var customFilters []*xlsxCustomFilter
		customFilters = append(customFilters, &customFilter)
		fc.CustomFilters = &xlsxCustomFilters{CustomFilter: customFilters}
	}
}

// parseFilterExpression provides a function to converts the tokens of a
// possibly conditional expression into 1 or 2 sub expressions for further
// parsing.
//
// Examples:
//
//	('x', '==', 2000) -> exp1
//	('x', '>',  2000, 'and', 'x', '<', 5000) -> exp1 and exp2
func (f *File) parseFilterExpression(expression string, tokens []string) ([]int, []string, error) {
	var expressions []int
	var t []string
	if len(tokens) == 7 {
		// The number of tokens will be either 3 (for 1 expression) or 7 (for 2
		// expressions).
		conditional := 0
		c := tokens[3]
		re, _ := regexp.Match(`(or|\|\|)`, []byte(c))
		if re {
			conditional = 1
		}
		expression1, token1, err := f.parseFilterTokens(expression, tokens[:3])
		if err != nil {
			return expressions, t, err
		}
		expression2, token2, err := f.parseFilterTokens(expression, tokens[4:7])
		if err != nil {
			return expressions, t, err
		}
		expressions = []int{expression1[0], conditional, expression2[0]}
		t = []string{token1, token2}
	} else {
		exp, token, err := f.parseFilterTokens(expression, tokens)
		if err != nil {
			return expressions, t, err
		}
		expressions = exp
		t = []string{token}
	}
	return expressions, t, nil
}

// parseFilterTokens provides a function to parse the 3 tokens of a filter
// expression and return the operator and token.
func (f *File) parseFilterTokens(expression string, tokens []string) ([]int, string, error) {
	operators := map[string]int{
		"==": 2,
		"=":  2,
		"=~": 2,
		"eq": 2,
		"!=": 5,
		"!~": 5,
		"ne": 5,
		"<>": 5,
		"<":  1,
		"<=": 3,
		">":  4,
		">=": 6,
	}
	operator, ok := operators[strings.ToLower(tokens[1])]
	if !ok {
		// Convert the operator from a number to a descriptive string.
		return []int{}, "", fmt.Errorf("unknown operator: %s", tokens[1])
	}
	token := tokens[2]
	// Special handling for Blanks/NonBlanks.
	re, _ := regexp.Match("blanks|nonblanks", []byte(strings.ToLower(token)))
	if re {
		// Only allow Equals or NotEqual in this context.
		if operator != 2 && operator != 5 {
			return []int{operator}, token, fmt.Errorf("the operator '%s' in expression '%s' is not valid in relation to Blanks/NonBlanks'", tokens[1], expression)
		}
		token = strings.ToLower(token)
		// The operator should always be 2 (=) to flag a "simple" equality in
		// the binary record. Therefore we convert <> to =.
		if token == "blanks" {
			if operator == 5 {
				token = " "
			}
		} else {
			if operator == 5 {
				operator = 2
				token = "blanks"
			} else {
				operator = 5
				token = " "
			}
		}
	}
	// If the string token contains an Excel match character then change the
	// operator type to indicate a non "simple" equality.
	re, _ = regexp.Match("[*?]", []byte(token))
	if operator == 2 && re {
		operator = 22
	}
	return []int{operator}, token, nil
}
