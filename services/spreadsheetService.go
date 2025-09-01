package services

import (
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/xuri/excelize/v2"
	"go.uber.org/zap"
)

type ISpreadsheetService interface {
	AddValueForCategory(sheet io.Reader, category string, value float32) error
}

type ExcelerizeSpreadsheetService struct{}

const (
	KEY_COLUMN_KEY       string = "KEY_COLUMN"
	VALUE_COLUMN_KEY     string = "VALUE_COLUMN"
	MAX_EMPTY_CELL_COUNT uint   = 5
)

func (s *ExcelerizeSpreadsheetService) AddValueForCategory(sheet io.Reader, category string, value float32) error {

	f, err := excelize.OpenReader(sheet, excelize.Options{})
	if err != nil {
		zap.L().Error("Failed to open spreadsheet", zap.Error(err))
		return fmt.Errorf("Failed to open spreadsheet")
	}

	defer func() {
		// Close the spreadsheet.
		if err := f.Close(); err != nil {
			zap.L().Error("Failed to close spreadsheet", zap.Error(err))
		}
	}()

	// currently default to last sheet
	sheetName := f.GetSheetName(f.SheetCount - 1)
	zap.L().Info("Editing sheet with name: ", zap.String("name", sheetName))

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
			return fmt.Errorf("Failed to get value for cell")
		}

		normVal := strings.ToLower(strings.ReplaceAll(val, " ", ""))
		if normVal == comparison {
			found = true
		}

		if len(normVal) == 0 {
			emptyCellCount++
		} else {
			emptyCellCount = 0
		}

		if emptyCellCount >= MAX_EMPTY_CELL_COUNT {
			zap.L().Warn("Category not found, quitting", zap.String("category", category))
			return fmt.Errorf("Category not found")
		}

		currentRow++
	}

	valueColumn := os.Getenv(VALUE_COLUMN_KEY)
	cell := fmt.Sprintf("%s%d", valueColumn, currentRow)
	val, err := f.GetCellValue(sheetName, cell)
	if err != nil {
		zap.L().Error("Failed to get value for cell", zap.String("cell", cell), zap.Error(err))
		return fmt.Errorf("Failed to get value for cell")
	}

	zap.L().Info("Got cell value", zap.String("cell", cell), zap.String("value", val))

	return nil
}
