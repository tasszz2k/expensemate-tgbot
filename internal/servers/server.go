package servers

import (
	"log"
	"log/slog"

	"expensemate-tgbot/internal/bots"
	"expensemate-tgbot/pkg/configs"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

type server struct {
	botAPI *tgbotapi.BotAPI
	bot    bots.Botter
}

type ServerConfig struct {
	AppConf configs.AppConfig
}

type Server interface {
	Start() error
}

func NewServer(config ServerConfig) Server {
	// init Telegram Bot API
	botAPI, err := tgbotapi.NewBotAPI(config.AppConf.TelegramBot.ApiToken)
	if err != nil {
		log.Panic(err)
	}

	botAPI.Debug = true
	log.Printf("Authorized on account %s", botAPI.Self.UserName)

	// init bot
	bot := bots.NewExpensemate(bots.ExpensemateConfig{BotAPI: botAPI})

	return &server{
		botAPI: botAPI,
		bot:    bot,
	}
}

func (b server) Start() error {
	// run bot
	b.run()
	return nil
}

func (b server) run() {
	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60
	updates := b.botAPI.GetUpdatesChan(u)

	for update := range updates {
		if update.Message == nil { // ignore any non-Message updates
			continue
		}

		if !update.Message.IsCommand() { // ignore any non-command Messages
			continue
		}

		err := b.bot.Handle(update)

		if err != nil {
			slog.Error("error handling update: ", err)
		}
	}
}
