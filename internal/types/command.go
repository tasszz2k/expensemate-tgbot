package types

import "strings"

// Command represents a Telegram bot command
type Command string

const (
	CommandStart    Command = "start"
	CommandHelp     Command = "help"
	CommandGSheets  Command = "gsheets"
	CommandSettings Command = "settings"
	CommandFeedback Command = "feedback"

	CommandExpenses    Command = "expenses"
	CommandExpenseAdd  Command = "expenses_add"
	CommandExpenseHelp Command = "expenses_help"
)

// ParseCallbackData parses callback data and returns command and sub-commands.
// Format: [command]:[sub_command_1]:[sub_command_n]
func ParseCallbackData(data string) (command Command, subCommands []string) {
	parts := strings.Split(data, ":")
	command = Command(parts[0])
	if len(parts) > 1 {
		subCommands = parts[1:]
	}
	return
}

// BuildCallbackData builds callback data from command and sub-commands
func BuildCallbackData(command Command, subCommands ...string) string {
	if len(subCommands) == 0 {
		return string(command)
	}
	return string(command) + ":" + strings.Join(subCommands, ":")
}
