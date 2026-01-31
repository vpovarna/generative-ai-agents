package main

import (
	"fmt"

	groupanagrams "github.com/povarna/generative-ai-with-go/fundamentals/leetcode/0049_group_anagrams"
)

func main() {
	input := []string{"eat", "tea", "tan", "ate", "nat", "bat"}
	result := groupanagrams.GroupAnagrams(input)
	fmt.Printf("%v", result)
}
