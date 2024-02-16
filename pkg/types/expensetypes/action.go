package expensetypes

type Action string

const (
	ActionAdd    Action = "add"
	ActionView   Action = "view"
	ActionUpdate Action = "update"
	ActionDelete Action = "delete"
	ActionReport Action = "report"
	ActionHelp   Action = "help"
)

func (a Action) String() string {
	return string(a)
}
