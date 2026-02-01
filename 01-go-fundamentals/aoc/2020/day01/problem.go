package aoc2020day01

import (
	"flag"
	"fmt"
	"os"
	"strings"

	"github.com/povarna/generative-ai-with-go/fundamentals/utils"
)

var inputFile = flag.String("inputFile", "input.txt", "Relative path to the input file")
var target = 2020

func Run() {
	flag.Parse()

	bytes, err := os.ReadFile(*inputFile)

	if err != nil {
		panic(err)
	}

	input := string(bytes)
	fmt.Printf("AOC 2020 Day1 Part1 solution is: %d\n", part1(&input, target))
	fmt.Printf("AOC 2020 Day1 Part2 solution is: %d\n", part2(&input, target))
}

func part1(input *string, target int) int {
	lines := strings.Split(*input, "\n")

	seen := make(map[int]int)

	for i, number := range lines {
		n := utils.ToInt(number)
		m := target - n
		if _, exists := seen[m]; exists {
			return n * m
		}

		seen[n] = i

	}

	return -1
}

func part2(input *string, target int) int {
	lines := strings.Split(*input, "\n")

	for i, number := range lines {
		seen := make(map[int]int)
		n := utils.ToInt(number)
		new_target := target - n

		for j, secondNumber := range lines[i+1:] {
			m := utils.ToInt(secondNumber)

			_, exists := seen[new_target-m]
			if exists {
				return n * m * (new_target - m)
			}
			seen[m] = j
		}
	}

	return -1
}
