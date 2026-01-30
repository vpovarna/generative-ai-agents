package twosum

import "testing"

func TestTwoSum(t *testing.T) {
	nums := []int{2, 7, 11, 15}
	target := 9
	result := TwoSum(nums, target)
	expected_result := []int{0, 1}

	if len(result) != len(expected_result) || result[0] != expected_result[0] || result[1] != expected_result[1] {
		t.Errorf("TwoSum(%v, %d) = %v; want %v", nums, target, result, expected_result)
	}
}

func TestTwoSumSuite(t *testing.T) {
	testCase := []struct {
		name     string
		nums     []int
		target   int
		expected []int
	}{
		{
			name:     "Example 1",
			nums:     []int{2, 7, 11, 15},
			target:   9,
			expected: []int{0, 1},
		},
		{
			name:     "Example 2",
			nums:     []int{3, 2, 4},
			target:   6,
			expected: []int{1, 2},
		},
		{
			name:     "Example 3",
			nums:     []int{3, 3},
			target:   6,
			expected: []int{0, 1},
		},
	}

	for _, tc := range testCase {
		result := TwoSum(tc.nums, tc.target)
		if len(result) != 2 || result[0] != tc.expected[0] || result[1] != tc.expected[1] {
			t.Errorf("TwoSum(%v, %d) = %v; want %v", tc.nums, tc.target, result, tc.expected)
		}
	}
}
