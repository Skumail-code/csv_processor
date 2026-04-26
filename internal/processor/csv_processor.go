package processor

import (
	"encoding/csv"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

type Transaction struct {
	ID          int
	Date        time.Time
	Description string
	Amount      float64
	Category    string
	IsValid     bool
	ErrorMsg    string
}

type CSVProcessor struct {
	FilePath      string
	ValidRows     []Transaction
	InvalidRows   []Transaction
	TotalRows     int
	ProcessedRows int
}

func NewCSVProcessor(filePath string) *CSVProcessor {
	return &CSVProcessor{
		FilePath:    filePath,
		ValidRows:   make([]Transaction, 0),
		InvalidRows: make([]Transaction, 0),
	}
}

func (p *CSVProcessor) Process(onProgress func(processed, valid, invalid int)) error {
	file, err := os.Open(p.FilePath)
	if err != nil {
		return fmt.Errorf("failed to open file: %w", err)
	}
	defer file.Close()

	reader := csv.NewReader(file)
	reader.TrimLeadingSpace = true
	reader.FieldsPerRecord = -1 // Allow variable fields

	// Read header
	header, err := reader.Read()
	if err != nil {
		return fmt.Errorf("failed to read header: %w", err)
	}

	// Validate headers
	expectedHeaders := []string{"id", "date", "description", "amount", "category"}
	if err := p.validateHeaders(header, expectedHeaders); err != nil {
		return err
	}

	// Process rows
	rowIndex := 1 // Start after header
	for {
		record, err := reader.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			p.addInvalidRow(rowIndex, fmt.Sprintf("parse error: %v", err))
			rowIndex++
			continue
		}

		// Process single row
		transaction := p.processRow(rowIndex, record)
		if transaction.IsValid {
			p.ValidRows = append(p.ValidRows, *transaction)
		} else {
			p.InvalidRows = append(p.InvalidRows, *transaction)
		}

		p.ProcessedRows = rowIndex
		rowIndex++

		// Report progress every 100 rows
		if rowIndex%100 == 0 && onProgress != nil {
			onProgress(p.ProcessedRows, len(p.ValidRows), len(p.InvalidRows))
		}
	}

	p.TotalRows = rowIndex - 1 // Subtract header row

	// Final progress update
	if onProgress != nil {
		onProgress(p.TotalRows, len(p.ValidRows), len(p.InvalidRows))
	}

	return nil
}

func (p *CSVProcessor) validateHeaders(actual, expected []string) error {
	if len(actual) < len(expected) {
		return fmt.Errorf("missing columns: expected %d columns, got %d", len(expected), len(actual))
	}

	for i, exp := range expected {
		if i < len(actual) && strings.ToLower(actual[i]) != exp {
			return fmt.Errorf("invalid column at position %d: expected '%s', got '%s'", i+1, exp, actual[i])
		}
	}
	return nil
}

func (p *CSVProcessor) processRow(rowNum int, record []string) *Transaction {
	trans := &Transaction{
		IsValid: true,
	}

	// Check column count
	if len(record) < 5 {
		trans.IsValid = false
		trans.ErrorMsg = fmt.Sprintf("expected 5 columns, got %d", len(record))
		return trans
	}

	// Parse ID
	id, err := strconv.Atoi(strings.TrimSpace(record[0]))
	if err != nil {
		trans.IsValid = false
		trans.ErrorMsg = fmt.Sprintf("invalid ID: %v", err)
		return trans
	}
	trans.ID = id

	// Parse Date
	dateStr := strings.TrimSpace(record[1])
	date, err := time.Parse("2006-01-02", dateStr)
	if err != nil {
		trans.IsValid = false
		trans.ErrorMsg = fmt.Sprintf("invalid date format (expected YYYY-MM-DD): %v", err)
		return trans
	}
	trans.Date = date

	// Parse Description
	trans.Description = strings.TrimSpace(record[2])
	if trans.Description == "" {
		trans.IsValid = false
		trans.ErrorMsg = "description cannot be empty"
		return trans
	}

	// Parse Amount (handle Indian Rupee symbol)
	amountStr := strings.TrimSpace(record[3])
	amountStr = strings.ReplaceAll(amountStr, "₹", "")
	amountStr = strings.ReplaceAll(amountStr, ",", "")
	amount, err := strconv.ParseFloat(amountStr, 64)
	if err != nil {
		trans.IsValid = false
		trans.ErrorMsg = fmt.Sprintf("invalid amount: %v", err)
		return trans
	}
	trans.Amount = amount

	// Parse Category
	trans.Category = strings.TrimSpace(record[4])
	validCategories := map[string]bool{
		"Income": true, "Food & Dining": true, "Transport": true,
		"Groceries": true, "Entertainment": true, "Transfer": true,
		"Shopping": true, "Donation": true, "Finance": true,
	}
	if !validCategories[trans.Category] {
		trans.IsValid = false
		trans.ErrorMsg = fmt.Sprintf("invalid category: %s", trans.Category)
		return trans
	}

	return trans
}

func (p *CSVProcessor) addInvalidRow(rowNum int, errMsg string) {
	trans := &Transaction{
		ID:       rowNum,
		IsValid:  false,
		ErrorMsg: errMsg,
	}
	p.InvalidRows = append(p.InvalidRows, *trans)
}

func (p *CSVProcessor) GenerateOutputFile(outputDir string) (string, error) {
	// Create output filename with timestamp
	timestamp := time.Now().Format("20060102_150405")
	outputPath := filepath.Join(outputDir, fmt.Sprintf("processed_%s.csv", timestamp))

	// Create output file
	outFile, err := os.Create(outputPath)
	if err != nil {
		return "", fmt.Errorf("failed to create output file: %w", err)
	}
	defer outFile.Close()

	writer := csv.NewWriter(outFile)
	defer writer.Flush()

	// Write headers
	headers := []string{"id", "date", "description", "amount", "category", "status", "error_message"}
	if err := writer.Write(headers); err != nil {
		return "", fmt.Errorf("failed to write headers: %w", err)
	}

	// Write valid rows with "VALID" status
	for _, row := range p.ValidRows {
		record := []string{
			strconv.Itoa(row.ID),
			row.Date.Format("2006-01-02"),
			row.Description,
			strconv.FormatFloat(row.Amount, 'f', 2, 64),
			row.Category,
			"VALID",
			"",
		}
		if err := writer.Write(record); err != nil {
			return "", fmt.Errorf("failed to write valid row: %w", err)
		}
	}

	// Write invalid rows with error messages
	for _, row := range p.InvalidRows {
		record := []string{
			strconv.Itoa(row.ID),
			"", // Date may be invalid
			row.Description,
			"0",
			"",
			"INVALID",
			row.ErrorMsg,
		}
		if err := writer.Write(record); err != nil {
			return "", fmt.Errorf("failed to write invalid row: %w", err)
		}
	}

	return outputPath, nil
}

// GetSummary returns processing summary
func (p *CSVProcessor) GetSummary() map[string]interface{} {
	return map[string]interface{}{
		"total_rows":   p.TotalRows,
		"valid_rows":   len(p.ValidRows),
		"invalid_rows": len(p.InvalidRows),
		"success_rate": float64(len(p.ValidRows)) / float64(p.TotalRows) * 100,
	}
}
