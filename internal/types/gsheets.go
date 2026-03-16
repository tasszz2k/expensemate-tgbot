package types

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"time"
)

// GSheetsAction represents a Google Sheets related action
type GSheetsAction string

const (
	GSheetsActionConfigure              GSheetsAction = "configure"
	GSheetsActionHelp                   GSheetsAction = "help"
	GSheetsActionUpdateActivePage       GSheetsAction = "update_current_page"
	GSheetsActionUpdateActivePageManual GSheetsAction = "update_current_page_manual"
	GSheetsActionCreateNewMonth         GSheetsAction = "create_new_month"
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

	ExpensesReportRange   = "I3:J10"
	ExpensesCategoryRange = "N3:P22"

	GroupBudgetCol                = "K"
	CategoryBudgetCol             = "Q"
	GroupReportWithBudgetRange    = "I3:L10"
	CategoryReportWithBudgetRange = "N3:R22"

	// New month sheet creation
	ExpenseDataStartRow    = 10
	NewMonthNextExpenseID  = 10
	SalaryCellRef          = "C4"
	InvestmentFormulaRange = "C5:C9"
	InvestmentNoteRange    = "AF16:AF30"
	AssetCurrentRange   = "T2:U9"
	AssetLastMonthRange = "T17:U24"
	SheetNameCell          = "A1"
)

// GroupBudgetRow maps groups to their sheet row numbers in the group report (I3:L10)
var GroupBudgetRow = map[Group]int{
	GroupIncome:        3,
	GroupInvestmentOut: 4,
	// Row 5 is INVESTMENT PROFIT (formula row, not user-settable)
	GroupMustHave:   6,
	GroupNiceToHave: 7,
	GroupWaste:      8,
	GroupFamily:     9,
	GroupLover:      10,
}

// CategoryBudgetRow maps categories to their sheet row numbers in the category report (N3:R22)
var CategoryBudgetRow = map[Category]int{
	CategoryUnclassified:  3,
	CategoryFood:          4,
	CategoryCafe:          5,
	CategoryGroceries:     6,
	CategoryTransport:     7,
	CategoryEntertainment: 8,
	CategoryMiscellaneous: 9,
	CategorySubscription:  10,
	CategoryHousing:       11,
	CategoryPersonalCare:  12,
	CategoryHealthcare:    13,
	CategoryClothing:      14,
	CategoryEducation:     15,
	CategoryTech:          16,
	CategoryTravel:        17,
	CategoryPresent:       18,
	CategoryLifeEvents:    19,
	CategoryLover:         20,
	CategoryFamily:        21,
	CategoryLostMoney:     22,
}

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

// parseSheetMonth parses a YYYY_MM sheet name into a time.Time
func parseSheetMonth(name string) (time.Time, error) {
	parts := strings.SplitN(name, "_", 2)
	if len(parts) != 2 {
		return time.Time{}, fmt.Errorf("invalid sheet name format: %s", name)
	}
	year, err := strconv.Atoi(parts[0])
	if err != nil {
		return time.Time{}, fmt.Errorf("invalid year in sheet name: %s", name)
	}
	month, err := strconv.Atoi(parts[1])
	if err != nil || month < 1 || month > 12 {
		return time.Time{}, fmt.Errorf("invalid month in sheet name: %s", name)
	}
	return time.Date(year, time.Month(month), 1, 0, 0, 0, 0, time.UTC), nil
}

// NextMonthName returns the next month sheet name (handles year rollover, e.g. 2025_12 -> 2026_01)
func NextMonthName(current string) (string, error) {
	t, err := parseSheetMonth(current)
	if err != nil {
		return "", err
	}
	next := t.AddDate(0, 1, 0)
	return fmt.Sprintf("%04d_%02d", next.Year(), next.Month()), nil
}

// PrevMonthName returns the previous month sheet name (handles year rollover, e.g. 2026_01 -> 2025_12)
func PrevMonthName(current string) (string, error) {
	t, err := parseSheetMonth(current)
	if err != nil {
		return "", err
	}
	prev := t.AddDate(0, -1, 0)
	return fmt.Sprintf("%04d_%02d", prev.Year(), prev.Month()), nil
}
