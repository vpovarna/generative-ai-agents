package validanagram

import "testing"

func TestIsAnagram(t *testing.T) {
	a := "rat"
	b := "cat"
	expected := false

	result := IsAnagram(a, b)

	if result != expected {
		t.Errorf("IsAnagram(%s, %s) = %v; Expected: %v", a, b, result, expected)
	}

}

func TestIsAnagramWithTestCase(t *testing.T) {
	testCase := []struct {
		string_a string
		string_b string
		expected bool
	}{
		{
			string_a: "rat",
			string_b: "cat",
			expected: false,
		},
		{
			string_a: "anagram",
			string_b: "nagaram",
			expected: true,
		},
	}

	for _, test := range testCase {
		result := IsAnagram(test.string_a, test.string_b)
		second_result := IsAnagramWithDict(test.string_a, test.string_b)
		if result != test.expected || second_result != test.expected {
			t.Errorf("IsAnagram(%s, %s) = %v; Expected: %v", test.string_a, test.string_b, second_result, test.expected)
		}
	}
}
