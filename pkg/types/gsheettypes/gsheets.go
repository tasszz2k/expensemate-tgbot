package gsheettypes

import (
	"fmt"
	"regexp"
)

// Expensemate Database Spreadsheet
const (
	UserSheetMappingSheetName  = "user_sheet_mappings"
	UserSheetMappingNextIdCell = "B1"
	UserSheetMappingTopRow     = 2
	UserSheetMappingLeftCol    = "A"
	UserSheetMappingRightCol   = "H"
)

// My Expensemate Spreadsheet > Expenses
const (
	ExpensemateDatabaseSheetName       = "Database"
	ExpensemateDatabaseCurrentPageCell = "B2"

	ExpensemateExpensesNextIdCell = "B2"
	ExpensemateExpensesTopRow     = 3
	ExpensemateExpensesLeftCol    = "A"
	ExpensemateExpensesRightCol   = "G"
)

// My Expensemate Spreadsheet > Report
const (
	ExpensemateExpensesReportRange   = "I3:J9"
	ExpensemateExpensesCategoryRange = "L3:N15"
)

// Regular expression pattern for "YYYY_MM" format
const pattern = `^\d{4}_\d{2}$`

func BuildCell(sheetName string, cell string) string {
	return sheetName + "!" + cell
}

func BuildRangeFromCells(
	sheetName string,
	startCol string,
	startRow int,
	endCol string,
	endRow int,
) string {
	return fmt.Sprintf("%s!%s%d:%s%d", sheetName, startCol, startRow, endCol, endRow)
}

func BuildRange(sheetName string, rangeString string) string {
	return sheetName + "!" + rangeString
}

func IsFormatValidSheetName(input string) bool {
	// Compile the regular expression pattern
	regex := regexp.MustCompile(pattern)

	// Check if the input string matches the pattern
	return regex.MatchString(input)
}
