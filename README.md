# ExpenseMate Telegram Chatbot

ExpenseMate is a Telegram bot for personal expense tracking with Google Sheets integration, voice input support, and budget management.

| Bot                               | Commands                            | /gsheets                                  |
|-----------------------------------|-------------------------------------|-------------------------------------------|
| ![bot.PNG](docs%2Fdemo%2Fbot.PNG) | ![cmds.PNG](docs%2Fdemo%2Fcmds.PNG) | ![gsheets.PNG](docs%2Fdemo%2Fgsheets.PNG) |

| /expenses_add                                       | /expenses_view                                        | /expenses_report                                          |
|-----------------------------------------------------|-------------------------------------------------------|-----------------------------------------------------------|
| ![expenses_add.PNG](docs%2Fdemo%2Fexpenses_add.PNG) | ![expenses_view.PNG](docs%2Fdemo%2Fexpenses_view.PNG) | ![expenses_report.PNG](docs%2Fdemo%2Fexpenses_report.PNG) |

| Google Sheets Report                                  |
|-------------------------------------------------------|
| ![sheets_report.png](docs%2Fdemo%2Fsheets_report.png) |

---

## Features

- Add expenses via text or voice input
- View, update, and delete expenses
- Expense reports by group and category
- Budget management with overspend alerts
- Google Sheets integration for data storage
- Monthly sheet creation with automated setup
- Voice expense input via OpenAI Whisper + ChatGPT

---

## Commands

| Command | Description |
|---------|-------------|
| `/start` | Welcome message and getting started |
| `/help` | Show available commands |
| `/expenses` | Expense manager menu (Add, View, Report, Budget, Update, Delete, Help) |
| `/expenses_add` | Quickly add an expense |
| `/expenses_help` | View supported groups & categories |
| `/gsheets` | Configure Google Sheets (Configure, View, Create New Month, Update Active Page) |
| `/settings` | Bot settings (Admin only) |
| `/feedback` | Send feedback |

---

## Expense Groups

| Emoji | Group | Aliases |
|-------|-------|---------|
| 💰 | INCOME | `i`, `tn` |
| 📈 | INVESTMENT OUT | `inv`, `dt`, `io` |
| 🔒 | MUST HAVE | `mh`, `ty` |
| ✨ | NICE TO HAVE | `nth`, `nc` |
| 🗑 | WASTED | `w`, `lp` |
| 👨‍👩‍👧‍👦 | FAMILY | `fam`, `gd` |
| ❤️ | LOVER | `lov`, `ny` |

## Expense Categories

| Emoji | Category | Aliases |
|-------|----------|---------|
| 🍜 | Food / Ăn ngoài | `f`, `an`, `cf` |
| 🛒 | Groceries / Đi chợ | `gr`, `dc` |
| 🚗 | Transport / Đi lại | `tr`, `dl` |
| 🎮 | Entertainment / Giải trí | `ent`, `gt` |
| 📦 | Miscellaneous / Linh tinh | `mis`, `lt` |
| 🔄 | Subscription / Đăng ký | `sub`, `dk` |
| 🏠 | Housing / Nhà ở | `hou`, `no` |
| 💆 | Personal Care / Chăm sóc | `pc`, `cs` |
| 🏥 | Healthcare / Sức khỏe | `hc`, `sk` |
| 👕 | Clothing / Quần áo | `clo`, `qa` |
| 📚 | Education / Giáo dục | `edu`, `hoc` |
| 💻 | Tech / Công nghệ | `tech`, `cn` |
| ✈️ | Travel / Du lịch | `tv`, `dul` |
| 🎁 | Present / Quà tặng | `pre`, `qt` |
| 🎊 | Life Events / Hiếu hỉ | `le`, `hh` |
| ❤️ | Lover / Người yêu | `lov`, `ny` |
| 👨‍👩‍👧‍👦 | Family / Gia đình | `fam`, `gd` |
| 💸 | Lost Money / Mất tiền | `lm`, `mat` |

---

## Installation

1. Clone the repository: `git clone <repository-url>`
2. Install dependencies: `go mod download`
3. Configure environment variables (see [installation guide](docs/installation/installation.md))
4. Build: `go build ./cmd/bot/`
5. Run: `./bot`

---

## Google Sheets Setup

1. Clone the [template spreadsheet](https://docs.google.com/spreadsheets/d/16jOEcyvHiHzW1GdRBvhHEadECojq0g3tzBT3a2MoLnI)
2. Use `/gsheets` in the bot and tap **Configure**
3. Paste your Google Sheets URL
4. Share **Editing access** with: `housematee-gsheets@housematee.iam.gserviceaccount.com`

---

## Contributing

1. Fork the repository
2. Create a branch: `git checkout -b feature-branch`
3. Commit changes: `git commit -am 'Add new feature'`
4. Push: `git push origin feature-branch`
5. Submit a pull request

---

## License

ExpenseMate Telegram Chatbot is licensed under the [MIT License](LICENSE).
