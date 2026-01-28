package handler

import (
	"context"
	"fmt"

	"expensemate-tgbot/internal/log"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

// StartHandler handles start, help, and other basic commands
type StartHandler struct{}

// NewStartHandler creates a new StartHandler
func NewStartHandler() *StartHandler {
	return &StartHandler{}
}

// HandleStart handles the /start command
func (h *StartHandler) HandleStart(ctx context.Context, msg *tgbotapi.Message) (tgbotapi.MessageConfig, error) {
	logger := log.WithMessage(msg)
	log.WithAction(logger, "start").Info("user started bot")

	fullName := msg.From.FirstName
	if msg.From.LastName != "" {
		fullName += " " + msg.From.LastName
	}

	reply := tgbotapi.NewMessage(msg.Chat.ID, "")
	reply.Text = fmt.Sprintf(
		"Hello %s! I am Expensemate bot. I can help you manage your expenses.\n"+
			"Please use /help to see the list of supported commands.", fullName,
	)
	return reply, nil
}

// HandleHelp handles the /help command
func (h *StartHandler) HandleHelp(ctx context.Context, msg *tgbotapi.Message) (tgbotapi.MessageConfig, error) {
	logger := log.WithMessage(msg)
	log.WithAction(logger, "help").Info("user requested help")

	reply := tgbotapi.NewMessage(msg.Chat.ID, "")
	reply.Text = `I can help you manage your expenses. You can use the following commands:

/start - Start the bot
/help - Show this help message
/expenses - Manage your expenses
/expenses_add - Quickly add an expense
/expenses_help - View supported groups and categories
/gsheets - Configure Google Sheets
/settings - Configure the bot settings (Admin only)
/feedback - Send feedback to the admin`
	return reply, nil
}

// HandleSettings handles the /settings command
func (h *StartHandler) HandleSettings(ctx context.Context, msg *tgbotapi.Message) (tgbotapi.MessageConfig, error) {
	logger := log.WithMessage(msg)
	log.WithAction(logger, "settings").Info("user accessed settings")

	reply := tgbotapi.NewMessage(msg.Chat.ID, "")
	reply.Text = "Settings are not yet implemented."
	return reply, nil
}

// HandleFeedback handles the /feedback command
func (h *StartHandler) HandleFeedback(ctx context.Context, msg *tgbotapi.Message) (tgbotapi.MessageConfig, error) {
	logger := log.WithMessage(msg)
	log.WithAction(logger, "feedback").Info("user accessed feedback")

	reply := tgbotapi.NewMessage(msg.Chat.ID, "")
	reply.Text = "Feedback feature is not yet implemented."
	return reply, nil
}

// HandleUnknown handles unknown commands
func (h *StartHandler) HandleUnknown(ctx context.Context, msg *tgbotapi.Message) (tgbotapi.MessageConfig, error) {
	reply := tgbotapi.NewMessage(msg.Chat.ID, "")
	reply.Text = "Unknown command. Please use /help to see the list of supported commands."
	return reply, nil
}
