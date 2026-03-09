package sheets

import (
	"context"
	"fmt"
	"strings"
	"time"

	"expensemate-tgbot/internal/log"
	"expensemate-tgbot/internal/model"
	"expensemate-tgbot/internal/types"

	"github.com/sirupsen/logrus"
	"github.com/spf13/cast"
	"google.golang.org/api/sheets/v4"
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

// GetActivePage reads the active page name from Database!B2
func (r *MappingRepository) GetActivePage(ctx context.Context, spreadsheetID string) (string, error) {
	cell := types.BuildCell(types.DatabaseSheetName, types.DatabaseActivePageCell)
	return r.client.GetValue(ctx, spreadsheetID, cell)
}

// CreateNewMonthSheet duplicates the current sheet, clears expense data, updates formulas, and sets the active page
func (r *MappingRepository) CreateNewMonthSheet(ctx context.Context, spreadsheetID, currentSheet, newSheet string) error {
	// The formulas in currentSheet reference (currentSheet - 1).
	// In the new sheet, they should reference currentSheet instead.
	prevOfCurrent, err := types.PrevMonthName(currentSheet)
	if err != nil {
		return fmt.Errorf("calculating previous month: %w", err)
	}

	// 1. Get sheet ID of currentSheet
	sheetID, err := r.client.GetSheetID(ctx, spreadsheetID, currentSheet)
	if err != nil {
		return fmt.Errorf("getting sheet ID for %s: %w", currentSheet, err)
	}

	log.Info("creating new month sheet", logrus.Fields{
		"current": currentSheet,
		"new":     newSheet,
	})

	// 2. Duplicate sheet
	dupReq := &sheets.Request{
		DuplicateSheet: &sheets.DuplicateSheetRequest{
			SourceSheetId: sheetID,
			NewSheetName:  newSheet,
		},
	}
	if _, err := r.client.BatchUpdate(ctx, spreadsheetID, []*sheets.Request{dupReq}); err != nil {
		return fmt.Errorf("duplicating sheet: %w", err)
	}

	// 3. Update A1 with new sheet name
	a1Cell := types.BuildCell(newSheet, types.SheetNameCell)
	if err := r.client.Update(ctx, spreadsheetID, a1Cell, [][]interface{}{{newSheet}}); err != nil {
		return fmt.Errorf("updating A1: %w", err)
	}

	// 4. Read current asset values P2:Q9 (unformatted to get raw numbers without currency symbols)
	assetRange := types.BuildRange(newSheet, types.AssetCurrentRange)
	assetResp, err := r.client.GetUnformatted(ctx, spreadsheetID, assetRange)
	if err != nil {
		return fmt.Errorf("reading asset data: %w", err)
	}

	// 5. Write asset values to P17:Q24 (last month's snapshot)
	lastMonthRange := types.BuildRange(newSheet, types.AssetLastMonthRange)
	if assetResp != nil && len(assetResp.Values) > 0 {
		if err := r.client.Update(ctx, spreadsheetID, lastMonthRange, assetResp.Values); err != nil {
			return fmt.Errorf("writing last month assets: %w", err)
		}
	}

	// 6. Clear C4 (monthly salary)
	salaryCell := types.BuildCell(newSheet, types.SalaryCellRef)
	if err := r.client.ClearValues(ctx, spreadsheetID, salaryCell); err != nil {
		return fmt.Errorf("clearing salary cell: %w", err)
	}

	// 7. Clear expense data rows 10+ (keep formatting)
	clearRange := types.BuildRangeFromCells(newSheet, types.ExpensesLeftCol, types.ExpenseDataStartRow, types.ExpensesRightCol, 1000)
	if err := r.client.ClearValues(ctx, spreadsheetID, clearRange); err != nil {
		return fmt.Errorf("clearing expense data: %w", err)
	}

	// 8. Reset B2 to initial next expense ID
	nextIDCell := types.BuildCell(newSheet, types.ExpensesNextIDCell)
	if err := r.client.Update(ctx, spreadsheetID, nextIDCell, [][]interface{}{{types.NewMonthNextExpenseID}}); err != nil {
		return fmt.Errorf("resetting next expense ID: %w", err)
	}

	// 9. Read formulas from current sheet C5:C9
	formulaRange := types.BuildRange(currentSheet, types.InvestmentFormulaRange)
	formulaResp, err := r.client.GetFormulas(ctx, spreadsheetID, formulaRange)
	if err != nil {
		return fmt.Errorf("reading investment formulas: %w", err)
	}

	// 10. Replace month reference in formulas: e.g. '2026_01' -> '2026_02'
	if formulaResp != nil && len(formulaResp.Values) > 0 {
		oldRef := fmt.Sprintf("'%s'", prevOfCurrent)
		newRef := fmt.Sprintf("'%s'", currentSheet)
		for i, row := range formulaResp.Values {
			for j, cell := range row {
				if s, ok := cell.(string); ok {
					formulaResp.Values[i][j] = strings.ReplaceAll(s, oldRef, newRef)
				}
			}
		}

		// 11. Write updated formulas to new sheet
		newFormulaRange := types.BuildRange(newSheet, types.InvestmentFormulaRange)
		if err := r.client.UpdateUserEntered(ctx, spreadsheetID, newFormulaRange, formulaResp.Values); err != nil {
			return fmt.Errorf("updating investment formulas: %w", err)
		}
	}

	// 12. Clear investment notes
	investmentNoteRange := types.BuildRange(newSheet, types.InvestmentNoteRange)
	if err := r.client.ClearValues(ctx, spreadsheetID, investmentNoteRange); err != nil {
		return fmt.Errorf("clearing investment notes: %w", err)
	}

	// 13. Update Database!B2 to the new sheet name
	if err := r.UpdateActivePage(ctx, spreadsheetID, newSheet); err != nil {
		return fmt.Errorf("updating active page: %w", err)
	}

	log.Info("new month sheet created", logrus.Fields{
		"sheet": newSheet,
	})

	return nil
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
