package expensetypes

import "strings"

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

func GetGroupByAlias(name string) (Group, bool) {
	name = strings.ToLower(name)
	if group, exists := groupAliases[name]; exists {
		return group, true
	}
	return Group(""), false
}

type Category string

const (
	CategoryUnclassified   Category = "Unclassified / Chưa phân loại"
	CategoryFood           Category = "Food / Ăn uống"
	CategoryHousing        Category = "Housing / Nhà ở"
	CategoryTransportation Category = "Transportation / Đi lại"
	CategoryUtilities      Category = "Utilities / Tiện ích"
	CategoryHealthCare     Category = "Healthcare / Sức khỏe"
	CategoryEntertainment  Category = "Entertainment / Giải trí"
	CategoryEducation      Category = "Education / Giáo dục"
	CategoryClothing       Category = "Clothing / Quần áo"
	CategoryPersonalCare   Category = "Personal Care / Chăm sóc cá nhân"
	CategoryMiscellaneous  Category = "Miscellaneous / Đồ linh tinh"
	CategoryTravel         Category = "Travel / Du lịch"
	CategoryOther          Category = "Other / Khác"
)

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
	"o":     CategoryOther,
}

func GetCategoryByAlias(name string) (Category, bool) {
	name = strings.ToLower(name)
	if category, exists := categoryAliases[name]; exists {
		return category, true
	}
	return Category(""), false
}
