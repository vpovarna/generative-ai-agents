package longestsubstring

func lengthOfLongestSubstring(s string) int {
	charSet := make(map[byte]bool)
	l, res := 0, 0

	for i := 0; i < len(s); i++ {
		for charSet[s[i]] {
			delete(charSet, s[l])
			l += 1
		}

		charSet[s[i]] = true
		if i-l+1 > res {
			res = i - l + 1
		}
	}

	return res

}
