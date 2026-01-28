package currency

import (
	"strconv"
	"strings"

	"expensemate-tgbot/internal/types"
)

// ParseAmount processes amount string and converts "k" and "m" to multipliers.
// Examples: "50k" -> 50000, "1.5m" -> 1500000, "50000" -> 50000
func ParseAmount(amountStr string) int64 {
	amountStr = strings.TrimSpace(amountStr)
	if amountStr == "" {
		return 0
	}

	var multiplier int64
	switch {
	case strings.HasSuffix(amountStr, "k"):
		multiplier = 1000
		amountStr = strings.TrimSuffix(amountStr, "k")
	case strings.HasSuffix(amountStr, "m"):
		multiplier = 1000000
		amountStr = strings.TrimSuffix(amountStr, "m")
	default:
		value, err := strconv.ParseInt(amountStr, 10, 64)
		if err != nil {
			return 0
		}
		return value
	}

	value, err := strconv.ParseFloat(amountStr, 64)
	if err != nil {
		return 0
	}

	return int64(value * float64(multiplier))
}

// FormatVND formats a number as Vietnamese Dong currency.
// Example: 100000 -> "100,000 d"
func FormatVND(amount types.Unsigned) string {
	amountStr := strconv.FormatUint(uint64(amount), 10)
	var result strings.Builder

	n := len(amountStr)
	for i, char := range amountStr {
		if i > 0 && (n-i)%3 == 0 {
			result.WriteRune(',')
		}
		result.WriteRune(char)
	}

	result.WriteString(" d")
	return result.String()
}

// ReverseFormatVND parses a formatted VND string back to Unsigned.
// Example: "100,000 d" -> 100000
func ReverseFormatVND(amountStr string) (types.Unsigned, error) {
	// Remove currency symbols and separators
	amountStr = strings.ReplaceAll(amountStr, " d", "")
	amountStr = strings.ReplaceAll(amountStr, ",", "")
	amountStr = strings.TrimSpace(amountStr)

	amount, err := strconv.ParseUint(amountStr, 10, 64)
	if err != nil {
		return 0, err
	}
	return types.Unsigned(amount), nil
}
