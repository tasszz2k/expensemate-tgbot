package service

import (
	"context"
	"fmt"
	"strings"

	"expensemate-tgbot/internal/log"
	"expensemate-tgbot/internal/repository/openai"
	"expensemate-tgbot/internal/types"
	"expensemate-tgbot/internal/util/currency"
	timepkg "expensemate-tgbot/internal/util/time"

	oailib "github.com/sashabaranov/go-openai"
	"github.com/sirupsen/logrus"
)

const maxConversationTurns = 10

// AskService handles AI-powered natural language queries about expenses
type AskService struct {
	openaiClient   *openai.Client
	expenseService *ExpenseService
}

// NewAskService creates a new AskService
func NewAskService(openaiClient *openai.Client, expenseService *ExpenseService) *AskService {
	return &AskService{
		openaiClient:   openaiClient,
		expenseService: expenseService,
	}
}

// Ask processes a natural language question with conversation history.
// On the first call (empty history), it fetches expense data and builds the system prompt.
// On follow-ups, it reuses the existing system prompt from history.
// Returns the AI answer and the updated conversation history.
func (s *AskService) Ask(ctx context.Context, userID types.ID, question string, history []openai.ChatMessage) (string, []openai.ChatMessage, error) {
	if len(history) == 0 {
		systemPrompt, err := s.buildSystemPrompt(ctx, userID)
		if err != nil {
			log.Warn("failed to build full system prompt, using fallback", logrus.Fields{
				"user_id": userID,
				"error":   err.Error(),
				"action":  "ask_system_prompt",
			})
			systemPrompt = askSystemPromptFallback
		}

		history = []openai.ChatMessage{
			{Role: oailib.ChatMessageRoleSystem, Content: systemPrompt},
		}
	}

	history = append(history, openai.ChatMessage{
		Role:    oailib.ChatMessageRoleUser,
		Content: question,
	})

	// Trim history if it exceeds the max turns (keep system prompt + last N turns)
	history = trimHistory(history, maxConversationTurns)

	answer, err := s.openaiClient.ChatWithHistory(ctx, history)
	if err != nil {
		return "", history, fmt.Errorf("AI chat: %w", err)
	}

	history = append(history, openai.ChatMessage{
		Role:    oailib.ChatMessageRoleAssistant,
		Content: answer,
	})

	log.Info("ask AI completed", logrus.Fields{
		"user_id":      userID,
		"history_len":  len(history),
		"question_len": len(question),
		"answer_len":   len(answer),
		"action":       "ask_ai",
	})

	return answer, history, nil
}

// IsEnabled returns true if the ask service is ready
func (s *AskService) IsEnabled() bool {
	return s.openaiClient != nil
}

// trimHistory keeps the system prompt (index 0) plus the last maxTurns
// user+assistant pairs. Each turn is 2 messages.
func trimHistory(history []openai.ChatMessage, maxTurns int) []openai.ChatMessage {
	if len(history) <= 1 {
		return history
	}

	conversationMessages := history[1:]
	maxMessages := maxTurns * 2
	if len(conversationMessages) <= maxMessages {
		return history
	}

	trimmed := make([]openai.ChatMessage, 0, 1+maxMessages)
	trimmed = append(trimmed, history[0])
	trimmed = append(trimmed, conversationMessages[len(conversationMessages)-maxMessages:]...)
	return trimmed
}

// buildSystemPrompt gathers expense data and assembles the system prompt
func (s *AskService) buildSystemPrompt(ctx context.Context, userID types.ID) (string, error) {
	var sections []string
	sections = append(sections, askSystemPromptBase)

	// Current month budget overview (group + category + summary)
	budgetSection, err := s.buildBudgetSection(ctx, userID)
	if err != nil {
		log.Warn("failed to build budget section for ask prompt", logrus.Fields{
			"user_id": userID,
			"error":   err.Error(),
		})
	} else if budgetSection != "" {
		sections = append(sections, budgetSection)
	}

	// Multi-month insights (3M averages)
	insightsSection, err := s.buildInsightsSection(ctx, userID)
	if err != nil {
		log.Warn("failed to build insights section for ask prompt", logrus.Fields{
			"user_id": userID,
			"error":   err.Error(),
		})
	} else if insightsSection != "" {
		sections = append(sections, insightsSection)
	}

	sections = append(sections, askBotHelpReference)

	today := timepkg.Now().Format("2006-01-02")
	sections = append(sections, fmt.Sprintf("Today's date: %s", today))

	return strings.Join(sections, "\n\n---\n\n"), nil
}

func (s *AskService) buildBudgetSection(ctx context.Context, userID types.ID) (string, error) {
	groups, categories, summary, _, err := s.expenseService.GetBudgetOverview(ctx, userID)
	if err != nil {
		return "", err
	}

	var b strings.Builder
	b.WriteString("CURRENT MONTH EXPENSE DATA:\n\n")

	if len(groups) > 0 {
		b.WriteString("Group Report:\n")
		for _, g := range groups {
			line := fmt.Sprintf("  %s: %s", g.Name, currency.FormatVNDSigned(g.Spent))
			if g.HasBudget {
				line += fmt.Sprintf(" (budget: %s, remaining: %s)",
					currency.FormatVNDSigned(g.Budget),
					currency.FormatVNDSigned(g.Remaining))
			}
			b.WriteString(line + "\n")
		}
	}

	if len(categories) > 0 {
		b.WriteString("\nCategory Report:\n")
		for _, c := range categories {
			if c.Spent == 0 && !c.HasBudget {
				continue
			}
			line := fmt.Sprintf("  %s: %s", c.Name, currency.FormatVNDSigned(c.Spent))
			if c.HasBudget {
				line += fmt.Sprintf(" (budget: %s, remaining: %s)",
					currency.FormatVNDSigned(c.Budget),
					currency.FormatVNDSigned(c.Remaining))
			}
			b.WriteString(line + "\n")
		}
	}

	if len(summary) > 0 {
		b.WriteString("\nSummary:\n")
		for _, s := range summary {
			b.WriteString(fmt.Sprintf("  %s: %s\n", s.Name, currency.FormatVNDSigned(s.Spent)))
		}
	}

	return b.String(), nil
}

func (s *AskService) buildInsightsSection(ctx context.Context, userID types.ID) (string, error) {
	reports, sortedMonths, err := s.expenseService.GetMonthlyReports(ctx, userID, 6)
	if err != nil {
		return "", err
	}

	var b strings.Builder
	b.WriteString(fmt.Sprintf("HISTORICAL DATA (per-month, months: %s):\n", strings.Join(sortedMonths, ", ")))

	for _, month := range sortedMonths {
		report := reports[month]
		b.WriteString(fmt.Sprintf("\n[%s]\n", month))

		if len(report.Groups) > 0 {
			b.WriteString("  Groups:\n")
			for _, g := range report.Groups {
				if g.Spent == 0 {
					continue
				}
				b.WriteString(fmt.Sprintf("    %s: %s\n", g.Name, currency.FormatVNDSigned(g.Spent)))
			}
		}
		if len(report.Categories) > 0 {
			b.WriteString("  Categories:\n")
			for _, c := range report.Categories {
				if c.Spent == 0 {
					continue
				}
				b.WriteString(fmt.Sprintf("    %s: %s\n", c.Name, currency.FormatVNDSigned(c.Spent)))
			}
		}
		if len(report.Summary) > 0 {
			b.WriteString("  Summary:\n")
			for _, sm := range report.Summary {
				b.WriteString(fmt.Sprintf("    %s: %s\n", sm.Name, currency.FormatVNDSigned(sm.Spent)))
			}
		}
	}

	return b.String(), nil
}

const askSystemPromptBase = `You are a personal finance assistant AND bot usage guide for ExpenseMate, a Telegram expense tracking bot. Your job is to:
1. Answer questions about the user's expense data (current month and historical trends)
2. Help users understand how to use the bot's features and commands
3. Provide financial insights and advice based on the data

CRITICAL FORMATTING - READ CAREFULLY:
Your output is rendered as Telegram HTML. You MUST follow these rules strictly:
- Use ONLY Telegram-supported HTML tags: <b>bold</b>, <i>italic</i>, <code>monospace</code>
- NEVER use markdown syntax. No **, no *, no #, no ##, no ` + "`" + `, no [](), no - for lists
- For bullet points use a simple dash character or the dot character, e.g. "- item" or ". item"
- For line breaks just use newlines
- NEVER wrap text in **asterisks** for bold. ALWAYS use <b>text</b> instead
- Keep responses under 4096 characters

CONTENT RULES:
- Answer in the SAME LANGUAGE as the user's question (Vietnamese or English)
- Use compact VND format: 50k, 1.5m, etc.
- Be concise but thorough
- Reference specific numbers from the provided data
- If the data doesn't contain enough info to answer, say so honestly

RANKING/COMPARISON QUESTIONS:
When the user asks "which month was highest/lowest?", "when did I spend the most on X?", or similar ranking questions:
- Show the top 3 months (or all if fewer than 3) with exact amounts
- Include the monthly average for that group/category
- Format as a ranked list, e.g.:
  1. 2025_11: 5,040k
  2. 2025_10: 3,200k
  3. 2026_01: 2,327k
  Avg: 2,650k/month`

const askSystemPromptFallback = askSystemPromptBase + "\n\n" + askBotHelpReference

const askBotHelpReference = `BOT COMMANDS AND FEATURES:

/expenses - Main expense menu with buttons: Add, View, Report, Budget, Update, Delete, Insights, Help
/expenses_add - Add a new expense. Input format (one per line):
  name (required)
  amount (required) - supports "k" (thousand), "m" (million). E.g. 50k = 50,000
  group (optional) - or select via buttons after adding
  category (optional) - or select via buttons after adding
  date (optional, default: today) - format: d/m/yyyy
  note (optional)
/budget - View and set monthly budgets per group or category
/expenses_insights - View average spending across 3M/6M/12M/YTD periods with emergency fund calculation
/ask (or /a, /q) - Ask AI about your expenses or how to use the bot (this feature)
/gsheets - Configure your Google Spreadsheet or create a new monthly sheet
/help - Show command list
/cancel - Cancel current conversation

Voice Input: Send a voice message anytime to add an expense hands-free (Vietnamese supported)

Expense Groups (7): INCOME, INVESTMENT OUT, MUST HAVE, NICE TO HAVE, WASTED, FAMILY, LOVER
Expense Categories (20): Food, Cafe, Groceries, Transport, Entertainment, Miscellaneous, Subscription, Housing, Personal Care, Healthcare, Clothing, Education, Tech, Travel, Present, Life Events, Lover, Family, Lost Money, Unclassified

Budget: Users can set monthly budgets per group and per category. The report shows spent vs budget vs remaining.
Insights: Shows monthly averages across multiple months and calculates emergency fund target (default 3x avg MUST HAVE spending).`
