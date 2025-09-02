package services

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/xuri/excelize/v2"
	"go.uber.org/zap"
)

type ISpreadsheetService interface {
	AddValueForCategory(sheet io.Reader, category string, value float32) (io.Reader, error)
}

type ExcelerizeSpreadsheetService struct{}

const (
	KEY_COLUMN_KEY       string = "KEY_COLUMN"
	VALUE_COLUMN_KEY     string = "VALUE_COLUMN"
	MAX_EMPTY_CELL_COUNT uint   = 5
)

func (s *ExcelerizeSpreadsheetService) AddValueForCategory(sheet io.Reader, category string, value float32) (io.Reader, error) {
	f, err := excelize.OpenReader(sheet, excelize.Options{})
	if err != nil {
		zap.L().Error("Failed to open spreadsheet", zap.Error(err))
		return nil, fmt.Errorf("Failed to open spreadsheet")
	}

	defer func() {
		// Close the spreadsheet.
		if err := f.Close(); err != nil {
			zap.L().Error("Failed to close spreadsheet", zap.Error(err))
		}
	}()

	// currently default to last sheet
	sheetName := f.GetSheetName(f.SheetCount - 1)

	// iterate key column until find the category
	keyColumn := os.Getenv(KEY_COLUMN_KEY)
	currentRow := uint(1)
	found := false
	comparison := strings.ToLower(strings.ReplaceAll(category, " ", ""))
	emptyCellCount := uint(0)

	for !found {
		cell := fmt.Sprintf("%s%d", keyColumn, currentRow)
		val, err := f.GetCellValue(sheetName, cell)
		if err != nil {
			zap.L().Error("Failed to get value for cell", zap.String("cell", cell), zap.Error(err))
			return nil, fmt.Errorf("Failed to get value for cell")
		}

		normVal := strings.ToLower(strings.ReplaceAll(val, " ", ""))
		if normVal == comparison {
			found = true
			break
		}

		if len(normVal) == 0 {
			emptyCellCount++
		} else {
			emptyCellCount = 0
		}

		if emptyCellCount >= MAX_EMPTY_CELL_COUNT {
			zap.L().Warn("Category not found, quitting", zap.String("category", category))
			return nil, fmt.Errorf("Category not found")
		}

		currentRow++
	}

	// get the cell formula
	valueColumn := os.Getenv(VALUE_COLUMN_KEY)
	cell := fmt.Sprintf("%s%d", valueColumn, currentRow)
	form, err := f.GetCellFormula(sheetName, cell)
	if err != nil {
		zap.L().Error("Failed to get cell formula", zap.Error(err))
		return nil, fmt.Errorf("Failed to get cell formula")
	}

	var updatedFormula string

	if len(form) == 0 {
		// set the first value
		updatedFormula = fmt.Sprintf("%.2f", value)
	} else {
		// add the value to the formula
		updatedFormula = fmt.Sprintf("%s+%.2f", form, value)
	}

	if err := f.SetCellFormula(sheetName, cell, updatedFormula); err != nil {
		zap.L().DPanic("Failed to set the cell's formula", zap.Error(err), zap.String("formula", updatedFormula))
		return nil, fmt.Errorf("Failed to set cell formula")
	}

	updatedVal, err := f.CalcCellValue(sheetName, cell)
	if err != nil {
		zap.L().DPanic("Failed to get updated cell value from formula", zap.Error(err))
		return nil, fmt.Errorf("Failed to get updated value from formula")
	}

	if err := f.SetCellValue(sheetName, cell, updatedVal); err != nil {
		zap.L().DPanic("Failed to update cell value", zap.Error(err))
		return nil, fmt.Errorf("Failed to update cell value")
	}

	// return the spreadhseet as an io.Reader
	var buffer bytes.Buffer
	if err := f.Write(&buffer); err != nil {
		zap.L().DPanic("Failed to write spreadsheet to buffer", zap.Error(err))
	}

	return bytes.NewReader(buffer.Bytes()), nil
}
