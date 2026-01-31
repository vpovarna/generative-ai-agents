package validpalindrome

import "unicode"

func IsPalindrome(s string) bool {
	chars := []rune(s)

	i := 0
	j := len(chars) - 1

	for i <= j {
		if !unicode.IsLetter(chars[i]) && !unicode.IsDigit(chars[i]) {
			i += 1
			continue
		}
		if !unicode.IsLetter(chars[j]) && !unicode.IsDigit(chars[j]) {
			j -= 1
			continue
		}

		if unicode.ToLower(chars[i]) != unicode.ToLower(chars[j]) {
			return false
		}
		i += 1
		j -= 1
	}

	return true
}
