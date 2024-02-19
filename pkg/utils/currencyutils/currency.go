package currencyutils

import (
	"strconv"
	"strings"

	"expensemate-tgbot/pkg/types/types"
)

// ParseAmount processes the amount string and converts "k" and "m" to their respective multipliers.
func ParseAmount(amountStr string) int64 {
	var multiplier int64

	switch {
	case strings.Contains(amountStr, "k"):
		multiplier = 1000
	case strings.Contains(amountStr, "m"):
		multiplier = 1000000
	default:
		// If no multiplier is found, return the input as is
		value, err := strconv.ParseInt(amountStr, 10, 64)
		if err != nil {
			return 0 // Return 0 in case of error
		}
		return value
	}

	valueStr := strings.ReplaceAll(amountStr, "k", "")
	valueStr = strings.ReplaceAll(valueStr, "m", "")

	value, err := strconv.ParseFloat(valueStr, 64)
	if err != nil {
		return 0 // Return 0 in case of error
	}

	return int64(value * float64(multiplier))
}

// FormatVND Format number to money format
// e.g., 100000 -> 100,000 ₫
func FormatVND(amount types.Unsigned) string {
	amountStr := strconv.FormatUint(uint64(amount), 10)
	var result string

	// Iterate over the amount string from the end and add commas after every third digit
	for i := len(amountStr) - 1; i >= 0; i-- {
		if (len(amountStr)-i-1)%3 == 0 && i != len(amountStr)-1 {
			result = "," + result
		}
		result = string(amountStr[i]) + result
	}

	return result + " ₫"
}

func ReverseFormatVND(amountStr string) (types.Unsigned, error) {
	amountStr = strings.ReplaceAll(amountStr, " ₫", "")
	amountStr = strings.ReplaceAll(amountStr, ",", "")
	amount, err := strconv.ParseUint(amountStr, 10, 64)
	if err != nil {
		return 0, err
	}
	return types.Unsigned(amount), nil
}
