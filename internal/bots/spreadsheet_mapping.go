package bots

import (
	"context"
	"errors"
	"log/slog"

	"expensemate-tgbot/pkg/clients/gsheetclients"
	"expensemate-tgbot/pkg/configs"
	"expensemate-tgbot/pkg/models"
	"expensemate-tgbot/pkg/types/gsheettypes"
	"expensemate-tgbot/pkg/types/types"
	"expensemate-tgbot/pkg/utils/httputils"
	"expensemate-tgbot/pkg/utils/timeutils"

	"github.com/spf13/cast"
	"google.golang.org/api/sheets/v4"
)

func (e *Expensemate) loadSpreadsheetMappings(ctx context.Context) {
	spreadsheetMappings := make(map[types.Id]models.UserSheetMapping)
	databaseSpreadsheetId := configs.Get().GoogleSheets.DatabaseSpreadsheetId

	nextId, err := e.getUserSheetMappingsNextId(ctx)
	if err != nil {
		slog.Error("Failed to load spreadsheet mappings, next id")
		return
	}
	if nextId > 1 { // load only if there are mappings
		lastRow := gsheettypes.UserSheetMappingTopRow + (nextId - 1) // load to the last id = next id - 1

		// build the range of cells to load
		readRange := gsheettypes.BuildRangeFromCells(
			gsheettypes.UserSheetMappingSheetName,
			gsheettypes.UserSheetMappingLeftCol,
			gsheettypes.UserSheetMappingTopRow+1, // load from id=1
			gsheettypes.UserSheetMappingRightCol,
			lastRow,
		)
		values, err := gsheetclients.GetInstance().Get(ctx, databaseSpreadsheetId, readRange)
		if err != nil {
			slog.Error("Failed to load spreadsheet mappings, values")
			return
		}

		for _, row := range values.Values {
			createdAt, err := timeutils.ParseDateOnly(cast.ToString(row[5]))
			if err != nil {
				createdAt = timeutils.GetCurrentDay()
			}
			updateAt, err := timeutils.ParseDateOnly(cast.ToString(row[6]))
			if err != nil {
				updateAt = timeutils.GetCurrentDay()
			}
			userSheetMapping := models.UserSheetMapping{
				ID:              types.Id(cast.ToInt(row[0])),
				UserID:          types.Id(cast.ToInt(row[1])),
				Username:        cast.ToString(row[2]),
				FullName:        cast.ToString(row[3]),
				SpreadSheetsURL: cast.ToString(row[4]),
				CreatedAt:       createdAt,
				UpdateAt:        updateAt,
				Status:          models.MappingStatus(cast.ToString(row[7])),
			}
			spreadsheetMappings[userSheetMapping.UserID] = userSheetMapping
		}
	}

	e.smMux.Lock()
	defer e.smMux.Unlock()
	e.spreadsheetMappings = spreadsheetMappings

	slog.Info("Spreadsheet mappings are loaded")

	return
}

func (e *Expensemate) getUserSheetMappingsNextId(ctx context.Context) (int, error) {
	nextIdCell := gsheettypes.BuildCell(
		gsheettypes.UserSheetMappingSheetName,
		gsheettypes.UserSheetMappingNextIdCell,
	)
	databaseSpreadsheetId := configs.Get().GoogleSheets.DatabaseSpreadsheetId
	value, err := gsheetclients.GetInstance().GetValue(
		ctx, databaseSpreadsheetId, nextIdCell,
	)
	if err != nil {
		slog.Error("Failed to load spreadsheet mappings, next id cell")
		return 0, err
	}
	nextId := cast.ToInt(value)
	return nextId, nil
}

func (e *Expensemate) getSpreadsheetMappingByUserID(
	_ context.Context,
	userID int64,
) (models.UserSheetMapping, error) {
	e.smMux.RLock()
	defer e.smMux.RUnlock()
	if mapping, exists := e.spreadsheetMappings[types.Id(userID)]; exists {
		return mapping, nil
	}
	return models.UserSheetMapping{}, errors.New("spreadsheet mapping not found")
}

func (e *Expensemate) upsertSpreadsheetMapping(
	ctx context.Context,
	mapping models.UserSheetMapping,
) error {
	// save the mapping to the database
	// if the mapping is new (without id), insert it
	// if the mapping is existing, update it
	if mapping.ID == 0 {
		// insert
		nextId, err := e.getUserSheetMappingsNextId(ctx)
		if err != nil {
			return err
		}
		mapping.ID = types.Id(nextId)
	}

	// update cache
	e.smMux.Lock()
	defer e.smMux.Unlock()
	e.spreadsheetMappings[mapping.UserID] = mapping

	row := int(mapping.ID + gsheettypes.UserSheetMappingTopRow)
	writeRange := gsheettypes.BuildRangeFromCells(
		gsheettypes.UserSheetMappingSheetName,
		gsheettypes.UserSheetMappingLeftCol, row,
		gsheettypes.UserSheetMappingRightCol, row,
	)

	values := [][]interface{}{
		{
			mapping.ID,
			mapping.UserID,
			mapping.Username,
			mapping.FullName,
			mapping.SpreadSheetsURL,
			timeutils.FormatDateOnly(mapping.CreatedAt),
			timeutils.FormatDateOnly(mapping.UpdateAt),
			mapping.Status,
		},
	}
	vr := &sheets.ValueRange{
		Values: values,
	}
	// update value range
	_, err := gsheetclients.GetInstance().Update(
		ctx,
		configs.Get().GoogleSheets.DatabaseSpreadsheetId,
		writeRange,
		vr,
	)
	if err != nil {
		return err
	}

	// update the next id
	nextId := int(mapping.ID) + 1
	err = e.UpdateUserSheetMappingNextId(ctx, nextId)
	if err != nil {
		return err
	}

	return nil
}

func (e *Expensemate) UpdateUserSheetMappingNextId(
	ctx context.Context,
	nextId int,
) error {
	nextIdCell := gsheettypes.BuildCell(
		gsheettypes.UserSheetMappingSheetName,
		gsheettypes.UserSheetMappingNextIdCell,
	)

	_, err := gsheetclients.GetInstance().Update(
		ctx,
		configs.Get().GoogleSheets.DatabaseSpreadsheetId,
		nextIdCell,
		&sheets.ValueRange{
			Values: [][]interface{}{
				{nextId},
			},
		},
	)
	if err != nil {
		return err
	}
	return nil
}

func (e *Expensemate) checkValidSpreadsheet(url string) error {
	docId := httputils.GetGoogleSheetsDocID(url)
	if docId == "" {
		return errors.New("invalid google sheets url")
	}
	// check if the spreadsheet exists
	e.smMux.RLock()
	defer e.smMux.RUnlock()
	for _, mapping := range e.spreadsheetMappings {
		if httputils.GetGoogleSheetsDocID(mapping.SpreadSheetsURL) == docId {
			return errors.New("spreadsheet already exists")
		}
	}
	return nil
}
