package model

import (
	"fmt"

	"expensemate-tgbot/internal/types"
	"expensemate-tgbot/internal/util/currency"

	"github.com/spf13/cast"
)

// BudgetEntry represents a single budget entry for a group or category
type BudgetEntry struct {
	Name       string
	Spent      int64
	Percentage string // Only for categories (e.g., "2.35%")
	Budget     int64
	Remaining  int64
	HasBudget  bool
	Row        int
}

// FormatBudgetLine formats a budget entry for Telegram display with emoji status indicators.
func (e BudgetEntry) FormatBudgetLine() string {
	if !e.HasBudget {
		return ""
	}

	spentStr := currency.FormatVND(types.Unsigned(abs64(e.Spent)))
	budgetStr := currency.FormatVND(types.Unsigned(e.Budget))

	if e.Remaining >= 0 {
		remainStr := currency.FormatVND(types.Unsigned(e.Remaining))
		icon := "✅"
		if e.Spent == 0 {
			icon = "⬜"
		}
		return fmt.Sprintf("%s %s: <b>%s</b> / %s\n      <i>%s left</i>", icon, e.Name, spentStr, budgetStr, remainStr)
	}

	overStr := currency.FormatVND(types.Unsigned(abs64(-e.Remaining)))
	return fmt.Sprintf("🔴 %s: <b>%s</b> / %s\n      <i>⚠️ OVER %s</i>", e.Name, spentStr, budgetStr, overStr)
}

// FormatShortBudgetLine formats a compact budget status for the after-add display.
func (e BudgetEntry) FormatShortBudgetLine() string {
	if !e.HasBudget {
		return ""
	}

	spentStr := currency.FormatVND(types.Unsigned(abs64(e.Spent)))
	budgetStr := currency.FormatVND(types.Unsigned(e.Budget))

	if e.Remaining >= 0 {
		return fmt.Sprintf("✅ %s <b>%s</b>/%s", e.Name, spentStr, budgetStr)
	}
	return fmt.Sprintf("🔴 %s <b>%s</b>/%s ⚠️", e.Name, spentStr, budgetStr)
}

// FormatTotalLine formats a compact total line for after-add display (e.g. "💰 This month: 646k / 13.4m").
func (e BudgetEntry) FormatTotalLine() string {
	spentStr := currency.FormatVND(types.Unsigned(abs64(e.Spent)))
	if !e.HasBudget {
		return fmt.Sprintf("💰 This month: <b>%s</b>", spentStr)
	}
	budgetStr := currency.FormatVND(types.Unsigned(e.Budget))
	return fmt.Sprintf("💰 This month: <b>%s</b> / %s", spentStr, budgetStr)
}

func abs64(x int64) int64 {
	if x < 0 {
		return -x
	}
	return x
}

// ParseGroupBudgetRow parses a row from I3:L10 range into a BudgetEntry.
// Row format: [Name, Amount, Budget, Remaining]
func ParseGroupBudgetRow(row []interface{}, rowIndex int) BudgetEntry {
	entry := BudgetEntry{
		Row: rowIndex,
	}

	if len(row) >= 1 {
		entry.Name = cast.ToString(row[0])
	}
	if len(row) >= 2 {
		entry.Spent = cast.ToInt64(row[1])
	}
	if len(row) >= 3 && row[2] != nil && cast.ToString(row[2]) != "" {
		entry.HasBudget = true
		entry.Budget = cast.ToInt64(row[2])
		if len(row) >= 4 && row[3] != nil {
			entry.Remaining = cast.ToInt64(row[3])
		} else {
			entry.Remaining = entry.Budget - entry.Spent
		}
	}

	return entry
}

// ParseCategoryBudgetRow parses a row from N3:R22 range into a BudgetEntry.
// Row format: [Name, Amount, Percentage, Budget, Remaining]
func ParseCategoryBudgetRow(row []interface{}, rowIndex int) BudgetEntry {
	entry := BudgetEntry{
		Row: rowIndex,
	}

	if len(row) >= 1 {
		entry.Name = cast.ToString(row[0])
	}
	if len(row) >= 2 {
		entry.Spent = cast.ToInt64(row[1])
	}
	if len(row) >= 3 {
		entry.Percentage = cast.ToString(row[2])
	}
	if len(row) >= 4 && row[3] != nil && cast.ToString(row[3]) != "" {
		entry.HasBudget = true
		entry.Budget = cast.ToInt64(row[3])
		if len(row) >= 5 && row[4] != nil {
			entry.Remaining = cast.ToInt64(row[4])
		} else {
			entry.Remaining = entry.Budget - entry.Spent
		}
	}

	return entry
}
