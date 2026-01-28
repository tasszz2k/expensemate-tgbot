package types

import "strings"

// ExpenseAction represents an expense-related action
type ExpenseAction string

const (
	ExpenseActionAdd         ExpenseAction = "add"
	ExpenseActionView        ExpenseAction = "view"
	ExpenseActionUpdate      ExpenseAction = "update"
	ExpenseActionDelete      ExpenseAction = "delete"
	ExpenseActionReport      ExpenseAction = "report"
	ExpenseActionHelp        ExpenseAction = "help"
	ExpenseActionSetGroup    ExpenseAction = "setgrp"
	ExpenseActionSetCategory ExpenseAction = "setcat"
)

func (a ExpenseAction) String() string {
	return string(a)
}

// Group represents an expense group
type Group string

const (
	GroupIncome     Group = "INCOME"
	GroupInvestment Group = "INVESTMENT"
	GroupMustHave   Group = "MUST HAVE"
	GroupNiceToHave Group = "NICE TO HAVE"
	GroupWaste      Group = "WASTED"
	GroupFamily     Group = "FAMILY"
	GroupLover      Group = "LOVER"
)

var groupAliases = map[string]Group{
	"income":   GroupIncome,
	"thu nhập": GroupIncome,
	"thu nhap": GroupIncome,
	"i":        GroupIncome,
	"tn":       GroupIncome,

	"investment": GroupInvestment,
	"đầu tư":     GroupInvestment,
	"dau tu":     GroupInvestment,
	"inv":        GroupInvestment,
	"dt":         GroupInvestment,

	"must have": GroupMustHave,
	"thiết yếu": GroupMustHave,
	"thiet yeu": GroupMustHave,
	"mh":        GroupMustHave,
	"ty":        GroupMustHave,

	"nice to have": GroupNiceToHave,
	"nên chi":      GroupNiceToHave,
	"nen chi":      GroupNiceToHave,
	"nth":          GroupNiceToHave,
	"nc":           GroupNiceToHave,

	"wasted":   GroupWaste,
	"waste":    GroupWaste,
	"lãng phí": GroupWaste,
	"lang phi": GroupWaste,
	"w":        GroupWaste,
	"lp":       GroupWaste,

	"family":   GroupFamily,
	"gia đình": GroupFamily,
	"gia dinh": GroupFamily,
	"fam":      GroupFamily,
	"gd":       GroupFamily,

	"lover":     GroupLover,
	"người yêu": GroupLover,
	"nguoi yeu": GroupLover,
	"lov":       GroupLover,
	"ny":        GroupLover,
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

// allGroups is the ordered list of all groups for UI display
var allGroups = []Group{
	GroupMustHave,
	GroupNiceToHave,
	GroupWaste,
	GroupIncome,
	GroupInvestment,
	GroupFamily,
	GroupLover,
}

// GetAllGroups returns all groups for UI selection
func GetAllGroups() []Group {
	return allGroups
}

// groupShortNames maps groups to short display names for buttons
var groupShortNames = map[Group]string{
	GroupIncome:     "Income",
	GroupInvestment: "Investment",
	GroupMustHave:   "Must Have",
	GroupNiceToHave: "Nice to Have",
	GroupWaste:      "Waste",
	GroupFamily:     "Family",
	GroupLover:      "Lover",
}

// groupAliasForCallback maps groups to their short alias for callbacks
var groupAliasForCallback = map[Group]string{
	GroupIncome:     "i",
	GroupInvestment: "inv",
	GroupMustHave:   "mh",
	GroupNiceToHave: "nth",
	GroupWaste:      "w",
	GroupFamily:     "fam",
	GroupLover:      "lov",
}

// GetGroupShortName returns a short display name for a group
func GetGroupShortName(g Group) string {
	if name, exists := groupShortNames[g]; exists {
		return name
	}
	return string(g)
}

// GetGroupAlias returns the short alias for a group (for callbacks)
func GetGroupAlias(g Group) string {
	if alias, exists := groupAliasForCallback[g]; exists {
		return alias
	}
	return "mh"
}

// Category represents an expense category
type Category string

const (
	CategoryUnclassified  Category = "Unclassified / Chưa phân loại"
	CategoryFood          Category = "Food / Ăn ngoài"
	CategoryGroceries     Category = "Groceries / Đi chợ"
	CategoryTransport     Category = "Transport / Đi lại"
	CategoryEntertainment Category = "Entertainment / Giải trí"
	CategoryMiscellaneous Category = "Miscellaneous / Linh tinh"
	CategorySubscription  Category = "Subscription / Đăng ký"
	CategoryHousing       Category = "Housing / Nhà ở"
	CategoryPersonalCare  Category = "Personal Care / Chăm sóc"
	CategoryHealthcare    Category = "Healthcare / Sức khỏe"
	CategoryClothing      Category = "Clothing / Quần áo"
	CategoryEducation     Category = "Education / Giáo dục"
	CategoryTech          Category = "Tech / Công nghệ"
	CategoryTravel        Category = "Travel / Du lịch"
	CategoryPresent       Category = "Present / Quà tặng"
	CategoryLifeEvents    Category = "Life Events / Hiếu hỉ"
	CategoryLover         Category = "Lover / Người yêu"
	CategoryFamily        Category = "Family / Gia đình"
	CategoryLostMoney     Category = "Lost Money / Mất tiền"
)

func (c Category) String() string {
	return string(c)
}

// allCategories is the ordered list of all categories for UI display
var allCategories = []Category{
	CategoryFood,
	CategoryGroceries,
	CategoryTransport,
	CategoryEntertainment,
	CategoryMiscellaneous,
	CategorySubscription,
	CategoryHousing,
	CategoryPersonalCare,
	CategoryHealthcare,
	CategoryClothing,
	CategoryEducation,
	CategoryTech,
	CategoryTravel,
	CategoryPresent,
	CategoryLifeEvents,
	CategoryLover,
	CategoryFamily,
	CategoryLostMoney,
}

// GetAllCategories returns all categories (excluding Unclassified) for UI selection
func GetAllCategories() []Category {
	return allCategories
}

// categoryAliasMap maps aliases to categories
var categoryAliases = map[string]Category{
	"unclassified":   CategoryUnclassified,
	"chưa phân loại": CategoryUnclassified,
	"chua phan loai": CategoryUnclassified,
	"uc":             CategoryUnclassified,
	"cpl":            CategoryUnclassified,

	"food":     CategoryFood,
	"ăn ngoài": CategoryFood,
	"an ngoai": CategoryFood,
	"f":        CategoryFood,
	"an":       CategoryFood,
	"cf":       CategoryFood,

	"groceries": CategoryGroceries,
	"đi chợ":    CategoryGroceries,
	"di cho":    CategoryGroceries,
	"gr":        CategoryGroceries,
	"dc":        CategoryGroceries,

	"transport": CategoryTransport,
	"đi lại":    CategoryTransport,
	"di lai":    CategoryTransport,
	"tr":        CategoryTransport,
	"dl":        CategoryTransport,

	"entertainment": CategoryEntertainment,
	"giải trí":      CategoryEntertainment,
	"giai tri":      CategoryEntertainment,
	"ent":           CategoryEntertainment,
	"gt":            CategoryEntertainment,

	"miscellaneous": CategoryMiscellaneous,
	"linh tinh":     CategoryMiscellaneous,
	"mis":           CategoryMiscellaneous,
	"lt":            CategoryMiscellaneous,

	"subscription": CategorySubscription,
	"đăng ký":      CategorySubscription,
	"dang ky":      CategorySubscription,
	"sub":          CategorySubscription,
	"dk":           CategorySubscription,

	"housing": CategoryHousing,
	"nhà ở":   CategoryHousing,
	"nha o":   CategoryHousing,
	"hou":     CategoryHousing,
	"no":      CategoryHousing,

	"personal care": CategoryPersonalCare,
	"chăm sóc":      CategoryPersonalCare,
	"cham soc":      CategoryPersonalCare,
	"pc":            CategoryPersonalCare,
	"cs":            CategoryPersonalCare,

	"healthcare": CategoryHealthcare,
	"sức khỏe":   CategoryHealthcare,
	"suc khoe":   CategoryHealthcare,
	"hc":         CategoryHealthcare,
	"sk":         CategoryHealthcare,

	"clothing": CategoryClothing,
	"quần áo":  CategoryClothing,
	"quan ao":  CategoryClothing,
	"clo":      CategoryClothing,
	"qa":       CategoryClothing,

	"education": CategoryEducation,
	"giáo dục":  CategoryEducation,
	"giao duc":  CategoryEducation,
	"edu":       CategoryEducation,
	"hoc":       CategoryEducation,

	"tech":      CategoryTech,
	"công nghệ": CategoryTech,
	"cong nghe": CategoryTech,
	"cn":        CategoryTech,

	"travel":  CategoryTravel,
	"du lịch": CategoryTravel,
	"du lich": CategoryTravel,
	"tv":      CategoryTravel,
	"dul":     CategoryTravel,

	"present":  CategoryPresent,
	"quà tặng": CategoryPresent,
	"qua tang": CategoryPresent,
	"pre":      CategoryPresent,
	"qt":       CategoryPresent,

	"life events": CategoryLifeEvents,
	"hiếu hỉ":     CategoryLifeEvents,
	"hieu hi":     CategoryLifeEvents,
	"le":          CategoryLifeEvents,
	"hh":          CategoryLifeEvents,

	"lover":     CategoryLover,
	"người yêu": CategoryLover,
	"nguoi yeu": CategoryLover,
	"lov":       CategoryLover,
	"ny":        CategoryLover,

	"family":   CategoryFamily,
	"gia đình": CategoryFamily,
	"gia dinh": CategoryFamily,
	"fam":      CategoryFamily,
	"gd":       CategoryFamily,

	"lost money": CategoryLostMoney,
	"mất tiền":   CategoryLostMoney,
	"mat tien":   CategoryLostMoney,
	"lm":         CategoryLostMoney,
	"mat":        CategoryLostMoney,
}

// categoryShortNames maps categories to short display names for buttons
var categoryShortNames = map[Category]string{
	CategoryFood:          "Food",
	CategoryGroceries:     "Groceries",
	CategoryTransport:     "Transport",
	CategoryEntertainment: "Entertainment",
	CategoryMiscellaneous: "Misc",
	CategorySubscription:  "Subscription",
	CategoryHousing:       "Housing",
	CategoryPersonalCare:  "Personal",
	CategoryHealthcare:    "Healthcare",
	CategoryClothing:      "Clothing",
	CategoryEducation:     "Education",
	CategoryTech:          "Tech",
	CategoryTravel:        "Travel",
	CategoryPresent:       "Present",
	CategoryLifeEvents:    "Life Events",
	CategoryLover:         "Lover",
	CategoryFamily:        "Family",
	CategoryLostMoney:     "Lost Money",
}

// categoryAliasForCallback maps categories to their short alias for callbacks
var categoryAliasForCallback = map[Category]string{
	CategoryFood:          "f",
	CategoryGroceries:     "gr",
	CategoryTransport:     "tr",
	CategoryEntertainment: "ent",
	CategoryMiscellaneous: "mis",
	CategorySubscription:  "sub",
	CategoryHousing:       "hou",
	CategoryPersonalCare:  "pc",
	CategoryHealthcare:    "hc",
	CategoryClothing:      "clo",
	CategoryEducation:     "edu",
	CategoryTech:          "tech",
	CategoryTravel:        "tv",
	CategoryPresent:       "pre",
	CategoryLifeEvents:    "le",
	CategoryLover:         "lov",
	CategoryFamily:        "fam",
	CategoryLostMoney:     "lm",
}

// GetCategoryByAlias returns a Category by its alias (case-insensitive)
func GetCategoryByAlias(name string) (Category, bool) {
	name = strings.ToLower(strings.TrimSpace(name))
	if category, exists := categoryAliases[name]; exists {
		return category, true
	}
	return Category(""), false
}

// GetCategoryShortName returns a short display name for a category
func GetCategoryShortName(c Category) string {
	if name, exists := categoryShortNames[c]; exists {
		return name
	}
	return string(c)
}

// GetCategoryAlias returns the short alias for a category (for callbacks)
func GetCategoryAlias(c Category) string {
	if alias, exists := categoryAliasForCallback[c]; exists {
		return alias
	}
	return "uc"
}
