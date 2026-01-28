package handler

import (
	"context"
	"fmt"
	"time"

	"expensemate-tgbot/internal/log"
	"expensemate-tgbot/internal/service"
	"expensemate-tgbot/internal/state"
	"expensemate-tgbot/internal/types"
	"expensemate-tgbot/internal/util/currency"
	timepkg "expensemate-tgbot/internal/util/time"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

// ExpenseHandler handles expense-related commands and callbacks
type ExpenseHandler struct {
	expenseService *service.ExpenseService
	mappingService *service.MappingService
	stateManager   *state.Manager
}

// NewExpenseHandler creates a new ExpenseHandler
func NewExpenseHandler(
	expenseService *service.ExpenseService,
	mappingService *service.MappingService,
	stateManager *state.Manager,
) *ExpenseHandler {
	return &ExpenseHandler{
		expenseService: expenseService,
		mappingService: mappingService,
		stateManager:   stateManager,
	}
}

// HandleExpensesCommand handles the /expenses command
func (h *ExpenseHandler) HandleExpensesCommand(ctx context.Context, msg *tgbotapi.Message) (tgbotapi.MessageConfig, error) {
	logger := log.WithMessage(msg)
	log.WithAction(logger, "expenses").Info("user accessed expenses menu")

	reply := tgbotapi.NewMessage(msg.Chat.ID, "")

	// Check if user has configured spreadsheet
	mapping, err := h.mappingService.GetByUserID(ctx, types.ID(msg.From.ID))
	if err != nil || mapping == nil {
		return h.getUnauthorizedMsg(reply), nil
	}

	reply.Text = "Manage your expenses:"
	reply.ReplyMarkup = tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("Add", "expenses:add"),
			tgbotapi.NewInlineKeyboardButtonData("View", "expenses:view"),
			tgbotapi.NewInlineKeyboardButtonData("Report", "expenses:report"),
		),
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("Update", "expenses:update"),
			tgbotapi.NewInlineKeyboardButtonData("Delete", "expenses:delete"),
			tgbotapi.NewInlineKeyboardButtonData("Help", "expenses:help"),
		),
	)

	return reply, nil
}

// HandleExpensesAddCommand handles the /expenses_add command
func (h *ExpenseHandler) HandleExpensesAddCommand(ctx context.Context, msg *tgbotapi.Message) (tgbotapi.MessageConfig, error) {
	logger := log.WithMessage(msg)
	log.WithAction(logger, "expenses_add").Info("starting add expense flow")

	reply := tgbotapi.NewMessage(msg.Chat.ID, "")

	// Check authorization
	mapping, err := h.mappingService.GetByUserID(ctx, types.ID(msg.From.ID))
	if err != nil || mapping == nil {
		return h.getUnauthorizedMsg(reply), nil
	}

	reply.Text = fmt.Sprintf(`Please provide the expense details:
---
[expense name] <b>(*)</b>
[amount] <b>(*)</b>
[group] <i>(default "MUST HAVE")</i> - /expenses_help for list
[category] <i>(default "Unclassified")</i> - /expenses_help for list
[date] <i>(default: %s)</i>
[note]`, timepkg.FormatDateOnly(time.Now()))

	// Start conversation
	h.stateManager.Start(msg.Chat.ID, state.StateExpensesAdd)

	return reply, nil
}

// HandleExpensesAdd handles expense input in conversation
func (h *ExpenseHandler) HandleExpensesAdd(ctx context.Context, msg *tgbotapi.Message) (tgbotapi.MessageConfig, error) {
	logger := log.WithMessage(msg)
	log.WithAction(logger, "expenses_add").Info("processing expense input")

	reply := tgbotapi.NewMessage(msg.Chat.ID, "")

	expense, err := h.expenseService.Add(ctx, types.ID(msg.From.ID), msg.Text)
	if err != nil {
		reply.Text = fmt.Sprintf("<b>Error:</b> %s", err.Error())
		return reply, nil
	}

	// Get spreadsheet URL for link
	spreadsheetURL, _ := h.mappingService.GetSpreadsheetURL(ctx, types.ID(msg.From.ID))

	reply.Text = fmt.Sprintf(`<b>Expense added successfully!</b>
%s

<a href="%s">View in Google Sheets</a>`, expense.FormatHTML(), spreadsheetURL)

	// Restart conversation for next expense
	go func() {
		time.Sleep(500 * time.Millisecond)
		h.stateManager.Start(msg.Chat.ID, state.StateExpensesAdd)
	}()

	return reply, nil
}

// HandleExpensesView handles viewing recent expenses
func (h *ExpenseHandler) HandleExpensesView(ctx context.Context, chatID int64, userID int64) (tgbotapi.MessageConfig, error) {
	log.Debug("viewing expenses", log.Fields{"user_id": userID, "chat_id": chatID, "action": "expenses_view"})

	reply := tgbotapi.NewMessage(chatID, "")

	expenses, spreadsheetURL, err := h.expenseService.GetRecent(ctx, types.ID(userID), 5)
	if err != nil {
		reply.Text = fmt.Sprintf("<b>Error:</b> %s", err.Error())
		return reply, nil
	}

	if len(expenses) == 0 {
		reply.Text = "No expenses found."
		return reply, nil
	}

	text := "<b>Recent Expenses:</b>\n\n"
	for _, e := range expenses {
		text += fmt.Sprintf("<b>%d.</b> %s - %s (%s)\n",
			e.ID, e.Name, currency.FormatVND(e.Amount), timepkg.FormatDateOnly(e.Date))
	}
	text += fmt.Sprintf("\n<a href=\"%s\">View all in Google Sheets</a>", spreadsheetURL)

	reply.Text = text
	return reply, nil
}

// HandleExpensesReport handles expense report
func (h *ExpenseHandler) HandleExpensesReport(ctx context.Context, chatID int64, userID int64) (tgbotapi.MessageConfig, error) {
	log.Debug("generating report", log.Fields{"user_id": userID, "chat_id": chatID, "action": "expenses_report"})

	reply := tgbotapi.NewMessage(chatID, "")

	groupReport, categoryReport, spreadsheetURL, err := h.expenseService.GetReport(ctx, types.ID(userID))
	if err != nil {
		reply.Text = fmt.Sprintf("<b>Error:</b> %s", err.Error())
		return reply, nil
	}

	text := "<b>Expense Report</b>\n\n"

	// Group report
	text += "<b>By Group:</b>\n"
	for _, row := range groupReport {
		if len(row) >= 2 {
			text += fmt.Sprintf("  %s: %s\n", row[0], row[1])
		}
	}

	// Category report
	text += "\n<b>By Category:</b>\n"
	for _, row := range categoryReport {
		if len(row) >= 2 && row[1] != "" && row[1] != "0" {
			text += fmt.Sprintf("  %s: %s\n", row[0], row[1])
		}
	}

	text += fmt.Sprintf("\n<a href=\"%s\">View full report in Google Sheets</a>", spreadsheetURL)

	reply.Text = text
	return reply, nil
}

// HandleExpensesHelp handles the /expenses_help command
func (h *ExpenseHandler) HandleExpensesHelp(ctx context.Context, msg *tgbotapi.Message) (tgbotapi.MessageConfig, error) {
	logger := log.WithMessage(msg)
	log.WithAction(logger, "expenses_help").Info("showing expense help")

	reply := tgbotapi.NewMessage(msg.Chat.ID, "")
	reply.Text = `<b>Expense Groups:</b>
- INCOME (i) - Thu nhap
- MUST HAVE (mh) - Chi tieu thiet yeu
- NICE TO HAVE (nth) - Khong thiet yeu nhung nen chi
- WASTED (w) - Lang phi
- OTHER (o) - Khac

<b>Expense Categories:</b>
- Food (f, au) - An uong
- Housing (h, no) - Nha o
- Transportation (t, dl) - Di lai
- Utilities (u, ti) - Tien ich
- Healthcare (hc, sk) - Suc khoe
- Entertainment (en, gt) - Giai tri
- Education (ed, gd) - Giao duc
- Clothing (c, qa) - Quan ao
- Personal Care (pc, cscn) - Cham soc ca nhan
- Miscellaneous (m, dlt) - Do linh tinh
- Travel (tv) - Du lich
- Other (o, k) - Khac
- Unclassified (uc, cpl) - Chua phan loai`

	return reply, nil
}

// HandleCallback handles expense-related callbacks
func (h *ExpenseHandler) HandleCallback(ctx context.Context, cb *tgbotapi.CallbackQuery, action string, subCommands []string) (tgbotapi.MessageConfig, error) {
	logger := log.WithCallback(cb)
	log.WithAction(logger, "expenses_callback").WithField("callback", cb.Data).Info("processing callback")

	chatID := cb.Message.Chat.ID
	userID := cb.From.ID // Use cb.From for the actual user who clicked

	switch types.ExpenseAction(action) {
	case types.ExpenseActionAdd:
		return h.handleExpensesAddCallback(ctx, cb)
	case types.ExpenseActionView:
		h.stateManager.End(chatID)
		return h.HandleExpensesView(ctx, chatID, userID)
	case types.ExpenseActionReport:
		h.stateManager.End(chatID)
		return h.HandleExpensesReport(ctx, chatID, userID)
	case types.ExpenseActionHelp:
		h.stateManager.End(chatID)
		return h.handleExpensesHelpCallback(ctx, chatID)
	default:
		h.stateManager.End(chatID)
		reply := tgbotapi.NewMessage(chatID, "")
		reply.Text = "This action is not yet implemented."
		return reply, nil
	}
}

// handleExpensesAddCallback handles add expense callback
func (h *ExpenseHandler) handleExpensesAddCallback(ctx context.Context, cb *tgbotapi.CallbackQuery) (tgbotapi.MessageConfig, error) {
	chatID := cb.Message.Chat.ID
	userID := cb.From.ID

	reply := tgbotapi.NewMessage(chatID, "")

	// Check authorization
	mapping, err := h.mappingService.GetByUserID(ctx, types.ID(userID))
	if err != nil || mapping == nil {
		return h.getUnauthorizedMsg(reply), nil
	}

	reply.Text = fmt.Sprintf(`Please provide the expense details:
---
[expense name] <b>(*)</b>
[amount] <b>(*)</b>
[group] <i>(default "MUST HAVE")</i> - /expenses_help for list
[category] <i>(default "Unclassified")</i> - /expenses_help for list
[date] <i>(default: %s)</i>
[note]`, timepkg.FormatDateOnly(time.Now()))

	// Start conversation
	h.stateManager.Start(chatID, state.StateExpensesAdd)

	return reply, nil
}

// handleExpensesHelpCallback handles help callback
func (h *ExpenseHandler) handleExpensesHelpCallback(ctx context.Context, chatID int64) (tgbotapi.MessageConfig, error) {
	reply := tgbotapi.NewMessage(chatID, "")
	reply.Text = `<b>Expense Groups:</b>
- INCOME (i) - Thu nhap
- MUST HAVE (mh) - Chi tieu thiet yeu
- NICE TO HAVE (nth) - Khong thiet yeu nhung nen chi
- WASTED (w) - Lang phi
- OTHER (o) - Khac

<b>Expense Categories:</b>
- Food (f, au) - An uong
- Housing (h, no) - Nha o
- Transportation (t, dl) - Di lai
- Utilities (u, ti) - Tien ich
- Healthcare (hc, sk) - Suc khoe
- Entertainment (en, gt) - Giai tri
- Education (ed, gd) - Giao duc
- Clothing (c, qa) - Quan ao
- Personal Care (pc, cscn) - Cham soc ca nhan
- Miscellaneous (m, dlt) - Do linh tinh
- Travel (tv) - Du lich
- Other (o, k) - Khac
- Unclassified (uc, cpl) - Chua phan loai`

	return reply, nil
}

// getUnauthorizedMsg returns the unauthorized message with configure button
func (h *ExpenseHandler) getUnauthorizedMsg(msg tgbotapi.MessageConfig) tgbotapi.MessageConfig {
	msg.Text = "Please configure Google Sheets first using /gsheets command."
	msg.ReplyMarkup = tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("Configure", "gsheets:configure"),
		),
	)
	return msg
}

// EndConversation ends the conversation for a chat
func (h *ExpenseHandler) EndConversation(chatID int64) {
	h.stateManager.End(chatID)
}
