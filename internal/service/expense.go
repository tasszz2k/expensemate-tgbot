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

// GetBudgetOverview returns all group and category budget entries with the sheet URL.
func (s *ExpenseService) GetBudgetOverview(ctx context.Context, userID types.ID) (groups []model.BudgetEntry, categories []model.BudgetEntry, spreadsheetURL string, err error) {
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

	groups, err = s.repo.GetGroupReportWithBudget(ctx, spreadsheetID, activePage)
	if err != nil {
		return nil, nil, "", fmt.Errorf("getting group budgets: %w", err)
	}

	categories, err = s.repo.GetCategoryReportWithBudget(ctx, spreadsheetID, activePage)
	if err != nil {
		return nil, nil, "", fmt.Errorf("getting category budgets: %w", err)
	}

	sheetURL := s.buildSheetURL(ctx, spreadsheetID, activePage, mapping.SpreadSheetsURL)
	return groups, categories, sheetURL, nil
}

// GetBudgetForExpense returns budget entries for a specific group and category (used after adding an expense).
func (s *ExpenseService) GetBudgetForExpense(ctx context.Context, userID types.ID, group types.Group, category types.Category) (groupBudget *model.BudgetEntry, categoryBudget *model.BudgetEntry, err error) {
	mapping, err := s.mappingService.GetByUserID(ctx, userID)
	if err != nil {
		return nil, nil, fmt.Errorf("getting user mapping: %w", err)
	}
	if mapping == nil {
		return nil, nil, nil
	}

	spreadsheetID := mapping.SpreadsheetDocID()
	activePage, err := s.repo.GetActivePage(ctx, spreadsheetID)
	if err != nil {
		return nil, nil, fmt.Errorf("getting active page: %w", err)
	}

	groups, err := s.repo.GetGroupReportWithBudget(ctx, spreadsheetID, activePage)
	if err != nil {
		return nil, nil, fmt.Errorf("getting group budgets: %w", err)
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
			return groupBudget, nil, fmt.Errorf("getting category budgets: %w", err)
		}

		for i := range categories {
			if types.Category(categories[i].Name) == category && categories[i].HasBudget {
				categoryBudget = &categories[i]
				break
			}
		}
	}

	return groupBudget, categoryBudget, nil
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
