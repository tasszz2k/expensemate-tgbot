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

// Add creates a new expense for a user
func (s *ExpenseService) Add(ctx context.Context, userID types.ID, text string) (*model.Expense, error) {
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
	expense, err := model.ParseTextToExpense(text)
	if err != nil {
		return nil, err
	}

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

	return &expense, nil
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

	return expenses, mapping.SpreadSheetsURL, nil
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

	return groupReport, categoryReport, mapping.SpreadSheetsURL, nil
}
