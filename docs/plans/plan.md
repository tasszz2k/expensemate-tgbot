# Project plan

## TODO List

### Telegram chatbot

- [x] Configure supported commands: `/start`, `/expenses`, `/expenses_add`, `gsheets`, `/settings`, `/feedback`, `/help`, ...
- [x] Configure API token
- [x] Command configurations:
```text
start - Greet user and guide them on using the bot. Configure Google Sheets with /gsheets.
expenses - Manage expenses with sub-commands
expenses_add - Quickly add an expense
gsheets - Manage Google Sheets integration with sub-commands
settings - (Admin only) Manage bot settings
feedback - Send feedback to the bot's admin
help - Show a list of available commands with descriptions
```

### Google sheets

- [x] Configure Google Sheets API
- [x] Configure Google Sheets credentials

## Requirements

### 0. Biz constants

I. Expense group:

1. INCOME / thu nhập
2. MUST HAVE / chi tiêu thiết yếu
3. NICE TO HAVE / không phải chi tiêu thiết yếu, nhưng nên chi, có thì tốt
4. WASTE / chi tiêu không cần thiết, lãng phí
5. OTHER / khác

II. Expense category:

1. Unclassified / Chưa phân loại
2. Food / Ăn uống
3. Housing / Nhà ở
4. Transportation / Giao thông vận tải
5. Utilities / Tiện ích
6. Healthcare / Sức khỏe
7. Entertainment / Giải trí
8. Education / Giáo dục
9. Clothing / Quần áo
10. Personal Care / Chăm sóc cá nhân
11. Miscellaneous / Đồ linh tinh
12. Other / Khác

### 1. Handle all commands when received

- [x] reply to user if command is supported or unsupported

### 2. Command: `/start` handler

- [ ] Greet user and guide user how to use the bot. Using `/gsheets` command to configure Google Sheets for themselves.

### 3. Command: `/expenses` handler

- [ ] Show the short description with sub-command buttons. There are sub-commands: `add`, `view`, `update`, `delete`, `report`
- [ ] **Handle `add` (or `/expenses_add` sub-command) button:**
    + User input: each on a new line: `expense name`, `amount`,`group`, `category`, `date`
      ```
      [expense name]: (*)
      [amount]: (*). support parse "k" -> thousand, "m" -> million
      [group]: default: "MUST HAVE"
      [category]: default: "Unclassified"
      [date]: default: current date (format: dd/mm/yyyy)
      [note]: default: empty
      ```
    + For example:
      ```
      Mua bánh mì
      10k
      MUST HAVE
      Food
      30/1/2024
      ```
    + Add new record to Google Sheets
    + Reply to user:
      ```
      Status: <status>
      --- 
      ID: <id> 
      Expense name: <name>
      Amount: <amount>
      Group: <group>
      Category: <category>
      Date: <date>
      Note: <note>
      ```

- [ ] **Handle `/view` button:**
    + Show the five last expenses:
    + Format:
       ```
       Here are the last 5 expenses:
       --- 
       ID: <id> 
       Expense name: <name>
       Amount: <amount>
       Group: <group>
       Category: <category>
       Date: <date>
       Note: <note>
       
       [...]
       ```

- [ ] **Handle `/update` button:**
- [ ] **Handle `/delete` button:**
- [ ] **Handle `/report` button:**

### 3. Command: `/gsheets` handler

- [ ] Show the list of buttons for Google Sheets management:  `list`, `select`, `configure`.
- [ ] **Handle `list` button:**
    + Show the list of Google Sheets that the bot can access.
    + Format:
      ```
      Here are the list of Google Sheets that the bot can access:
      <spreadsheets_id>: <spreadsheets_name> 
      ---
      1. <sheet_name>
      2. <sheet_name>
      3. <sheet_name>
      [...]
      ```
- [ ] **Handle `select` button:**
    + Show the list of Google Sheets that the bot can access.
    + User input: `index of the sheet`
    + Format:
      ```
      Please input the index of the sheet that you want to select:
      ```
    + Reply to user:
      ```
      Status: <status>
      ---
      You have selected <spreadsheets_id>: <spreadsheets_name>
      ```    
- [ ] **Handle `configure` button:**
    + Show the current configuration of Google Sheets for the user.
    + If the user has not configured Google Sheets for themselves, show a message to guide them how to configure.
    + Format:
      ```
      Here is your current configuration of Google Sheets:
      ---
      Spreadsheets ID: <spreadsheets_id>
      Spreadsheets name: <spreadsheets_name>
      Sheet name: <sheet_name>
      ---
      If you want to configure Google Sheets for yourself, please follow these steps:
      1. Clone the template Google Sheets: <template_url>
      2. Share the Google Sheets with the bot's email with edit permission: <bot_email>
      3. Input the Spreadsheets URL here:
      ```
    + User input: `spreadsheets_url`
    + Reply to user:
      ```
      Status: <status>
      ---
      You have configured Google Sheets for yourself:
      Spreadsheets ID: <spreadsheets_id>
      Spreadsheets name: <spreadsheets_name>
      Sheet name: <sheet_name>
      ```

### 4. Command: `/settings` handler

- **For admin only**

### 5. Command: `/feedback` handler
- [ ] Show the list of buttons for feedback: `report bug`, `suggest feature`, `say thanks`
- [ ] **Handle `report bug` button:**
    + User input: `bug description`
    + Reply to user:
      ```
      Status: <status>
      ---
      Thank you for your feedback!
      ```
- [ ] **Handle `suggest feature` button:**
    + User input: `feature description`
    + Reply to user:
      ```
      Status: <status>
      ---
      Thank you for your suggestion!
      ```
- [ ] **Handle `say thanks` button:**
    + User input: `thanks message`
    + Reply to user:
        ```
        Status: <status>
        ---
        Your message has been sent to the bot's admin.
        ```

### 6. Command: `/help` handler
- [ ] Show the list of buttons for help with short descriptions: `/hello`, `/expenses`, `/gsheets`, `/settings`, `/feedback`, `/help`, ...