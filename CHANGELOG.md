# Changelog

All notable changes to the ExpenseMate Telegram Bot project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

### Added
- Auto-create monthly sheet via `/gsheets` > "Create New Month" button
  - Duplicates current sheet with next month name (handles year rollover, e.g. 2025_12 -> 2026_01)
  - Clears salary (C4), expense data (rows 10+), and investment notes (AB16:AB30)
  - Snapshots current asset values (P2:Q9) to last month section (P17:Q24)
  - Updates investment formulas (C5:C9) to reference the correct previous month
  - Resets next expense ID to 10 and sets new sheet as active page
  - Two-step confirmation flow to prevent accidental creation
- Active page name shown in `/gsheets` menu alongside spreadsheet URL
- Google Sheets API methods: BatchUpdate, ClearValues, GetFormulas, GetUnformatted, GetSheetID, UpdateUserEntered
- Month name helpers (`NextMonthName`, `PrevMonthName`) with year boundary rollover

### Fixed
- Voice expense date bug - OpenAI hallucinated dates (e.g., 4/10/2023) because the prompt didn't include today's date; now injects current date into user message
- All date parsing/formatting now uses Asia/Ho_Chi_Minh timezone consistently instead of UTC, preventing date shifts near midnight

### Changed
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
