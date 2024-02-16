package bots

import (
	"context"
	"fmt"
	"time"

	"expensemate-tgbot/pkg/models"
	"expensemate-tgbot/pkg/types/expensetypes"
	"expensemate-tgbot/pkg/types/tgtypes"
	"expensemate-tgbot/pkg/utils/timeutils"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
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
	_ context.Context,
	incomingMessage *tgbotapi.Message,
) (tgbotapi.MessageConfig, error) {
	msg := tgbotapi.NewMessage(incomingMessage.Chat.ID, "")
	msg.Text = "You can quickly add an expense."
	return msg, nil
}

func (e *Expensemate) handleExpensesCallback(
	ctx context.Context,
	query *tgbotapi.CallbackQuery,
) (tgbotapi.MessageConfig, error) {
	callbackQueryData := query.Data
	_, subCommand := tgtypes.ParseCallbackData(callbackQueryData)
	action := expensetypes.Action(subCommand)

	msg := tgbotapi.NewMessage(query.Message.Chat.ID, "")
	switch action {
	case expensetypes.ActionAdd:
		msg.Text = fmt.Sprintf(
			`
Please provide the details of the expense in the following format: 📑
---
[expense name] <b>(*)</b> 🖋️ 
[amount] <b>(*)</b> 🖋️
[group] <i>(default "OTHER")</i> 🖋️ <i> click /expenses_groups to see the list of supported groups</i>
[category] <i>(default "Unclassified")</i> 🖋️ <i> click /expenses_categories to see the list of supported categories</i>
[date] <i>(auto-filled: %s)</i>
[note] 
`, timeutils.FormatDateOnly(time.Now()),
		)
		// update conversation state
		e.updateConversationState(
			ctx, query.Message.Chat.ID, tgtypes.BuildCallbackData(
				tgtypes.
					CommandExpenses, expensetypes.ActionAdd.String(),
			),
		)
	case expensetypes.ActionView:
		msg.Text = "Here are your 5 latest expenses:"
	case expensetypes.ActionReport:
		msg.Text = "Here is your expense report:"
	case expensetypes.ActionHelp:
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
	default:
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
	msg.ReplyToMessageID = message.MessageID

	expense, err := models.ParseTextToExpense(text)
	if err != nil {
		msg.Text = fmt.Sprintf("<b>Invalid expense input!</b>\n%s", err)
		return msg, nil
	}

	// todo: save the expense to the database
	expense.Id = 1
	// return the saved expense
	msg.Text = expense.String()

	e.removeConversationState(ctx, message.Chat.ID)
	return msg, nil
}
