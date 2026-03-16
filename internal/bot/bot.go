package bot

import (
	"context"
	"fmt"
	"strings"

	"expensemate-tgbot/internal/handler"
	"expensemate-tgbot/internal/log"
	"expensemate-tgbot/internal/state"
	"expensemate-tgbot/internal/types"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/sirupsen/logrus"
)

const parseModeHTML = "HTML"

// Bot handles Telegram updates
type Bot struct {
	api            *tgbotapi.BotAPI
	stateManager   *state.Manager
	startHandler   *handler.StartHandler
	expenseHandler *handler.ExpenseHandler
	gsheetsHandler *handler.GSheetsHandler
}

// Config holds bot configuration
type Config struct {
	API            *tgbotapi.BotAPI
	StateManager   *state.Manager
	StartHandler   *handler.StartHandler
	ExpenseHandler *handler.ExpenseHandler
	GSheetsHandler *handler.GSheetsHandler
}

// New creates a new Bot
func New(cfg Config) *Bot {
	return &Bot{
		api:            cfg.API,
		stateManager:   cfg.StateManager,
		startHandler:   cfg.StartHandler,
		expenseHandler: cfg.ExpenseHandler,
		gsheetsHandler: cfg.GSheetsHandler,
	}
}

// Handle processes a Telegram update
func (b *Bot) Handle(ctx context.Context, update tgbotapi.Update) error {
	var (
		resp   handler.Response
		err    error
		logger *logrus.Entry
	)

	switch {
	case update.Message != nil:
		logger = log.WithMessage(update.Message)
		log.DebugInput(logger, "message", update.Message.Text)
		var msg tgbotapi.MessageConfig
		msg, err = b.handleMessage(ctx, update.Message)
		resp = handler.NewMessageResponse(msg)
	case update.CallbackQuery != nil:
		logger = log.WithCallback(update.CallbackQuery)
		log.DebugInput(logger, "callback", update.CallbackQuery.Data)
		resp, err = b.handleCallback(ctx, update.CallbackQuery)
	default:
		return nil
	}

	if err != nil {
		log.Error("error handling update", err, logrus.Fields{})
		return err
	}

	// Log output in debug mode
	if logger != nil {
		log.DebugOutput(logger, resp.Text())
	}

	// Send response
	if resp.IsEdit() {
		resp.Edit.ParseMode = parseModeHTML
		if _, err := b.api.Send(resp.Edit); err != nil {
			log.Error("error editing message", err, logrus.Fields{})
			return err
		}
	} else if resp.Message != nil && resp.Message.Text != "" {
		resp.Message.ParseMode = parseModeHTML
		if update.Message != nil && update.Message.MessageID != 0 {
			resp.Message.ReplyToMessageID = update.Message.MessageID
		}
		if _, err := b.api.Send(resp.Message); err != nil {
			log.Error("error sending message", err, logrus.Fields{})
			return err
		}
	}

	return nil
}

// handleMessage processes incoming messages
func (b *Bot) handleMessage(ctx context.Context, msg *tgbotapi.Message) (tgbotapi.MessageConfig, error) {
	chatID := msg.Chat.ID

	// Check conversation state first
	currentState := b.stateManager.Get(chatID)

	// Handle voice messages - works in any context if voice is enabled
	if msg.Voice != nil {
		return b.handleVoiceMessage(ctx, msg, currentState)
	}

	if msg.IsCommand() {
		return b.handleCommand(ctx, msg)
	}

	// Handle conversation state
	if currentState != "" {
		return b.handleConversation(ctx, msg, currentState)
	}

	// Ignore non-command messages when not in conversation
	return tgbotapi.NewMessage(chatID, ""), nil
}

// handleVoiceMessage processes voice messages for expense input
func (b *Bot) handleVoiceMessage(ctx context.Context, msg *tgbotapi.Message, currentState string) (tgbotapi.MessageConfig, error) {
	logger := log.WithMessage(msg)
	log.WithAction(logger, "voice_message").
		WithField("duration", msg.Voice.Duration).
		WithField("file_size", msg.Voice.FileSize).
		Info("processing voice message")

	// Check if voice is enabled
	if !b.expenseHandler.IsVoiceEnabled() {
		return tgbotapi.NewMessage(msg.Chat.ID, "Voice input is not available. Please use text input with /expenses_add command."), nil
	}

	// Process voice message for expense
	return b.expenseHandler.HandleVoiceExpense(ctx, msg)
}

// handleCommand processes command messages
func (b *Bot) handleCommand(ctx context.Context, msg *tgbotapi.Message) (tgbotapi.MessageConfig, error) {
	command := types.Command(msg.Command())
	chatID := msg.Chat.ID

	logger := log.WithMessage(msg)
	log.WithAction(logger, string(command)).Info("processing command")

	switch command {
	case types.CommandStart:
		b.stateManager.End(chatID)
		return b.startHandler.HandleStart(ctx, msg)

	case types.CommandHelp:
		b.stateManager.End(chatID)
		return b.startHandler.HandleHelp(ctx, msg)

	case types.CommandExpenses:
		return b.expenseHandler.HandleExpensesCommand(ctx, msg)

	case types.CommandExpenseAdd:
		return b.expenseHandler.HandleExpensesAddCommand(ctx, msg)

	case types.CommandExpenseHelp:
		b.stateManager.End(chatID)
		return b.expenseHandler.HandleExpensesHelp(ctx, msg)

	case types.CommandBudget:
		return b.expenseHandler.HandleBudgetCommand(ctx, msg)

	case types.CommandExpenseInsights:
		return b.expenseHandler.HandleExpensesInsightsCommand(ctx, msg)

	case types.CommandAsk, types.CommandAskShortA, types.CommandAskShortQ:
		return b.expenseHandler.HandleAskCommand(ctx, msg)

	case types.CommandGSheets:
		return b.gsheetsHandler.HandleGSheetsCommand(ctx, msg)

	case types.CommandSettings:
		b.stateManager.End(chatID)
		return b.startHandler.HandleSettings(ctx, msg)

	case types.CommandFeedback:
		b.stateManager.End(chatID)
		return b.startHandler.HandleFeedback(ctx, msg)

	default:
		b.stateManager.End(chatID)
		return b.startHandler.HandleUnknown(ctx, msg)
	}
}

// handleConversation processes messages in a conversation state
func (b *Bot) handleConversation(ctx context.Context, msg *tgbotapi.Message, currentState string) (tgbotapi.MessageConfig, error) {
	logger := log.WithMessage(msg)
	log.WithAction(logger, currentState).Info("processing conversation")

	switch currentState {
	case state.StateExpensesAdd:
		return b.expenseHandler.HandleExpensesAdd(ctx, msg)

	case state.StateExpensesVoiceClarify:
		return b.expenseHandler.HandleVoiceClarification(ctx, msg)

	case state.StateGSheetsConfig:
		return b.gsheetsHandler.HandleConfigure(ctx, msg)

	case state.StateGSheetsUpdateActivePage:
		return b.gsheetsHandler.HandleUpdateActivePage(ctx, msg)

	case state.StateInsightsCustomEF:
		return b.expenseHandler.HandleInsightsCustomEF(ctx, msg)

	case state.StateAskConversation:
		return b.expenseHandler.HandleAskFollowUp(ctx, msg)

	default:
		// Check for budget states (prefix-based)
		if strings.HasPrefix(currentState, state.StateBudgetSetGroupPrefix) ||
			strings.HasPrefix(currentState, state.StateBudgetSetCategoryPrefix) {
			return b.expenseHandler.HandleBudgetAmountInput(ctx, msg)
		}
		b.stateManager.End(msg.Chat.ID)
		return tgbotapi.NewMessage(msg.Chat.ID, "Unknown conversation state."), fmt.Errorf("unknown state: %s", currentState)
	}
}

// handleCallback processes callback queries
func (b *Bot) handleCallback(ctx context.Context, cb *tgbotapi.CallbackQuery) (handler.Response, error) {
	command, subCommands := types.ParseCallbackData(cb.Data)

	logger := log.WithCallback(cb)
	log.WithAction(logger, string(command)).WithField("callback", cb.Data).Info("processing callback")

	// Acknowledge callback
	callback := tgbotapi.NewCallback(cb.ID, "")
	b.api.Request(callback)

	if len(subCommands) == 0 {
		msg := tgbotapi.NewMessage(cb.Message.Chat.ID, "Invalid callback data")
		return handler.NewMessageResponse(msg), nil
	}

	action := subCommands[0]

	switch command {
	case types.CommandExpenses:
		return b.expenseHandler.HandleCallback(ctx, cb, action, subCommands)

	case types.CommandGSheets:
		msg, err := b.gsheetsHandler.HandleCallback(ctx, cb, action, subCommands)
		return handler.NewMessageResponse(msg), err

	default:
		msg := tgbotapi.NewMessage(cb.Message.Chat.ID, "Unknown callback command")
		return handler.NewMessageResponse(msg), nil
	}
}

// GetAPI returns the bot API (for server use)
func (b *Bot) GetAPI() *tgbotapi.BotAPI {
	return b.api
}
