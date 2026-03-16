package currency

import (
	"fmt"
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

// FormatVND formats a number as compact Vietnamese Dong for Telegram display.
// >= 1M and divisible by 100k: millions (6.9m, 32m)
// >= 1k: truncated thousands with commas (516k, 6,383k)
// < 1k: exact with dong symbol (1 ₫, 0 ₫)
func FormatVND(amount types.Unsigned) string {
	v := uint64(amount)

	if v >= 1_000_000 && v%100_000 == 0 {
		if v%1_000_000 == 0 {
			return fmt.Sprintf("%sm", addCommas(v/1_000_000))
		}
		return fmt.Sprintf("%.1fm", float64(v)/1_000_000)
	}

	if v >= 1_000 {
		return addCommas(v/1_000) + "k"
	}

	return fmt.Sprintf("%d ₫", v)
}

// FormatVNDSigned formats a signed int64 as compact Vietnamese Dong.
// Negative values are prefixed with a minus sign.
func FormatVNDSigned(amount int64) string {
	if amount < 0 {
		return "-" + FormatVND(types.Unsigned(-amount))
	}
	return FormatVND(types.Unsigned(amount))
}

func addCommas(n uint64) string {
	s := strconv.FormatUint(n, 10)
	var b strings.Builder
	for i, ch := range s {
		if i > 0 && (len(s)-i)%3 == 0 {
			b.WriteRune(',')
		}
		b.WriteRune(ch)
	}
	return b.String()
}

// ReverseFormatVND parses a formatted VND string back to Unsigned.
// Example: "100,000 d" -> 100000
func ReverseFormatVND(amountStr string) (types.Unsigned, error) {
	// Remove currency symbols and separators
	// Handle both " d" suffix and Unicode dong symbol
	amountStr = strings.ReplaceAll(amountStr, " d", "")
	amountStr = strings.ReplaceAll(amountStr, "d", "")
	amountStr = strings.ReplaceAll(amountStr, " ₫", "") // space + dong symbol
	amountStr = strings.ReplaceAll(amountStr, "₫", "")  // dong symbol alone
	amountStr = strings.ReplaceAll(amountStr, ",", "")
	amountStr = strings.TrimSpace(amountStr)

	amount, err := strconv.ParseUint(amountStr, 10, 64)
	if err != nil {
		return 0, err
	}
	return types.Unsigned(amount), nil
}
