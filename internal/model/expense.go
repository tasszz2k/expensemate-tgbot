package model

import (
	"errors"
	"fmt"
	"strings"
	"time"

	"expensemate-tgbot/internal/types"
	"expensemate-tgbot/internal/util/currency"
	timepkg "expensemate-tgbot/internal/util/time"

	"github.com/spf13/cast"
)

// Expense represents a single expense entry
type Expense struct {
	ID       types.ID       `json:"id"`
	Name     string         `json:"name"`
	Amount   types.Unsigned `json:"amount"`
	Group    types.Group    `json:"group"`
	Category types.Category `json:"category"`
	Date     time.Time      `json:"date"`
	Note     string         `json:"note"`
}

// ParseRowToExpense parses a Google Sheets row into an Expense
func ParseRowToExpense(row []interface{}) (Expense, error) {
	// Pad row to ensure at least 7 columns
	for len(row) < 7 {
		row = append(row, "")
	}

	// Parse amount - try as number first, then formatted string
	var amount types.Unsigned
	if num := cast.ToUint64(row[2]); num > 0 {
		amount = types.Unsigned(num)
	} else {
		// Fallback for formatted strings (legacy data)
		var err error
		amount, err = currency.ReverseFormatVND(cast.ToString(row[2]))
		if err != nil {
			return Expense{}, fmt.Errorf("invalid amount %q: %w", row[2], err)
		}
	}

	date, err := timepkg.ParseDateOnly(cast.ToString(row[5]))
	if err != nil {
		date = timepkg.Now()
	}

	// Column order: ID(0), Name(1), Amount(2), Group(3), Category(4), Date(5), Note(6)
	return Expense{
		ID:       types.ID(cast.ToInt64(row[0])),
		Name:     cast.ToString(row[1]),
		Amount:   amount,
		Group:    types.Group(cast.ToString(row[3])),
		Category: types.Category(cast.ToString(row[4])),
		Date:     date,
		Note:     cast.ToString(row[6]),
	}, nil
}

// ParseResult contains the parsed expense and flags indicating what was explicitly provided
type ParseResult struct {
	Expense          Expense
	GroupProvided    bool // true if user explicitly provided group
	CategoryProvided bool // true if user explicitly provided category
}

// ParseTextToExpense parses user input text into an Expense
func ParseTextToExpense(text string) (ParseResult, error) {
	lines := strings.Split(text, "\n")
	if len(lines) < 2 {
		return ParseResult{}, errors.New("invalid format: need at least name and amount")
	}

	// Pad to 6 lines
	for len(lines) < 6 {
		lines = append(lines, "")
	}

	// Parse name (required)
	name := strings.TrimSpace(lines[0])
	if name == "" {
		return ParseResult{}, errors.New("expense name is required")
	}

	// Parse amount (required)
	amount := currency.ParseAmount(lines[1])
	if amount == 0 {
		return ParseResult{}, fmt.Errorf("invalid amount: %q", lines[1])
	}

	// Parse group (optional, default: MUST HAVE)
	// Note: Group is before Category to match spreadsheet column order (D=Group, E=Category)
	var group types.Group
	groupInput := strings.TrimSpace(lines[2])
	groupProvided := groupInput != ""
	if !groupProvided {
		group = types.GroupMustHave
	} else {
		var ok bool
		group, ok = types.GetGroupByAlias(groupInput)
		if !ok {
			return ParseResult{}, fmt.Errorf("invalid group: %q (use /expenses_help for list)", groupInput)
		}
	}

	// Parse category (optional, default: Unclassified)
	var category types.Category
	categoryInput := strings.TrimSpace(lines[3])
	categoryProvided := categoryInput != ""
	if !categoryProvided {
		category = types.CategoryUnclassified
	} else {
		var ok bool
		category, ok = types.GetCategoryByAlias(categoryInput)
		if !ok {
			return ParseResult{}, fmt.Errorf("invalid category: %q (use /expenses_help for list)", categoryInput)
		}
	}

	// Parse date (optional, default: today in local timezone)
	date := timepkg.Now()
	if lines[4] != "" {
		parsed, err := time.ParseInLocation(timepkg.DateOnlyFormat, lines[4], timepkg.LocalLocation)
		if err == nil {
			date = parsed
		}
	}

	return ParseResult{
		Expense: Expense{
			Name:     name,
			Amount:   types.Unsigned(amount),
			Group:    group,
			Category: category,
			Date:     date,
			Note:     strings.TrimSpace(lines[5]),
		},
		GroupProvided:    groupProvided,
		CategoryProvided: categoryProvided,
	}, nil
}

// FormatHTML returns the expense as HTML-formatted string
func (e Expense) FormatHTML() string {
	// Base fields
	result := fmt.Sprintf(
		`<b>ID</b>: <i>%d</i>
<b>Name</b>: <i>%s</i>
<b>Amount</b>: <i>%s</i>
<b>Group</b>: <i>%s</i>`,
		e.ID,
		e.Name,
		currency.FormatVND(e.Amount),
		e.Group,
	)

	// Only show category if this group type needs it (not Income/Investment Out)
	if e.Group.NeedsCategory() {
		result += fmt.Sprintf("\n<b>Category</b>: <i>%s</i>", e.Category)
	}

	// Add date and note
	result += fmt.Sprintf(`
<b>Date</b>: <i>%s</i>
<b>Note</b>: <i>%s</i>`,
		timepkg.FormatDateOnly(e.Date),
		e.Note,
	)

	return result
}

// ToRow converts Expense to a Google Sheets row
// Column order: ID, Name, Amount, Group, Category, Date, Note
func (e Expense) ToRow() []interface{} {
	// Leave category empty for Income/Investment Out groups
	var category interface{} = e.Category
	if !e.Group.NeedsCategory() {
		category = ""
	}

	return []interface{}{
		e.ID, // types.ID (int64)
		e.Name,
		e.Amount, // types.Unsigned (uint64) - spreadsheet formats it
		e.Group,  // types.Group
		category, // types.Category or empty string
		timepkg.FormatDateTime(e.Date),
		e.Note,
	}
}
