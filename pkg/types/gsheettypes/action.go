package gsheettypes

type Action string

const (
	ActionConfigure Action = "configure"
	ActionHelp      Action = "help"
)

func (a Action) String() string {
	return string(a)
}
