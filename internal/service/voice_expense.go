package service

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"time"

	"expensemate-tgbot/internal/log"
	"expensemate-tgbot/internal/model"
	"expensemate-tgbot/internal/repository/openai"
	"expensemate-tgbot/internal/state"
	"expensemate-tgbot/internal/types"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/sirupsen/logrus"
)

// VoiceExpenseService handles voice expense processing
type VoiceExpenseService struct {
	openaiClient   *openai.Client
	expenseService *ExpenseService
	bot            *tgbotapi.BotAPI
}

// VoiceProcessResult represents the result of voice processing
type VoiceProcessResult struct {
	Success               bool
	Expense               *model.Expense
	NeedsClarification    bool
	ClarificationQuestion string
	TranscribedText       string
	ParsedData            *openai.ParsedExpense
	SpreadsheetURL        string
	Error                 error
}

// NewVoiceExpenseService creates a new VoiceExpenseService
func NewVoiceExpenseService(
	openaiClient *openai.Client,
	expenseService *ExpenseService,
	bot *tgbotapi.BotAPI,
) *VoiceExpenseService {
	return &VoiceExpenseService{
		openaiClient:   openaiClient,
		expenseService: expenseService,
		bot:            bot,
	}
}

// ProcessVoiceMessage processes a voice message and returns the result
func (s *VoiceExpenseService) ProcessVoiceMessage(ctx context.Context, userID types.ID, voice *tgbotapi.Voice) (*VoiceProcessResult, error) {
	log.Info("processing voice message", logrus.Fields{
		"user_id":   userID,
		"file_id":   voice.FileID,
		"duration":  voice.Duration,
		"file_size": voice.FileSize,
		"action":    "voice_expense_start",
	})

	// Download voice file from Telegram
	audioData, err := s.downloadVoiceFile(voice.FileID)
	if err != nil {
		return nil, fmt.Errorf("downloading voice file: %w", err)
	}

	// Transcribe audio using Whisper
	transcribedText, err := s.openaiClient.TranscribeAudio(ctx, audioData, "ogg")
	if err != nil {
		return nil, fmt.Errorf("transcribing audio: %w", err)
	}

	log.Info("voice transcribed", logrus.Fields{
		"user_id":          userID,
		"transcribed_text": transcribedText,
		"action":           "voice_transcribe",
	})

	// Parse transcribed text using ChatGPT
	parsed, err := s.openaiClient.ParseExpenseText(ctx, transcribedText)
	if err != nil {
		return nil, fmt.Errorf("parsing expense text: %w", err)
	}

	log.Info("expense parsed", logrus.Fields{
		"user_id":  userID,
		"parsed":   parsed,
		"is_clear": parsed.IsClear,
		"action":   "voice_parse",
	})

	// Get spreadsheet URL for response
	spreadsheetURL, _ := s.expenseService.GetSpreadsheetURL(ctx, userID)

	result := &VoiceProcessResult{
		TranscribedText: transcribedText,
		ParsedData:      parsed,
		SpreadsheetURL:  spreadsheetURL,
	}

	// Check if clarification is needed
	if !parsed.IsClear {
		result.NeedsClarification = true
		if parsed.ClarificationQuestion != nil {
			result.ClarificationQuestion = *parsed.ClarificationQuestion
		} else {
			result.ClarificationQuestion = "Could you please provide more details about this expense?"
		}
		return result, nil
	}

	// Create expense from parsed data
	expense, err := s.createExpenseFromParsed(ctx, userID, parsed)
	if err != nil {
		result.Error = err
		return result, nil
	}

	result.Success = true
	result.Expense = expense
	return result, nil
}

// ProcessClarification processes a clarification response from the user
func (s *VoiceExpenseService) ProcessClarification(ctx context.Context, userID types.ID, clarificationText string, pendingData *state.VoicePendingData) (*VoiceProcessResult, error) {
	log.Info("processing voice clarification", logrus.Fields{
		"user_id":       userID,
		"original_text": pendingData.OriginalText,
		"clarification": clarificationText,
		"action":        "voice_clarification",
	})

	// Re-parse with additional context
	parsed, err := s.openaiClient.ParseExpenseWithContext(ctx, pendingData.OriginalText, clarificationText)
	if err != nil {
		return nil, fmt.Errorf("parsing with clarification: %w", err)
	}

	// Get spreadsheet URL
	spreadsheetURL, _ := s.expenseService.GetSpreadsheetURL(ctx, userID)

	result := &VoiceProcessResult{
		TranscribedText: pendingData.OriginalText,
		ParsedData:      parsed,
		SpreadsheetURL:  spreadsheetURL,
	}

	// Check if still needs clarification
	if !parsed.IsClear {
		result.NeedsClarification = true
		if parsed.ClarificationQuestion != nil {
			result.ClarificationQuestion = *parsed.ClarificationQuestion
		} else {
			result.ClarificationQuestion = "Could you please provide more details?"
		}
		return result, nil
	}

	// Create expense
	expense, err := s.createExpenseFromParsed(ctx, userID, parsed)
	if err != nil {
		result.Error = err
		return result, nil
	}

	result.Success = true
	result.Expense = expense
	return result, nil
}

// createExpenseFromParsed creates an expense from parsed data
func (s *VoiceExpenseService) createExpenseFromParsed(ctx context.Context, userID types.ID, parsed *openai.ParsedExpense) (*model.Expense, error) {
	name, amount, group, category, date, note, err := parsed.ToExpenseFields()
	if err != nil {
		return nil, err
	}

	expense := &model.Expense{
		Name:     name,
		Amount:   amount,
		Group:    group,
		Category: category,
		Date:     date,
		Note:     note,
	}

	return s.expenseService.AddFromModel(ctx, userID, expense)
}

// downloadVoiceFile downloads a voice file from Telegram
func (s *VoiceExpenseService) downloadVoiceFile(fileID string) ([]byte, error) {
	// Get file info from Telegram
	file, err := s.bot.GetFile(tgbotapi.FileConfig{FileID: fileID})
	if err != nil {
		return nil, fmt.Errorf("getting file info: %w", err)
	}

	// Download file
	fileURL := file.Link(s.bot.Token)

	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Get(fileURL)
	if err != nil {
		return nil, fmt.Errorf("downloading file: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("reading file data: %w", err)
	}

	return data, nil
}

// IsEnabled returns true if voice expense feature is enabled (OpenAI client configured)
func (s *VoiceExpenseService) IsEnabled() bool {
	return s.openaiClient != nil
}
