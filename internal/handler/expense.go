package handler

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"time"

	"expensemate-tgbot/internal/log"
	"expensemate-tgbot/internal/model"
	"expensemate-tgbot/internal/service"
	"expensemate-tgbot/internal/state"
	"expensemate-tgbot/internal/types"
	"expensemate-tgbot/internal/util/currency"
	timepkg "expensemate-tgbot/internal/util/time"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

// ExpenseHandler handles expense-related commands and callbacks
type ExpenseHandler struct {
	expenseService      *service.ExpenseService
	mappingService      *service.MappingService
	voiceExpenseService *service.VoiceExpenseService
	stateManager        *state.Manager
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

// SetVoiceExpenseService sets the voice expense service (optional, for voice support)
func (h *ExpenseHandler) SetVoiceExpenseService(voiceService *service.VoiceExpenseService) {
	h.voiceExpenseService = voiceService
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

	reply.Text = "💰 <b>Expense Manager</b>"
	reply.ReplyMarkup = tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("➕ Add", "expenses:add"),
			tgbotapi.NewInlineKeyboardButtonData("📋 View", "expenses:view"),
			tgbotapi.NewInlineKeyboardButtonData("📊 Report", "expenses:report"),
		),
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("💵 Budget", "expenses:budget"),
			tgbotapi.NewInlineKeyboardButtonData("✏️ Update", "expenses:update"),
			tgbotapi.NewInlineKeyboardButtonData("🗑 Delete", "expenses:delete"),
		),
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("❓ Help", "expenses:help"),
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

	reply.Text = fmt.Sprintf(`➕ <b>Add Expense</b>

📝 name <b>(*)</b>
💵 amount <b>(*)</b>
💼 group <i>(or select below)</i>
📂 category <i>(or select below)</i>
📅 date <i>(default: %s)</i>
🗒 note`, timepkg.FormatDateOnly(time.Now()))

	// Start conversation
	h.stateManager.Start(msg.Chat.ID, state.StateExpensesAdd)

	return reply, nil
}

// HandleExpensesAdd handles expense input in conversation
func (h *ExpenseHandler) HandleExpensesAdd(ctx context.Context, msg *tgbotapi.Message) (tgbotapi.MessageConfig, error) {
	logger := log.WithMessage(msg)
	log.WithAction(logger, "expenses_add").Info("processing expense input")

	reply := tgbotapi.NewMessage(msg.Chat.ID, "")

	result, err := h.expenseService.Add(ctx, types.ID(msg.From.ID), msg.Text)
	if err != nil {
		reply.Text = fmt.Sprintf("<b>Error:</b> %s", err.Error())
		return reply, nil
	}

	expense := result.Expense

	// Get spreadsheet URL pointing to active sheet
	spreadsheetURL, _ := h.expenseService.GetSpreadsheetURL(ctx, types.ID(msg.From.ID))

	reply.Text = fmt.Sprintf("✅ <b>Expense added!</b>\n\n%s", expense.FormatHTML())

	// Append budget status if budgets exist
	groupBudget, categoryBudget, _ := h.expenseService.GetBudgetForExpense(ctx, types.ID(msg.From.ID), expense.Group, expense.Category)
	budgetLine := formatBudgetStatus(groupBudget, categoryBudget)
	if budgetLine != "" {
		reply.Text += "\n\n" + budgetLine
	}

	reply.Text += fmt.Sprintf("\n\n🔗 <a href=\"%s\">View in Google Sheets</a>", spreadsheetURL)

	// Only show group buttons if user didn't explicitly provide group
	if !result.GroupProvided {
		reply.Text += "\n\n💼 <b>Select a group:</b>"
		reply.ReplyMarkup = h.buildGroupKeyboard(expense.ID, !result.CategoryProvided)
	} else if !result.CategoryProvided && expense.Group.NeedsCategory() {
		// Group was provided, but category wasn't - show category selection
		// Skip category selection for Income/Investment groups
		reply.Text += "\n\n📂 <b>Select a category:</b>"
		reply.ReplyMarkup = h.buildCategoryKeyboard(expense.ID)
	}

	// Restart conversation for next expense
	go func() {
		time.Sleep(500 * time.Millisecond)
		h.stateManager.Start(msg.Chat.ID, state.StateExpensesAdd)
	}()

	return reply, nil
}

// buildGroupKeyboard builds inline keyboard with group buttons
// needsCategoryAfter indicates if category selection should follow group selection
func (h *ExpenseHandler) buildGroupKeyboard(expenseID types.ID, needsCategoryAfter bool) tgbotapi.InlineKeyboardMarkup {
	groups := types.GetAllGroups()
	var rows [][]tgbotapi.InlineKeyboardButton

	// Build rows with 2 buttons each
	for i := 0; i < len(groups); i += 2 {
		var row []tgbotapi.InlineKeyboardButton

		// First button
		grp1 := groups[i]
		alias1 := types.GetGroupAlias(grp1)
		name1 := types.GetGroupShortName(grp1)
		// Callback format: expenses:setgrp:expense_id:group_alias:needs_category (0 or 1)
		needsCat := "0"
		if needsCategoryAfter {
			needsCat = "1"
		}
		callback1 := fmt.Sprintf("expenses:setgrp:%d:%s:%s", expenseID, alias1, needsCat)
		row = append(row, tgbotapi.NewInlineKeyboardButtonData(name1, callback1))

		// Second button (if exists)
		if i+1 < len(groups) {
			grp2 := groups[i+1]
			alias2 := types.GetGroupAlias(grp2)
			name2 := types.GetGroupShortName(grp2)
			callback2 := fmt.Sprintf("expenses:setgrp:%d:%s:%s", expenseID, alias2, needsCat)
			row = append(row, tgbotapi.NewInlineKeyboardButtonData(name2, callback2))
		}

		rows = append(rows, row)
	}

	// Add Skip button - if skipping group, still ask for category if needed
	needsCat := "0"
	if needsCategoryAfter {
		needsCat = "1"
	}
	skipCallback := fmt.Sprintf("expenses:setgrp:%d:skip:%s", expenseID, needsCat)
	rows = append(rows, tgbotapi.NewInlineKeyboardRow(
		tgbotapi.NewInlineKeyboardButtonData("⏭ Skip (Keep Must Have)", skipCallback),
	))

	return tgbotapi.NewInlineKeyboardMarkup(rows...)
}

// buildCategoryKeyboard builds inline keyboard with category buttons
func (h *ExpenseHandler) buildCategoryKeyboard(expenseID types.ID) tgbotapi.InlineKeyboardMarkup {
	categories := types.GetAllCategories()
	var rows [][]tgbotapi.InlineKeyboardButton

	// Build rows with 2 buttons each
	for i := 0; i < len(categories); i += 2 {
		var row []tgbotapi.InlineKeyboardButton

		// First button
		cat1 := categories[i]
		alias1 := types.GetCategoryAlias(cat1)
		name1 := types.GetCategoryShortName(cat1)
		callback1 := fmt.Sprintf("expenses:setcat:%d:%s", expenseID, alias1)
		row = append(row, tgbotapi.NewInlineKeyboardButtonData(name1, callback1))

		// Second button (if exists)
		if i+1 < len(categories) {
			cat2 := categories[i+1]
			alias2 := types.GetCategoryAlias(cat2)
			name2 := types.GetCategoryShortName(cat2)
			callback2 := fmt.Sprintf("expenses:setcat:%d:%s", expenseID, alias2)
			row = append(row, tgbotapi.NewInlineKeyboardButtonData(name2, callback2))
		}

		rows = append(rows, row)
	}

	// Add Skip button - include expense ID so we can show final record
	rows = append(rows, tgbotapi.NewInlineKeyboardRow(
		tgbotapi.NewInlineKeyboardButtonData("⏭ Skip (Keep Unclassified)", fmt.Sprintf("expenses:setcat:%d:skip", expenseID)),
	))

	return tgbotapi.NewInlineKeyboardMarkup(rows...)
}

// HandleExpensesView handles viewing recent expenses
func (h *ExpenseHandler) HandleExpensesView(ctx context.Context, chatID int64, userID int64) (tgbotapi.MessageConfig, error) {
	return h.buildExpensesView(ctx, chatID, userID, 5)
}

// buildExpensesView builds the expenses view message with the given count
func (h *ExpenseHandler) buildExpensesView(ctx context.Context, chatID int64, userID int64, count int) (tgbotapi.MessageConfig, error) {
	log.Debug("viewing expenses", log.Fields{"user_id": userID, "chat_id": chatID, "action": "expenses_view", "count": count})

	const maxCount = 25
	if count > maxCount {
		count = maxCount
	}

	reply := tgbotapi.NewMessage(chatID, "")

	expenses, spreadsheetURL, err := h.expenseService.GetRecent(ctx, types.ID(userID), count)
	if err != nil {
		reply.Text = fmt.Sprintf("<b>Error:</b> %s", err.Error())
		return reply, nil
	}

	if len(expenses) == 0 {
		reply.Text = "📭 No expenses found."
		return reply, nil
	}

	text := "📋 <b>Recent Expenses:</b>\n\n"
	for _, e := range expenses {
		text += fmt.Sprintf("💰 <b>%d.</b> %s — <b>%s</b> (%s)\n",
			e.ID, e.Name, currency.FormatVND(e.Amount), timepkg.FormatDateOnly(e.Date))
	}
	text += fmt.Sprintf("\n🔗 <a href=\"%s\">View all in Google Sheets</a>", spreadsheetURL)

	// Add "Show more" button unless we've hit the max
	if count < maxCount {
		nextCount := count + 5
		if nextCount > maxCount {
			nextCount = maxCount
		}
		keyboard := tgbotapi.NewInlineKeyboardMarkup(
			tgbotapi.NewInlineKeyboardRow(
				tgbotapi.NewInlineKeyboardButtonData("Show more ⬇️", fmt.Sprintf("expenses:viewmore:%d", nextCount)),
			),
		)
		reply.ReplyMarkup = keyboard
	}

	reply.Text = text
	return reply, nil
}

// HandleExpensesReport handles expense report with budget comparison
func (h *ExpenseHandler) HandleExpensesReport(ctx context.Context, chatID int64, userID int64) (tgbotapi.MessageConfig, error) {
	log.Debug("generating report", log.Fields{"user_id": userID, "chat_id": chatID, "action": "expenses_report"})

	reply := tgbotapi.NewMessage(chatID, "")

	groups, categories, spreadsheetURL, err := h.expenseService.GetBudgetOverview(ctx, types.ID(userID))
	if err != nil {
		reply.Text = fmt.Sprintf("<b>Error:</b> %s", err.Error())
		return reply, nil
	}

	text := "📊 <b>Expense Report</b>\n\n"
	text += "💼 <b>By Group:</b>\n"
	text += sortedReportLines(groups, "🔸")

	text += "\n📂 <b>By Category:</b>\n"
	text += sortedReportLines(categories, "🔹")

	text += fmt.Sprintf("\n🔗 <a href=\"%s\">View full report in Google Sheets</a>", spreadsheetURL)

	reply.Text = text
	return reply, nil
}

// HandleExpensesHelp handles the /expenses_help command
func (h *ExpenseHandler) HandleExpensesHelp(ctx context.Context, msg *tgbotapi.Message) (tgbotapi.MessageConfig, error) {
	logger := log.WithMessage(msg)
	log.WithAction(logger, "expenses_help").Info("showing expense help")

	reply := tgbotapi.NewMessage(msg.Chat.ID, "")
	reply.Text = getHelpText()
	return reply, nil
}

// HandleCallback handles expense-related callbacks
func (h *ExpenseHandler) HandleCallback(ctx context.Context, cb *tgbotapi.CallbackQuery, action string, subCommands []string) (Response, error) {
	logger := log.WithCallback(cb)
	log.WithAction(logger, "expenses_callback").WithField("callback", cb.Data).Info("processing callback")

	chatID := cb.Message.Chat.ID
	messageID := cb.Message.MessageID
	userID := cb.From.ID // Use cb.From for the actual user who clicked

	switch types.ExpenseAction(action) {
	case types.ExpenseActionAdd:
		msg, err := h.handleExpensesAddCallback(ctx, cb)
		return NewMessageResponse(msg), err
	case types.ExpenseActionView:
		h.stateManager.End(chatID)
		msg, err := h.HandleExpensesView(ctx, chatID, userID)
		return NewMessageResponse(msg), err
	case types.ExpenseActionViewMore:
		return h.handleViewMoreCallback(ctx, subCommands, chatID, messageID, userID)
	case types.ExpenseActionReport:
		h.stateManager.End(chatID)
		msg, err := h.HandleExpensesReport(ctx, chatID, userID)
		return NewMessageResponse(msg), err
	case types.ExpenseActionHelp:
		h.stateManager.End(chatID)
		msg, err := h.handleExpensesHelpCallback(ctx, chatID)
		return NewMessageResponse(msg), err
	case types.ExpenseActionSetGroup:
		return h.handleSetGroupCallback(ctx, cb, subCommands, chatID, messageID, userID)
	case types.ExpenseActionSetCategory:
		return h.handleSetCategoryCallback(ctx, cb, subCommands, chatID, messageID, userID)
	case types.ExpenseActionQuickDelete:
		return h.handleQuickDeleteCallback(ctx, cb, subCommands, chatID, messageID, userID)
	case types.ExpenseActionBudget:
		return h.handleBudgetCallback(ctx, cb, subCommands[1:], chatID, messageID, userID)
	default:
		h.stateManager.End(chatID)
		reply := tgbotapi.NewMessage(chatID, "")
		reply.Text = "🚧 This action is not yet available."
		return NewMessageResponse(reply), nil
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

	reply.Text = fmt.Sprintf(`➕ <b>Add Expense</b>

📝 name <b>(*)</b>
💵 amount <b>(*)</b>
💼 group <i>(or select below)</i>
📂 category <i>(or select below)</i>
📅 date <i>(default: %s)</i>
🗒 note`, timepkg.FormatDateOnly(time.Now()))

	// Start conversation
	h.stateManager.Start(chatID, state.StateExpensesAdd)

	return reply, nil
}

// handleExpensesHelpCallback handles help callback
func (h *ExpenseHandler) handleExpensesHelpCallback(ctx context.Context, chatID int64) (tgbotapi.MessageConfig, error) {
	reply := tgbotapi.NewMessage(chatID, "")
	reply.Text = getHelpText()
	return reply, nil
}

// handleViewMoreCallback handles "Show more" button clicks on the expenses view
func (h *ExpenseHandler) handleViewMoreCallback(ctx context.Context, subCommands []string, chatID int64, messageID int, userID int64) (Response, error) {
	if len(subCommands) < 2 {
		return NewEditResponse(chatID, messageID, "Invalid callback data"), nil
	}

	count, err := strconv.Atoi(subCommands[1])
	if err != nil || count <= 0 {
		return NewEditResponse(chatID, messageID, "Invalid count"), nil
	}

	msg, err := h.buildExpensesView(ctx, chatID, userID, count)
	if err != nil {
		return NewEditResponse(chatID, messageID, fmt.Sprintf("<b>Error:</b> %s", err.Error())), nil
	}

	// Edit the existing message instead of sending a new one
	keyboard, _ := msg.ReplyMarkup.(tgbotapi.InlineKeyboardMarkup)
	if len(keyboard.InlineKeyboard) > 0 {
		return NewEditWithKeyboardResponse(chatID, messageID, msg.Text, keyboard), nil
	}
	return NewEditResponse(chatID, messageID, msg.Text), nil
}

// handleSetGroupCallback handles group selection callback
func (h *ExpenseHandler) handleSetGroupCallback(ctx context.Context, cb *tgbotapi.CallbackQuery, subCommands []string, chatID int64, messageID int, userID int64) (Response, error) {
	// subCommands: [action, expense_id, group_alias, needs_category]
	// action is at index 0, so we need indices 1, 2, 3
	if len(subCommands) < 4 {
		return NewEditResponse(chatID, messageID, "Invalid callback data"), nil
	}

	needsCategory := subCommands[3] == "1"
	transcription := extractTranscription(cb.Message.Text)

	expenseID, err := strconv.ParseInt(subCommands[1], 10, 64)
	if err != nil {
		return NewEditResponse(chatID, messageID, "Invalid expense ID"), nil
	}

	// Check for skip - group stays as default
	if subCommands[2] == "skip" {
		if needsCategory && expenseID > 0 {
			text := formatTranscriptionHTML(transcription) + "📂 <b>Select a category:</b>"
			return NewEditWithKeyboardResponse(chatID, messageID, text, h.buildCategoryKeyboard(types.ID(expenseID))), nil
		}
		return h.buildFinalExpenseResponse(ctx, userID, expenseID, chatID, messageID, transcription)
	}

	groupAlias := subCommands[2]
	group, ok := types.GetGroupByAlias(groupAlias)
	if !ok {
		return NewEditResponse(chatID, messageID, "Invalid group"), nil
	}

	// Update group
	if err := h.expenseService.UpdateGroup(ctx, types.ID(userID), types.ID(expenseID), group); err != nil {
		return NewEditResponse(chatID, messageID, fmt.Sprintf("<b>Error:</b> %s", err.Error())), nil
	}

	// If category still needs to be selected and this group type needs categories, show category buttons
	// Income and Investment don't need category selection
	if needsCategory && group.NeedsCategory() {
		text := formatTranscriptionHTML(transcription) + "📂 <b>Select a category:</b>"
		return NewEditWithKeyboardResponse(chatID, messageID, text, h.buildCategoryKeyboard(types.ID(expenseID))), nil
	}

	return h.buildFinalExpenseResponse(ctx, userID, expenseID, chatID, messageID, transcription)
}

// handleSetCategoryCallback handles category selection callback
func (h *ExpenseHandler) handleSetCategoryCallback(ctx context.Context, cb *tgbotapi.CallbackQuery, subCommands []string, chatID int64, messageID int, userID int64) (Response, error) {
	// subCommands: [action, expense_id, category_alias]
	// action is at index 0, so we need indices 1, 2
	if len(subCommands) < 3 {
		return NewEditResponse(chatID, messageID, "Invalid callback data"), nil
	}

	transcription := extractTranscription(cb.Message.Text)

	expenseID, err := strconv.ParseInt(subCommands[1], 10, 64)
	if err != nil {
		return NewEditResponse(chatID, messageID, "Invalid expense ID"), nil
	}

	// Check for skip - category stays as default
	if subCommands[2] == "skip" {
		return h.buildFinalExpenseResponse(ctx, userID, expenseID, chatID, messageID, transcription)
	}

	categoryAlias := subCommands[2]
	category, ok := types.GetCategoryByAlias(categoryAlias)
	if !ok {
		return NewEditResponse(chatID, messageID, "Invalid category"), nil
	}

	// Update category
	if err := h.expenseService.UpdateCategory(ctx, types.ID(userID), types.ID(expenseID), category); err != nil {
		return NewEditResponse(chatID, messageID, fmt.Sprintf("<b>Error:</b> %s", err.Error())), nil
	}

	return h.buildFinalExpenseResponse(ctx, userID, expenseID, chatID, messageID, transcription)
}

// buildFinalExpenseResponse retrieves the expense and builds the final response message.
// transcription is included in the response when non-empty (voice expenses).
func (h *ExpenseHandler) buildFinalExpenseResponse(ctx context.Context, userID, expenseID int64, chatID int64, messageID int, transcription string) (Response, error) {
	expense, err := h.expenseService.GetByID(ctx, types.ID(userID), types.ID(expenseID))
	if err != nil {
		return NewEditResponse(chatID, messageID, fmt.Sprintf("✅ <b>Expense saved!</b>\n🆔 %d\n\n<i>(Could not load details)</i>", expenseID)), nil
	}
	if expense == nil {
		return NewEditResponse(chatID, messageID, fmt.Sprintf("✅ <b>Expense saved!</b>\n🆔 %d\n\n<i>(Expense not found)</i>", expenseID)), nil
	}

	sheetURL, _ := h.expenseService.GetSpreadsheetURL(ctx, types.ID(userID))

	text := fmt.Sprintf("✅ <b>Expense saved!</b>\n%s%s", formatTranscriptionHTML(transcription), expense.FormatHTML())

	groupBudget, categoryBudget, _ := h.expenseService.GetBudgetForExpense(ctx, types.ID(userID), expense.Group, expense.Category)
	budgetLine := formatBudgetStatus(groupBudget, categoryBudget)
	if budgetLine != "" {
		text += "\n\n" + budgetLine
	}

	text += fmt.Sprintf("\n\n🔗 <a href=\"%s\">View in Google Sheets</a>", sheetURL)
	return NewEditResponse(chatID, messageID, text), nil
}

// handleQuickDeleteCallback handles quick delete callback for voice expenses
func (h *ExpenseHandler) handleQuickDeleteCallback(ctx context.Context, cb *tgbotapi.CallbackQuery, subCommands []string, chatID int64, messageID int, userID int64) (Response, error) {
	// subCommands: [action, expense_id]
	if len(subCommands) < 2 {
		return NewEditResponse(chatID, messageID, "Invalid callback data"), nil
	}

	expenseID, err := strconv.ParseInt(subCommands[1], 10, 64)
	if err != nil {
		return NewEditResponse(chatID, messageID, "Invalid expense ID"), nil
	}

	// Get expense info for confirmation message
	expense, err := h.expenseService.GetByID(ctx, types.ID(userID), types.ID(expenseID))
	if err != nil {
		return NewEditResponse(chatID, messageID, fmt.Sprintf("<b>Error:</b> %s", err.Error())), nil
	}

	// Delete the expense
	username := cb.From.UserName
	if username == "" {
		username = fmt.Sprintf("user_%d", userID)
	}
	if err := h.expenseService.Delete(ctx, types.ID(userID), types.ID(expenseID), "@"+username); err != nil {
		return NewEditResponse(chatID, messageID, fmt.Sprintf("<b>Error deleting:</b> %s", err.Error())), nil
	}

	text := "🗑 <b>Expense deleted!</b>"
	if expense != nil {
		text = fmt.Sprintf("🗑 <b>Deleted:</b> %s — <b>%s</b>",
			expense.Name, currency.FormatVND(expense.Amount))
	}

	return NewEditResponse(chatID, messageID, text), nil
}

// buildVoiceExpenseKeyboard builds keyboard with group/category buttons plus delete
func (h *ExpenseHandler) buildVoiceExpenseKeyboard(expenseID types.ID, needsGroup, needsCategory bool) tgbotapi.InlineKeyboardMarkup {
	var rows [][]tgbotapi.InlineKeyboardButton

	if needsGroup {
		// Add group buttons (2 per row)
		groups := types.GetAllGroups()
		for i := 0; i < len(groups); i += 2 {
			var row []tgbotapi.InlineKeyboardButton
			grp1 := groups[i]
			alias1 := types.GetGroupAlias(grp1)
			name1 := types.GetGroupShortName(grp1)
			needsCat := "0"
			if needsCategory {
				needsCat = "1"
			}
			callback1 := fmt.Sprintf("expenses:setgrp:%d:%s:%s", expenseID, alias1, needsCat)
			row = append(row, tgbotapi.NewInlineKeyboardButtonData(name1, callback1))

			if i+1 < len(groups) {
				grp2 := groups[i+1]
				alias2 := types.GetGroupAlias(grp2)
				name2 := types.GetGroupShortName(grp2)
				callback2 := fmt.Sprintf("expenses:setgrp:%d:%s:%s", expenseID, alias2, needsCat)
				row = append(row, tgbotapi.NewInlineKeyboardButtonData(name2, callback2))
			}
			rows = append(rows, row)
		}
		// Skip group button
		needsCat := "0"
		if needsCategory {
			needsCat = "1"
		}
		skipCallback := fmt.Sprintf("expenses:setgrp:%d:skip:%s", expenseID, needsCat)
		rows = append(rows, tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("⏭ Skip Group", skipCallback),
		))
	} else if needsCategory {
		// Add category buttons (2 per row)
		categories := types.GetAllCategories()
		for i := 0; i < len(categories); i += 2 {
			var row []tgbotapi.InlineKeyboardButton
			cat1 := categories[i]
			alias1 := types.GetCategoryAlias(cat1)
			name1 := types.GetCategoryShortName(cat1)
			callback1 := fmt.Sprintf("expenses:setcat:%d:%s", expenseID, alias1)
			row = append(row, tgbotapi.NewInlineKeyboardButtonData(name1, callback1))

			if i+1 < len(categories) {
				cat2 := categories[i+1]
				alias2 := types.GetCategoryAlias(cat2)
				name2 := types.GetCategoryShortName(cat2)
				callback2 := fmt.Sprintf("expenses:setcat:%d:%s", expenseID, alias2)
				row = append(row, tgbotapi.NewInlineKeyboardButtonData(name2, callback2))
			}
			rows = append(rows, row)
		}
		// Skip category button
		rows = append(rows, tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("⏭ Skip Category", fmt.Sprintf("expenses:setcat:%d:skip", expenseID)),
		))
	}

	// Always add delete button at the bottom
	deleteCallback := fmt.Sprintf("expenses:qdel:%d", expenseID)
	rows = append(rows, tgbotapi.NewInlineKeyboardRow(
		tgbotapi.NewInlineKeyboardButtonData("🗑 Delete", deleteCallback),
	))

	return tgbotapi.NewInlineKeyboardMarkup(rows...)
}

// extractTranscription extracts the transcribed text from a voice expense message.
// Returns the transcription formatted as HTML, or empty string if not found.
func extractTranscription(messageText string) string {
	const prefix = "Transcribed: \""
	startIdx := strings.Index(messageText, prefix)
	if startIdx == -1 {
		return ""
	}
	startIdx += len(prefix)
	endIdx := strings.Index(messageText[startIdx:], "\"")
	if endIdx == -1 {
		return ""
	}
	return messageText[startIdx : startIdx+endIdx]
}

// formatTranscriptionHTML wraps transcription text in HTML for message display.
func formatTranscriptionHTML(transcription string) string {
	if transcription == "" {
		return ""
	}
	return fmt.Sprintf("<b>Transcribed:</b> <i>\"%s\"</i>\n\n", transcription)
}

// getHelpText returns the help text with all groups and categories
func getHelpText() string {
	return `💼 <b>Expense Groups:</b>
💰 INCOME (i, tn) - Thu nhập
📈 INVESTMENT (inv, dt) - Đầu tư
🔒 MUST HAVE (mh, ty) - Thiết yếu
✨ NICE TO HAVE (nth, nc) - Nên chi
🗑 WASTE (w, lp) - Lãng phí
👨‍👩‍👧‍👦 FAMILY (fam, gd) - Gia đình
❤️ LOVER (lov, ny) - Người yêu

📂 <b>Expense Categories:</b>
🍜 Food (f, an, cf) - Ăn ngoài
🛒 Groceries (gr, dc) - Đi chợ
🚗 Transport (tr, dl) - Đi lại
🎮 Entertainment (ent, gt) - Giải trí
📦 Miscellaneous (mis, lt) - Linh tinh
🔄 Subscription (sub, dk) - Đăng ký
🏠 Housing (hou, no) - Nhà ở
💆 Personal Care (pc, cs) - Chăm sóc
🏥 Healthcare (hc, sk) - Sức khỏe
👕 Clothing (clo, qa) - Quần áo
📚 Education (edu, hoc) - Giáo dục
💻 Tech (tech, cn) - Công nghệ
✈️ Travel (tv, dul) - Du lịch
🎁 Present (pre, qt) - Quà tặng
🎊 Life Events (le, hh) - Hiếu hỉ
❤️ Lover (lov, ny) - Người yêu
👨‍👩‍👧‍👦 Family (fam, gd) - Gia đình
💸 Lost Money (lm, mat) - Mất tiền
📋 Unclassified (uc, cpl) - Chưa phân loại`
}

// getUnauthorizedMsg returns the unauthorized message with configure button
func (h *ExpenseHandler) getUnauthorizedMsg(msg tgbotapi.MessageConfig) tgbotapi.MessageConfig {
	msg.Text = "⚠️ Please configure Google Sheets first using /gsheets."
	msg.ReplyMarkup = tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("⚙️ Configure", "gsheets:configure"),
		),
	)
	return msg
}

// EndConversation ends the conversation for a chat
func (h *ExpenseHandler) EndConversation(chatID int64) {
	h.stateManager.End(chatID)
}

// HandleVoiceExpense handles voice message input for expenses
func (h *ExpenseHandler) HandleVoiceExpense(ctx context.Context, msg *tgbotapi.Message) (tgbotapi.MessageConfig, error) {
	logger := log.WithMessage(msg)
	log.WithAction(logger, "voice_expense").Info("processing voice expense")

	reply := tgbotapi.NewMessage(msg.Chat.ID, "")

	// Check if voice service is enabled
	if h.voiceExpenseService == nil || !h.voiceExpenseService.IsEnabled() {
		reply.Text = "🎤 Voice input is not available. Please use text input."
		return reply, nil
	}

	// Check authorization
	mapping, err := h.mappingService.GetByUserID(ctx, types.ID(msg.From.ID))
	if err != nil || mapping == nil {
		return h.getUnauthorizedMsg(reply), nil
	}

	// Send processing message
	reply.Text = "🎤 Processing voice message..."

	// Process voice message
	result, err := h.voiceExpenseService.ProcessVoiceMessage(ctx, types.ID(msg.From.ID), msg.Voice)
	if err != nil {
		log.Error("voice processing failed", err, log.Fields{
			"user_id": msg.From.ID,
			"action":  "voice_expense_error",
		})
		reply.Text = fmt.Sprintf("❌ <b>Error processing voice:</b> %s\n\nPlease try again or use text input.", err.Error())
		return reply, nil
	}

	// Handle clarification needed
	if result.NeedsClarification {
		// Store pending data for clarification
		h.stateManager.SetVoicePendingData(msg.Chat.ID, &state.VoicePendingData{
			OriginalText: result.TranscribedText,
			ParsedName:   result.ParsedData.Name,
			ParsedAmount: result.ParsedData.Amount,
			ParsedGroup:  result.ParsedData.Group,
			ParsedCat:    result.ParsedData.Category,
			ParsedDate:   result.ParsedData.Date,
			ParsedNote:   result.ParsedData.Note,
		})
		h.stateManager.Start(msg.Chat.ID, state.StateExpensesVoiceClarify)

		reply.Text = fmt.Sprintf("🎤 <b>I heard:</b> <i>\"%s\"</i>\n\n%s", result.TranscribedText, result.ClarificationQuestion)
		return reply, nil
	}

	// Handle error
	if result.Error != nil {
		reply.Text = fmt.Sprintf("<b>Error:</b> %s", result.Error.Error())
		return reply, nil
	}

	// Success - expense created
	expense := result.Expense
	needsGroup := expense.Group == types.GroupMustHave
	// Always allow category selection for voice expenses since ChatGPT infers
	// the category from context, and the user should be able to confirm or change it
	needsCategory := expense.Group.NeedsCategory()

	groupBudget, categoryBudget, _ := h.expenseService.GetBudgetForExpense(ctx, types.ID(msg.From.ID), expense.Group, expense.Category)
	budgetLine := formatBudgetStatus(groupBudget, categoryBudget)

	reply.Text = fmt.Sprintf("🎤 <b>Voice expense added!</b>\n\n🗣 <i>\"%s\"</i>\n\n%s",
		result.TranscribedText, expense.FormatHTML())
	if budgetLine != "" {
		reply.Text += "\n\n" + budgetLine
	}
	reply.Text += fmt.Sprintf("\n\n🔗 <a href=\"%s\">View in Google Sheets</a>", result.SpreadsheetURL)

	// Show keyboard with group/category selection and delete button
	if needsGroup {
		reply.Text += "\n\n💼 <b>Select a group:</b>"
	} else if needsCategory {
		reply.Text += "\n\n📂 <b>Select a category:</b>"
	}
	reply.ReplyMarkup = h.buildVoiceExpenseKeyboard(expense.ID, needsGroup, needsCategory)

	return reply, nil
}

// HandleVoiceClarification handles clarification response for voice expense
func (h *ExpenseHandler) HandleVoiceClarification(ctx context.Context, msg *tgbotapi.Message) (tgbotapi.MessageConfig, error) {
	logger := log.WithMessage(msg)
	log.WithAction(logger, "voice_clarification").Info("processing voice clarification")

	reply := tgbotapi.NewMessage(msg.Chat.ID, "")

	// Get pending data
	pendingData := h.stateManager.GetVoicePendingData(msg.Chat.ID)
	if pendingData == nil {
		h.stateManager.End(msg.Chat.ID)
		reply.Text = "⚠️ No pending voice expense. Please send a new voice message."
		return reply, nil
	}

	// Process clarification
	result, err := h.voiceExpenseService.ProcessClarification(ctx, types.ID(msg.From.ID), msg.Text, pendingData)
	if err != nil {
		log.Error("clarification processing failed", err, log.Fields{
			"user_id": msg.From.ID,
			"action":  "voice_clarification_error",
		})
		reply.Text = fmt.Sprintf("<b>Error:</b> %s\n\nPlease try again.", err.Error())
		return reply, nil
	}

	// Still needs clarification
	if result.NeedsClarification {
		// Update pending data with new parsed values
		h.stateManager.SetVoicePendingData(msg.Chat.ID, &state.VoicePendingData{
			OriginalText: pendingData.OriginalText + "\n" + msg.Text,
			ParsedName:   result.ParsedData.Name,
			ParsedAmount: result.ParsedData.Amount,
			ParsedGroup:  result.ParsedData.Group,
			ParsedCat:    result.ParsedData.Category,
			ParsedDate:   result.ParsedData.Date,
			ParsedNote:   result.ParsedData.Note,
		})

		reply.Text = result.ClarificationQuestion
		return reply, nil
	}

	// Clear pending data and end conversation
	h.stateManager.ClearVoicePendingData(msg.Chat.ID)
	h.stateManager.End(msg.Chat.ID)

	// Handle error
	if result.Error != nil {
		reply.Text = fmt.Sprintf("<b>Error:</b> %s", result.Error.Error())
		return reply, nil
	}

	// Success - expense created
	expense := result.Expense
	needsGroup := expense.Group == types.GroupMustHave
	// Always allow category selection for voice expenses since ChatGPT infers
	// the category from context, and the user should be able to confirm or change it
	needsCategory := expense.Group.NeedsCategory()

	groupBudget, categoryBudget, _ := h.expenseService.GetBudgetForExpense(ctx, types.ID(msg.From.ID), expense.Group, expense.Category)
	budgetLine := formatBudgetStatus(groupBudget, categoryBudget)

	reply.Text = fmt.Sprintf("🎤 <b>Voice expense added!</b>\n\n%s", expense.FormatHTML())
	if budgetLine != "" {
		reply.Text += "\n\n" + budgetLine
	}
	reply.Text += fmt.Sprintf("\n\n🔗 <a href=\"%s\">View in Google Sheets</a>", result.SpreadsheetURL)

	// Show keyboard with group/category selection and delete button
	if needsGroup {
		reply.Text += "\n\n💼 <b>Select a group:</b>"
	} else if needsCategory {
		reply.Text += "\n\n📂 <b>Select a category:</b>"
	}
	reply.ReplyMarkup = h.buildVoiceExpenseKeyboard(expense.ID, needsGroup, needsCategory)

	return reply, nil
}

// isNonZeroAmount checks if an amount string represents a non-zero value.
func isNonZeroAmount(amount string) bool {
	for _, c := range amount {
		if c >= '1' && c <= '9' {
			return true
		}
	}
	return false
}

// abs64 returns the absolute value of x.
func abs64(x int64) int64 {
	if x < 0 {
		return -x
	}
	return x
}

// formatBudgetStatus formats budget status for after-add display.
func formatBudgetStatus(groupBudget, categoryBudget *model.BudgetEntry) string {
	var parts []string
	if groupBudget != nil {
		parts = append(parts, groupBudget.FormatShortBudgetLine())
	}
	if categoryBudget != nil {
		parts = append(parts, categoryBudget.FormatShortBudgetLine())
	}
	if len(parts) == 0 {
		return ""
	}
	return "📊 <b>Budget:</b>\n" + strings.Join(parts, "\n")
}

// --- Budget handlers ---

func (h *ExpenseHandler) handleBudgetCallback(ctx context.Context, cb *tgbotapi.CallbackQuery, subCommands []string, chatID int64, messageID int, userID int64) (Response, error) {
	if len(subCommands) == 0 {
		return h.handleBudgetMenu(ctx, chatID, messageID, userID)
	}

	action := subCommands[0]
	switch action {
	case "view":
		return h.handleBudgetView(ctx, chatID, messageID, userID)
	case "setgrp":
		return h.handleBudgetSetGroup(ctx, subCommands[1:], chatID, messageID, userID)
	case "setcat":
		return h.handleBudgetSetCategory(ctx, subCommands[1:], chatID, messageID, userID)
	case "cleargrp":
		return h.handleBudgetClearGroup(ctx, subCommands[1:], chatID, messageID, userID)
	case "clearcat":
		return h.handleBudgetClearCategory(ctx, subCommands[1:], chatID, messageID, userID)
	case "back":
		h.stateManager.End(chatID)
		return h.handleBudgetMenu(ctx, chatID, messageID, userID)
	default:
		return NewEditResponse(chatID, messageID, "Unknown budget action"), nil
	}
}

// HandleBudgetCommand handles the /budget command (sends a new message)
func (h *ExpenseHandler) HandleBudgetCommand(ctx context.Context, msg *tgbotapi.Message) (tgbotapi.MessageConfig, error) {
	logger := log.WithMessage(msg)
	log.WithAction(logger, "budget").Info("user accessed budget menu")

	reply := tgbotapi.NewMessage(msg.Chat.ID, "")

	mapping, err := h.mappingService.GetByUserID(ctx, types.ID(msg.From.ID))
	if err != nil || mapping == nil {
		return h.getUnauthorizedMsg(reply), nil
	}

	reply.Text = "💵 <b>Budget Manager</b>\nManage your spending caps:"
	reply.ReplyMarkup = tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("📊 View Budgets", "expenses:budget:view"),
		),
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("💼 Set Group", "expenses:budget:setgrp"),
			tgbotapi.NewInlineKeyboardButtonData("📂 Set Category", "expenses:budget:setcat"),
		),
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("⬅️ Back", "expenses:budget:back"),
		),
	)

	return reply, nil
}

func (h *ExpenseHandler) handleBudgetMenu(ctx context.Context, chatID int64, messageID int, userID int64) (Response, error) {
	h.stateManager.End(chatID)

	keyboard := tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("📊 View Budgets", "expenses:budget:view"),
		),
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("💼 Set Group", "expenses:budget:setgrp"),
			tgbotapi.NewInlineKeyboardButtonData("📂 Set Category", "expenses:budget:setcat"),
		),
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("⬅️ Back", "expenses:budget:back"),
		),
	)

	return NewEditWithKeyboardResponse(chatID, messageID, "💵 <b>Budget Manager</b>\nManage your spending caps:", keyboard), nil
}

func (h *ExpenseHandler) handleBudgetView(ctx context.Context, chatID int64, messageID int, userID int64) (Response, error) {
	groups, categories, sheetURL, err := h.expenseService.GetBudgetOverview(ctx, types.ID(userID))
	if err != nil {
		return NewEditResponse(chatID, messageID, fmt.Sprintf("<b>Error:</b> %s", err.Error())), nil
	}

	text := "📊 <b>Budget Overview</b>\n\n"

	// Group section -- sorted: over -> ok -> empty
	groupOver, groupOk := 0, 0
	var groupLinesOver, groupLinesOk, groupLinesEmpty []string
	for _, g := range groups {
		if !g.HasBudget {
			continue
		}
		line := g.FormatBudgetLine()
		if g.Remaining < 0 {
			groupOver++
			groupLinesOver = append(groupLinesOver, line)
		} else if g.Spent > 0 {
			groupOk++
			groupLinesOk = append(groupLinesOk, line)
		} else {
			groupOk++
			groupLinesEmpty = append(groupLinesEmpty, line)
		}
	}

	total := len(groupLinesOver) + len(groupLinesOk) + len(groupLinesEmpty)
	if total > 0 {
		text += fmt.Sprintf("💼 <b>By Group</b> <i>(%d ok, %d over)</i>\n", groupOk, groupOver)
		for _, line := range groupLinesOver {
			text += line + "\n"
		}
		for _, line := range groupLinesOk {
			text += line + "\n"
		}
		for _, line := range groupLinesEmpty {
			text += line + "\n"
		}
	}

	// Category section -- sorted: over -> ok -> empty
	catOver, catOk := 0, 0
	var catLinesOver, catLinesOk, catLinesEmpty []string
	for _, c := range categories {
		if !c.HasBudget {
			continue
		}
		line := c.FormatBudgetLine()
		if c.Remaining < 0 {
			catOver++
			catLinesOver = append(catLinesOver, line)
		} else if c.Spent > 0 {
			catOk++
			catLinesOk = append(catLinesOk, line)
		} else {
			catOk++
			catLinesEmpty = append(catLinesEmpty, line)
		}
	}

	total = len(catLinesOver) + len(catLinesOk) + len(catLinesEmpty)
	if total > 0 {
		text += fmt.Sprintf("\n📂 <b>By Category</b> <i>(%d ok, %d over)</i>\n", catOk, catOver)
		for _, line := range catLinesOver {
			text += line + "\n"
		}
		for _, line := range catLinesOk {
			text += line + "\n"
		}
		for _, line := range catLinesEmpty {
			text += line + "\n"
		}
	}

	text += fmt.Sprintf("\n🔗 <a href=\"%s\">View in Google Sheets</a>", sheetURL)

	keyboard := tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("⬅️ Back", "expenses:budget:back"),
		),
	)

	return NewEditWithKeyboardResponse(chatID, messageID, text, keyboard), nil
}

func (h *ExpenseHandler) handleBudgetSetGroup(ctx context.Context, subCommands []string, chatID int64, messageID int, userID int64) (Response, error) {
	// If a specific row was selected, prompt for amount
	if len(subCommands) > 0 {
		row, err := strconv.Atoi(subCommands[0])
		if err != nil {
			return NewEditResponse(chatID, messageID, "Invalid row"), nil
		}

		// Find group name for this row
		groupName := ""
		for g, r := range types.GroupBudgetRow {
			if r == row {
				groupName = types.GetGroupShortName(g)
				break
			}
		}

		h.stateManager.Start(chatID, state.StateBudgetSetGroupPrefix+subCommands[0])
		h.stateManager.SetBudgetPendingData(chatID, &state.BudgetPendingData{
			Col: types.GroupBudgetCol,
			Row: row,
		})

		return NewEditResponse(chatID, messageID,
			fmt.Sprintf("💵 Enter budget for <b>%s</b>:\n\n<i>Supports: 100k, 1.5m, 1500000</i>", groupName)), nil
	}

	// Show group selection buttons
	var rows [][]tgbotapi.InlineKeyboardButton
	allGroups := types.GetAllGroups()

	for i := 0; i < len(allGroups); i += 2 {
		var row []tgbotapi.InlineKeyboardButton
		g := allGroups[i]
		if r, ok := types.GroupBudgetRow[g]; ok {
			row = append(row, tgbotapi.NewInlineKeyboardButtonData(
				types.GetGroupShortName(g),
				fmt.Sprintf("expenses:budget:setgrp:%d", r),
			))
		}
		if i+1 < len(allGroups) {
			g2 := allGroups[i+1]
			if r, ok := types.GroupBudgetRow[g2]; ok {
				row = append(row, tgbotapi.NewInlineKeyboardButtonData(
					types.GetGroupShortName(g2),
					fmt.Sprintf("expenses:budget:setgrp:%d", r),
				))
			}
		}
		if len(row) > 0 {
			rows = append(rows, row)
		}
	}

	rows = append(rows, tgbotapi.NewInlineKeyboardRow(
		tgbotapi.NewInlineKeyboardButtonData("Back", "expenses:budget:back"),
	))

	keyboard := tgbotapi.NewInlineKeyboardMarkup(rows...)
	return NewEditWithKeyboardResponse(chatID, messageID, "💼 <b>Select a group to set budget:</b>", keyboard), nil
}

func (h *ExpenseHandler) handleBudgetSetCategory(ctx context.Context, subCommands []string, chatID int64, messageID int, userID int64) (Response, error) {
	// If a specific row was selected, prompt for amount
	if len(subCommands) > 0 {
		row, err := strconv.Atoi(subCommands[0])
		if err != nil {
			return NewEditResponse(chatID, messageID, "Invalid row"), nil
		}

		// Find category name for this row
		catName := ""
		for c, r := range types.CategoryBudgetRow {
			if r == row {
				catName = types.GetCategoryShortName(c)
				break
			}
		}

		h.stateManager.Start(chatID, state.StateBudgetSetCategoryPrefix+subCommands[0])
		h.stateManager.SetBudgetPendingData(chatID, &state.BudgetPendingData{
			Col: types.CategoryBudgetCol,
			Row: row,
		})

		return NewEditResponse(chatID, messageID,
			fmt.Sprintf("💵 Enter budget for <b>%s</b>:\n\n<i>Supports: 100k, 1.5m, 1500000</i>", catName)), nil
	}

	// Show category selection buttons
	var rows [][]tgbotapi.InlineKeyboardButton
	allCategories := types.GetAllCategories()

	for i := 0; i < len(allCategories); i += 2 {
		var row []tgbotapi.InlineKeyboardButton
		c := allCategories[i]
		if r, ok := types.CategoryBudgetRow[c]; ok {
			row = append(row, tgbotapi.NewInlineKeyboardButtonData(
				types.GetCategoryShortName(c),
				fmt.Sprintf("expenses:budget:setcat:%d", r),
			))
		}
		if i+1 < len(allCategories) {
			c2 := allCategories[i+1]
			if r, ok := types.CategoryBudgetRow[c2]; ok {
				row = append(row, tgbotapi.NewInlineKeyboardButtonData(
					types.GetCategoryShortName(c2),
					fmt.Sprintf("expenses:budget:setcat:%d", r),
				))
			}
		}
		if len(row) > 0 {
			rows = append(rows, row)
		}
	}

	rows = append(rows, tgbotapi.NewInlineKeyboardRow(
		tgbotapi.NewInlineKeyboardButtonData("Back", "expenses:budget:back"),
	))

	keyboard := tgbotapi.NewInlineKeyboardMarkup(rows...)
	return NewEditWithKeyboardResponse(chatID, messageID, "📂 <b>Select a category to set budget:</b>", keyboard), nil
}

func (h *ExpenseHandler) handleBudgetClearGroup(ctx context.Context, subCommands []string, chatID int64, messageID int, userID int64) (Response, error) {
	if len(subCommands) == 0 {
		// Show group selection for clearing
		var rows [][]tgbotapi.InlineKeyboardButton
		allGroups := types.GetAllGroups()
		for i := 0; i < len(allGroups); i += 2 {
			var row []tgbotapi.InlineKeyboardButton
			g := allGroups[i]
			if r, ok := types.GroupBudgetRow[g]; ok {
				row = append(row, tgbotapi.NewInlineKeyboardButtonData(
					types.GetGroupShortName(g),
					fmt.Sprintf("expenses:budget:cleargrp:%d", r),
				))
			}
			if i+1 < len(allGroups) {
				g2 := allGroups[i+1]
				if r, ok := types.GroupBudgetRow[g2]; ok {
					row = append(row, tgbotapi.NewInlineKeyboardButtonData(
						types.GetGroupShortName(g2),
						fmt.Sprintf("expenses:budget:cleargrp:%d", r),
					))
				}
			}
			if len(row) > 0 {
				rows = append(rows, row)
			}
		}
		rows = append(rows, tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("⬅️ Back", "expenses:budget:back"),
		))
		keyboard := tgbotapi.NewInlineKeyboardMarkup(rows...)
		return NewEditWithKeyboardResponse(chatID, messageID, "🧹 <b>Select a group to clear budget:</b>", keyboard), nil
	}

	row, err := strconv.Atoi(subCommands[0])
	if err != nil {
		return NewEditResponse(chatID, messageID, "Invalid row"), nil
	}

	if err := h.expenseService.ClearBudget(ctx, types.ID(userID), types.GroupBudgetCol, row); err != nil {
		return NewEditResponse(chatID, messageID, fmt.Sprintf("<b>Error:</b> %s", err.Error())), nil
	}

	keyboard := tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("⬅️ Back", "expenses:budget:back"),
		),
	)
	return NewEditWithKeyboardResponse(chatID, messageID, "✅ <b>Budget cleared!</b>", keyboard), nil
}

func (h *ExpenseHandler) handleBudgetClearCategory(ctx context.Context, subCommands []string, chatID int64, messageID int, userID int64) (Response, error) {
	if len(subCommands) == 0 {
		var rows [][]tgbotapi.InlineKeyboardButton
		allCategories := types.GetAllCategories()
		for i := 0; i < len(allCategories); i += 2 {
			var row []tgbotapi.InlineKeyboardButton
			c := allCategories[i]
			if r, ok := types.CategoryBudgetRow[c]; ok {
				row = append(row, tgbotapi.NewInlineKeyboardButtonData(
					types.GetCategoryShortName(c),
					fmt.Sprintf("expenses:budget:clearcat:%d", r),
				))
			}
			if i+1 < len(allCategories) {
				c2 := allCategories[i+1]
				if r, ok := types.CategoryBudgetRow[c2]; ok {
					row = append(row, tgbotapi.NewInlineKeyboardButtonData(
						types.GetCategoryShortName(c2),
						fmt.Sprintf("expenses:budget:clearcat:%d", r),
					))
				}
			}
			if len(row) > 0 {
				rows = append(rows, row)
			}
		}
		rows = append(rows, tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("⬅️ Back", "expenses:budget:back"),
		))
		keyboard := tgbotapi.NewInlineKeyboardMarkup(rows...)
		return NewEditWithKeyboardResponse(chatID, messageID, "🧹 <b>Select a category to clear budget:</b>", keyboard), nil
	}

	row, err := strconv.Atoi(subCommands[0])
	if err != nil {
		return NewEditResponse(chatID, messageID, "Invalid row"), nil
	}

	if err := h.expenseService.ClearBudget(ctx, types.ID(userID), types.CategoryBudgetCol, row); err != nil {
		return NewEditResponse(chatID, messageID, fmt.Sprintf("<b>Error:</b> %s", err.Error())), nil
	}

	keyboard := tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("⬅️ Back", "expenses:budget:back"),
		),
	)
	return NewEditWithKeyboardResponse(chatID, messageID, "✅ <b>Budget cleared!</b>", keyboard), nil
}

// HandleBudgetAmountInput handles text input for budget amount when in budget set state.
func (h *ExpenseHandler) HandleBudgetAmountInput(ctx context.Context, msg *tgbotapi.Message) (tgbotapi.MessageConfig, error) {
	reply := tgbotapi.NewMessage(msg.Chat.ID, "")

	pending := h.stateManager.GetBudgetPendingData(msg.Chat.ID)
	if pending == nil {
		reply.Text = "⚠️ No budget selection found. Please try again."
		h.stateManager.End(msg.Chat.ID)
		return reply, nil
	}

	// Parse amount using currency.ParseAmount (supports 500k, 1.5m, 1500000)
	amount := currency.ParseAmount(msg.Text)
	if amount <= 0 {
		reply.Text = "❌ <b>Invalid amount.</b> Please enter a valid amount (e.g., 500k, 1.5m, 1000000)"
		return reply, nil
	}

	// Set budget
	if err := h.expenseService.SetBudget(ctx, types.ID(msg.From.ID), pending.Col, pending.Row, uint64(amount)); err != nil {
		reply.Text = fmt.Sprintf("<b>Error:</b> %s", err.Error())
		h.stateManager.End(msg.Chat.ID)
		h.stateManager.ClearBudgetPendingData(msg.Chat.ID)
		return reply, nil
	}

	h.stateManager.End(msg.Chat.ID)
	h.stateManager.ClearBudgetPendingData(msg.Chat.ID)

	reply.Text = fmt.Sprintf("✅ <b>Budget updated to %s</b>", currency.FormatVND(types.Unsigned(amount)))
	return reply, nil
}

// IsVoiceEnabled returns true if voice expense feature is enabled
func (h *ExpenseHandler) IsVoiceEnabled() bool {
	return h.voiceExpenseService != nil && h.voiceExpenseService.IsEnabled()
}
