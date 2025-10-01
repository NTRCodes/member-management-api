package privacy

import "testing"

func TestIsValidSSN(t *testing.T) {
	testCases := []struct {
		val      string
		expected bool
	}{
		{
			val:      "123456789",
			expected: true,
		},
		{
			val:      "12345678",
			expected: false,
		},
		{
			val:      "",
			expected: false,
		},
	}

	for _, tc := range testCases {
		if isValid := IsValidSSN(tc.val); isValid != tc.expected {
			t.Errorf("Expected: %v\nGot: %v", tc.expected, isValid)
		}
	}
}
