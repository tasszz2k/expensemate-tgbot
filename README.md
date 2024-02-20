# ExpenseMate Telegram Chatbot

ExpenseMate Telegram Chatbot is a versatile tool designed to help users manage their expenses efficiently within the Telegram messaging platform. With ExpenseMate, users can easily add, view, update, and delete expenses, as well as generate expense reports and configure Google Sheets integration for seamless expense tracking.

---

## Table of Contents

- [Project Overview](#project-overview)
- [Requirements](#requirements)
- [Commands](#commands)
    - [/start](#start)
    - [/expenses](#expenses)
    - [/gsheets](#gsheets)
    - [/settings](#settings)
    - [/feedback](#feedback)
    - [/help](#help)
- [Installation](#installation)
- [Usage](#usage)
- [Contributing](#contributing)
- [License](#license)

---

## Project Overview

ExpenseMate Telegram Chatbot is designed to facilitate expense management and tracking for users directly within the Telegram messaging platform. The chatbot provides a user-friendly interface for adding, viewing, updating, and deleting expenses, as well as configuring Google Sheets integration for comprehensive expense tracking and reporting.

---

## Requirements

ExpenseMate Telegram Chatbot has the following requirements:

### Business Constants

- **Expense Groups**:
    - INCOME / thu nhập
    - MUST HAVE / chi tiêu thiết yếu
    - NICE TO HAVE / không phải chi tiêu thiết yếu, nhưng nên chi, có thì tốt
    - WASTE / chi tiêu không cần thiết, lãng phí
    - OTHER / khác

- **Expense Categories**:
    - Unclassified / Chưa phân loại
    - Food / Ăn uống
    - Housing / Nhà ở
    - Transportation / Đi lại
    - Utilities / Tiện ích
    - Healthcare / Sức khỏe
    - Entertainment / Giải trí
    - Education / Giáo dục
    - Clothing / Quần áo
    - Personal Care / Chăm sóc cá nhân
    - Miscellaneous / Đồ linh tinh
    - Travel / Du lịch
    - Other / Khác

---

## Commands

ExpenseMate Telegram Chatbot supports the following commands:

### /start

- Greet the user and provide a guide on using the bot.
- Configure Google Sheets integration with `/gsheets` command.

### /expenses

- Manage expenses with sub-commands:
    - add
    - view
    - update
    - delete
    - report
    - help

### /gsheets

- Manage Google Sheets integration with sub-commands:
    - list
    - select
    - configure

### /settings

- *(Admin only)* Manage bot settings.

### /feedback

- Send feedback to the bot's admin:
    - report bug
    - suggest feature
    - say thanks

### /help

- Show a list of available commands with descriptions.

---

## Installation

To install ExpenseMate Telegram Chatbot, follow these steps:

1. Clone the repository: `git clone <repository-url>`
2. Install dependencies: `go mod download`
3. Build the project: `go build`

---

## Usage

To use ExpenseMate Telegram Chatbot, follow these steps:

1. Start the bot by running the executable.
2. Interact with the bot using the supported commands listed above.

---

## Installation
- Please follow the [installation guide](docs/installation/installation.md) to install the bot for your own use.


---

## Contributing

Contributions are welcome! To contribute to ExpenseMate Telegram Chatbot, follow these steps:

1. Fork the repository.
2. Create a new branch: `git checkout -b feature-branch`
3. Make your changes and commit them: `git commit -am 'Add new feature'`
4. Push to the branch: `git push origin feature-branch`
5. Submit a pull request.

---

## License

ExpenseMate Telegram Chatbot is licensed under the [MIT License](LICENSE).
