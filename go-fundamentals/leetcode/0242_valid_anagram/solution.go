package validanagram

import (
	"sort"
)

func IsAnagram(t string, s string) bool {
	sortedT := []rune(t)
	sortedS := []rune(s)

	sort.Slice(sortedT, func(i, j int) bool {
		return sortedT[i] < sortedT[j]
	})

	sort.Slice(sortedS, func(i, j int) bool {
		return sortedS[i] < sortedS[j]
	})

	return string(sortedS) == string(sortedT)

}

func IsAnagramWithDict(s, t string) bool {
	dictS := getOccurrence(s)
	dictT := getOccurrence(t)

	if len(dictS) != len(dictT) {
		return false
	}

	for k, v := range dictS {
		c, exists := dictT[k]
		if !exists {
			return false
		}
		if v != c {
			return false
		}
	}

	return true
}

func getOccurrence(s string) map[rune]int {
	dictS := make(map[rune]int)

	for _, c := range s {
		if v, exists := dictS[c]; exists {
			v += 1
			dictS[c] = v
		}
		dictS[c] = 1
	}
	return dictS
}
