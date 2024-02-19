package models

import (
	"time"

	"expensemate-tgbot/pkg/types/types"
	"expensemate-tgbot/pkg/utils/httputils"
)

type UserSheetMapping struct {
	ID              types.Id      `json:"id"`
	UserID          types.Id      `json:"user_id"`
	Username        string        `json:"username"`
	FullName        string        `json:"full_name"`
	SpreadSheetsURL string        `json:"spread_sheets_url"`
	CreatedAt       time.Time     `json:"created_at"`
	UpdateAt        time.Time     `json:"update_at"`
	Status          MappingStatus `json:"status"`
}

type MappingStatus string

const (
	MappingStatusMapped  MappingStatus = "MAPPED"
	MappingStatusSuccess MappingStatus = "SUCCESS"
	MappingStatusFailed  MappingStatus = "FAILED"
)

func (u UserSheetMapping) SpreadsheetDocId() string {
	return httputils.GetGoogleSheetsDocID(u.SpreadSheetsURL)
}
