package openai

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"time"

	"expensemate-tgbot/internal/config"
	"expensemate-tgbot/internal/types"

	"github.com/sashabaranov/go-openai"
)

// Client wraps the OpenAI API for Whisper and ChatGPT
type Client struct {
	client       *openai.Client
	whisperModel string
	chatModel    string
}

// ParsedExpense represents the structured output from ChatGPT
type ParsedExpense struct {
	Name                  string  `json:"name"`
	Amount                uint64  `json:"amount"`
	Group                 string  `json:"group"`
	Category              string  `json:"category"`
	Date                  string  `json:"date"`
	Note                  string  `json:"note"`
	IsClear               bool    `json:"is_clear"`
	ClarificationQuestion *string `json:"clarification_question"`
}

// NewClient creates a new OpenAI client
func NewClient(cfg *config.Config) (*Client, error) {
	if cfg.OpenAI.APIKey == "" {
		return nil, fmt.Errorf("OpenAI API key is required")
	}

	client := openai.NewClient(cfg.OpenAI.APIKey)

	return &Client{
		client:       client,
		whisperModel: cfg.OpenAI.GetWhisperModel(),
		chatModel:    cfg.OpenAI.GetChatModel(),
	}, nil
}

// TranscribeAudio transcribes audio data using Whisper API
func (c *Client) TranscribeAudio(ctx context.Context, audioData []byte, format string) (string, error) {
	// Create audio request with reader
	req := openai.AudioRequest{
		Model:    c.whisperModel,
		Reader:   bytes.NewReader(audioData),
		FilePath: fmt.Sprintf("audio.%s", format),
	}

	resp, err := c.client.CreateTranscription(ctx, req)
	if err != nil {
		return "", fmt.Errorf("transcribing audio: %w", err)
	}

	return resp.Text, nil
}

// systemPrompt is the prompt for ChatGPT to parse expense text
const systemPrompt = `You are an expense parser. Parse the user's voice transcription into expense data.

Available Groups (use the alias in lowercase for output):
- INCOME (alias: i) - Thu nhap
- INVESTMENT (alias: inv) - Dau tu
- MUST HAVE (alias: mh) - Thiet yeu (DEFAULT if not specified)
- NICE TO HAVE (alias: nth) - Nen chi
- WASTE (alias: w) - Lang phi
- FAMILY (alias: fam) - Gia dinh
- LOVER (alias: lov) - Nguoi yeu

Available Categories (use the alias in lowercase for output):
- Unclassified (alias: uc) - Chua phan loai (DEFAULT if not specified)
- Food (alias: f) - An ngoai, an trua, an toi, com, bun, pho, cafe, tra sua
- Groceries (alias: gr) - Di cho, sieu thi
- Transport (alias: tr) - Di lai, Grab, xe om, taxi, xang, giu xe
- Entertainment (alias: ent) - Giai tri, xem phim, game
- Miscellaneous (alias: mis) - Linh tinh
- Subscription (alias: sub) - Dang ky, Spotify, Netflix, YouTube
- Housing (alias: hou) - Nha o, tien nha, dien, nuoc
- Personal Care (alias: pc) - Cham soc, cat toc, my pham
- Healthcare (alias: hc) - Suc khoe, thuoc, kham benh
- Clothing (alias: clo) - Quan ao, giay dep
- Education (alias: edu) - Giao duc, hoc, sach
- Tech (alias: tech) - Cong nghe, dien thoai, laptop
- Travel (alias: tv) - Du lich
- Present (alias: pre) - Qua tang
- Life Events (alias: le) - Hieu hi, dam cuoi, tang le
- Lover (alias: lov) - Nguoi yeu
- Family (alias: fam) - Gia dinh
- Lost Money (alias: lm) - Mat tien

IMPORTANT RULES:
1. Extract the expense name - clean it up but keep the meaning
2. Extract the amount - support "k" (thousand) and "m" (million) suffixes. Examples: "50k" = 50000, "1.5m" = 1500000
3. Infer the group if mentioned (mh, must have, thiet yeu, etc.) - DEFAULT to "mh" if not clear
4. Infer the category based on context (f for food/dining, tr for transport/Grab, etc.) - DEFAULT to "uc" if unclear
5. Date defaults to today if not mentioned
6. If the amount is missing or unclear, set is_clear=false and ask for it

Return ONLY valid JSON (no markdown, no explanation):
{
  "name": "expense name",
  "amount": 150000,
  "group": "mh",
  "category": "f",
  "date": "YYYY-MM-DD",
  "note": "",
  "is_clear": true,
  "clarification_question": null
}

If amount or critical info is missing, set:
{
  "is_clear": false,
  "clarification_question": "What was the amount?" 
}

Examples:
- "Lunch meeting 150k must have food" -> name: "Lunch meeting", amount: 150000, group: "mh", category: "f", is_clear: true
- "Grab di Vincom" -> name: "Grab di Vincom", group: "mh", category: "tr", is_clear: false, clarification_question: "How much was the Grab ride?"
- "Mua Spotify subscription 59k" -> name: "Spotify subscription", amount: 59000, group: "mh", category: "sub", is_clear: true
- "An trua 50k" -> name: "An trua", amount: 50000, group: "mh", category: "f", is_clear: true`

// ParseExpenseText parses natural language text into expense structure using ChatGPT
func (c *Client) ParseExpenseText(ctx context.Context, text string) (*ParsedExpense, error) {
	resp, err := c.client.CreateChatCompletion(ctx, openai.ChatCompletionRequest{
		Model: c.chatModel,
		Messages: []openai.ChatCompletionMessage{
			{
				Role:    openai.ChatMessageRoleSystem,
				Content: systemPrompt,
			},
			{
				Role:    openai.ChatMessageRoleUser,
				Content: text,
			},
		},
		Temperature: 0.1, // Low temperature for consistent parsing
	})
	if err != nil {
		return nil, fmt.Errorf("calling ChatGPT: %w", err)
	}

	if len(resp.Choices) == 0 {
		return nil, fmt.Errorf("no response from ChatGPT")
	}

	content := resp.Choices[0].Message.Content

	var parsed ParsedExpense
	if err := json.Unmarshal([]byte(content), &parsed); err != nil {
		return nil, fmt.Errorf("parsing ChatGPT response: %w (response: %s)", err, content)
	}

	// Set default date if empty
	if parsed.Date == "" {
		parsed.Date = time.Now().Format("2006-01-02")
	}

	return &parsed, nil
}

// ParseExpenseWithContext re-parses with additional context from user clarification
func (c *Client) ParseExpenseWithContext(ctx context.Context, originalText, clarification string) (*ParsedExpense, error) {
	combinedText := fmt.Sprintf("Original: %s\nClarification: %s", originalText, clarification)
	return c.ParseExpenseText(ctx, combinedText)
}

// ToExpenseFields converts ParsedExpense to validated expense fields
func (p *ParsedExpense) ToExpenseFields() (name string, amount types.Unsigned, group types.Group, category types.Category, date time.Time, note string, err error) {
	name = p.Name
	if name == "" {
		err = fmt.Errorf("expense name is required")
		return
	}

	amount = types.Unsigned(p.Amount)
	if amount == 0 {
		err = fmt.Errorf("expense amount is required")
		return
	}

	// Parse group
	if p.Group != "" {
		var ok bool
		group, ok = types.GetGroupByAlias(p.Group)
		if !ok {
			group = types.GroupMustHave
		}
	} else {
		group = types.GroupMustHave
	}

	// Parse category
	if p.Category != "" {
		var ok bool
		category, ok = types.GetCategoryByAlias(p.Category)
		if !ok {
			category = types.CategoryUnclassified
		}
	} else {
		category = types.CategoryUnclassified
	}

	// Parse date
	if p.Date != "" {
		date, err = time.Parse("2006-01-02", p.Date)
		if err != nil {
			date = time.Now()
			err = nil
		}
	} else {
		date = time.Now()
	}

	note = p.Note
	return
}
