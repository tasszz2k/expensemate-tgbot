package bots

import (
	"context"
	"log/slog"

	"expensemate-tgbot/pkg/clients/gsheetclients"
	"expensemate-tgbot/pkg/configs"
	"expensemate-tgbot/pkg/models"
	"expensemate-tgbot/pkg/types/gsheettypes"
	"expensemate-tgbot/pkg/types/types"
	"expensemate-tgbot/pkg/utils/timeutils"

	"github.com/spf13/cast"
)

func (e *Expensemate) loadSpreadsheetMappings(ctx context.Context) {
	spreadsheetMappings := make(map[string]models.UserSheetMapping)

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
		return
	}
	lastRow := gsheettypes.UserSheetMappingTopRow + cast.ToInt(value) - 2

	// build the range of cells to load
	readRange := gsheettypes.BuildRange(
		gsheettypes.UserSheetMappingSheetName,
		gsheettypes.UserSheetMappingLeftCol+cast.ToString(gsheettypes.UserSheetMappingTopRow),
		gsheettypes.UserSheetMappingRightCol+cast.ToString(lastRow),
	)
	values, err := gsheetclients.GetInstance().Get(ctx, databaseSpreadsheetId, readRange)
	if err != nil {
		slog.Error("Failed to load spreadsheet mappings, values")
		return
	}

	for _, row := range values.Values {
		createdAt, err := timeutils.ParseDateOnly(cast.ToString(row[3]))
		if err != nil {
			createdAt = timeutils.GetCurrentDay()
		}
		updateAt, err := timeutils.ParseDateOnly(cast.ToString(row[4]))
		if err != nil {
			updateAt = timeutils.GetCurrentDay()
		}
		userSheetMapping := models.UserSheetMapping{
			ID:              types.Id(cast.ToInt(row[0])),
			Username:        cast.ToString(row[1]),
			SpreadSheetsURL: cast.ToString(row[2]),
			CreatedAt:       createdAt,
			UpdateAt:        updateAt,
			Status:          models.MappingStatus(cast.ToString(row[5])),
		}
		spreadsheetMappings[userSheetMapping.Username] = userSheetMapping
	}

	e.smMux.Lock()
	defer e.smMux.Unlock()
	e.spreadsheetMappings = spreadsheetMappings

	slog.Info("Spreadsheet mappings are loaded")

	return
}
