package validpalindrome

import "testing"

func TestIsPalindrome(t *testing.T) {
	s := "A man, a plan, a canal: Panama"
	expected := true

	result := IsPalindrome(s)
	if result != expected {
		t.Errorf("IsPalindrome(%v)=%v. Expected = %v", s, result, expected)
	}
}

func TestIsPalindromeSmall(t *testing.T) {
	s := "0P"
	expected := false

	result := IsPalindrome(s)
	if result != expected {
		t.Errorf("IsPalindrome(%v)=%v. Expected = %v", s, result, expected)
	}
}
