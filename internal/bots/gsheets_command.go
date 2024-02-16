package bots

import (
	"context"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

func (e *Expensemate) handleGSheetsCommand(
	ctx context.Context,
	incomingMessage *tgbotapi.Message,
) (tgbotapi.MessageConfig, error) {
	msg := tgbotapi.NewMessage(incomingMessage.Chat.ID, "")
	msg.Text = "You can configure Google Sheets to manage your expenses."
	return msg, nil
}
