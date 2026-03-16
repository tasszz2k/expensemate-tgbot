# Changelog

All notable changes to the ExpenseMate Telegram Bot project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

### Added
- AI Ask feature for natural language expense queries (`/ask`, `/a`, `/q` commands)
  - Ask about expenses in Vietnamese or English with follow-up conversation support
  - GPT-powered with current month budget data + 6-month per-month historical breakdowns
  - Serves as unified help system -- AI can answer bot usage questions too
  - In-memory conversation history per chat (capped at 10 turns, lost on restart)
  - Supports inline usage: `/ask how much did I spend this month?`
  - System prompt with Telegram HTML formatting rules, ranking/comparison instructions
- `ChatMessage` type and `ChatWithHistory` method on OpenAI client for multi-turn conversations
- `AskService` with system prompt builder, expense data gatherer, conversation orchestrator
- `AskPendingData` conversation state for storing chat history
- `GetMonthlyReports` method on ExpenseService for per-month group/category/summary data
- Expense Insights feature (`/expenses_insights`) for multi-month spending analysis
  - Average spending by group/category across 3M/6M/12M/YTD periods
  - Emergency fund calculator with configurable multiplier (3x/6x/Custom)
  - Period and EF multiplier selection via inline buttons
  - Option to exclude current (incomplete) month from averages (default: excluded)
  - Custom EF multiplier input via conversation state
  - `FormatVNDSigned` for correct display of negative currency (e.g., investment losses)
  - Summary icons for Self Expenses, Total Expenses, Net Change
  - Multi-month batch reading via `GetMultiMonthReports` repository method
  - `InsightsResult` and `AverageEntry` models
  - Sheet name helpers: `RecentSheetNames`, `YTDSheetNames`, `SortSheetNames`, `ParseSheetMonth`
- New "Cafe / Cafe" expense category separated from Food (20 categories total)
  - Aliases: `cf`, `caphe`, `cafe`, `ca phe`
  - Cafe-related items (coffee, milk tea, juice, smoothie) moved from Food to Cafe
  - OpenAI voice prompt updated with CAFE/DRINKS speech corrections section
  - Category report range expanded from `N3:R21` to `N3:R22`
- Group alias shortcuts: `f` for Family, `l` for Lover (synced from docs to code)
- Compact currency formatting for Telegram messages (`516k`, `6.9m` instead of `516,002 ₫`, `6,900,000 ₫`)
  - Amounts >= 1M and divisible by 100k shown in millions (e.g., `6.9m`, `32m`)
  - Amounts >= 1k shown in truncated thousands (e.g., `516k`, `6,383k`)
  - Small amounts shown exact with dong symbol (e.g., `1 ₫`)
- Monthly spending summary in report and after-add response (Self Expenses, Total Expenses, Net Change from rows I11:L13)
  - Summary section shown in Expense Report and Budget Overview after categories
  - "This month" total line appended to after-add budget status
  - `FormatTotalLine()` method on `BudgetEntry` for compact display
- Category-specific icons in expense view (e.g., food shows fork, cafe shows coffee cup instead of generic money bag)
- Unit tests for `FormatVND` compact currency formatting
- Budget (Expense Plan) feature for setting monthly spending caps per group and category
  - `/budget` standalone command for quick access to budget management
  - Budget menu accessible from `/expenses` > "Budget" button
  - View budget overview with emoji status indicators and sorted display (over -> ok -> empty)
  - Set/update budget caps for individual groups and categories via Telegram
  - Budget data stored in Google Sheets alongside existing reports (columns K for groups, Q for categories)
  - Budget status shown after adding expenses (text, voice, and group/category button selection)
  - Budget status shown in expense report with spent/budget/remaining breakdown
  - `BudgetEntry` model with `FormatBudgetLine()` and `FormatShortBudgetLine()` display methods
  - `BatchGet` method on Sheets client for efficient multi-range reads
  - Budget conversation states (`budget:set_group:{row}`, `budget:set_category:{row}`)
- Auto-create monthly sheet via `/gsheets` > "Create New Month" button
  - Duplicates current sheet with next month name (handles year rollover, e.g. 2025_12 -> 2026_01)
  - Clears salary (C4), expense data (rows 10+), and investment notes (AF16:AF30)
  - Snapshots current asset values (T2:U9) to last month section (T17:U24)
  - Updates investment formulas (C5:C9) to reference the correct previous month
  - Resets next expense ID to 10 and sets new sheet as active page
  - Two-step confirmation flow to prevent accidental creation
  - "View in Google Sheets" link shown in success message
- Active page name shown in `/gsheets` menu alongside spreadsheet URL
- Google Sheets API methods: BatchUpdate, ClearValues, GetFormulas, GetUnformatted, GetSheetID, UpdateUserEntered
- Month name helpers (`NextMonthName`, `PrevMonthName`) with year boundary rollover
- "Show more" button on expense view to load additional expenses (5 at a time, up to 25)
- `viewmore` callback action for paginated expense viewing

### Fixed
- All "View in Google Sheets" links now navigate directly to the active sheet tab using canonical `/edit#gid=` URL format
- Category report range expanded from `L3:N15` to `L3:N21` to include all 19 categories (Travel, Present, Life Events, etc. were missing)
- `InvestmentNoteRange` updated from `AB16:AB30` to `AF16:AF30` to match column shift from budget feature
- Budget status now shown after group/category selection via inline buttons (was only shown on initial add)
- Budget status now shown for voice expenses and voice clarification responses

### Changed
- Zero-spent budget entries hidden from Expense Report to reduce clutter
- Column layout shifted for budget columns: categories report N3:P21 (+2), assets T2:U9 (+4), investment notes AF16:AF30 (+4)
- Expense report formatting: non-zero amounts shown in bold, zero amounts shown with muted bullet
- Expense view formatting: amounts shown in bold with dash separator
- Voice expense date bug - OpenAI hallucinated dates (e.g., 4/10/2023) because the prompt didn't include today's date; now injects current date into user message
- All date parsing/formatting now uses Asia/Ho_Chi_Minh timezone consistently instead of UTC, preventing date shifts near midnight
- Expanded Whisper vocabulary prompt from food-only to all expense categories (tech, personal care, housing, transport, investments, events)
- Expanded ChatGPT system prompt with comprehensive Vietnamese speech-to-text corrections across all categories
- Category selection buttons now show bilingual labels (e.g., "Food / Ăn ngoài", "Housing / Nhà ở")
- Added `run-prod` target to Makefile for local development with config file

## [1.1.0] - 2026-01-29

### Added
- Voice input expense feature using OpenAI Whisper (speech-to-text) and ChatGPT (natural language parsing)
- Voice messages can be sent in any context to add expenses
- Clarification flow for ambiguous voice inputs (e.g., missing amount)
- Quick delete button for voice expenses - allows immediate deletion if parsing was incorrect
- Group and category selection via inline buttons after adding expense
- Conditional button display - only shows selection buttons if user didn't provide group/category in input
- Response wrapper for flexible message editing vs new message sending
- Separate debug logging control for bot API vs application logs
- Local Dockerfile build for deployment instead of remote builder
- Automatic expense add conversation continuation - users can add multiple expenses without typing `/add_expense` again
- Support for both currency symbols: VND dong sign and lowercase d with stroke
- Document ID validation for Google Sheets configuration
- Demo screenshots and documentation
- Installation guide documentation
- Deployment configuration with Fly.io
- Environment-based configuration reading for secrets
- Sample configuration files
- Google Sheets list sheets functionality
- Expense view and report features
- Expense add functionality with multi-line input parsing
- Spreadsheet mappings loader for user-specific sheets
- `/expenses_help` command showing supported groups and categories
- Cursor rules for project context, Go standards, and Telegram bot patterns

### Changed
- Renamed "Dining" category to "Food" with aliases: f, an, cf
- Income and Investment Out groups no longer show/save category (not expense types)
- Complete codebase refactoring following Go standards and layered architecture
- Replaced slog with logrus for structured logging with context fields
- Dependency injection pattern replacing global singletons
- Message editing instead of sending new messages for group/category selection flow
- Date column now uses datetime format (d/m/yyyy HH:mm) for automatic input
- Expense groups updated to 7: INCOME, INVESTMENT OUT, MUST HAVE, NICE TO HAVE, WASTED, FAMILY, LOVER
- Renamed WASTE group to WASTED
- Column order updated: ID, Name, Amount, Group, Category, Date, Note
- Categories and groups now use exact Vietnamese diacritics
- Renamed "current page" to "active page" for clarity

### Fixed
- Currency parsing for Unicode dong symbol (U+20AB)
- Callback handler using bot ID instead of user ID
- Callback data indexing for group/category selection
- Expense retrieval for final response display after selection
- Skip button now includes expense ID for proper record display
- Linting issues in codebase
- Default category assignment bug

### Technical
- Added `internal/repository/openai/client.go` - OpenAI Whisper + ChatGPT client
- Added `internal/service/voice_expense.go` - Voice processing orchestration
- Added voice clarification conversation state `expenses:voice_clarify`
- Added `VoicePendingData` struct for storing pending voice expense during clarification
- Added `NeedsCategory()` method on Group type
- Added `ExpenseActionQuickDelete` action for instant expense deletion
- OpenAI configuration section in config (api_key, whisper_model, chat_model)
- `buildFinalExpenseResponse` helper for consistent expense display

## [0.1.0] - Initial Release

### Added
- Initial project structure with Go 1.25
- Telegram Bot API integration using `go-telegram-bot-api/v5`
- Google Sheets API integration for expense storage
- Core expense management features:
  - Add expenses with name, amount, category, group, date, and note
  - View recent expenses
  - Generate expense reports
- Expense categorization with 13 categories:
  - Unclassified, Food, Housing, Transportation, Utilities
  - Healthcare, Entertainment, Education, Clothing
  - Personal Care, Miscellaneous, Travel, Other
- Expense grouping with 5 groups:
  - INCOME, MUST HAVE, NICE TO HAVE, WASTED, OTHER
- Google Sheets configuration per user
- Amount parsing with support for 'k' (thousands) and 'm' (millions) suffixes
- Bilingual support (English/Vietnamese) for categories and groups
- User-to-spreadsheet mapping in central database
- Conversation state management for multi-step interactions
- Commands:
  - `/start` - Bot greeting and introduction
  - `/expenses` - Expense management menu
  - `/expenses_add` - Quick expense addition
  - `/expenses_help` - Show groups and categories
  - `/gsheets` - Google Sheets configuration
  - `/help` - Command list

### Technical
- Viper-based configuration management
- slog structured logging
- JWT service account authentication for Google APIs
- Singleton pattern for Google Sheets client
- HTTP server with Telegram update polling

[Unreleased]: https://github.com/user/expensemate-tgbot/compare/v1.1.0...HEAD
[1.1.0]: https://github.com/user/expensemate-tgbot/compare/v0.1.0...v1.1.0
[0.1.0]: https://github.com/user/expensemate-tgbot/releases/tag/v0.1.0
