package bots

import (
	"context"
	"fmt"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

func (e *Expensemate) handleStartCommand(
	_ context.Context,
	incomingMessage *tgbotapi.Message,
) (tgbotapi.MessageConfig, error) {
	msg := tgbotapi.NewMessage(incomingMessage.Chat.ID, "")
	fullName := incomingMessage.From.FirstName + " " + incomingMessage.From.LastName
	msg.Text = fmt.Sprintf(
		"Hello %s! I am Expensemate bot. I can help you manage your expenses.\n"+
			"Please use /help to see the list of supported commands.", fullName,
	)
	return msg, nil
}

func (e *Expensemate) handleHelpCommand(
	_ context.Context,
	incomingMessage *tgbotapi.Message,
) (tgbotapi.MessageConfig, error) {
	msg := tgbotapi.NewMessage(incomingMessage.Chat.ID, "")
	msg.Text = "I can help you manage your expenses. You can use the following commands:\n" +
		"/start - Start the bot\n" +
		"/help - Show this help message\n" +
		"/expenses - Manage your expenses\n" +
		"/expenses_add - Quickly add an expense\n" +
		"/gsheets - Configure Google Sheets\n" +
		"/settings - Configure the bot settings (Admin only)\n" +
		"/feedback - Send feedback to the admin\n"
	return msg, nil
}

func (e *Expensemate) handleSettingsCommand(
	_ context.Context,
	incomingMessage *tgbotapi.Message,
) (tgbotapi.MessageConfig, error) {
	msg := tgbotapi.NewMessage(incomingMessage.Chat.ID, "")
	msg.Text = "You can configure the bot settings."
	return msg, nil
}

func (e *Expensemate) handleFeedbackCommand(
	_ context.Context,
	incomingMessage *tgbotapi.Message,
) (tgbotapi.MessageConfig, error) {
	msg := tgbotapi.NewMessage(incomingMessage.Chat.ID, "")
	msg.Text = "You can send feedback to the bot."
	return msg, nil
}

func (e *Expensemate) handleDefaultCommand(
	_ context.Context,
	incomingMessage *tgbotapi.Message,
) (tgbotapi.MessageConfig, error) {
	msg := tgbotapi.NewMessage(incomingMessage.Chat.ID, "")
	msg.Text = "Unfortunately, We have not supported this command yet.\n" +
		"Please use /help to see the list of supported commands."
	return msg, nil
}
