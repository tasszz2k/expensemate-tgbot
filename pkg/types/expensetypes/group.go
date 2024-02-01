package expensetypes

type Group string

const (
	GroupIncome     Group = "INCOME"
	GroupMustHave   Group = "MUST_HAVE"
	GroupNiceToHave Group = "NICE_TO_HAVE"
	GroupWasted     Group = "WASTED"
	GroupOther      Group = "OTHER"
)

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
