package services

import (
	"bytes"
	"fmt"
	"io"
	"jarvis/tool_spreadsheet_editor/model"
	"regexp"
	"strconv"
	"strings"

	"github.com/xuri/excelize/v2"
	"go.uber.org/zap"
)

type ISpreadsheetService interface {
	ListCategoriesAndValues(source model.SpreadsheetSource, sheet io.Reader) (*[]model.Entry, error)
	AddValueForCategory(source model.SpreadsheetSource, sheet io.Reader, category string, value float32) (io.Reader, *string, error)
	ReadValueForCategory(source model.SpreadsheetSource, sheet io.Reader, category string, details bool) (*string, error)
	RemoveLastValueForCategory(source model.SpreadsheetSource, sheet io.Reader, category string) (*RemovedResult, error)
}

type ExcelerizeSpreadsheetService struct{}

const (
	MAX_EMPTY_CELL_COUNT uint = 5
)

func (s *ExcelerizeSpreadsheetService) ListCategoriesAndValues(source model.SpreadsheetSource, sheet io.Reader) (*[]model.Entry, error) {
	bs := getBaseSpreadsheetSource(source)

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

	entries := []model.Entry{}

	// currently default to last sheet
	sheetName := f.GetSheetName(f.SheetCount - 1)

	// iterate key column until getting empty cells
	emptyCellCount := uint(0)
	currentRow := uint(1)
	for {
		if emptyCellCount >= MAX_EMPTY_CELL_COUNT {
			break
		}

		categoryCell := fmt.Sprintf("%s%d", bs.CostNameColumn, currentRow)
		category, err := f.GetCellValue(sheetName, categoryCell)
		if err != nil {
			zap.L().Error("Failed to get value for category cell", zap.String("cell", categoryCell), zap.Error(err))
			return nil, fmt.Errorf("Failed to get value for cell")
		}

		trimmed := strings.ReplaceAll(category, " ", "")
		if len(trimmed) == 0 {
			emptyCellCount++
			currentRow++
			continue
		}

		emptyCellCount = 0

		valueCell := fmt.Sprintf("%s%d", bs.CostValueColumn, currentRow)
		value, err := f.CalcCellValue(sheetName, valueCell)
		if err != nil {
			zap.L().Error("Failed to get value for value cell", zap.String("cell", categoryCell), zap.Error(err))
			return nil, fmt.Errorf("Failed to get value for cell")
		}

		entries = append(entries, model.Entry{
			Category: category,
			Value:    value,
		})

		currentRow++
	}

	return &entries, nil
}

func (s *ExcelerizeSpreadsheetService) AddValueForCategory(source model.SpreadsheetSource, sheet io.Reader, category string, value float32) (io.Reader, *string, error) {
	bs := getBaseSpreadsheetSource(source)

	f, err := excelize.OpenReader(sheet, excelize.Options{})
	if err != nil {
		zap.L().Error("Failed to open spreadsheet", zap.Error(err))
		return nil, nil, fmt.Errorf("Failed to open spreadsheet")
	}

	defer func() {
		// Close the spreadsheet.
		if err := f.Close(); err != nil {
			zap.L().Error("Failed to close spreadsheet", zap.Error(err))
		}
	}()

	// currently default to last sheet
	sheetName := f.GetSheetName(f.SheetCount - 1)

	// get correct row
	row, err := getRowForCategory(bs, f, category, sheetName)
	if err != nil {
		return nil, nil, err
	}

	// get the cell formula
	cell := fmt.Sprintf("%s%d", bs.CostValueColumn, *row)
	form, err := f.GetCellFormula(sheetName, cell)
	if err != nil {
		zap.L().Error("Failed to get cell formula", zap.Error(err))
		return nil, nil, fmt.Errorf("Failed to get cell formula")
	}

	var valueToAdd *string
	if len(form) == 0 {
		// this is either an empty cell or it has a value which we must not lose
		val, err := f.GetCellValue(sheetName, cell)
		if err != nil {
			zap.L().Error("Failed to get cell value, borking", zap.Error(err))
			return nil, nil, fmt.Errorf("Failed to get cell value")
		}

		numericRegex := regexp.MustCompile(`-?\d+(?:,\d{3})*(?:\.\d+)?`)
		numericStr := numericRegex.FindString(val)
		numericStr = strings.ReplaceAll(numericStr, ",", "")

		if len(numericStr) > 0 {
			floatVal, err := strconv.ParseFloat(numericStr, 32)
			if err != nil {
				zap.L().Warn("Failed to parse numeric value from cell", zap.Error(err), zap.String("string", numericStr))
			} else if floatVal != 0 {
				valueToAdd = &numericStr
				zap.L().Info("Found existing value to add to fomula", zap.String("value", numericStr))
			}
		}
	}

	zap.L().Info("Staring formula update", zap.String("current formula", form))

	var updatedFormula string
	isValue := false // whether to use set value rather than set formula

	if (len(form) == 0 || strings.Compare(form, "0") == 0) && valueToAdd == nil {
		// set the first value
		updatedFormula = fmt.Sprintf("%.2f", value)
		isValue = true
	} else if valueToAdd != nil {
		// include the old value
		updatedFormula = fmt.Sprintf("%s+%.2f", *valueToAdd, value)
	} else {
		// add the value to the formula
		updatedFormula = fmt.Sprintf("%s+%.2f", form, value)
	}

	if isValue {
		zap.L().Info(
			"Setting cell value as there was nothing there. Will change to formula next time.",
			zap.Float32("value", value),
			zap.String("sheet", sheetName),
			zap.String("cell", cell),
		)

		if err := f.SetCellValue(sheetName, cell, value); err != nil {
			zap.L().Error("Failed to set the cell's value", zap.Error(err))
			return nil, nil, fmt.Errorf("Failed to set cell value")
		}

		// We can skip the rest as it is all related to formulas
		updatedValue, err := f.GetCellValue(sheetName, cell)
		if err != nil {
			zap.L().Error("Failed to get updated cell value", zap.Error(err))
			return nil, nil, fmt.Errorf("Failed to get updated value")
		}

		// return the spreadhseet as an io.Reader
		var buffer bytes.Buffer
		if err := f.Write(&buffer); err != nil {
			zap.L().Error("Failed to write spreadsheet to buffer", zap.Error(err))
		}

		return bytes.NewReader(buffer.Bytes()), &updatedValue, nil
	}

	zap.L().Info("Setting cell formula", zap.String("formula", updatedFormula), zap.String("Sheet", sheetName), zap.String("cell", cell))

	if err := f.SetCellFormula(sheetName, cell, updatedFormula); err != nil {
		zap.L().Error("Failed to set the cell's formula", zap.Error(err), zap.String("formula", updatedFormula))
		return nil, nil, fmt.Errorf("Failed to set cell formula")
	}

	updatedVal, err := f.CalcCellValue(sheetName, cell)
	if err != nil {
		zap.L().Error("Failed to get updated cell value from formula", zap.Error(err))
		return nil, nil, fmt.Errorf("Failed to get updated value from formula")
	}

	zap.L().Info("Calculated new value", zap.String("val", updatedVal))

	if err := f.UpdateLinkedValue(); err != nil {
		zap.L().Warn("Failed to updated linked values", zap.Error(err))
	}

	// return the spreadhseet as an io.Reader
	var buffer bytes.Buffer
	if err := f.Write(&buffer); err != nil {
		zap.L().Error("Failed to write spreadsheet to buffer", zap.Error(err))
	}

	return bytes.NewReader(buffer.Bytes()), &updatedVal, nil
}

func (s *ExcelerizeSpreadsheetService) ReadValueForCategory(source model.SpreadsheetSource, sheet io.Reader, category string, details bool) (*string, error) {
	bs := getBaseSpreadsheetSource(source)

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

	// get correct row
	row, err := getRowForCategory(bs, f, category, sheetName)
	if err != nil {
		return nil, err
	}

	// get the cell value
	cell := fmt.Sprintf("%s%d", bs.CostValueColumn, *row)

	if details {
		// get the formula
		form, err := f.GetCellFormula(sheetName, cell)
		if err != nil {
			zap.L().Error("Failed to get cell formula", zap.Error(err))
			return nil, fmt.Errorf("Failed to get cell formula")
		}

		if len(form) > 0 {
			return &form, nil
		}
	}

	val, err := f.CalcCellValue(sheetName, cell)
	if err != nil {
		zap.L().Error("Failed to get cell value", zap.Error(err))
		return nil, fmt.Errorf("Failed to get cell value")
	}

	if details {
		// there was no formula but strip £ so it looks like a sum
		stripped := strings.ReplaceAll(val, "£", "")
		return &stripped, nil
	}

	return &val, nil
}

type RemovedResult struct {
	ModifiedSheet io.Reader
	OldValue      string
	RemovedValue  string
	NewValue      string
}

func (s *ExcelerizeSpreadsheetService) RemoveLastValueForCategory(source model.SpreadsheetSource, sheet io.Reader, category string) (*RemovedResult, error) {
	bs := getBaseSpreadsheetSource(source)

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

	// get correct row
	row, err := getRowForCategory(bs, f, category, sheetName)
	if err != nil {
		return nil, err
	}

	// get the cell formula
	cell := fmt.Sprintf("%s%d", bs.CostValueColumn, *row)
	form, err := f.GetCellFormula(sheetName, cell)
	if err != nil {
		zap.L().Error("Failed to get cell formula", zap.Error(err))
		return nil, fmt.Errorf("Failed to get cell formula")
	}

	if len(form) == 0 || !strings.Contains(form, "+") {
		// set the cell value to 0
		val, err := f.CalcCellValue(sheetName, cell)
		if err != nil {
			zap.L().Error("Failed to calc cell value", zap.Error(err))
			return nil, fmt.Errorf("Failed to calc cell value")
		}

		if err := f.SetCellValue(sheetName, cell, 0); err != nil {
			zap.L().Error("Failed to set cell value", zap.Error(err))
			return nil, fmt.Errorf("Failed to set cell value")
		}

		// return the spreadhseet as an io.Reader
		var buffer bytes.Buffer
		if err := f.Write(&buffer); err != nil {
			zap.L().Error("Failed to write spreadsheet to buffer", zap.Error(err))
		}

		return &RemovedResult{
			ModifiedSheet: bytes.NewBuffer(buffer.Bytes()),
			OldValue:      val,
			RemovedValue:  val,
			NewValue:      "£0",
		}, nil
	}

	// get the current cell value
	val, err := f.CalcCellValue(sheetName, cell)
	if err != nil {
		zap.L().Error("Failed to calc cell value", zap.Error(err))
		return nil, fmt.Errorf("Failed to calc cell value")
	}

	// remove after the last +
	lastIdx := strings.LastIndex(form, "+")
	toRemove := form[lastIdx+1:]
	newForm := form[:lastIdx]

	// set the new formula
	if err := f.SetCellFormula(sheetName, cell, newForm); err != nil {
		zap.L().Error("Failed to set cell formula", zap.Error(err))
		return nil, fmt.Errorf("Failed to set cell formula")
	}

	updatedVal, err := f.CalcCellValue(sheetName, cell)
	if err != nil {
		zap.L().Error("Failed to get updated cell value from formula", zap.Error(err))
		return nil, fmt.Errorf("Failed to get updated value from formula")
	}

	if err := f.UpdateLinkedValue(); err != nil {
		zap.L().Warn("Failed to updated linked values", zap.Error(err))
	}

	// return the spreadhseet as an io.Reader
	var buffer bytes.Buffer
	if err := f.Write(&buffer); err != nil {
		zap.L().Error("Failed to write spreadsheet to buffer", zap.Error(err))
	}

	return &RemovedResult{
		ModifiedSheet: bytes.NewBuffer(buffer.Bytes()),
		OldValue:      val,
		RemovedValue:  fmt.Sprintf("£%s", toRemove),
		NewValue:      updatedVal,
	}, nil
}

// private

func getBaseSpreadsheetSource(source model.SpreadsheetSource) *model.BaseSpreadsheetSource {
	switch s := source.(type) {
	case *model.NextcloudSpreadsheetSource:
		return &s.BaseSpreadsheetSource
	case *model.BaseSpreadsheetSource:
		return s
	default:
		zap.L().Panic("Spreadsheet source does not inherit from base spreadsheet source", zap.String("type", source.GetType()))
		return nil
	}
}

func getRowForCategory(source *model.BaseSpreadsheetSource, file *excelize.File, category string, sheetName string) (*uint, error) {
	// iterate key column until find the category
	currentRow := uint(1)
	comparison := strings.ToLower(strings.ReplaceAll(category, " ", ""))
	emptyCellCount := uint(0)

	for {
		cell := fmt.Sprintf("%s%d", source.CostNameColumn, currentRow)
		val, err := file.GetCellValue(sheetName, cell)
		if err != nil {
			zap.L().Error("Failed to get value for cell", zap.String("cell", cell), zap.Error(err))
			return nil, fmt.Errorf("Failed to get value for cell")
		}

		normVal := strings.ToLower(strings.ReplaceAll(val, " ", ""))
		if normVal == comparison {
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

	return &currentRow, nil
}
