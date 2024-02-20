package gsheettypes

type Action string

const (
	ActionConfigure        Action = "configure"
	ActionHelp             Action = "help"
	ActionUpdateActivePage Action = "update_current_page"
)

func (a Action) String() string {
	return string(a)
}
