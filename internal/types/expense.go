package types

import "strings"

// ExpenseAction represents an expense-related action
type ExpenseAction string

const (
	ExpenseActionAdd    ExpenseAction = "add"
	ExpenseActionView   ExpenseAction = "view"
	ExpenseActionUpdate ExpenseAction = "update"
	ExpenseActionDelete ExpenseAction = "delete"
	ExpenseActionReport ExpenseAction = "report"
	ExpenseActionHelp   ExpenseAction = "help"
)

func (a ExpenseAction) String() string {
	return string(a)
}

// Group represents an expense group
type Group string

const (
	GroupIncome     Group = "INCOME"
	GroupMustHave   Group = "MUST HAVE"
	GroupNiceToHave Group = "NICE TO HAVE"
	GroupWasted     Group = "WASTED"
	GroupOther      Group = "OTHER"
)

var groupAliases = map[string]Group{
	"income": GroupIncome,
	"i":      GroupIncome,

	"must have": GroupMustHave,
	"must":      GroupMustHave,
	"mh":        GroupMustHave,

	"nice to have": GroupNiceToHave,
	"nice":         GroupNiceToHave,
	"nth":          GroupNiceToHave,

	"wasted": GroupWasted,
	"waste":  GroupWasted,
	"w":      GroupWasted,

	"other": GroupOther,
	"o":     GroupOther,
}

func (g Group) String() string {
	return string(g)
}

// GetGroupByAlias returns a Group by its alias (case-insensitive)
func GetGroupByAlias(name string) (Group, bool) {
	name = strings.ToLower(strings.TrimSpace(name))
	if group, exists := groupAliases[name]; exists {
		return group, true
	}
	return Group(""), false
}

// Category represents an expense category
type Category string

const (
	CategoryUnclassified   Category = "Unclassified / Chua phan loai"
	CategoryFood           Category = "Food / An uong"
	CategoryHousing        Category = "Housing / Nha o"
	CategoryTransportation Category = "Transportation / Di lai"
	CategoryUtilities      Category = "Utilities / Tien ich"
	CategoryHealthCare     Category = "Healthcare / Suc khoe"
	CategoryEntertainment  Category = "Entertainment / Giai tri"
	CategoryEducation      Category = "Education / Giao duc"
	CategoryClothing       Category = "Clothing / Quan ao"
	CategoryPersonalCare   Category = "Personal Care / Cham soc ca nhan"
	CategoryMiscellaneous  Category = "Miscellaneous / Do linh tinh"
	CategoryTravel         Category = "Travel / Du lich"
	CategoryOther          Category = "Other / Khac"
)

func (c Category) String() string {
	return string(c)
}

var categoryAliases = map[string]Category{
	"unclassified":   CategoryUnclassified,
	"chua phan loai": CategoryUnclassified,
	"cpl":            CategoryUnclassified,
	"uc":             CategoryUnclassified,

	"food":    CategoryFood,
	"an uong": CategoryFood,
	"au":      CategoryFood,
	"f":       CategoryFood,

	"housing": CategoryHousing,
	"nha o":   CategoryHousing,
	"no":      CategoryHousing,
	"h":       CategoryHousing,

	"transportation": CategoryTransportation,
	"di lai":         CategoryTransportation,
	"dl":             CategoryTransportation,
	"t":              CategoryTransportation,

	"utilities": CategoryUtilities,
	"tien ich":  CategoryUtilities,
	"ti":        CategoryUtilities,
	"u":         CategoryUtilities,

	"healthcare": CategoryHealthCare,
	"suc khoe":   CategoryHealthCare,
	"sk":         CategoryHealthCare,
	"hc":         CategoryHealthCare,

	"entertainment": CategoryEntertainment,
	"giai tri":      CategoryEntertainment,
	"gt":            CategoryEntertainment,
	"en":            CategoryEntertainment,

	"education": CategoryEducation,
	"giao duc":  CategoryEducation,
	"gd":        CategoryEducation,
	"ed":        CategoryEducation,

	"clothing": CategoryClothing,
	"quan ao":  CategoryClothing,
	"qa":       CategoryClothing,
	"c":        CategoryClothing,

	"personal care":    CategoryPersonalCare,
	"cham soc ca nhan": CategoryPersonalCare,
	"cscn":             CategoryPersonalCare,
	"pc":               CategoryPersonalCare,

	"miscellaneous": CategoryMiscellaneous,
	"do linh tinh":  CategoryMiscellaneous,
	"dlt":           CategoryMiscellaneous,
	"lt":            CategoryMiscellaneous,
	"m":             CategoryMiscellaneous,

	"travel":  CategoryTravel,
	"du lich": CategoryTravel,
	"tv":      CategoryTravel,

	"other": CategoryOther,
	"khac":  CategoryOther,
	"k":     CategoryOther,
}

// GetCategoryByAlias returns a Category by its alias (case-insensitive)
func GetCategoryByAlias(name string) (Category, bool) {
	name = strings.ToLower(strings.TrimSpace(name))
	if category, exists := categoryAliases[name]; exists {
		return category, true
	}
	return Category(""), false
}
