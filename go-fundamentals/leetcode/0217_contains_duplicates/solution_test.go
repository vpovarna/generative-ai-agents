package containsduplicates

import "testing"

func TestIsDuplicateCase(t *testing.T) {
	testCases := []struct {
		i        []int
		expected bool
	}{
		{
			i:        []int{1, 2, 3, 1},
			expected: true,
		},
		{
			i:        []int{1, 2, 3, 4},
			expected: false,
		},
		{
			i:        []int{1, 1, 1, 3, 3, 3, 4, 3, 4, 2},
			expected: true,
		},
	}

	for _, testCase := range testCases {
		result := ContainsDuplicates(testCase.i)
		if result != testCase.expected {
			t.Errorf("ContainsDuplicates(%v) = %v, Want: %v", testCase.i, result, testCase.expected)
		}
	}
}
