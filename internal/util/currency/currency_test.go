package currency

import (
	"testing"

	"expensemate-tgbot/internal/types"
)

func TestFormatVND(t *testing.T) {
	tests := []struct {
		amount uint64
		want   string
	}{
		{0, "0 ₫"},
		{1, "1 ₫"},
		{999, "999 ₫"},
		{1_000, "1k"},
		{1_500, "1k"},
		{50_000, "50k"},
		{516_002, "516k"},
		{999_999, "999k"},
		{1_000_000, "1m"},
		{1_100_000, "1.1m"},
		{1_150_000, "1,150k"},
		{6_383_998, "6,383k"},
		{6_900_000, "6.9m"},
		{32_000_000, "32m"},
		{31_999_999, "31,999k"},
		{100_000_000, "100m"},
	}

	for _, tt := range tests {
		got := FormatVND(types.Unsigned(tt.amount))
		if got != tt.want {
			t.Errorf("FormatVND(%d) = %q, want %q", tt.amount, got, tt.want)
		}
	}
}
