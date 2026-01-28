# Changelog

All notable changes to the ExpenseMate Telegram Bot project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

### Added
- Local Dockerfile build for deployment instead of remote builder
- Automatic expense add conversation continuation - users can add multiple expenses without typing `/add_expense` again
- Support for both currency symbols: VND dong sign and lowercase d with stroke
- Document ID validation for Google Sheets configuration
- Demo screenshots and documentation
- Installation guide documentation
- Google Sheets authorization requirement
- Deployment configuration with Fly.io
- Environment-based configuration reading for secrets
- Sample configuration files
- Google Sheets list sheets functionality
- Expense view and report features
- Expense add functionality with multi-line input parsing
- Spreadsheet mappings loader for user-specific sheets
- `/expenses_help` command showing supported groups and categories
- Base command handlers for Telegram bot
- Project documentation and planning docs

### Changed
- Renamed "current page" to "active page" for clarity
- Updated default category handling
- Updated current/active page value management
- Updated start command behavior
- Improved date format handling (date only format)

### Fixed
- Linting issues in codebase
- Default category assignment bug

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

[Unreleased]: https://github.com/user/expensemate-tgbot/compare/v0.1.0...HEAD
[0.1.0]: https://github.com/user/expensemate-tgbot/releases/tag/v0.1.0
