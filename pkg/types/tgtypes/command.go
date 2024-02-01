package tgtypes

type Command string

const (
	CommandStart      Command = "start"
	CommandHelp       Command = "help"
	CommandExpenses   Command = "expenses"
	CommandExpenseAdd Command = "expenses_add"
	CommandGSheets    Command = "gsheets"
	CommandSettings   Command = "settings"
	CommandFeedback   Command = "feedback"
)
