package bots

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"expensemate-tgbot/pkg/clients/gsheetclients"
	"expensemate-tgbot/pkg/models"
	"expensemate-tgbot/pkg/types/expensetypes"
	"expensemate-tgbot/pkg/types/gsheettypes"
	"expensemate-tgbot/pkg/types/tgtypes"
	"expensemate-tgbot/pkg/types/types"
	"expensemate-tgbot/pkg/utils/currencyutils"
	"expensemate-tgbot/pkg/utils/timeutils"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/spf13/cast"
	"google.golang.org/api/sheets/v4"
)

func (e *Expensemate) handleExpensesCommand(
	_ context.Context,
	incomingMessage *tgbotapi.Message,
) (tgbotapi.MessageConfig, error) {
	msg := tgbotapi.NewMessage(incomingMessage.Chat.ID, "")
	msg.Text = "You can manage your expenses."

	expensesKeyboard := tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("Add", "expenses:add"),
			tgbotapi.NewInlineKeyboardButtonData("View", "expenses:view"),
			tgbotapi.NewInlineKeyboardButtonData("Report", "expenses:report"),
		),
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("Update_", "expenses:update"),
			tgbotapi.NewInlineKeyboardButtonData("Delete", "expenses:delete"),
			tgbotapi.NewInlineKeyboardButtonData("Help", "expenses:help"),
		),
	)
	msg.ReplyMarkup = expensesKeyboard

	return msg, nil
}

func (e *Expensemate) handleExpensesAddCommand(
	ctx context.Context,
	incomingMessage *tgbotapi.Message,
) (tgbotapi.MessageConfig, error) {
	msg := tgbotapi.NewMessage(incomingMessage.Chat.ID, "")
	msg.Text = fmt.Sprintf(
		`
Please provide the details of the expense in the following format: 📑
---
[expense name] <b>(*)</b> 🖋️ 
[amount] <b>(*)</b> 🖋️
[group] <i>(default "OTHER")</i> 🖋️ <i> click /expenses_help to see the list of supported groups</i>
[category] <i>(default "Unclassified")</i> 🖋️ <i> click /expenses_help to see the list of supported categories</i>
[date] <i>(auto-filled: %s)</i>
[note] 
`, timeutils.FormatDateOnly(time.Now()),
	)
	// update conversation state
	e.startConversation(
		ctx, incomingMessage.Chat.ID, tgtypes.BuildCallbackData(
			tgtypes.
				CommandExpenses, expensetypes.ActionAdd.String(),
		),
	)
	return msg, nil
}

func (e *Expensemate) handleExpensesCallback(
	ctx context.Context,
	query *tgbotapi.CallbackQuery,
) (tgbotapi.MessageConfig, error) {
	callbackQueryData := query.Data
	_, subCommand := tgtypes.ParseCallbackData(callbackQueryData)
	action := expensetypes.Action(subCommand)
	var msg tgbotapi.MessageConfig

	switch action {
	case expensetypes.ActionAdd:
		msg, _ = e.handleExpensesAddCommand(ctx, query.Message)
	case expensetypes.ActionView:
		msg, _ = e.handleExpensesViewCommand(ctx, query.Message)
	case expensetypes.ActionReport:
		msg, _ = e.handleExpensesReportCommand(ctx, query.Message)
	case expensetypes.ActionHelp:
		msg, _ = e.handleExpenseHelpCommand(ctx, query.Message)
	default:
		msg = tgbotapi.NewMessage(query.Message.Chat.ID, "")
		msg.Text = "Unfortunately, We have not supported this action yet."
	}

	return msg, nil
}

func (e *Expensemate) handleExpensesAdd(
	ctx context.Context,
	message *tgbotapi.Message,
) (tgbotapi.MessageConfig, error) {
	text := message.Text
	msg := tgbotapi.NewMessage(message.Chat.ID, "")

	expense, err := models.ParseTextToExpense(text)
	if err != nil {
		msg.Text = fmt.Sprintf("<b>Invalid expense input!</b>\n%s", err)
		return msg, nil
	}

	// save the expense to the database
	mapping, err := e.getSpreadsheetMappingByUserID(ctx, message.From.ID)
	if err != nil {
		msg.Text = "You haven't configured a Google Sheets yet or the URL is invalid." +
			" Click the button below to configure it."
		msg.ReplyMarkup = tgbotapi.NewInlineKeyboardMarkup(
			tgbotapi.NewInlineKeyboardRow(
				tgbotapi.NewInlineKeyboardButtonData("Configure", "gsheets:configure"),
				tgbotapi.NewInlineKeyboardButtonData("Help", "gsheets:help"),
			),
		)
		return msg, nil
	}
	spreadsheetDocId := mapping.SpreadsheetDocId()

	currentPage, err := e.getCurrentPage(ctx, spreadsheetDocId)
	if err != nil {
		msg.Text = "Failed to get current page for expenses. Make sure the database is set up correctly."
		return msg, nil
	}
	nextId, err := e.getExpensesNextId(ctx, spreadsheetDocId, currentPage)
	if err != nil {
		msg.Text = "Failed to get next id for expenses. Make sure the database is set up correctly."
		return msg, nil
	}

	expense.Id = types.Id(nextId)
	if err = e.insertNewExpense(ctx, spreadsheetDocId, currentPage, expense); err != nil {
		msg.Text = "Failed to save the expense to the database."
		return msg, nil
	}

	// return the saved expense
	msg.Text = expense.String()

	e.endConversation(ctx, message.Chat.ID)
	return msg, nil
}

func (e *Expensemate) handleExpenseHelpCommand(
	_ context.Context,
	message *tgbotapi.Message,
) (tgbotapi.MessageConfig, error) {
	msg := tgbotapi.NewMessage(message.Chat.ID, "")
	msg.Text = `
Here are the supported aliases for expense groups and categories:
<b>I. Expense groups:</b>

1. INCOME / thu nhập
    - Alias: I
2. MUST HAVE / chi tiêu thiết yếu
    - Alias: MH
3. NICE TO HAVE / không phải chi tiêu thiết yếu, nhưng nên chi, có thì tốt
    - Alias: NTH
4. WASTE / chi tiêu không cần thiết, lãng phí
    - Alias: W
5. OTHER / khác
    - Alias: O

<b>II. Expense categories:</b>

1. Unclassified / Chưa phân loại
    - Vietnamese Alias: CPL
    - English Alias: UC

2. Food / Ăn uống
    - Vietnamese Alias: AU
    - English Alias: F

3. Housing / Nhà ở
    - Vietnamese Alias: NO
    - English Alias: H

4. Transportation / Đi lại
    - Vietnamese Alias: DL
    - English Alias: T

5. Utilities / Tiện ích
    - Vietnamese Alias: TI
    - English Alias: U

6. Healthcare / Sức khỏe
    - Vietnamese Alias: SK
    - English Alias: H

7. Entertainment / Giải trí
    - Vietnamese Alias: GT
    - English Alias: EN

8. Education / Giáo dục
    - Vietnamese Alias: GD
    - English Alias: ED

9. Clothing / Quần áo
    - Vietnamese Alias: QA
    - English Alias: C

10. Personal Care / Chăm sóc cá nhân
    - Vietnamese Alias: CSCN
    - English Alias: PC

11. Miscellaneous / Đồ linh tinh
    - Vietnamese Alias: DLT/LT
    - English Alias: M

12. Travel / Du lịch
    - Vietnamese Alias: Du Lich
    - English Alias: TV

13. Other / Khác
    - Vietnamese Alias: K
    - English Alias: O
`
	return msg, nil
}

func (e *Expensemate) getExpensesNextId(
	ctx context.Context, spreadsheetDocId,
	currentPage string,
) (int, error) {
	nextIdCell := gsheettypes.BuildCell(
		currentPage,
		gsheettypes.ExpensemateExpensesNextIdCell,
	)
	nextId, err := gsheetclients.GetInstance().GetValue(
		ctx, spreadsheetDocId, nextIdCell,
	)
	if err != nil {
		slog.Error(
			fmt.Sprintf(
				"Failed to get next id for expenses in page %s: %s",
				spreadsheetDocId, currentPage,
			), err,
		)
		return 0, err
	}
	return cast.ToInt(nextId), nil
}

func (e *Expensemate) getCurrentPage(ctx context.Context, spreadsheetDocId string) (string, error) {
	currentPageCell := gsheettypes.BuildCell(
		gsheettypes.ExpensemateDatabaseSheetName,
		gsheettypes.ExpensemateDatabaseCurrentPageCell,
	)
	currentPage, err := gsheetclients.GetInstance().GetValue(
		ctx, spreadsheetDocId, currentPageCell,
	)
	if err != nil {
		slog.Error("Failed to load Expensemate database, current page cell")
		return "", err
	}
	return currentPage, nil
}

func (e *Expensemate) insertNewExpense(
	ctx context.Context,
	spreadsheetDocId, currentPage string,
	expense models.Expense,
) error {
	row := int(gsheettypes.ExpensemateExpensesTopRow + expense.Id)
	writeRange := gsheettypes.BuildRangeFromCells(
		currentPage,
		gsheettypes.ExpensemateExpensesLeftCol, row,
		gsheettypes.ExpensemateExpensesRightCol, row,
	)

	values := [][]interface{}{
		{
			expense.Id,
			expense.Name,
			expense.Amount,
			expense.Category,
			expense.Group,
			timeutils.FormatDateOnly(expense.Date),
			expense.Note,
		},
	}
	vr := &sheets.ValueRange{
		Values: values,
	}
	if _, err := gsheetclients.GetInstance().Update(
		ctx,
		spreadsheetDocId,
		writeRange,
		vr,
	); err != nil {
		slog.Error("Failed to insert new expense", err)
		return err
	}

	// update next id
	nextId := int(expense.Id) + 1
	if err := e.updateExpensesNextId(
		ctx, spreadsheetDocId, currentPage, nextId,
	); err != nil {
		return err
	}

	return nil
}

func (e *Expensemate) updateExpensesNextId(
	ctx context.Context, spreadsheetDocId, currentPage string,
	nextId int,
) error {
	nextIdCell := gsheettypes.BuildCell(
		currentPage,
		gsheettypes.ExpensemateExpensesNextIdCell,
	)
	if _, err := gsheetclients.GetInstance().Update(
		ctx,
		spreadsheetDocId,
		nextIdCell,
		&sheets.ValueRange{
			Values: [][]interface{}{
				{nextId},
			},
		},
	); err != nil {
		slog.Error("Failed to update next id for expenses", err)
		return err
	}
	return nil
}

func (e *Expensemate) handleExpensesViewCommand(
	ctx context.Context,
	message *tgbotapi.Message,
) (tgbotapi.MessageConfig, error) {
	msg := tgbotapi.NewMessage(message.Chat.ID, "")
	mapping, err := e.getSpreadsheetMappingByUserID(ctx, message.Chat.ID)
	if err != nil {
		msg.Text = "You haven't configured a Google Sheets yet or the URL is invalid." +
			" Click the button below to configure it."
		msg.ReplyMarkup = tgbotapi.NewInlineKeyboardMarkup(
			tgbotapi.NewInlineKeyboardRow(
				tgbotapi.NewInlineKeyboardButtonData("Configure", "gsheets:configure"),
				tgbotapi.NewInlineKeyboardButtonData("Help", "gsheets:help"),
			),
		)
		return msg, nil
	}
	spreadsheetDocId := mapping.SpreadsheetDocId()
	currentPage, err := e.getCurrentPage(ctx, spreadsheetDocId)
	if err != nil {
		msg.Text = "Failed to get current page for expenses. Make sure the database is set up correctly."
		return msg, nil
	}

	expenses, err := e.get5MostRecentExpenses(ctx, spreadsheetDocId, currentPage)
	if err != nil {
		msg.Text = "Failed to get expenses from the database."
		return msg, nil
	}

	msg.Text = "Here are your 5 most recent expenses:\n-----\n"
	for _, expense := range expenses {
		msg.Text += expense.String()
	}
	return msg, nil
}

func (e *Expensemate) get5MostRecentExpenses(
	ctx context.Context, spreadsheetDocId,
	currentPage string,
) ([]models.Expense, error) {
	expensesNextId, err := e.getExpensesNextId(ctx, spreadsheetDocId, currentPage)
	if err != nil {
		return nil, err
	}

	lastExpenseId := int(expensesNextId) - 1
	if lastExpenseId < 1 {
		return []models.Expense{}, nil
	}
	topRow := max(
		gsheettypes.ExpensemateExpensesTopRow+lastExpenseId-5,
		gsheettypes.ExpensemateExpensesTopRow+1,
	)

	readRange := gsheettypes.BuildRangeFromCells(
		currentPage,
		gsheettypes.ExpensemateExpensesLeftCol,
		topRow,
		gsheettypes.ExpensemateExpensesRightCol,
		gsheettypes.ExpensemateExpensesTopRow+lastExpenseId,
	)
	values, err := gsheetclients.GetInstance().Get(ctx, spreadsheetDocId, readRange)
	if err != nil {
		return nil, err
	}

	expenses := make([]models.Expense, 0, len(values.Values))
	for _, row := range values.Values {
		expense, err := models.ParseRowToExpense(row)
		if err != nil {
			continue
		}
		expenses = append(expenses, expense)
	}
	return expenses, nil
}

func (e *Expensemate) handleExpensesReportCommand(
	ctx context.Context,
	message *tgbotapi.Message,
) (tgbotapi.MessageConfig, error) {
	msg := tgbotapi.NewMessage(message.Chat.ID, "")
	mapping, err := e.getSpreadsheetMappingByUserID(ctx, message.Chat.ID)
	if err != nil {
		msg.Text = "You haven't configured a Google Sheets yet or the URL is invalid." +
			" Click the button below to configure it."
		msg.ReplyMarkup = tgbotapi.NewInlineKeyboardMarkup(
			tgbotapi.NewInlineKeyboardRow(
				tgbotapi.NewInlineKeyboardButtonData("Configure", "gsheets:configure"),
				tgbotapi.NewInlineKeyboardButtonData("Help", "gsheets:help"),
			),
		)
		return msg, nil
	}
	spreadsheetDocId := mapping.SpreadsheetDocId()
	currentPage, err := e.getCurrentPage(ctx, spreadsheetDocId)
	if err != nil {
		msg.Text = "Failed to get current page for expenses. Make sure the database is set up correctly."
		return msg, nil
	}

	// get report
	reportReadRange := gsheettypes.BuildRange(
		currentPage,
		gsheettypes.ExpensemateExpensesReportRange,
	)
	reportValues, err := gsheetclients.GetInstance().Get(ctx, spreadsheetDocId, reportReadRange)
	if err != nil {
		msg.Text = "Failed to get report from the database."
		return msg, nil
	}

	categoryReadRange := gsheettypes.BuildRange(
		currentPage,
		gsheettypes.ExpensemateExpensesCategoryRange,
	)
	categoryValues, err := gsheetclients.GetInstance().Get(ctx, spreadsheetDocId, categoryReadRange)
	if err != nil {
		msg.Text = "Failed to get category from the database."
		return msg, nil
	}

	// format report and category
	msg.Text = "Here is your expense report, categorized by groups:\n-----\n"
	for _, row := range reportValues.Values {
		msg.Text += fmt.Sprintf("<b>%s</b>: %s\n", row[0], row[1])
	}
	msg.Text += "\nHere are your expense report, categorized by categories:\n-----\n"
	for _, row := range categoryValues.Values {
		totalAmount, _ := currencyutils.ReverseFormatVND(cast.ToString(row[1]))
		if totalAmount > 0 {
			msg.Text += fmt.Sprintf("<b>%s</b>: %s | %s \n", row[0], row[1], row[2])
		}
	}
	return msg, nil
}
