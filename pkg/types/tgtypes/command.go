package tgtypes

import (
	"strings"
)

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

// ParseCallbackData parses the callback data and returns the command and the data.
// callback data format: [command]:[sub_command_1]:[sub_command_1.n]
func ParseCallbackData(data string) (command Command, subCommand string) {
	commands := strings.Split(data, ":")
	command = Command(commands[0])
	if len(commands) > 1 {
		subCommand = strings.Join(commands[1:], ":")
	}
	return
}

func BuildCallbackData(command Command, subCommand ...string) string {
	return string(command) + ":" + strings.Join(subCommand, ":")
}
