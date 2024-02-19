package gsheettypes

const (
	// User sheet Mapping
	UserSheetMappingSheetName  = "user_sheet_mappings"
	UserSheetMappingNextIdCell = "B1"
	UserSheetMappingTopRow     = 3
	UserSheetMappingLeftCol    = "A"
	UserSheetMappingRightCol   = "F"
)

func BuildCell(sheetName string, cell string) string {
	return sheetName + "!" + cell
}

func BuildRange(sheetName string, startCell string, endCell string) string {
	return sheetName + "!" + startCell + ":" + endCell
}
