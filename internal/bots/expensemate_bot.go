package bots

import (
	"log/slog"

	"expensemate-tgbot/pkg/types/tgtypes"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

type Expensemate struct {
	botAPI *tgbotapi.BotAPI
}

type ExpensemateConfig struct {
	BotAPI *tgbotapi.BotAPI
}

type Botter interface {
	Handle(update tgbotapi.Update) error
}

func (e Expensemate) Handle(update tgbotapi.Update) error {
	// Create a new MessageConfig. We don't have text yet,
	// so we leave it empty.
	incomingMessage := update.Message
	msg := tgbotapi.NewMessage(incomingMessage.Chat.ID, "")

	// Extract the command from the Message.
	commandType := tgtypes.Command(incomingMessage.Command())
	switch commandType {
	case tgtypes.CommandStart:
		msg.Text = "Hello! I am Expensemate bot. I can help you manage your expenses."
	case tgtypes.CommandHelp:
		msg.Text = "I can help you manage your expenses. You can use the following commands:\n" +
			"/start - Start the bot\n" +
			"/help - Show this help message\n" +
			"/expenses - Manage your expenses\n" +
			"/expenses_add - Quickly add an expense\n" +
			"/gsheets - Configure Google Sheets\n" +
			"/settings - Configure the bot settings (Admin only)\n" +
			"/feedback - Send feedback to the admin\n"
	case tgtypes.CommandExpenses:
		msg.Text = "You can use the following commands to manage your expenses:\n"
	case tgtypes.CommandExpenseAdd:
		msg.Text = "You can quickly add an expense."
	case tgtypes.CommandGSheets:
		msg.Text = "You can configure Google Sheets to manage your expenses."
	case tgtypes.CommandSettings:
		msg.Text = "You can configure the bot settings."
	case tgtypes.CommandFeedback:
		msg.Text = "You can send feedback to the bot."
	default:
		msg.Text = "Unfortunately, We have not supported this command yet.\n" +
			"Please use /help to see the list of supported commands."
	}

	// Respond the message to the user.
	if _, err := e.botAPI.Send(msg); err != nil {
		slog.Error("Error while sending message: %v", err)
	}

	return nil
}

func NewExpensemate(config ExpensemateConfig) Botter {
	return &Expensemate{
		botAPI: config.BotAPI,
	}

}
