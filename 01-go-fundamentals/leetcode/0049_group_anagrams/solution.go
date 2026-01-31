package groupanagrams

import (
	"slices"
)

func GroupAnagrams(strs []string) [][]string {
	dict := make(map[string][]string)

	for _, s := range strs {
		runes := []rune(s)
		slices.Sort(runes)

		key := string(runes)

		_, exists := dict[key]
		if !exists {
			dict[key] = []string{}
		}

		dict[key] = append(dict[key], s)
	}

	result := [][]string{}
	for _, group := range dict {
		result = append(result, group)
	}

	return result
}
