package log

import (
	"os"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/sirupsen/logrus"
)

// Logger is the global logger instance
var Logger *logrus.Logger

func init() {
	Logger = logrus.New()
	Logger.SetOutput(os.Stdout)
	Logger.SetFormatter(&logrus.TextFormatter{
		FullTimestamp: true,
	})
	Logger.SetLevel(logrus.InfoLevel)
}

// SetDebug enables or disables debug logging
func SetDebug(enabled bool) {
	if enabled {
		Logger.SetLevel(logrus.DebugLevel)
	} else {
		Logger.SetLevel(logrus.InfoLevel)
	}
}

// WithMessage creates a logger entry with standard Telegram message context
func WithMessage(msg *tgbotapi.Message) *logrus.Entry {
	if msg == nil {
		return Logger.WithFields(logrus.Fields{})
	}

	fields := logrus.Fields{
		"chat_id":   msg.Chat.ID,
		"chat_type": msg.Chat.Type,
	}

	if msg.From != nil {
		fields["user_id"] = msg.From.ID
		if msg.From.UserName != "" {
			fields["username"] = msg.From.UserName
		}
	}

	return Logger.WithFields(fields)
}

// WithCallback creates a logger entry with callback query context
func WithCallback(cb *tgbotapi.CallbackQuery) *logrus.Entry {
	if cb == nil {
		return Logger.WithFields(logrus.Fields{})
	}

	fields := logrus.Fields{}

	if cb.Message != nil {
		fields["chat_id"] = cb.Message.Chat.ID
		fields["chat_type"] = cb.Message.Chat.Type
	}

	if cb.From != nil {
		fields["user_id"] = cb.From.ID
		if cb.From.UserName != "" {
			fields["username"] = cb.From.UserName
		}
	}

	return Logger.WithFields(fields)
}

// WithAction adds action field to existing entry
func WithAction(entry *logrus.Entry, action string) *logrus.Entry {
	return entry.WithField("action", action)
}

// WithExpense adds expense fields to existing entry
func WithExpense(entry *logrus.Entry, expenseID int64, name string) *logrus.Entry {
	return entry.WithFields(logrus.Fields{
		"expense_id": expenseID,
		"name":       name,
	})
}

// WithSpreadsheet adds spreadsheet field to existing entry
func WithSpreadsheet(entry *logrus.Entry, spreadsheetID string) *logrus.Entry {
	return entry.WithField("spreadsheet_id", spreadsheetID)
}

// Info logs an info message
func Info(msg string, fields ...logrus.Fields) {
	if len(fields) > 0 {
		Logger.WithFields(fields[0]).Info(msg)
	} else {
		Logger.Info(msg)
	}
}

// Error logs an error message
func Error(msg string, err error, fields ...logrus.Fields) {
	entry := Logger.WithError(err)
	if len(fields) > 0 {
		entry = entry.WithFields(fields[0])
	}
	entry.Error(msg)
}

// Debug logs a debug message
func Debug(msg string, fields ...logrus.Fields) {
	if len(fields) > 0 {
		Logger.WithFields(fields[0]).Debug(msg)
	} else {
		Logger.Debug(msg)
	}
}

// Warn logs a warning message
func Warn(msg string, fields ...logrus.Fields) {
	if len(fields) > 0 {
		Logger.WithFields(fields[0]).Warn(msg)
	} else {
		Logger.Warn(msg)
	}
}

// Fields is an alias for logrus.Fields for convenience
type Fields = logrus.Fields

// IsDebugEnabled returns true if debug logging is enabled
func IsDebugEnabled() bool {
	return Logger.GetLevel() >= logrus.DebugLevel
}

// DebugInput logs input message/callback data at debug level
func DebugInput(entry *logrus.Entry, inputType string, content string) {
	if !IsDebugEnabled() {
		return
	}
	entry.WithFields(logrus.Fields{
		"direction":  "input",
		"input_type": inputType,
		"content":    truncate(content, 500),
	}).Debug("received input")
}

// DebugOutput logs output response at debug level
func DebugOutput(entry *logrus.Entry, responseText string) {
	if !IsDebugEnabled() {
		return
	}
	entry.WithFields(logrus.Fields{
		"direction": "output",
		"response":  truncate(responseText, 500),
	}).Debug("sending response")
}

// truncate limits string length for logging
func truncate(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen] + "...[truncated]"
}
