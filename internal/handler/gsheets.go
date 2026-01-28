package handler

import (
	"context"
	"fmt"
	"strings"
	"time"

	"expensemate-tgbot/internal/log"
	"expensemate-tgbot/internal/service"
	"expensemate-tgbot/internal/state"
	"expensemate-tgbot/internal/types"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

// GSheetsHandler handles Google Sheets related commands and callbacks
type GSheetsHandler struct {
	mappingService *service.MappingService
	stateManager   *state.Manager
}

// NewGSheetsHandler creates a new GSheetsHandler
func NewGSheetsHandler(
	mappingService *service.MappingService,
	stateManager *state.Manager,
) *GSheetsHandler {
	return &GSheetsHandler{
		mappingService: mappingService,
		stateManager:   stateManager,
	}
}

// HandleGSheetsCommand handles the /gsheets command
func (h *GSheetsHandler) HandleGSheetsCommand(ctx context.Context, msg *tgbotapi.Message) (tgbotapi.MessageConfig, error) {
	logger := log.WithMessage(msg)
	log.WithAction(logger, "gsheets").Info("user accessed gsheets")

	reply := tgbotapi.NewMessage(msg.Chat.ID, "")

	// Check if user has configured spreadsheet
	spreadsheetURL, err := h.mappingService.GetSpreadsheetURL(ctx, types.ID(msg.From.ID))
	if err != nil || spreadsheetURL == "" {
		reply.Text = "You haven't configured a Google Sheets yet. Click Configure to set one up."
		reply.ReplyMarkup = tgbotapi.NewInlineKeyboardMarkup(
			tgbotapi.NewInlineKeyboardRow(
				tgbotapi.NewInlineKeyboardButtonData("Configure", "gsheets:configure"),
				tgbotapi.NewInlineKeyboardButtonData("Help", "gsheets:help"),
			),
		)
		return reply, nil
	}

	reply.Text = fmt.Sprintf("Your Google Sheets: %s", spreadsheetURL)
	reply.ReplyMarkup = tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("Configure", "gsheets:configure"),
			tgbotapi.NewInlineKeyboardButtonURL("View", spreadsheetURL),
			tgbotapi.NewInlineKeyboardButtonData("Help", "gsheets:help"),
		),
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("Update Active Page", "gsheets:update_current_page"),
		),
	)

	return reply, nil
}

// HandleConfigureCallback handles the configure callback
func (h *GSheetsHandler) HandleConfigureCallback(ctx context.Context, cb *tgbotapi.CallbackQuery) (tgbotapi.MessageConfig, error) {
	logger := log.WithCallback(cb)
	log.WithAction(logger, "gsheets_configure").Info("starting configure flow")

	reply := tgbotapi.NewMessage(cb.Message.Chat.ID, "")
	reply.Text = `Please provide the URL of your Google Sheets.
Example: https://docs.google.com/spreadsheets/d/YOUR_SPREADSHEET_ID/edit`

	h.stateManager.Start(cb.Message.Chat.ID, state.StateGSheetsConfig)

	return reply, nil
}

// HandleConfigure handles the spreadsheet URL input
func (h *GSheetsHandler) HandleConfigure(ctx context.Context, msg *tgbotapi.Message) (tgbotapi.MessageConfig, error) {
	logger := log.WithMessage(msg)
	log.WithAction(logger, "gsheets_configure").Info("processing spreadsheet URL")

	reply := tgbotapi.NewMessage(msg.Chat.ID, "")

	url := strings.TrimSpace(msg.Text)

	// Get user info
	username := ""
	if msg.From.UserName != "" {
		username = "@" + msg.From.UserName
	}
	fullName := msg.From.FirstName
	if msg.From.LastName != "" {
		fullName += " " + msg.From.LastName
	}

	_, err := h.mappingService.Configure(ctx, types.ID(msg.From.ID), username, fullName, url)
	if err != nil {
		reply.Text = fmt.Sprintf("<b>Error:</b> %s", err.Error())
		return reply, nil
	}

	h.stateManager.End(msg.Chat.ID)

	reply.Text = `<b>Google Sheets configured successfully!</b>

<b>Important:</b> Make sure to share <b>Editing access</b> with:
<code>housematee-gsheets@housematee.iam.gserviceaccount.com</code>

Use /expenses to start tracking your expenses.`

	return reply, nil
}

// HandleHelpCallback handles the help callback
func (h *GSheetsHandler) HandleHelpCallback(ctx context.Context, cb *tgbotapi.CallbackQuery) (tgbotapi.MessageConfig, error) {
	logger := log.WithCallback(cb)
	log.WithAction(logger, "gsheets_help").Info("showing gsheets help")

	h.stateManager.End(cb.Message.Chat.ID)

	reply := tgbotapi.NewMessage(cb.Message.Chat.ID, "")
	reply.Text = `<b>Google Sheets Setup:</b>

1. Clone template: <a href="https://docs.google.com/spreadsheets/d/16jOEcyvHiHzW1GdRBvhHEadECojq0g3tzBT3a2MoLnI">Template</a>
2. Use /gsheets and click "Configure"
3. Paste your Google Sheets URL
4. Share <b>Editing access</b> with:
   <code>housematee-gsheets@housematee.iam.gserviceaccount.com</code>

<i>This is a service account - no one else can access your data.</i>`

	return reply, nil
}

// HandleUpdateActivePageCallback handles the update active page callback
func (h *GSheetsHandler) HandleUpdateActivePageCallback(ctx context.Context, cb *tgbotapi.CallbackQuery, subCommands []string) (tgbotapi.MessageConfig, error) {
	logger := log.WithCallback(cb)
	log.WithAction(logger, "gsheets_update_page").Info("updating active page")

	reply := tgbotapi.NewMessage(cb.Message.Chat.ID, "")

	// If a page was selected from the list
	if len(subCommands) > 1 {
		pageName := subCommands[1]
		if err := h.mappingService.UpdateActivePage(ctx, types.ID(cb.From.ID), pageName); err != nil {
			reply.Text = fmt.Sprintf("<b>Error:</b> %s", err.Error())
			return reply, nil
		}
		h.stateManager.End(cb.Message.Chat.ID)
		reply.Text = fmt.Sprintf("Active page updated to: <b>%s</b>", pageName)
		return reply, nil
	}

	// Show available sheets
	sheetNames, err := h.mappingService.GetValidSheetNames(ctx, types.ID(cb.From.ID))
	if err != nil {
		reply.Text = fmt.Sprintf("<b>Error:</b> %s", err.Error())
		return reply, nil
	}

	if len(sheetNames) == 0 {
		// Suggest current month format
		currentMonth := time.Now().Format("2006_01")
		reply.Text = fmt.Sprintf("No sheets found with YYYY_MM format.\nEnter a page name (e.g., %s):", currentMonth)
		h.stateManager.Start(cb.Message.Chat.ID, state.StateGSheetsUpdateActivePage)
		return reply, nil
	}

	// Build keyboard with available sheets
	var rows [][]tgbotapi.InlineKeyboardButton
	for _, name := range sheetNames {
		rows = append(rows, tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData(name, fmt.Sprintf("gsheets:update_current_page:%s", name)),
		))
	}

	reply.Text = "Select the active page:"
	reply.ReplyMarkup = tgbotapi.NewInlineKeyboardMarkup(rows...)

	return reply, nil
}

// HandleUpdateActivePage handles the active page input
func (h *GSheetsHandler) HandleUpdateActivePage(ctx context.Context, msg *tgbotapi.Message) (tgbotapi.MessageConfig, error) {
	logger := log.WithMessage(msg)
	log.WithAction(logger, "gsheets_update_page").Info("processing page name")

	reply := tgbotapi.NewMessage(msg.Chat.ID, "")

	pageName := strings.TrimSpace(msg.Text)

	if err := h.mappingService.UpdateActivePage(ctx, types.ID(msg.From.ID), pageName); err != nil {
		reply.Text = fmt.Sprintf("<b>Error:</b> %s", err.Error())
		return reply, nil
	}

	h.stateManager.End(msg.Chat.ID)
	reply.Text = fmt.Sprintf("Active page updated to: <b>%s</b>", pageName)

	return reply, nil
}

// HandleCallback handles gsheets-related callbacks
func (h *GSheetsHandler) HandleCallback(ctx context.Context, cb *tgbotapi.CallbackQuery, action string, subCommands []string) (tgbotapi.MessageConfig, error) {
	logger := log.WithCallback(cb)
	log.WithAction(logger, "gsheets_callback").WithField("callback", cb.Data).Info("processing callback")

	switch types.GSheetsAction(action) {
	case types.GSheetsActionConfigure:
		return h.HandleConfigureCallback(ctx, cb)
	case types.GSheetsActionHelp:
		return h.HandleHelpCallback(ctx, cb)
	case types.GSheetsActionUpdateActivePage:
		return h.HandleUpdateActivePageCallback(ctx, cb, subCommands)
	default:
		h.stateManager.End(cb.Message.Chat.ID)
		reply := tgbotapi.NewMessage(cb.Message.Chat.ID, "")
		reply.Text = "This action is not yet implemented."
		return reply, nil
	}
}

// EndConversation ends the conversation for a chat
func (h *GSheetsHandler) EndConversation(chatID int64) {
	h.stateManager.End(chatID)
}
