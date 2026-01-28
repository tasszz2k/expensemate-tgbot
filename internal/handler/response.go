package handler

import tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"

// Response wraps a telegram response that can be either a new message or an edit
type Response struct {
	// Message is used for sending new messages
	Message *tgbotapi.MessageConfig
	// Edit is used for editing existing messages (callbacks)
	Edit *tgbotapi.EditMessageTextConfig
}

// NewMessageResponse creates a response with a new message
func NewMessageResponse(msg tgbotapi.MessageConfig) Response {
	return Response{Message: &msg}
}

// NewEditResponse creates a response that edits an existing message
func NewEditResponse(chatID int64, messageID int, text string) Response {
	edit := tgbotapi.NewEditMessageText(chatID, messageID, text)
	return Response{Edit: &edit}
}

// NewEditWithKeyboardResponse creates a response that edits an existing message with a keyboard
func NewEditWithKeyboardResponse(chatID int64, messageID int, text string, keyboard tgbotapi.InlineKeyboardMarkup) Response {
	edit := tgbotapi.NewEditMessageTextAndMarkup(chatID, messageID, text, keyboard)
	return Response{Edit: &edit}
}

// IsEdit returns true if this is an edit response
func (r Response) IsEdit() bool {
	return r.Edit != nil
}

// Text returns the text content of the response
func (r Response) Text() string {
	if r.Edit != nil {
		return r.Edit.Text
	}
	if r.Message != nil {
		return r.Message.Text
	}
	return ""
}
