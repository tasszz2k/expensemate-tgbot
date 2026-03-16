package service

import (
	"context"
	"fmt"

	"expensemate-tgbot/internal/log"
	"expensemate-tgbot/internal/model"
	"expensemate-tgbot/internal/repository/sheets"
	"expensemate-tgbot/internal/types"

	"github.com/sirupsen/logrus"
)

// ExpenseService handles expense business logic
type ExpenseService struct {
	repo           *sheets.ExpenseRepository
	mappingService *MappingService
}

// NewExpenseService creates a new ExpenseService
func NewExpenseService(repo *sheets.ExpenseRepository, mappingService *MappingService) *ExpenseService {
	return &ExpenseService{
		repo:           repo,
		mappingService: mappingService,
	}
}

// AddResult contains the added expense and flags from parsing
type AddResult struct {
	Expense          *model.Expense
	GroupProvided    bool
	CategoryProvided bool
}

// Add creates a new expense for a user
func (s *ExpenseService) Add(ctx context.Context, userID types.ID, text string) (*AddResult, error) {
	// Get user's spreadsheet mapping
	mapping, err := s.mappingService.GetByUserID(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("getting user mapping: %w", err)
	}
	if mapping == nil {
		return nil, fmt.Errorf("no spreadsheet configured for user %d", userID)
	}

	spreadsheetID := mapping.SpreadsheetDocID()

	// Get active page
	activePage, err := s.repo.GetActivePage(ctx, spreadsheetID)
	if err != nil {
		return nil, fmt.Errorf("getting active page: %w", err)
	}

	// Get next ID
	nextID, err := s.repo.GetNextID(ctx, spreadsheetID, activePage)
	if err != nil {
		return nil, fmt.Errorf("getting next ID: %w", err)
	}

	// Parse expense from text
	parseResult, err := model.ParseTextToExpense(text)
	if err != nil {
		return nil, err
	}

	expense := parseResult.Expense
	expense.ID = nextID

	// Save expense
	if err := s.repo.Create(ctx, spreadsheetID, activePage, &expense); err != nil {
		return nil, fmt.Errorf("saving expense: %w", err)
	}

	// Update next ID
	if err := s.repo.UpdateNextID(ctx, spreadsheetID, activePage, nextID+1); err != nil {
		return nil, fmt.Errorf("updating next ID: %w", err)
	}

	log.WithExpense(log.Logger.WithFields(logrus.Fields{
		"user_id":        userID,
		"spreadsheet_id": spreadsheetID,
		"action":         "expenses_add",
	}), int64(expense.ID), expense.Name).
		WithField("amount", expense.Amount).
		Info("expense added successfully")

	return &AddResult{
		Expense:          &expense,
		GroupProvided:    parseResult.GroupProvided,
		CategoryProvided: parseResult.CategoryProvided,
	}, nil
}

// GetRecent retrieves recent expenses for a user
func (s *ExpenseService) GetRecent(ctx context.Context, userID types.ID, limit int) ([]model.Expense, string, error) {
	mapping, err := s.mappingService.GetByUserID(ctx, userID)
	if err != nil {
		return nil, "", fmt.Errorf("getting user mapping: %w", err)
	}
	if mapping == nil {
		return nil, "", fmt.Errorf("no spreadsheet configured for user %d", userID)
	}

	spreadsheetID := mapping.SpreadsheetDocID()

	activePage, err := s.repo.GetActivePage(ctx, spreadsheetID)
	if err != nil {
		return nil, "", fmt.Errorf("getting active page: %w", err)
	}

	lastID, err := s.repo.GetNextID(ctx, spreadsheetID, activePage)
	if err != nil {
		return nil, "", fmt.Errorf("getting last ID: %w", err)
	}

	expenses, err := s.repo.GetRecent(ctx, spreadsheetID, activePage, lastID-1, limit)
	if err != nil {
		return nil, "", fmt.Errorf("getting recent expenses: %w", err)
	}

	sheetURL := s.buildSheetURL(ctx, spreadsheetID, activePage, mapping.SpreadSheetsURL)
	return expenses, sheetURL, nil
}

// GetByID retrieves an expense by ID
func (s *ExpenseService) GetByID(ctx context.Context, userID types.ID, expenseID types.ID) (*model.Expense, error) {
	mapping, err := s.mappingService.GetByUserID(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("getting user mapping: %w", err)
	}
	if mapping == nil {
		return nil, fmt.Errorf("no spreadsheet configured for user %d", userID)
	}

	spreadsheetID := mapping.SpreadsheetDocID()

	activePage, err := s.repo.GetActivePage(ctx, spreadsheetID)
	if err != nil {
		return nil, fmt.Errorf("getting active page: %w", err)
	}

	return s.repo.GetByID(ctx, spreadsheetID, activePage, expenseID)
}

// Update updates an expense
func (s *ExpenseService) Update(ctx context.Context, userID types.ID, expense *model.Expense, updatedBy string) error {
	mapping, err := s.mappingService.GetByUserID(ctx, userID)
	if err != nil {
		return fmt.Errorf("getting user mapping: %w", err)
	}
	if mapping == nil {
		return fmt.Errorf("no spreadsheet configured for user %d", userID)
	}

	spreadsheetID := mapping.SpreadsheetDocID()

	activePage, err := s.repo.GetActivePage(ctx, spreadsheetID)
	if err != nil {
		return fmt.Errorf("getting active page: %w", err)
	}

	if err := s.repo.Update(ctx, spreadsheetID, activePage, expense); err != nil {
		return fmt.Errorf("updating expense: %w", err)
	}

	log.WithExpense(log.Logger.WithFields(logrus.Fields{
		"user_id":    userID,
		"updated_by": updatedBy,
		"action":     "expenses_update",
	}), int64(expense.ID), expense.Name).
		WithField("amount", expense.Amount).
		Info("expense updated")

	return nil
}

// Delete soft-deletes an expense
func (s *ExpenseService) Delete(ctx context.Context, userID types.ID, expenseID types.ID, deletedBy string) error {
	mapping, err := s.mappingService.GetByUserID(ctx, userID)
	if err != nil {
		return fmt.Errorf("getting user mapping: %w", err)
	}
	if mapping == nil {
		return fmt.Errorf("no spreadsheet configured for user %d", userID)
	}

	spreadsheetID := mapping.SpreadsheetDocID()

	activePage, err := s.repo.GetActivePage(ctx, spreadsheetID)
	if err != nil {
		return fmt.Errorf("getting active page: %w", err)
	}

	// Get expense first for logging
	expense, err := s.repo.GetByID(ctx, spreadsheetID, activePage, expenseID)
	if err != nil {
		return fmt.Errorf("getting expense: %w", err)
	}

	note := fmt.Sprintf("deleted by %s", deletedBy)
	if err := s.repo.SoftDelete(ctx, spreadsheetID, activePage, expenseID, note); err != nil {
		return fmt.Errorf("deleting expense: %w", err)
	}

	if expense != nil {
		log.WithExpense(log.Logger.WithFields(logrus.Fields{
			"user_id":    userID,
			"deleted_by": deletedBy,
			"action":     "expenses_delete",
		}), int64(expense.ID), expense.Name).
			Info("expense soft deleted")
	}

	return nil
}

// UpdateCategory updates only the category of an expense
func (s *ExpenseService) UpdateCategory(ctx context.Context, userID types.ID, expenseID types.ID, category types.Category) error {
	mapping, err := s.mappingService.GetByUserID(ctx, userID)
	if err != nil {
		return fmt.Errorf("getting user mapping: %w", err)
	}
	if mapping == nil {
		return fmt.Errorf("no spreadsheet configured for user %d", userID)
	}

	spreadsheetID := mapping.SpreadsheetDocID()

	activePage, err := s.repo.GetActivePage(ctx, spreadsheetID)
	if err != nil {
		return fmt.Errorf("getting active page: %w", err)
	}

	if err := s.repo.UpdateCategory(ctx, spreadsheetID, activePage, expenseID, category); err != nil {
		return fmt.Errorf("updating category: %w", err)
	}

	log.Info("expense category updated", logrus.Fields{
		"user_id":    userID,
		"expense_id": expenseID,
		"category":   category,
		"action":     "expenses_setcat",
	})

	return nil
}

// UpdateGroup updates only the group of an expense
func (s *ExpenseService) UpdateGroup(ctx context.Context, userID types.ID, expenseID types.ID, group types.Group) error {
	mapping, err := s.mappingService.GetByUserID(ctx, userID)
	if err != nil {
		return fmt.Errorf("getting user mapping: %w", err)
	}
	if mapping == nil {
		return fmt.Errorf("no spreadsheet configured for user %d", userID)
	}

	spreadsheetID := mapping.SpreadsheetDocID()

	activePage, err := s.repo.GetActivePage(ctx, spreadsheetID)
	if err != nil {
		return fmt.Errorf("getting active page: %w", err)
	}

	if err := s.repo.UpdateGroup(ctx, spreadsheetID, activePage, expenseID, group); err != nil {
		return fmt.Errorf("updating group: %w", err)
	}

	log.Info("expense group updated", logrus.Fields{
		"user_id":    userID,
		"expense_id": expenseID,
		"group":      group,
		"action":     "expenses_setgrp",
	})

	return nil
}

// AddFromModel creates a new expense from an Expense model directly
func (s *ExpenseService) AddFromModel(ctx context.Context, userID types.ID, expense *model.Expense) (*model.Expense, error) {
	// Get user's spreadsheet mapping
	mapping, err := s.mappingService.GetByUserID(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("getting user mapping: %w", err)
	}
	if mapping == nil {
		return nil, fmt.Errorf("no spreadsheet configured for user %d", userID)
	}

	spreadsheetID := mapping.SpreadsheetDocID()

	// Get active page
	activePage, err := s.repo.GetActivePage(ctx, spreadsheetID)
	if err != nil {
		return nil, fmt.Errorf("getting active page: %w", err)
	}

	// Get next ID
	nextID, err := s.repo.GetNextID(ctx, spreadsheetID, activePage)
	if err != nil {
		return nil, fmt.Errorf("getting next ID: %w", err)
	}

	expense.ID = nextID

	// Save expense
	if err := s.repo.Create(ctx, spreadsheetID, activePage, expense); err != nil {
		return nil, fmt.Errorf("saving expense: %w", err)
	}

	// Update next ID
	if err := s.repo.UpdateNextID(ctx, spreadsheetID, activePage, nextID+1); err != nil {
		return nil, fmt.Errorf("updating next ID: %w", err)
	}

	log.WithExpense(log.Logger.WithFields(logrus.Fields{
		"user_id":        userID,
		"spreadsheet_id": spreadsheetID,
		"action":         "expenses_add_voice",
	}), int64(expense.ID), expense.Name).
		WithField("amount", expense.Amount).
		Info("expense added from voice successfully")

	return expense, nil
}

// GetSpreadsheetURL returns the spreadsheet URL pointing to the active sheet for a user
func (s *ExpenseService) GetSpreadsheetURL(ctx context.Context, userID types.ID) (string, error) {
	mapping, err := s.mappingService.GetByUserID(ctx, userID)
	if err != nil {
		return "", fmt.Errorf("getting user mapping: %w", err)
	}
	if mapping == nil {
		return "", fmt.Errorf("no spreadsheet configured for user %d", userID)
	}

	spreadsheetID := mapping.SpreadsheetDocID()
	activePage, err := s.repo.GetActivePage(ctx, spreadsheetID)
	if err != nil {
		return mapping.SpreadSheetsURL, nil
	}

	return s.buildSheetURL(ctx, spreadsheetID, activePage, mapping.SpreadSheetsURL), nil
}

// GetReport retrieves expense reports for a user
func (s *ExpenseService) GetReport(ctx context.Context, userID types.ID) (groupReport, categoryReport [][]interface{}, spreadsheetURL string, err error) {
	mapping, err := s.mappingService.GetByUserID(ctx, userID)
	if err != nil {
		return nil, nil, "", fmt.Errorf("getting user mapping: %w", err)
	}
	if mapping == nil {
		return nil, nil, "", fmt.Errorf("no spreadsheet configured for user %d", userID)
	}

	spreadsheetID := mapping.SpreadsheetDocID()

	activePage, err := s.repo.GetActivePage(ctx, spreadsheetID)
	if err != nil {
		return nil, nil, "", fmt.Errorf("getting active page: %w", err)
	}

	groupReport, err = s.repo.GetGroupReport(ctx, spreadsheetID, activePage)
	if err != nil {
		return nil, nil, "", fmt.Errorf("getting group report: %w", err)
	}

	categoryReport, err = s.repo.GetCategoryReport(ctx, spreadsheetID, activePage)
	if err != nil {
		return nil, nil, "", fmt.Errorf("getting category report: %w", err)
	}

	sheetURL := s.buildSheetURL(ctx, spreadsheetID, activePage, mapping.SpreadSheetsURL)
	return groupReport, categoryReport, sheetURL, nil
}

// GetBudgetOverview returns all group and category budget entries, summary rows, and the sheet URL.
func (s *ExpenseService) GetBudgetOverview(ctx context.Context, userID types.ID) (groups, categories, summary []model.BudgetEntry, spreadsheetURL string, err error) {
	mapping, err := s.mappingService.GetByUserID(ctx, userID)
	if err != nil {
		return nil, nil, nil, "", fmt.Errorf("getting user mapping: %w", err)
	}
	if mapping == nil {
		return nil, nil, nil, "", fmt.Errorf("no spreadsheet configured for user %d", userID)
	}

	spreadsheetID := mapping.SpreadsheetDocID()
	activePage, err := s.repo.GetActivePage(ctx, spreadsheetID)
	if err != nil {
		return nil, nil, nil, "", fmt.Errorf("getting active page: %w", err)
	}

	groups, err = s.repo.GetGroupReportWithBudget(ctx, spreadsheetID, activePage)
	if err != nil {
		return nil, nil, nil, "", fmt.Errorf("getting group budgets: %w", err)
	}

	categories, err = s.repo.GetCategoryReportWithBudget(ctx, spreadsheetID, activePage)
	if err != nil {
		return nil, nil, nil, "", fmt.Errorf("getting category budgets: %w", err)
	}

	summary, err = s.repo.GetExpenseSummary(ctx, spreadsheetID, activePage)
	if err != nil {
		log.Warn("failed to read expense summary", logrus.Fields{"error": err.Error()})
		summary = nil
	}

	sheetURL := s.buildSheetURL(ctx, spreadsheetID, activePage, mapping.SpreadSheetsURL)
	return groups, categories, summary, sheetURL, nil
}

// GetBudgetForExpense returns budget entries for a specific group and category, plus the "Total Expenses" summary entry.
func (s *ExpenseService) GetBudgetForExpense(ctx context.Context, userID types.ID, group types.Group, category types.Category) (groupBudget, categoryBudget, totalExpenses *model.BudgetEntry, err error) {
	mapping, err := s.mappingService.GetByUserID(ctx, userID)
	if err != nil {
		return nil, nil, nil, fmt.Errorf("getting user mapping: %w", err)
	}
	if mapping == nil {
		return nil, nil, nil, nil
	}

	spreadsheetID := mapping.SpreadsheetDocID()
	activePage, err := s.repo.GetActivePage(ctx, spreadsheetID)
	if err != nil {
		return nil, nil, nil, fmt.Errorf("getting active page: %w", err)
	}

	groups, err := s.repo.GetGroupReportWithBudget(ctx, spreadsheetID, activePage)
	if err != nil {
		return nil, nil, nil, fmt.Errorf("getting group budgets: %w", err)
	}

	for i := range groups {
		if types.Group(groups[i].Name) == group && groups[i].HasBudget {
			groupBudget = &groups[i]
			break
		}
	}

	if group.NeedsCategory() {
		categories, err := s.repo.GetCategoryReportWithBudget(ctx, spreadsheetID, activePage)
		if err != nil {
			return groupBudget, nil, nil, fmt.Errorf("getting category budgets: %w", err)
		}

		for i := range categories {
			if types.Category(categories[i].Name) == category && categories[i].HasBudget {
				categoryBudget = &categories[i]
				break
			}
		}
	}

	summary, err := s.repo.GetExpenseSummary(ctx, spreadsheetID, activePage)
	if err == nil {
		for i := range summary {
			if summary[i].Name == "Total Expenses" && summary[i].HasBudget {
				totalExpenses = &summary[i]
				break
			}
		}
	}

	return groupBudget, categoryBudget, totalExpenses, nil
}

// SetBudget sets a budget value for a group or category.
func (s *ExpenseService) SetBudget(ctx context.Context, userID types.ID, col string, row int, amount uint64) error {
	mapping, err := s.mappingService.GetByUserID(ctx, userID)
	if err != nil {
		return fmt.Errorf("getting user mapping: %w", err)
	}
	if mapping == nil {
		return fmt.Errorf("no spreadsheet configured for user %d", userID)
	}

	spreadsheetID := mapping.SpreadsheetDocID()
	activePage, err := s.repo.GetActivePage(ctx, spreadsheetID)
	if err != nil {
		return fmt.Errorf("getting active page: %w", err)
	}

	return s.repo.SetBudget(ctx, spreadsheetID, activePage, col, row, amount)
}

// ClearBudget clears a budget value.
func (s *ExpenseService) ClearBudget(ctx context.Context, userID types.ID, col string, row int) error {
	mapping, err := s.mappingService.GetByUserID(ctx, userID)
	if err != nil {
		return fmt.Errorf("getting user mapping: %w", err)
	}
	if mapping == nil {
		return fmt.Errorf("no spreadsheet configured for user %d", userID)
	}

	spreadsheetID := mapping.SpreadsheetDocID()
	activePage, err := s.repo.GetActivePage(ctx, spreadsheetID)
	if err != nil {
		return fmt.Errorf("getting active page: %w", err)
	}

	return s.repo.ClearBudget(ctx, spreadsheetID, activePage, col, row)
}

// GetInsights aggregates group/category spending across multiple monthly sheets
// and returns per-month averages.
func (s *ExpenseService) GetInsights(ctx context.Context, userID types.ID, months int, ytd bool, efMultiplier int, excludeCurrent bool) (*model.InsightsResult, error) {
	mapping, err := s.mappingService.GetByUserID(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("getting user mapping: %w", err)
	}
	if mapping == nil {
		return nil, fmt.Errorf("no spreadsheet configured for user %d", userID)
	}

	spreadsheetID := mapping.SpreadsheetDocID()

	activePage, err := s.repo.GetActivePage(ctx, spreadsheetID)
	if err != nil {
		return nil, fmt.Errorf("getting active page: %w", err)
	}

	allNames, err := s.mappingService.GetValidSheetNames(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("getting valid sheet names: %w", err)
	}

	var target, found, missing []string
	var periodLabel string
	if ytd {
		target, found, missing = types.YTDSheetNames(allNames, activePage)
		periodLabel = "YTD"
	} else {
		target, found, missing = types.RecentSheetNames(allNames, months, activePage)
		periodLabel = fmt.Sprintf("%d months", months)
	}

	if excludeCurrent {
		var filtered []string
		for _, name := range found {
			if name != activePage {
				filtered = append(filtered, name)
			}
		}
		found = filtered
	}

	if len(found) == 0 {
		return nil, fmt.Errorf("no monthly sheets found for the requested period (target: %v)", target)
	}

	sorted := types.SortSheetNames(found)

	reports, err := s.repo.GetMultiMonthReports(ctx, spreadsheetID, sorted)
	if err != nil {
		return nil, fmt.Errorf("getting multi-month reports: %w", err)
	}

	monthCount := len(sorted)

	groupTotals := make(map[string]int64)
	categoryTotals := make(map[string]int64)
	summaryTotals := make(map[string]int64)

	var groupOrder []string
	var categoryOrder []string
	var summaryOrder []string
	groupSeen := make(map[string]bool)
	categorySeen := make(map[string]bool)
	summarySeen := make(map[string]bool)

	for _, name := range sorted {
		report := reports[name]
		for _, g := range report.Groups {
			groupTotals[g.Name] += g.Spent
			if !groupSeen[g.Name] {
				groupSeen[g.Name] = true
				groupOrder = append(groupOrder, g.Name)
			}
		}
		for _, c := range report.Categories {
			categoryTotals[c.Name] += c.Spent
			if !categorySeen[c.Name] {
				categorySeen[c.Name] = true
				categoryOrder = append(categoryOrder, c.Name)
			}
		}
		for _, s := range report.Summary {
			summaryTotals[s.Name] += s.Spent
			if !summarySeen[s.Name] {
				summarySeen[s.Name] = true
				summaryOrder = append(summaryOrder, s.Name)
			}
		}
	}

	buildAvgs := func(order []string, totals map[string]int64) []model.AverageEntry {
		var entries []model.AverageEntry
		for _, name := range order {
			total := totals[name]
			avg := total / int64(monthCount)
			if avg != 0 {
				entries = append(entries, model.AverageEntry{
					Name:       name,
					Total:      total,
					Average:    avg,
					MonthCount: monthCount,
				})
			}
		}
		return entries
	}

	groupAvgs := buildAvgs(groupOrder, groupTotals)
	categoryAvgs := buildAvgs(categoryOrder, categoryTotals)
	summaryAvgs := buildAvgs(summaryOrder, summaryTotals)

	var mustHaveAvg int64
	for _, g := range groupAvgs {
		if g.Name == string(types.GroupMustHave) {
			mustHaveAvg = g.Average
			break
		}
	}

	if efMultiplier <= 0 {
		efMultiplier = 3
	}

	var excludedName string
	if excludeCurrent {
		excludedName = activePage
	}

	result := &model.InsightsResult{
		Period:                  periodLabel,
		MonthsFound:            sorted,
		MonthsMissing:          missing,
		ExcludedCurrent:        excludedName,
		GroupAvgs:               groupAvgs,
		CategoryAvgs:            categoryAvgs,
		SummaryAvgs:             summaryAvgs,
		EmergencyFundMultiplier: efMultiplier,
		EmergencyFund:           mustHaveAvg * int64(efMultiplier),
	}

	return result, nil
}

// GetMonthlyReports returns per-month group/category/summary data for the
// requested number of recent months. Used by the AI ask feature to provide
// per-month breakdowns (not just averages).
func (s *ExpenseService) GetMonthlyReports(ctx context.Context, userID types.ID, months int) (map[string]sheets.MonthReport, []string, error) {
	mapping, err := s.mappingService.GetByUserID(ctx, userID)
	if err != nil {
		return nil, nil, fmt.Errorf("getting user mapping: %w", err)
	}
	if mapping == nil {
		return nil, nil, fmt.Errorf("no spreadsheet configured for user %d", userID)
	}

	spreadsheetID := mapping.SpreadsheetDocID()

	activePage, err := s.repo.GetActivePage(ctx, spreadsheetID)
	if err != nil {
		return nil, nil, fmt.Errorf("getting active page: %w", err)
	}

	allNames, err := s.mappingService.GetValidSheetNames(ctx, userID)
	if err != nil {
		return nil, nil, fmt.Errorf("getting valid sheet names: %w", err)
	}

	_, found, _ := types.RecentSheetNames(allNames, months, activePage)
	if len(found) == 0 {
		return nil, nil, fmt.Errorf("no monthly sheets found")
	}

	sorted := types.SortSheetNames(found)

	reports, err := s.repo.GetMultiMonthReports(ctx, spreadsheetID, sorted)
	if err != nil {
		return nil, nil, fmt.Errorf("getting multi-month reports: %w", err)
	}

	return reports, sorted, nil
}

// buildSheetURL builds a URL that navigates directly to the active sheet tab.
// Uses the canonical /edit#gid= format so the fragment survives Google Sheets redirects.
// Falls back to the base URL if the sheet ID lookup fails.
func (s *ExpenseService) buildSheetURL(ctx context.Context, spreadsheetID, sheetName, baseURL string) string {
	gid, err := s.repo.GetSheetID(ctx, spreadsheetID, sheetName)
	if err != nil {
		return baseURL
	}
	return fmt.Sprintf("https://docs.google.com/spreadsheets/d/%s/edit#gid=%d", spreadsheetID, gid)
}
