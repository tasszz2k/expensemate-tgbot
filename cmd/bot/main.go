package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"

	"expensemate-tgbot/internal/bot"
	"expensemate-tgbot/internal/config"
	"expensemate-tgbot/internal/handler"
	"expensemate-tgbot/internal/log"
	"expensemate-tgbot/internal/repository/openai"
	"expensemate-tgbot/internal/repository/sheets"
	"expensemate-tgbot/internal/service"
	"expensemate-tgbot/internal/state"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/sirupsen/logrus"
)

func main() {
	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		log.Logger.Fatalf("failed to load config: %v", err)
	}

	// Set app debug mode (input/output logging)
	log.SetDebug(cfg.TelegramBot.Debug)

	// Initialize Telegram Bot API
	botAPI, err := tgbotapi.NewBotAPI(cfg.TelegramBot.APIToken)
	if err != nil {
		log.Logger.Fatalf("failed to create bot API: %v", err)
	}
	// Bot API debug (endpoint requests/responses) - separate from app debug
	botAPI.Debug = cfg.TelegramBot.BotDebug

	log.Info("bot initialized", logrus.Fields{
		"bot_id":       botAPI.Self.ID,
		"bot_username": botAPI.Self.UserName,
	})

	// Initialize Google Sheets client
	sheetsClient, err := sheets.NewClient(cfg)
	if err != nil {
		log.Logger.Fatalf("failed to create sheets client: %v", err)
	}

	// Initialize repositories
	expenseRepo := sheets.NewExpenseRepository(sheetsClient)
	mappingRepo := sheets.NewMappingRepository(sheetsClient, cfg.GoogleSheets.DatabaseSpreadsheetID)

	// Initialize state manager
	stateManager := state.NewManager()

	// Initialize services
	mappingService := service.NewMappingService(mappingRepo)
	expenseService := service.NewExpenseService(expenseRepo, mappingService)

	// Load mapping cache
	ctx := context.Background()
	if err := mappingService.LoadCache(ctx); err != nil {
		log.Warn("failed to load mappings cache", logrus.Fields{"error": err.Error()})
	}

	// Initialize handlers
	startHandler := handler.NewStartHandler()
	expenseHandler := handler.NewExpenseHandler(expenseService, mappingService, stateManager)
	gsheetsHandler := handler.NewGSheetsHandler(mappingService, stateManager)

	// Initialize OpenAI client for voice and AI features (optional)
	if cfg.OpenAI.APIKey != "" {
		openaiClient, err := openai.NewClient(cfg)
		if err != nil {
			log.Warn("failed to create OpenAI client, AI features disabled", logrus.Fields{"error": err.Error()})
		} else {
			voiceExpenseService := service.NewVoiceExpenseService(openaiClient, expenseService, botAPI)
			expenseHandler.SetVoiceExpenseService(voiceExpenseService)
			log.Info("voice expense feature enabled", logrus.Fields{
				"whisper_model": cfg.OpenAI.GetWhisperModel(),
				"chat_model":    cfg.OpenAI.GetChatModel(),
			})

			askService := service.NewAskService(openaiClient, expenseService)
			expenseHandler.SetAskService(askService)
			log.Info("ask AI feature enabled", logrus.Fields{
				"chat_model": cfg.OpenAI.GetChatModel(),
			})
		}
	} else {
		log.Info("OpenAI features disabled (no API key configured)", logrus.Fields{})
	}

	// Initialize bot
	expensemateBot := bot.New(bot.Config{
		API:            botAPI,
		StateManager:   stateManager,
		StartHandler:   startHandler,
		ExpenseHandler: expenseHandler,
		GSheetsHandler: gsheetsHandler,
	})

	log.Info("bot has been started", logrus.Fields{
		"bot_id":       botAPI.Self.ID,
		"bot_username": botAPI.Self.UserName,
	})

	// Start update loop
	go runUpdateLoop(ctx, expensemateBot, cfg.TelegramBot.Timeout)

	// Wait for shutdown signal
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Info("shutting down bot...", logrus.Fields{})
}

func runUpdateLoop(ctx context.Context, b *bot.Bot, timeout int) {
	u := tgbotapi.NewUpdate(0)
	u.Timeout = timeout

	updates := b.GetAPI().GetUpdatesChan(u)

	for update := range updates {
		go func(update tgbotapi.Update) {
			if err := b.Handle(ctx, update); err != nil {
				log.Error("error handling update", err, logrus.Fields{})
			}
		}(update)
	}
}
