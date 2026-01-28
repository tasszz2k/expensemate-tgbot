package model

import (
	"time"

	"expensemate-tgbot/internal/types"
	httputil "expensemate-tgbot/internal/util/http"
)

// MappingStatus represents the status of a user-spreadsheet mapping
type MappingStatus string

const (
	MappingStatusMapped  MappingStatus = "MAPPED"
	MappingStatusSuccess MappingStatus = "SUCCESS"
	MappingStatusFailed  MappingStatus = "FAILED"
)

// UserSheetMapping represents a user-to-spreadsheet mapping
type UserSheetMapping struct {
	ID              types.ID      `json:"id"`
	UserID          types.ID      `json:"user_id"`
	Username        string        `json:"username"`
	FullName        string        `json:"full_name"`
	SpreadSheetsURL string        `json:"spread_sheets_url"`
	CreatedAt       time.Time     `json:"created_at"`
	UpdatedAt       time.Time     `json:"updated_at"`
	Status          MappingStatus `json:"status"`
}

// SpreadsheetDocID extracts the document ID from the spreadsheet URL
func (m UserSheetMapping) SpreadsheetDocID() string {
	return httputil.GetGoogleSheetsDocID(m.SpreadSheetsURL)
}

// ToRow converts UserSheetMapping to a Google Sheets row
func (m UserSheetMapping) ToRow() []interface{} {
	return []interface{}{
		int64(m.ID),
		int64(m.UserID),
		m.Username,
		m.FullName,
		m.SpreadSheetsURL,
		m.CreatedAt.Format("2/1/2006"),
		m.UpdatedAt.Format("2/1/2006"),
		string(m.Status),
	}
}
