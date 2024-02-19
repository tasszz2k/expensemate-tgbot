package gsheettypes

import "fmt"

// Expensemate Database Spreadsheet
const (
	UserSheetMappingSheetName  = "user_sheet_mappings"
	UserSheetMappingNextIdCell = "B1"
	UserSheetMappingTopRow     = 2
	UserSheetMappingLeftCol    = "A"
	UserSheetMappingRightCol   = "H"
)

// My Expensemate Spreadsheet
const (
	ExpensemateDatabaseSheetName       = "Database"
	ExpensemateDatabaseCurrentPageCell = "B2"

	ExpensemateExpensesNextIdCell = "B2"
	ExpensemateExpensesTopRow     = 3
	ExpensemateExpensesLeftCol    = "A"
	ExpensemateExpensesRightCol   = "G"
)

func BuildCell(sheetName string, cell string) string {
	return sheetName + "!" + cell
}

func BuildRange(sheetName string, startCol string, startRow int, endCol string, endRow int) string {
	return fmt.Sprintf("%s!%s%d:%s%d", sheetName, startCol, startRow, endCol, endRow)
}
