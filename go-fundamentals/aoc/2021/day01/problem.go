package aoc2021day01

import (
	"flag"
	"fmt"
	"os"
	"strings"

	"github.com/povarna/generative-ai-with-go/fundamentals/utils"
)

var inputFile = flag.String("inputFile", "input.txt", "Relative path to the input file")

func part1(input *string) int {
	lines := strings.Split(*input, "\n")
	length := len(lines)
	count := 0
	for i := 1; i < length; i++ {
		if utils.ToInt(lines[i-1]) < utils.ToInt(lines[i]) {
			count += 1
		}
	}

	return count
}

func part2(input *string) int {
	lines := strings.Split(*input, "\n")
	length := len(lines)
	count := 0

	prev_sum := utils.ToInt(lines[0]) + utils.ToInt(lines[1]) + utils.ToInt(lines[2])
	for i := 3; i < length; i++ {
		current_sum := utils.ToInt(lines[i-2]) + utils.ToInt(lines[i-1]) + utils.ToInt(lines[i])
		if current_sum > prev_sum {
			count += 1
		}
		prev_sum = current_sum
	}

	return count
}

func Run() {
	flag.Parse()

	bytes, err := os.ReadFile(*inputFile)
	if err != nil {
		panic("Unable to read the input file")
	}

	input := string(bytes)

	fmt.Printf("AoC2021, Day01, Part1 solution is: %d\n", part1(&input))
	fmt.Printf("AoC2021, Day01, Part2 solution is: %d\n", part2(&input))
}
