package models

import (
	"errors"
	"fmt"
	"strings"
	"time"

	"expensemate-tgbot/pkg/types/expensetypes"
	"expensemate-tgbot/pkg/types/types"
	"expensemate-tgbot/pkg/utils/currencyutils"
	"expensemate-tgbot/pkg/utils/timeutils"

	"github.com/spf13/cast"
)

type Expense struct {
	Id       types.Id              `json:"id"`
	Name     string                `json:"name"`
	Amount   types.Unsigned        `json:"amount"`
	Group    expensetypes.Group    `json:"group"`
	Category expensetypes.Category `json:"category"`
	Date     time.Time             `json:"date"`
	Note     string                `json:"note"`
}

func ParseTextToExpense(text string) (Expense, error) {
	rows := strings.Split(text, "\n")
	if len(rows) < 2 {
		return Expense{}, errors.New("invalid expense text")
	}
	// fill to fix a length array with size = 6
	rows = append(rows, make([]string, 6-len(rows))...)

	amount, err := cast.ToUint64E(rows[1])
	if err != nil {
		return Expense{}, fmt.Errorf("invalid amount: %w", err)
	}

	group, ok := expensetypes.GetGroupByAlias(strings.TrimSpace(rows[2]))
	if !ok {
		return Expense{}, fmt.Errorf("invalid group: %s", rows[2])
	}
	category, ok := expensetypes.GetCategoryByAlias(strings.TrimSpace(rows[3]))
	if !ok {
		return Expense{}, fmt.Errorf("invalid category: %s", rows[3])
	}

	date := time.Now()

	if rows[4] != "" {
		date, err = time.Parse(timeutils.DateOnlyFormat, rows[4])
		if err != nil {
			date = time.Now()
		}
	}

	expense := Expense{
		Name:     strings.TrimSpace(rows[0]),
		Amount:   types.Unsigned(amount),
		Group:    group,
		Category: category,
		Date:     date,
		Note:     rows[5],
	}

	return expense, nil
}

func (e Expense) String() string {
	return fmt.Sprintf(
		`
<b>ID</b>: <i>%d</i>
<b>Name</b>: <i>%s</i>
<b>Amount</b>: <i>%s</i>
<b>Group</b>: <i>%s</i>
<b>Category</b>: <i>%s</i>
<b>Date</b>: <i>%s</i>
<b>Note</b>: <i>%s</i>
`,
		e.Id,
		e.Name,
		currencyutils.FormatVND(e.Amount),
		e.Group,
		e.Category,
		timeutils.FormatDateOnly(e.Date),
		e.Note,
	)
}
