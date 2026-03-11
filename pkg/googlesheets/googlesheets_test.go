package googlesheets

import (
	"context"
	"fmt"
	"reflect"
	"testing"
)

func TestA1ToIndices(t *testing.T) {
	tests := []struct {
		cell    string
		wantCol int
		wantRow int
		wantErr bool
	}{
		{"A1", 0, 0, false},
		{"B2", 1, 1, false},
		{"Z1", 25, 0, false},
		{"AA1", 26, 0, false},
		{"AB3", 27, 2, false},
		{"a1", 0, 0, false},
		{"", 0, 0, true},
		{"1A", 0, 0, true},
		{"A", 0, 0, true},
		{"A0", 0, 0, true},
	}

	for _, tc := range tests {
		col, row, err := a1ToIndices(tc.cell)
		if tc.wantErr {
			if err == nil {
				t.Errorf("a1ToIndices(%q): expected error, got col=%d row=%d", tc.cell, col, row)
			}
			continue
		}
		if err != nil {
			t.Errorf("a1ToIndices(%q): unexpected error: %v", tc.cell, err)
			continue
		}
		if col != tc.wantCol || row != tc.wantRow {
			t.Errorf("a1ToIndices(%q) = (%d,%d), want (%d,%d)",
				tc.cell, col, row, tc.wantCol, tc.wantRow)
		}
	}
}

// derefValue mirrors the pointer-unwrapping logic inside buildCellRequest.
// It is defined locally here so the test exercises the same logic without
// depending on the Google API types.
func derefValue(val any) any {
	if val == nil {
		return nil
	}
	rv := reflect.ValueOf(val)
	for rv.Kind() == reflect.Ptr {
		if rv.IsNil() {
			return nil
		}
		rv = rv.Elem()
		val = rv.Interface()
	}
	return val
}

func TestDerefValue(t *testing.T) {
	strVal := "hello"
	f64Val := float64(3.14)
	intVal := 42
	boolVal := true

	tests := []struct {
		name string
		in   any
		want any
	}{
		{"plain string", strVal, strVal},
		{"plain int", intVal, intVal},
		{"*string", &strVal, strVal},
		{"*float64", &f64Val, f64Val},
		{"*int", &intVal, intVal},
		{"*bool", &boolVal, boolVal},
		{"**int", func() any { p := &intVal; return &p }(), intVal},
		{"nil *int", (*int)(nil), nil},
		{"nil any", nil, nil},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := derefValue(tc.in)
			if !reflect.DeepEqual(got, tc.want) {
				t.Errorf("derefValue(%v) = %v (%T), want %v (%T)",
					tc.in, got, got, tc.want, tc.want)
			}
		})
	}
}

func TestConfigDefaults(t *testing.T) {
	cfg := Config{}
	cfg.applyDefaults()

	if cfg.CredentialsFile != "credentials.json" {
		t.Errorf("CredentialsFile default = %q, want %q", cfg.CredentialsFile, "credentials.json")
	}
	if cfg.TokenFile != "token.json" {
		t.Errorf("TokenFile default = %q, want %q", cfg.TokenFile, "token.json")
	}

	cfg2 := Config{CredentialsFile: "my-creds.json", TokenFile: "my-token.json"}
	cfg2.applyDefaults()
	if cfg2.CredentialsFile != "my-creds.json" {
		t.Errorf("CredentialsFile was overwritten to %q", cfg2.CredentialsFile)
	}
	if cfg2.TokenFile != "my-token.json" {
		t.Errorf("TokenFile was overwritten to %q", cfg2.TokenFile)
	}
}

// Regression test for the bug that produced the "cannot delete all
// non-frozen columns" error.  A FreezePane call can be queued after one or
// more RemoveCol operations; flush orders the freeze request before the
// deletes, which means the frozen count must be adjusted to account for the
// pending removals.  buildBatchRequests (and by extension flush) now clumps
// the final freeze count to at most colCount-1 when deletes are present.
func TestFreezeClampDuringBatch(t *testing.T) {
	s := &Sheet{
		ctx:          context.Background(),
		cellsBySheet: make(map[int64][]pendingCell),
		colCount:     map[int64]int{1: 3},
		frozenCols:   make(map[int64]int),
		frozenRows:   make(map[int64]int),
		sheetID:      1,
		sheetName:    "foo",
	}

	// simulate freezing more columns than we will end up with later
	if err := s.FreezePane("D1"); err != nil {
		t.Fatalf("FreezePane: %v", err)
	}
	// remove a column, decreasing colCount to 2
	if err := s.RemoveCol(1); err != nil {
		t.Fatalf("RemoveCol: %v", err)
	}

	reqs, err := s.buildBatchRequests()
	if err != nil {
		t.Fatalf("buildBatchRequests: %v", err)
	}

	// there should be a freeze request and a delete request
	var foundFreeze bool
	for _, r := range reqs {
		if r.UpdateSheetProperties != nil &&
			r.UpdateSheetProperties.Properties != nil &&
			r.UpdateSheetProperties.Properties.GridProperties != nil {
			foundFreeze = true
			frozen := r.UpdateSheetProperties.Properties.GridProperties.FrozenColumnCount
			if frozen != 1 {
				t.Errorf("expected frozen count 1 after clamp, got %d", frozen)
			}
		}
	}
	if !foundFreeze {
		t.Error("expected a freeze request in batch")
	}
}

func TestIsAllUnfrozenDeletionError(t *testing.T) {
	cases := []struct {
		name string
		err  error
		want bool
	}{
		{"Portuguese message", fmt.Errorf("googleapi: Error 400: Invalid requests[1].deleteDimension: Não é possível excluir todas as colunas não congeladas., badRequest"), true},
		{"English message", fmt.Errorf("googleapi: Error 400: Invalid requests[2].deleteDimension: Cannot delete all the non-frozen columns., badRequest"), true},
		{"unrelated error", fmt.Errorf("some other error"), false},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := isAllUnfrozenDeletionError(tc.err); got != tc.want {
				t.Errorf("got %v, want %v", got, tc.want)
			}
		})
	}
}
