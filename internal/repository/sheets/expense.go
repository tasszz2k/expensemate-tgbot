package sheets

import (
	"context"
	"fmt"

	"expensemate-tgbot/internal/log"
	"expensemate-tgbot/internal/model"
	"expensemate-tgbot/internal/types"

	"github.com/sirupsen/logrus"
	"github.com/spf13/cast"
)

// ExpenseRepository handles expense data in Google Sheets
type ExpenseRepository struct {
	client *Client
}

// NewExpenseRepository creates a new ExpenseRepository
func NewExpenseRepository(client *Client) *ExpenseRepository {
	return &ExpenseRepository{client: client}
}

// GetActivePage returns the active page name from Database sheet
func (r *ExpenseRepository) GetActivePage(ctx context.Context, spreadsheetID string) (string, error) {
	cell := types.BuildCell(types.DatabaseSheetName, types.DatabaseActivePageCell)

	log.Debug("reading active page", logrus.Fields{
		"spreadsheet_id": spreadsheetID,
		"cell":           cell,
	})

	value, err := r.client.GetValue(ctx, spreadsheetID, cell)
	if err != nil {
		return "", fmt.Errorf("getting active page: %w", err)
	}

	log.Debug("active page value", logrus.Fields{
		"value": value,
	})

	return value, nil
}

// GetNextID returns the next expense ID from the active page
func (r *ExpenseRepository) GetNextID(ctx context.Context, spreadsheetID, sheetName string) (types.ID, error) {
	cell := types.BuildCell(sheetName, types.ExpensesNextIDCell)
	value, err := r.client.GetValue(ctx, spreadsheetID, cell)
	if err != nil {
		return 0, fmt.Errorf("getting next ID: %w", err)
	}
	return types.ID(cast.ToInt64(value)), nil
}

// UpdateNextID updates the next expense ID counter
func (r *ExpenseRepository) UpdateNextID(ctx context.Context, spreadsheetID, sheetName string, nextID types.ID) error {
	cell := types.BuildCell(sheetName, types.ExpensesNextIDCell)
	return r.client.Update(ctx, spreadsheetID, cell, [][]interface{}{{int64(nextID)}})
}

// Create creates a new expense in the spreadsheet
func (r *ExpenseRepository) Create(ctx context.Context, spreadsheetID, sheetName string, expense *model.Expense) error {
	row := types.ExpensesTopRow + int(expense.ID)
	writeRange := types.BuildRangeFromCells(sheetName, types.ExpensesLeftCol, row, types.ExpensesRightCol, row)

	log.Debug("writing expense", logrus.Fields{
		"spreadsheet_id": spreadsheetID,
		"range":          writeRange,
		"expense_id":     expense.ID,
	})

	return r.client.Update(ctx, spreadsheetID, writeRange, [][]interface{}{expense.ToRow()})
}

// GetByID retrieves an expense by ID
func (r *ExpenseRepository) GetByID(ctx context.Context, spreadsheetID, sheetName string, id types.ID) (*model.Expense, error) {
	row := types.ExpensesTopRow + int(id)
	readRange := types.BuildRangeFromCells(sheetName, types.ExpensesLeftCol, row, types.ExpensesRightCol, row)

	resp, err := r.client.Get(ctx, spreadsheetID, readRange)
	if err != nil {
		return nil, fmt.Errorf("reading expense %d: %w", id, err)
	}

	if len(resp.Values) == 0 {
		return nil, nil
	}

	expense, err := model.ParseRowToExpense(resp.Values[0])
	if err != nil {
		return nil, fmt.Errorf("parsing expense %d: %w", id, err)
	}

	return &expense, nil
}

// GetRecent retrieves the most recent expenses
func (r *ExpenseRepository) GetRecent(ctx context.Context, spreadsheetID, sheetName string, lastID types.ID, limit int) ([]model.Expense, error) {
	startRow := max(types.ExpensesTopRow+int(lastID)-limit+1, types.ExpensesTopRow+1)
	endRow := types.ExpensesTopRow + int(lastID)

	readRange := types.BuildRangeFromCells(sheetName, types.ExpensesLeftCol, startRow, types.ExpensesRightCol, endRow)

	log.Debug("reading recent expenses", logrus.Fields{
		"spreadsheet_id": spreadsheetID,
		"range":          readRange,
	})

	resp, err := r.client.Get(ctx, spreadsheetID, readRange)
	if err != nil {
		return nil, fmt.Errorf("reading recent expenses: %w", err)
	}

	var expenses []model.Expense
	for _, row := range resp.Values {
		if len(row) == 0 || cast.ToString(row[0]) == "" {
			continue
		}
		expense, err := model.ParseRowToExpense(row)
		if err != nil {
			continue
		}
		// Skip soft-deleted expenses (empty name)
		if expense.Name == "" {
			continue
		}
		expenses = append(expenses, expense)
	}

	return expenses, nil
}

// Update updates an existing expense
func (r *ExpenseRepository) Update(ctx context.Context, spreadsheetID, sheetName string, expense *model.Expense) error {
	return r.Create(ctx, spreadsheetID, sheetName, expense)
}

// SoftDelete marks an expense as deleted by clearing fields except ID and Note
func (r *ExpenseRepository) SoftDelete(ctx context.Context, spreadsheetID, sheetName string, id types.ID, note string) error {
	row := types.ExpensesTopRow + int(id)
	writeRange := types.BuildRangeFromCells(sheetName, types.ExpensesLeftCol, row, types.ExpensesRightCol, row)

	// Keep ID, clear other fields, update note with deletion info
	values := [][]interface{}{{
		int64(id), // ID
		"",        // Name (cleared)
		"",        // Amount (cleared)
		"",        // Group (cleared)
		"",        // Category (cleared)
		"",        // Date (cleared)
		note,      // Note (audit log)
	}}

	return r.client.Update(ctx, spreadsheetID, writeRange, values)
}

// GetGroupReport retrieves the group-based expense report
func (r *ExpenseRepository) GetGroupReport(ctx context.Context, spreadsheetID, sheetName string) ([][]interface{}, error) {
	readRange := types.BuildRange(sheetName, types.ExpensesReportRange)
	resp, err := r.client.Get(ctx, spreadsheetID, readRange)
	if err != nil {
		return nil, fmt.Errorf("reading group report: %w", err)
	}
	return resp.Values, nil
}

// GetGroupReportWithBudget reads the group report with budget data (range I3:L10)
func (r *ExpenseRepository) GetGroupReportWithBudget(ctx context.Context, spreadsheetID, sheetName string) ([]model.BudgetEntry, error) {
	readRange := types.BuildRange(sheetName, types.GroupReportWithBudgetRange)
	resp, err := r.client.GetUnformatted(ctx, spreadsheetID, readRange)
	if err != nil {
		return nil, fmt.Errorf("reading group report with budget: %w", err)
	}

	var entries []model.BudgetEntry
	for i, row := range resp.Values {
		entry := model.ParseGroupBudgetRow(row, 3+i) // rows start at 3
		if entry.Name != "" {
			entries = append(entries, entry)
		}
	}
	return entries, nil
}

// GetCategoryReport retrieves the category-based expense report
func (r *ExpenseRepository) GetCategoryReport(ctx context.Context, spreadsheetID, sheetName string) ([][]interface{}, error) {
	readRange := types.BuildRange(sheetName, types.ExpensesCategoryRange)
	resp, err := r.client.Get(ctx, spreadsheetID, readRange)
	if err != nil {
		return nil, fmt.Errorf("reading category report: %w", err)
	}
	return resp.Values, nil
}

// GetCategoryReportWithBudget reads the category report with budget data (range N3:R21)
func (r *ExpenseRepository) GetCategoryReportWithBudget(ctx context.Context, spreadsheetID, sheetName string) ([]model.BudgetEntry, error) {
	readRange := types.BuildRange(sheetName, types.CategoryReportWithBudgetRange)
	resp, err := r.client.GetUnformatted(ctx, spreadsheetID, readRange)
	if err != nil {
		return nil, fmt.Errorf("reading category report with budget: %w", err)
	}

	var entries []model.BudgetEntry
	for i, row := range resp.Values {
		entry := model.ParseCategoryBudgetRow(row, 3+i) // rows start at 3
		if entry.Name != "" {
			entries = append(entries, entry)
		}
	}
	return entries, nil
}

// GetExpenseSummary reads the summary rows (I11:L13) and parses them using ParseGroupBudgetRow.
// Returns entries for Self Expenses, Total Expenses, and Net Change.
func (r *ExpenseRepository) GetExpenseSummary(ctx context.Context, spreadsheetID, sheetName string) ([]model.BudgetEntry, error) {
	readRange := types.BuildRange(sheetName, types.ExpensesSummaryRange)
	resp, err := r.client.GetUnformatted(ctx, spreadsheetID, readRange)
	if err != nil {
		return nil, fmt.Errorf("reading expense summary: %w", err)
	}

	var entries []model.BudgetEntry
	for i, row := range resp.Values {
		entry := model.ParseGroupBudgetRow(row, 11+i)
		if entry.Name != "" {
			entries = append(entries, entry)
		}
	}
	return entries, nil
}

// SetBudget writes a budget amount to a specific cell (column K for groups, Q for categories)
func (r *ExpenseRepository) SetBudget(ctx context.Context, spreadsheetID, sheetName, col string, row int, amount uint64) error {
	writeRange := fmt.Sprintf("%s!%s%d", sheetName, col, row)
	return r.client.Update(ctx, spreadsheetID, writeRange, [][]interface{}{{amount}})
}

// ClearBudget clears a budget cell
func (r *ExpenseRepository) ClearBudget(ctx context.Context, spreadsheetID, sheetName, col string, row int) error {
	clearRange := fmt.Sprintf("%s!%s%d", sheetName, col, row)
	return r.client.ClearValues(ctx, spreadsheetID, clearRange)
}

// UpdateCategory updates only the category column for an expense
func (r *ExpenseRepository) UpdateCategory(ctx context.Context, spreadsheetID, sheetName string, expenseID types.ID, category types.Category) error {
	row := types.ExpensesTopRow + int(expenseID)
	// Category is in column E (Group is D, Category is E)
	writeRange := fmt.Sprintf("%s!E%d", sheetName, row)

	log.Debug("updating expense category", logrus.Fields{
		"spreadsheet_id": spreadsheetID,
		"range":          writeRange,
		"expense_id":     expenseID,
		"category":       category,
	})

	return r.client.Update(ctx, spreadsheetID, writeRange, [][]interface{}{{string(category)}})
}

// UpdateGroup updates only the group column for an expense
func (r *ExpenseRepository) UpdateGroup(ctx context.Context, spreadsheetID, sheetName string, expenseID types.ID, group types.Group) error {
	row := types.ExpensesTopRow + int(expenseID)
	// Group is in column D (Group is D, Category is E)
	writeRange := fmt.Sprintf("%s!D%d", sheetName, row)

	log.Debug("updating expense group", logrus.Fields{
		"spreadsheet_id": spreadsheetID,
		"range":          writeRange,
		"expense_id":     expenseID,
		"group":          group,
	})

	return r.client.Update(ctx, spreadsheetID, writeRange, [][]interface{}{{string(group)}})
}

// GetSheetID returns the numeric sheet ID for a given sheet name
func (r *ExpenseRepository) GetSheetID(ctx context.Context, spreadsheetID, sheetName string) (int64, error) {
	return r.client.GetSheetID(ctx, spreadsheetID, sheetName)
}

// MonthReport holds parsed report data for a single monthly sheet
type MonthReport struct {
	Groups   []model.BudgetEntry
	Categories []model.BudgetEntry
	Summary  []model.BudgetEntry
}

// GetMultiMonthReports reads group, category, and summary reports from multiple
// monthly sheets in a single BatchGet call.
func (r *ExpenseRepository) GetMultiMonthReports(ctx context.Context, spreadsheetID string, sheetNames []string) (map[string]MonthReport, error) {
	if len(sheetNames) == 0 {
		return nil, nil
	}

	var ranges []string
	for _, name := range sheetNames {
		ranges = append(ranges,
			types.BuildRange(name, types.InsightsGroupAndSummaryRange),
			types.BuildRange(name, types.InsightsCategoryRange),
		)
	}

	log.Debug("batch reading multi-month reports", logrus.Fields{
		"spreadsheet_id": spreadsheetID,
		"sheet_count":    len(sheetNames),
		"range_count":    len(ranges),
	})

	results, err := r.client.BatchGet(ctx, spreadsheetID, ranges)
	if err != nil {
		return nil, fmt.Errorf("batch reading reports: %w", err)
	}

	reports := make(map[string]MonthReport, len(sheetNames))
	for i, name := range sheetNames {
		report := MonthReport{}

		groupIdx := i * 2
		catIdx := i*2 + 1

		if groupIdx < len(results) && results[groupIdx] != nil {
			for j, row := range results[groupIdx].Values {
				if j < 8 {
					entry := model.ParseGroupBudgetRow(row, 3+j)
					if entry.Name != "" {
						report.Groups = append(report.Groups, entry)
					}
				} else {
					entry := model.ParseGroupBudgetRow(row, 3+j)
					if entry.Name != "" {
						report.Summary = append(report.Summary, entry)
					}
				}
			}
		}

		if catIdx < len(results) && results[catIdx] != nil {
			for j, row := range results[catIdx].Values {
				entry := model.ParseCategoryBudgetRow(row, 3+j)
				if entry.Name != "" {
					report.Categories = append(report.Categories, entry)
				}
			}
		}

		reports[name] = report
	}

	return reports, nil
}
