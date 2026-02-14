package aoc2022day01

import (
	"flag"
	"fmt"
	"os"
	"slices"
	"strings"

	"github.com/povarna/generative-ai-with-go/fundamentals/utils"
)

var inputFile = flag.String("inputFile", "input.txt", "Relative path to the input file")

func part1(input *string) int {
	acc := getCalories(input)

	return slices.Max(acc)
}

func part2(input *string) int {
	acc := getCalories(input)

	slices.SortFunc(acc, func(a int, b int) int {
		return b - a
	})

	return acc[0] + acc[1] + acc[2]
}

func getCalories(input *string) []int {
	elfCalories := strings.Split(*input, "\n\n")

	acc := []int{}

	for _, group := range elfCalories {
		calories := strings.Split(group, "\n")
		total := 0
		for _, calorie := range calories {
			c := utils.ToInt(calorie)
			total += c
		}
		acc = append(acc, total)
	}
	return acc
}

func Run() {
	flag.Parse()
	bytes, err := os.ReadFile(*inputFile)
	if err != nil {
		panic("Unable to read the input file")
	}

	input := string(bytes)
	fmt.Printf("AoC 2022, Day01 part1 solution is: %d\n", part1(&input))
	fmt.Printf("AoC 2022, Day01 part2 solution is: %d\n", part2(&input))
}
