package currencyutils_test

import (
	"testing"

	"expensemate-tgbot/pkg/types/types"
	"expensemate-tgbot/pkg/utils/currencyutils"

	"github.com/stretchr/testify/assert"
)

func TestFormatMoney(t *testing.T) {
	t.Parallel()
	type testCase struct {
		name     string
		input    types.Unsigned
		expected string
	}

	testCases := []testCase{
		{
			name:     "less than 1000",
			input:    999,
			expected: "999 ₫",
		},
		{
			name:     "1000",
			input:    1000,
			expected: "1,000 ₫",
		},
		{
			name:     "1000000",
			input:    1000000,
			expected: "1,000,000 ₫",
		},
		{
			name:     "1000000000",
			input:    1000000000,
			expected: "1,000,000,000 ₫",
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(
			tc.name, func(t *testing.T) {
				t.Parallel()
				actual := currencyutils.FormatVND(tc.input)
				assert.Equal(t, tc.expected, actual)
			},
		)
	}
}
