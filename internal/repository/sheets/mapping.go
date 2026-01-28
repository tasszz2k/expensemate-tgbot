package sheets

import (
	"context"
	"fmt"
	"time"

	"expensemate-tgbot/internal/log"
	"expensemate-tgbot/internal/model"
	"expensemate-tgbot/internal/types"

	"github.com/sirupsen/logrus"
	"github.com/spf13/cast"
)

// MappingRepository handles user-spreadsheet mappings in Google Sheets
type MappingRepository struct {
	client                *Client
	databaseSpreadsheetID string
}

// NewMappingRepository creates a new MappingRepository
func NewMappingRepository(client *Client, databaseSpreadsheetID string) *MappingRepository {
	return &MappingRepository{
		client:                client,
		databaseSpreadsheetID: databaseSpreadsheetID,
	}
}

// GetAll retrieves all user-spreadsheet mappings
func (r *MappingRepository) GetAll(ctx context.Context) ([]model.UserSheetMapping, error) {
	// First get the next ID to know how many rows to read
	nextID, err := r.GetNextID(ctx)
	if err != nil {
		return nil, err
	}

	if nextID <= 1 {
		return nil, nil
	}

	startRow := types.UserMappingTopRow + 1
	endRow := types.UserMappingTopRow + int(nextID) - 1
	readRange := types.BuildRangeFromCells(
		types.UserMappingSheetName,
		types.UserMappingLeftCol,
		startRow,
		types.UserMappingRightCol,
		endRow,
	)

	log.Debug("reading all mappings", logrus.Fields{
		"range": readRange,
	})

	resp, err := r.client.Get(ctx, r.databaseSpreadsheetID, readRange)
	if err != nil {
		return nil, fmt.Errorf("reading mappings: %w", err)
	}

	var mappings []model.UserSheetMapping
	for _, row := range resp.Values {
		if len(row) < 8 {
			continue
		}
		mapping := parseRowToMapping(row)
		mappings = append(mappings, mapping)
	}

	return mappings, nil
}

// GetByUserID retrieves a mapping by Telegram user ID
func (r *MappingRepository) GetByUserID(ctx context.Context, userID types.ID) (*model.UserSheetMapping, error) {
	mappings, err := r.GetAll(ctx)
	if err != nil {
		return nil, err
	}

	for _, m := range mappings {
		if m.UserID == userID {
			return &m, nil
		}
	}

	return nil, nil
}

// GetNextID returns the next mapping ID
func (r *MappingRepository) GetNextID(ctx context.Context) (types.ID, error) {
	cell := types.BuildCell(types.UserMappingSheetName, types.UserMappingNextIDCell)
	value, err := r.client.GetValue(ctx, r.databaseSpreadsheetID, cell)
	if err != nil {
		return 0, fmt.Errorf("getting next mapping ID: %w", err)
	}
	return types.ID(cast.ToInt64(value)), nil
}

// UpdateNextID updates the next mapping ID counter
func (r *MappingRepository) UpdateNextID(ctx context.Context, nextID types.ID) error {
	cell := types.BuildCell(types.UserMappingSheetName, types.UserMappingNextIDCell)
	return r.client.Update(ctx, r.databaseSpreadsheetID, cell, [][]interface{}{{int64(nextID)}})
}

// Create creates a new user-spreadsheet mapping
func (r *MappingRepository) Create(ctx context.Context, mapping *model.UserSheetMapping) error {
	row := types.UserMappingTopRow + int(mapping.ID)
	writeRange := types.BuildRangeFromCells(
		types.UserMappingSheetName,
		types.UserMappingLeftCol,
		row,
		types.UserMappingRightCol,
		row,
	)

	log.Debug("creating mapping", logrus.Fields{
		"range":   writeRange,
		"user_id": mapping.UserID,
	})

	return r.client.Update(ctx, r.databaseSpreadsheetID, writeRange, [][]interface{}{mapping.ToRow()})
}

// Update updates an existing mapping
func (r *MappingRepository) Update(ctx context.Context, mapping *model.UserSheetMapping) error {
	mapping.UpdatedAt = time.Now()
	return r.Create(ctx, mapping)
}

// Upsert creates or updates a mapping
func (r *MappingRepository) Upsert(ctx context.Context, mapping *model.UserSheetMapping) error {
	existing, err := r.GetByUserID(ctx, mapping.UserID)
	if err != nil {
		return err
	}

	if existing != nil {
		mapping.ID = existing.ID
		mapping.CreatedAt = existing.CreatedAt
		return r.Update(ctx, mapping)
	}

	// New mapping
	nextID, err := r.GetNextID(ctx)
	if err != nil {
		return err
	}

	mapping.ID = nextID
	mapping.CreatedAt = time.Now()
	mapping.UpdatedAt = time.Now()

	if err := r.Create(ctx, mapping); err != nil {
		return err
	}

	return r.UpdateNextID(ctx, nextID+1)
}

// parseRowToMapping converts a spreadsheet row to UserSheetMapping
func parseRowToMapping(row []interface{}) model.UserSheetMapping {
	// Pad row to ensure 8 columns
	for len(row) < 8 {
		row = append(row, "")
	}

	createdAt, _ := time.Parse("2/1/2006", cast.ToString(row[5]))
	updatedAt, _ := time.Parse("2/1/2006", cast.ToString(row[6]))

	return model.UserSheetMapping{
		ID:              types.ID(cast.ToInt64(row[0])),
		UserID:          types.ID(cast.ToInt64(row[1])),
		Username:        cast.ToString(row[2]),
		FullName:        cast.ToString(row[3]),
		SpreadSheetsURL: cast.ToString(row[4]),
		CreatedAt:       createdAt,
		UpdatedAt:       updatedAt,
		Status:          model.MappingStatus(cast.ToString(row[7])),
	}
}

// UpdateActivePage updates the active page in a user's spreadsheet
func (r *MappingRepository) UpdateActivePage(ctx context.Context, spreadsheetID, pageName string) error {
	cell := types.BuildCell(types.DatabaseSheetName, types.DatabaseActivePageCell)
	return r.client.Update(ctx, spreadsheetID, cell, [][]interface{}{{pageName}})
}

// GetValidSheetNames returns sheet names matching YYYY_MM format
func (r *MappingRepository) GetValidSheetNames(ctx context.Context, spreadsheetID string) ([]string, error) {
	sheets, err := r.client.GetSheets(ctx, spreadsheetID)
	if err != nil {
		return nil, err
	}

	var validNames []string
	for _, sheet := range sheets {
		if types.IsValidSheetName(sheet.Properties.Title) {
			validNames = append(validNames, sheet.Properties.Title)
		}
	}

	return validNames, nil
}
