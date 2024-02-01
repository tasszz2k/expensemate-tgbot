package expensetypes

type NameWithAliases struct {
	Name    string   `json:"name"`
	Aliases []string `json:"aliases"`
}
