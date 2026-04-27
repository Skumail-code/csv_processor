package processor

import (
	"encoding/csv"
	"os"
	"testing"
	"time"
)

func TestProcessRowRejectsExtraColumns(t *testing.T) {
	trans := (&CSVProcessor{}).processRow(1, []string{
		"1",
		"2024-01-01",
		"Coffee",
		"100.00",
		"Food",
		"unexpected",
	})

	if trans.IsValid {
		t.Fatalf("expected invalid transaction")
	}

	want := "expected 5 columns, got 6 (extra columns are not allowed)"
	if trans.ErrorMsg != want {
		t.Fatalf("unexpected error message: got %q want %q", trans.ErrorMsg, want)
	}
}

func TestProcessRowRejectsAllEmptyFieldsTogether(t *testing.T) {
	trans := (&CSVProcessor{}).processRow(1, []string{
		"1",
		"2024-01-01",
		"",
		"",
		"",
	})

	if trans.IsValid {
		t.Fatalf("expected invalid transaction")
	}

	want := "description, amount, and category cannot be empty"
	if trans.ErrorMsg != want {
		t.Fatalf("unexpected error message: got %q want %q", trans.ErrorMsg, want)
	}
}

func TestProcessRowRejectsNonNumericAmount(t *testing.T) {
	trans := (&CSVProcessor{}).processRow(1, []string{
		"1",
		"2024-01-01",
		"Coffee",
		"abc",
		"Food",
	})

	if trans.IsValid {
		t.Fatalf("expected invalid transaction")
	}

	want := "invalid amount: 'abc' must be numeric"
	if trans.ErrorMsg != want {
		t.Fatalf("unexpected error message: got %q want %q", trans.ErrorMsg, want)
	}
}

func TestProcessRowRejectsIntegerCategory(t *testing.T) {
	trans := (&CSVProcessor{}).processRow(1, []string{
		"1",
		"2024-01-01",
		"Coffee",
		"100.50",
		"123",
	})

	if trans.IsValid {
		t.Fatalf("expected invalid transaction")
	}

	want := "invalid category: '123' cannot be an integer"
	if trans.ErrorMsg != want {
		t.Fatalf("unexpected error message: got %q want %q", trans.ErrorMsg, want)
	}
}

func TestProcessRowAcceptsValidRecord(t *testing.T) {
	trans := (&CSVProcessor{}).processRow(1, []string{
		"1",
		"2024-01-01",
		"Coffee",
		"100.50",
		"Food",
	})

	if !trans.IsValid {
		t.Fatalf("expected valid transaction, got error: %s", trans.ErrorMsg)
	}
}

func TestProcessRowRejectsNullDescription(t *testing.T) {
	trans := (&CSVProcessor{}).processRow(1, []string{
		"1",
		"2024-01-01",
		"null",
		"100.00",
		"Food",
	})

	if trans.IsValid {
		t.Fatalf("expected null description to be invalid")
	}
}

func TestProcessRowAcceptsNegativeAmount(t *testing.T) {
	trans := (&CSVProcessor{}).processRow(1, []string{
		"1",
		"2024-01-01",
		"Refund",
		"-100.00",
		"Adjustment",
	})

	if !trans.IsValid {
		t.Fatalf("expected negative amount to be valid, got error: %s", trans.ErrorMsg)
	}
}

func TestProcessRowRejectsZeroAmount(t *testing.T) {
	trans := (&CSVProcessor{}).processRow(1, []string{
		"1",
		"2024-01-01",
		"Coffee",
		"0.00",
		"Food",
	})

	if trans.IsValid {
		t.Fatalf("expected zero amount to be invalid")
	}

	want := "invalid amount: '0.00' cannot be 0"
	if trans.ErrorMsg != want {
		t.Fatalf("unexpected error message: got %q want %q", trans.ErrorMsg, want)
	}
}

func TestProcessRowRejectsPositiveIntegerAmountForNonIncome(t *testing.T) {
	trans := (&CSVProcessor{}).processRow(1, []string{
		"1",
		"2024-01-01",
		"Salary",
		"100",
		"Expense",
	})

	if trans.IsValid {
		t.Fatalf("expected positive integer amount to be invalid for non-income category")
	}

	want := "invalid amount: '100' positive integer amounts are only allowed for category 'Income'"
	if trans.ErrorMsg != want {
		t.Fatalf("unexpected error message: got %q want %q", trans.ErrorMsg, want)
	}
}

func TestProcessRowAcceptsPositiveIntegerAmountForIncome(t *testing.T) {
	trans := (&CSVProcessor{}).processRow(1, []string{
		"1",
		"2024-01-01",
		"Salary",
		"100",
		"Income",
	})

	if !trans.IsValid {
		t.Fatalf("expected positive integer amount to be valid for income, got error: %s", trans.ErrorMsg)
	}
}

func TestGenerateOutputFilePreservesRowOrder(t *testing.T) {
	p := &CSVProcessor{
		AllRows: []Transaction{
			{
				ID:             1,
				Date:           mustParseDate(t, "2024-01-01"),
				RawID:          "1",
				RawDate:        "2024-01-01",
				RawDescription: "Coffee",
				RawAmount:      "100.50",
				RawCategory:    "Food",
				Description:    "Coffee",
				Amount:         100.50,
				Category:       "Food",
				IsValid:        true,
			},
			{
				ID:             2,
				Date:           mustParseDate(t, "2024-01-02"),
				RawID:          "2",
				RawDate:        "2024-01-02",
				RawDescription: "Placeholder entry",
				RawAmount:      "0.00",
				RawCategory:    "Food",
				Description:    "Placeholder entry",
				Amount:         0,
				Category:       "Food",
				IsValid:        false,
				ErrorMsg:       "invalid description: 'Placeholder entry' is not a valid description",
			},
		},
	}

	outputPath, err := p.GenerateOutputFile(t.TempDir())
	if err != nil {
		t.Fatalf("GenerateOutputFile failed: %v", err)
	}

	f, err := os.Open(outputPath)
	if err != nil {
		t.Fatalf("failed to open output file: %v", err)
	}
	defer f.Close()

	rows, err := csv.NewReader(f).ReadAll()
	if err != nil {
		t.Fatalf("failed to read output file: %v", err)
	}

	if len(rows) != 3 {
		t.Fatalf("expected 3 rows, got %d", len(rows))
	}

	if rows[1][0] != "1" || rows[1][1] != "2024-01-01" || rows[1][2] != "Coffee" || rows[1][3] != "100.50" || rows[1][4] != "Food" || rows[1][5] != "VALID" || rows[1][6] != "" {
		t.Fatalf("unexpected first data row: %#v", rows[1])
	}
	if rows[2][0] != "2" || rows[2][1] != "2024-01-02" || rows[2][2] != "Placeholder entry" || rows[2][3] != "0.00" || rows[2][4] != "Food" || rows[2][5] != "INVALID" || rows[2][6] == "" {
		t.Fatalf("unexpected second data row: %#v", rows[2])
	}
}

func mustParseDate(t *testing.T, value string) time.Time {
	t.Helper()

	d, err := time.Parse("2006-01-02", value)
	if err != nil {
		t.Fatalf("failed to parse date %q: %v", value, err)
	}
	return d
}
