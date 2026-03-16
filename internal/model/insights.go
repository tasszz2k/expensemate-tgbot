package model

// AverageEntry holds aggregated spending data for a single group or category
type AverageEntry struct {
	Name       string
	Total      int64
	Average    int64
	MonthCount int
}

// InsightsResult contains the full insights report across multiple months
type InsightsResult struct {
	Period                  string
	MonthsFound             []string
	MonthsMissing           []string
	ExcludedCurrent         string // active page name if excluded, empty if included
	GroupAvgs               []AverageEntry
	CategoryAvgs            []AverageEntry
	SummaryAvgs             []AverageEntry
	EmergencyFundMultiplier int
	EmergencyFund           int64
}
