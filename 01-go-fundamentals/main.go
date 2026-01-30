package main

import (
	"fmt"

	twosum "github.com/povarna/generative-ai-with-go/fundamentals/leetcode/001_two_sum"
)

func main() {
	nums := []int{2, 7, 11, 15}
	target := 9

	newVar := twosum.TwoSum(nums, target)
	fmt.Printf("%v\n", newVar)
}
