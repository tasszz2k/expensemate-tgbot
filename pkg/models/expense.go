package models

import (
	"expensemate-tgbot/pkg/types/expensetypes"
	"expensemate-tgbot/pkg/types/types"
)

type Expense struct {
	Id       types.Id              `json:"id"`
	Name     string                `json:"name"`
	Amount   types.Unsigned        `json:"amount"`
	Group    expensetypes.Group    `json:"group"` // todo: change to enum
	Category expensetypes.Category `json:"category"`
	Date     types.Time            `json:"date"`
	Note     string                `json:"note"`
}
