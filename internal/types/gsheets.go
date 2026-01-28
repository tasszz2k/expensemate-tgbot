package types

import (
	"fmt"
	"regexp"
)

// GSheetsAction represents a Google Sheets related action
type GSheetsAction string

const (
	GSheetsActionConfigure        GSheetsAction = "configure"
	GSheetsActionHelp             GSheetsAction = "help"
	GSheetsActionUpdateActivePage GSheetsAction = "update_current_page"
)

func (a GSheetsAction) String() string {
	return string(a)
}

// Database Spreadsheet (Central - user mappings)
const (
	UserMappingSheetName  = "user_sheet_mappings"
	UserMappingNextIDCell = "B1"
	UserMappingTopRow     = 2
	UserMappingLeftCol    = "A"
	UserMappingRightCol   = "H"
)

// User Expense Spreadsheet (Per-user)
const (
	DatabaseSheetName      = "Database"
	DatabaseActivePageCell = "B2"

	ExpensesNextIDCell = "B2"
	ExpensesTopRow     = 3
	ExpensesLeftCol    = "A"
	ExpensesRightCol   = "G"

	ExpensesReportRange   = "I3:J9"
	ExpensesCategoryRange = "L3:N15"
)

// sheetNamePattern validates YYYY_MM format
var sheetNamePattern = regexp.MustCompile(`^\d{4}_\d{2}$`)

// BuildCell creates a cell reference like "Sheet!A1"
func BuildCell(sheetName, cell string) string {
	return sheetName + "!" + cell
}

// BuildRange creates a range reference like "Sheet!A1:G10"
func BuildRange(sheetName, rangeStr string) string {
	return sheetName + "!" + rangeStr
}

// BuildRangeFromCells creates a range from start/end columns and rows
func BuildRangeFromCells(sheetName, startCol string, startRow int, endCol string, endRow int) string {
	return fmt.Sprintf("%s!%s%d:%s%d", sheetName, startCol, startRow, endCol, endRow)
}

// IsValidSheetName validates if input matches YYYY_MM format
func IsValidSheetName(input string) bool {
	return sheetNamePattern.MatchString(input)
}
