package service

import (
	"context"
	"fmt"
	"slices"
	"sync"
	"time"

	"expensemate-tgbot/internal/log"
	"expensemate-tgbot/internal/model"
	"expensemate-tgbot/internal/repository/sheets"
	"expensemate-tgbot/internal/types"
	httputil "expensemate-tgbot/internal/util/http"

	"github.com/sirupsen/logrus"
)

// MappingService handles user-spreadsheet mapping business logic
type MappingService struct {
	repo  *sheets.MappingRepository
	cache map[types.ID]*model.UserSheetMapping
	mu    sync.RWMutex
}

// NewMappingService creates a new MappingService
func NewMappingService(repo *sheets.MappingRepository) *MappingService {
	return &MappingService{
		repo:  repo,
		cache: make(map[types.ID]*model.UserSheetMapping),
	}
}

// LoadCache loads all mappings into the cache
func (s *MappingService) LoadCache(ctx context.Context) error {
	mappings, err := s.repo.GetAll(ctx)
	if err != nil {
		return fmt.Errorf("loading mappings: %w", err)
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	for i := range mappings {
		s.cache[mappings[i].UserID] = &mappings[i]
	}

	log.Info("mappings cache loaded", logrus.Fields{
		"count": len(mappings),
	})

	return nil
}

// GetByUserID retrieves a mapping by user ID (from cache first)
func (s *MappingService) GetByUserID(ctx context.Context, userID types.ID) (*model.UserSheetMapping, error) {
	s.mu.RLock()
	if mapping, exists := s.cache[userID]; exists {
		s.mu.RUnlock()
		return mapping, nil
	}
	s.mu.RUnlock()

	// Not in cache, fetch from repository
	mapping, err := s.repo.GetByUserID(ctx, userID)
	if err != nil {
		return nil, err
	}

	if mapping != nil {
		s.mu.Lock()
		s.cache[userID] = mapping
		s.mu.Unlock()
	}

	return mapping, nil
}

// Configure configures a user's spreadsheet mapping
func (s *MappingService) Configure(ctx context.Context, userID types.ID, username, fullName, spreadsheetURL string) (*model.UserSheetMapping, error) {
	// Validate URL
	if !httputil.IsValidGoogleSheetsURL(spreadsheetURL) {
		return nil, fmt.Errorf("invalid Google Sheets URL")
	}

	mapping := &model.UserSheetMapping{
		UserID:          userID,
		Username:        username,
		FullName:        fullName,
		SpreadSheetsURL: spreadsheetURL,
		Status:          model.MappingStatusMapped,
		CreatedAt:       time.Now(),
		UpdatedAt:       time.Now(),
	}

	if err := s.repo.Upsert(ctx, mapping); err != nil {
		return nil, fmt.Errorf("saving mapping: %w", err)
	}

	// Update cache
	s.mu.Lock()
	s.cache[userID] = mapping
	s.mu.Unlock()

	log.Info("spreadsheet configured", logrus.Fields{
		"user_id":  userID,
		"username": username,
		"action":   "gsheets_configure",
	})

	return mapping, nil
}

// UpdateActivePage updates the active page for a user's spreadsheet
func (s *MappingService) UpdateActivePage(ctx context.Context, userID types.ID, pageName string) error {
	if !types.IsValidSheetName(pageName) {
		return fmt.Errorf("invalid page name format (expected YYYY_MM): %s", pageName)
	}

	mapping, err := s.GetByUserID(ctx, userID)
	if err != nil {
		return err
	}
	if mapping == nil {
		return fmt.Errorf("no spreadsheet configured for user %d", userID)
	}

	spreadsheetID := mapping.SpreadsheetDocID()

	if err := s.repo.UpdateActivePage(ctx, spreadsheetID, pageName); err != nil {
		return fmt.Errorf("updating active page: %w", err)
	}

	log.Info("active page updated", logrus.Fields{
		"user_id":   userID,
		"page_name": pageName,
		"action":    "gsheets_update_active_page",
	})

	return nil
}

// GetValidSheetNames returns valid sheet names for a user's spreadsheet
func (s *MappingService) GetValidSheetNames(ctx context.Context, userID types.ID) ([]string, error) {
	mapping, err := s.GetByUserID(ctx, userID)
	if err != nil {
		return nil, err
	}
	if mapping == nil {
		return nil, fmt.Errorf("no spreadsheet configured for user %d", userID)
	}

	spreadsheetID := mapping.SpreadsheetDocID()
	return s.repo.GetValidSheetNames(ctx, spreadsheetID)
}

// GetActivePage returns the current active page name for a user's spreadsheet
func (s *MappingService) GetActivePage(ctx context.Context, userID types.ID) (string, error) {
	mapping, err := s.GetByUserID(ctx, userID)
	if err != nil {
		return "", err
	}
	if mapping == nil {
		return "", fmt.Errorf("no spreadsheet configured for user %d", userID)
	}
	return s.repo.GetActivePage(ctx, mapping.SpreadsheetDocID())
}

// CreateNewMonthSheet creates a new monthly sheet from the current active page
func (s *MappingService) CreateNewMonthSheet(ctx context.Context, userID types.ID) (string, error) {
	mapping, err := s.GetByUserID(ctx, userID)
	if err != nil {
		return "", err
	}
	if mapping == nil {
		return "", fmt.Errorf("no spreadsheet configured for user %d", userID)
	}

	spreadsheetID := mapping.SpreadsheetDocID()

	currentPage, err := s.repo.GetActivePage(ctx, spreadsheetID)
	if err != nil {
		return "", fmt.Errorf("reading active page: %w", err)
	}
	if !types.IsValidSheetName(currentPage) {
		return "", fmt.Errorf("current active page %q is not a valid YYYY_MM format", currentPage)
	}

	newSheet, err := types.NextMonthName(currentPage)
	if err != nil {
		return "", fmt.Errorf("calculating next month: %w", err)
	}

	existingSheets, err := s.repo.GetValidSheetNames(ctx, spreadsheetID)
	if err != nil {
		return "", fmt.Errorf("checking existing sheets: %w", err)
	}
	if slices.Contains(existingSheets, newSheet) {
		return "", fmt.Errorf("sheet %q already exists", newSheet)
	}

	if err := s.repo.CreateNewMonthSheet(ctx, spreadsheetID, currentPage, newSheet); err != nil {
		return "", fmt.Errorf("creating new month sheet: %w", err)
	}

	log.Info("new month sheet created", logrus.Fields{
		"user_id":  userID,
		"from":     currentPage,
		"to":       newSheet,
		"action":   "gsheets_create_new_month",
	})

	return newSheet, nil
}

// GetSpreadsheetURL returns the spreadsheet URL for a user
func (s *MappingService) GetSpreadsheetURL(ctx context.Context, userID types.ID) (string, error) {
	mapping, err := s.GetByUserID(ctx, userID)
	if err != nil {
		return "", err
	}
	if mapping == nil {
		return "", nil
	}
	return mapping.SpreadSheetsURL, nil
}
