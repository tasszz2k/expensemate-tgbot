package bots

import (
	"context"
	"fmt"
	"log/slog"
	"sync"

	"expensemate-tgbot/pkg/models"
	"expensemate-tgbot/pkg/types/expensetypes"
	"expensemate-tgbot/pkg/types/gsheettypes"
	"expensemate-tgbot/pkg/types/tgtypes"
	"expensemate-tgbot/pkg/types/types"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

const parseModeHTML = "HTML"

type Expensemate struct {
	botAPI *tgbotapi.BotAPI
	// Map to store conversation states for each user
	conversationStates map[int64]string
	csMux              sync.RWMutex
	// Map to store user spreadsheet mappings
	spreadsheetMappings map[types.Id]models.UserSheetMapping
	smMux               sync.RWMutex
}

type ExpensemateConfig struct {
	BotAPI *tgbotapi.BotAPI
}

type BotHandler interface {
	Handle(ctx context.Context, update tgbotapi.Update) error
}

type ConversationStateHandler interface {
	startConversation(ctx context.Context, chatID int64, state string)
	getConversationState(ctx context.Context, chatID int64) string
	endConversation(ctx context.Context, chatID int64)
}

type SpreadsheetMappingHandler interface {
	loadSpreadsheetMappings(ctx context.Context)
	getSpreadsheetMappingByUserID(ctx context.Context, username string) (
		models.UserSheetMapping,
		error,
	)
	upsertSpreadsheetMapping(ctx context.Context, mapping models.UserSheetMapping) error
}

func (e *Expensemate) Handle(ctx context.Context, update tgbotapi.Update) error {
	var (
		err error
		msg tgbotapi.MessageConfig
	)

	// Extract chat ID and text from the message
	incomingMessage := update.Message
	switch {
	case incomingMessage != nil:

		chatID := incomingMessage.Chat.ID
		//text := incomingMessage.Text

		state := e.getConversationState(ctx, chatID)
		if state == "" && !incomingMessage.IsCommand() {
			// if the user is not in a conversation and the message is not a command
			// ignore the message
			return nil
		}

		if incomingMessage.IsCommand() {
			// Extract the command from the Message.
			commandType := tgtypes.Command(incomingMessage.Command())

			switch commandType {
			case tgtypes.CommandStart:
				msg, err = e.handleStartCommand(ctx, incomingMessage)
				e.endConversation(ctx, chatID)
			case tgtypes.CommandHelp:
				msg, err = e.handleHelpCommand(ctx, incomingMessage)
				e.endConversation(ctx, chatID)
			case tgtypes.CommandExpenses:
				msg, err = e.handleExpensesCommand(ctx, incomingMessage)
			case tgtypes.CommandExpenseAdd:
				msg, err = e.handleExpensesAddCommand(ctx, incomingMessage)
			case tgtypes.CommandExpenseHelp:
				msg, err = e.handleExpenseHelpCommand(ctx, incomingMessage)
			case tgtypes.CommandGSheets:
				msg, err = e.handleGSheetsCommand(ctx, incomingMessage)
			case tgtypes.CommandSettings:
				msg, err = e.handleSettingsCommand(ctx, incomingMessage)
				e.endConversation(ctx, chatID)
			case tgtypes.CommandFeedback:
				msg, err = e.handleFeedbackCommand(ctx, incomingMessage)
			default:
				msg, err = e.handleDefaultCommand(ctx, incomingMessage)
				e.endConversation(ctx, chatID)
			}
		} else if state != "" {
			// if the user is in a conversation
			// handle the message based on the conversation state
			switch state {
			case fmt.Sprintf("%s:%s", tgtypes.CommandExpenses, expensetypes.ActionAdd):
				msg, err = e.handleExpensesAdd(ctx, incomingMessage)
			case fmt.Sprintf("%s:%s", tgtypes.CommandGSheets, gsheettypes.ActionConfigure):
				msg, err = e.handleGSheetsConfigure(ctx, incomingMessage)
			default:
				slog.Error("Unsupported conversation state")
				e.endConversation(ctx, chatID)
				return fmt.Errorf("unsupported conversation state")
			}
		}

		if err != nil {
			slog.Error("Error while handling the command: %v", err)
			return err
		}

	case update.CallbackQuery != nil:
		callbackQuery := update.CallbackQuery
		data := callbackQuery.Data

		// callback data format: [command].[action]
		command, _ := tgtypes.ParseCallbackData(data)
		switch command {
		case tgtypes.CommandExpenses:
			msg, err = e.handleExpensesCallback(ctx, callbackQuery)
		case tgtypes.CommandGSheets:
			msg, err = e.handleGSheetsCallback(ctx, callbackQuery)
		default:
			slog.Error("Unsupported callback command")
			return fmt.Errorf("unsupported callback command")
		}

		if err != nil {
			slog.Error("Error while handling the callback: %v", err)
			return err
		}
	default:
		slog.Error("Unsupported update type")
		return fmt.Errorf("unsupported update type")
	}

	// format the message
	msg.ParseMode = parseModeHTML
	if incomingMessage != nil && incomingMessage.MessageID != 0 {
		msg.ReplyToMessageID = incomingMessage.MessageID
	}
	// Respond the message to the user.
	if _, err := e.botAPI.Send(msg); err != nil {
		slog.Error("Error while sending message: %v", err)
		return err
	}

	return nil
}

func NewExpensemate(config ExpensemateConfig) BotHandler {
	bot := &Expensemate{
		botAPI:              config.BotAPI,
		conversationStates:  make(map[int64]string),
		spreadsheetMappings: make(map[types.Id]models.UserSheetMapping),
	}

	// todo: change to scheduled job
	bot.loadSpreadsheetMappings(context.Background())
	return bot
}
