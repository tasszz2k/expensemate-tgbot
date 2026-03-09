package openai

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"time"

	"expensemate-tgbot/internal/config"
	"expensemate-tgbot/internal/types"
	timepkg "expensemate-tgbot/internal/util/time"

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

// whisperPrompt provides vocabulary hints for Vietnamese expense transcription
const whisperPrompt = `bún bò, bún đậu, cơm tấm, phở, sủi cảo, chả cá Lã Vọng, bánh mì, mì hoành, trà sữa, cà phê, KFC,
xăng xe, đổ xăng, Grab, grab sân bay, xe khách, taxi, bus,
ốp điện thoại, ốp airpod, sạc điện thoại, kính cường lực, giá đỡ laptop, sửa kính camera,
lì xì, biếu, đám cưới, sinh nhật,
vesaf, ssisca, YouTube Premium,
tiền trọ, đồ dọn nhà, thùng rác, túi rác, ổ cắm điện,
dầu gội, dầu xả, sữa rửa mặt, xịt khử mùi, dao cạo râu,
giày, dép, quần, khăn tắm, bình nước,
học piano, vé concert, vé máy bay,
nghìn, triệu, ngàn, đồng`

// TranscribeAudio transcribes audio data using Whisper API
func (c *Client) TranscribeAudio(ctx context.Context, audioData []byte, format string) (string, error) {
	// Create audio request with reader
	// Language: "vi" forces Vietnamese detection to prevent Thai/Chinese/other language output
	// Prompt: provides common Vietnamese expense vocabulary to improve transcription accuracy
	req := openai.AudioRequest{
		Model:    c.whisperModel,
		Reader:   bytes.NewReader(audioData),
		FilePath: fmt.Sprintf("audio.%s", format),
		Language: "vi",
		Prompt:   whisperPrompt,
	}

	resp, err := c.client.CreateTranscription(ctx, req)
	if err != nil {
		return "", fmt.Errorf("transcribing audio: %w", err)
	}

	return resp.Text, nil
}

// systemPrompt is the prompt for ChatGPT to parse expense text
const systemPrompt = `You are an expense parser for Vietnamese voice input. Speech-to-text often produces SEVERELY INCORRECT transcriptions due to Vietnamese tones and homophones. Your job is to:
1. INTERPRET the garbled transcription to understand what the user likely said
2. Correct it to proper Vietnamese
3. Parse it into expense data

Think about what makes sense in an EXPENSE CONTEXT (food, transport, shopping, bills, etc.)

VIETNAMESE SPEECH-TO-TEXT CORRECTIONS:
Whisper often badly mishears Vietnamese. It may even output WRONG LANGUAGES (Thai, Chinese, etc.) or meaningless text. Use expense context to interpret:

CRITICAL - WRONG LANGUAGE HANDLING:
- If you see Thai, Chinese, or other non-Vietnamese characters, IGNORE the foreign text
- Extract any numbers you can find and ask for the expense name
- Example: "45,000" with foreign script -> amount is 45000, but need to ask what was purchased

COMMON VIETNAMESE DISHES (Food category - "f"):
- "bun bo/bun ba/bum bo/uong bao" -> "Bún bò"
- "bun dau/bun dao" -> "Bún đậu"
- "cuom tam/cuom tan/com tam" -> "Cơm tấm"
- "suoi cao/sui cao/xui cao/mi suoi cao" -> "Sủi cảo" or "Mì sủi cảo"
- "cha ca/tra ca/tra cai" + "la vong" -> "Chả cá Lã Vọng"
- "mi hoanh/mi hoan" -> "Mì hoành"
- "banh mi/bang mi" -> "Bánh mì"
- "cafe/ca phe/ga phe/cf" -> "Cà phê"
- "an uong/an uoong" -> "Ăn uống"
- "com trua/cuom trua" -> "Cơm trưa"
- "gs25/ji es/GS" -> "GS25" (convenience store)

TRANSPORT (category - "tr"):
- "xang xe/xang se/sang xe/do xang" -> "Đổ xăng" or "Xăng xe"
- "giu xe/giu se/diu xe" -> "Giữ xe"
- "Grab/Grap/Gap/Gop" -> "Grab"
- "grab sb/grab san bay" -> "Grab sân bay"
- "grab cty/grab cong ty" -> "Grab công ty"
- "xe khach/se khach" -> "Xe khách"
- "taxiu/taxi/tac xi" -> "Taxi"

TECH (category - "tech"):
- "op dien thoai/oi bien thoai/op dt" -> "Ốp điện thoại" (phone case)
- "op airpod/op air pot" -> "Ốp Airpod"
- "sac dt/sac dien thoai/xac dien thoai" -> "Sạc điện thoại"
- "cuong luc/kinh cuong luc/king cuong luc" -> "Kính cường lực" (screen protector)
- "gia do laptop/gia đo laptop" -> "Giá đỡ laptop"
- "sua kinh camera/sua king camera" -> "Sửa kính camera điện thoại"
- "o cam dien/o cam" -> "Ổ cắm điện"

PERSONAL CARE (category - "pc"):
- "dau goi/dao goi" -> "Dầu gội"
- "dau xa/dao xa/dau xa thai duong" -> "Dầu xả"
- "sua rua mat/xua rua mat" -> "Sữa rửa mặt"
- "xit khu mui/sit khu mui" -> "Xịt khử mùi"
- "dao cao rau/dao cao" -> "Dao cạo râu"
- "ddvs" -> "Đồ dùng vệ sinh"

FAMILY/EVENTS (group - "fam" or category - "le"):
- "li xi/li si/ly xi" -> "Lì xì"
- "bieu/biu" -> "Biếu"
- "dam cuoi/dam cuoi" -> "Đám cưới"
- "sinh nhat/xinh nhat" -> "Sinh nhật"

HOUSING (category - "hou"):
- "tien tro/tien cho" -> "Tiền trọ"
- "do don nha/do don" -> "Đồ dọn nhà"

INVESTMENT:
- "vesaf/ve saf/vê sáp" -> "VESAF" (fund name, keep as-is)
- "ssisca/si si ca/xi xi ca" -> "SSISCA" (fund name, keep as-is)

VIETNAMESE NUMBERS (extract amounts even from garbled text):
- "nghin/ngan/nghìn/ngàn" = nghìn/ngàn = thousand (1,000)
- "trieu/triệu" = triệu = million (1,000,000)
- "mot tram/một trăm" = một trăm = 100
- "hai muoi/hai mươi" = hai mươi = 20
- Numeric digits like "45,000" or "45000" should be parsed directly
- Example: "ba mươi nghìn" = 30,000, "một trăm hai mươi nghìn" = 120,000

INTERPRETATION STRATEGY:
1. Look for any recognizable Vietnamese food/transport/expense words
2. Look for any amount (numbers, k suffix, "nghìn", "triệu")
3. If name is unrecognizable but amount exists, try to guess from phonetic similarity
4. Only ask for clarification if you truly cannot interpret the expense name AND amount

Available Groups (use the alias in lowercase for output):
- MUST HAVE (alias: mh) - Thiết yếu (DEFAULT if not specified)
- NICE TO HAVE (alias: nth) - Nên chi
- WASTE (alias: w) - Lãng phí
- FAMILY (alias: fam) - Gia đình
- LOVER (alias: lov) - Người yêu

Available Categories (use the alias in lowercase for output):
- Unclassified (alias: uc) - Chưa phân loại (DEFAULT if not specified)
- Food (alias: f) - Ăn ngoài, ăn trưa, ăn tối, cơm, bún, phở, cà phê, trà sữa, GS25, KFC, ăn uống
- Groceries (alias: gr) - Đi chợ, siêu thị, thùng rác, túi rác, bình nước, khăn tắm
- Transport (alias: tr) - Đi lại, Grab, xe ôm, taxi, xăng, đổ xăng, giữ xe, xe khách, bus, vé máy bay
- Entertainment (alias: ent) - Giải trí, vé concert, xem phim, game
- Miscellaneous (alias: mis) - Linh tinh, bình sịt giày, gang tay, lót mũ bảo hiểm
- Subscription (alias: sub) - Đăng ký, Spotify, Netflix, YouTube Premium, VESAF, SSISCA
- Housing (alias: hou) - Nhà ở, tiền trọ, điện, nước, đồ dọn nhà, ổ cắm điện, kệ máy giặt
- Personal Care (alias: pc) - Dầu gội, dầu xả, sữa rửa mặt, xịt khử mùi, dao cạo râu, ddvs, thuốc nhuộm tóc
- Healthcare (alias: hc) - Sức khỏe, thuốc, khám bệnh
- Clothing (alias: clo) - Quần áo, giày dép, dép sục, quần short, quần lót
- Education (alias: edu) - Học piano, học, sách
- Tech (alias: tech) - Ốp điện thoại, ốp airpod, sạc điện thoại, kính cường lực, giá đỡ laptop, sửa máy ảnh, sửa kính camera
- Travel (alias: tv) - Du lịch, vé máy bay, Vipassana
- Present (alias: pre) - Quà tặng, bông tai tặng, sách tặng, charm, quà biếu
- Life Events (alias: le) - Đám cưới, sinh nhật, tang lễ, lì xì
- Lover (alias: lov) - Người yêu
- Family (alias: fam) - Gia đình, biếu ông nội, đưa mẹ tiền, cho trẻ con

IMPORTANT RULES:
1. INTERPRET the garbled Vietnamese transcription - figure out what the user likely said based on expense context, then extract the corrected expense name
2. Extract the amount - support "k" (thousand) and "m" (million) suffixes. Examples: "50k" = 50000, "1.5m" = 1500000
3. Infer the group if mentioned (mh, must have, thiết yếu, etc.) - DEFAULT to "mh" if not clear
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
- "Uong bao 45000" -> name: "Bún bò", amount: 45000, group: "mh", category: "f", is_clear: true
- "Tra cai la vong 178k" -> name: "Chả cá Lã Vọng", amount: 178000, group: "mh", category: "f", is_clear: true
- "oi bien thoai 85k" -> name: "Ốp điện thoại", amount: 85000, group: "mh", category: "tech", is_clear: true
- "cuong luc cho me 50k" -> name: "Kính cường lực cho mẹ", amount: 50000, group: "fam", category: "tech", is_clear: true
- "do xang 100k" -> name: "Đổ xăng", amount: 100000, group: "mh", category: "tr", is_clear: true
- "Grap san bay 200k" -> name: "Grab sân bay", amount: 200000, group: "mh", category: "tr", is_clear: true
- "li xi e Quyen 500k" -> name: "Lì xì e Quyền", amount: 500000, group: "fam", category: "le", is_clear: true
- "vesaf 4 trieu" -> name: "VESAF", amount: 4000000, group: "mh", category: "sub", is_clear: true
- "tien tro 3.5m" -> name: "Tiền trọ", amount: 3500000, group: "mh", category: "hou", is_clear: true
- "An trua 50k" -> name: "Ăn trưa", amount: 50000, group: "mh", category: "f", is_clear: true
- "dam cuoi a Thong 500k" -> name: "Đám cưới a Thông", amount: 500000, group: "mh", category: "le", is_clear: true
- "youtube premium 79k" -> name: "YouTube Premium", amount: 79000, group: "mh", category: "sub", is_clear: true`

// ParseExpenseText parses natural language text into expense structure using ChatGPT
func (c *Client) ParseExpenseText(ctx context.Context, text string) (*ParsedExpense, error) {
	// Include today's date so the model knows what "today" means and doesn't hallucinate dates
	today := timepkg.Now().Format("2006-01-02")
	userMessage := fmt.Sprintf("Today is %s.\n\n%s", today, text)

	resp, err := c.client.CreateChatCompletion(ctx, openai.ChatCompletionRequest{
		Model: c.chatModel,
		Messages: []openai.ChatCompletionMessage{
			{
				Role:    openai.ChatMessageRoleSystem,
				Content: systemPrompt,
			},
			{
				Role:    openai.ChatMessageRoleUser,
				Content: userMessage,
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
		parsed.Date = timepkg.Now().Format("2006-01-02")
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

	// Parse date in local timezone so FormatDateTime doesn't shift the date
	if p.Date != "" {
		date, err = time.ParseInLocation("2006-01-02", p.Date, timepkg.LocalLocation)
		if err != nil {
			date = timepkg.Now()
			err = nil
		}
	} else {
		date = timepkg.Now()
	}

	note = p.Note
	return
}
