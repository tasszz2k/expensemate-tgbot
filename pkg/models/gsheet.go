package models

import (
	"time"

	"expensemate-tgbot/pkg/types/types"
)

type UserSheetMapping struct {
	ID              types.Id      `json:"id"`
	Username        string        `json:"username"`
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
