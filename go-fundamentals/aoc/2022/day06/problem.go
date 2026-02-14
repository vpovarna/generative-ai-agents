package aoc2022day06

import (
	"flag"
	"fmt"
	"os"
)

var inputFile = flag.String("inputFile", "input.txt", "Relative path to the input file")

func Run() {
	flag.Parse()

	bytes, err := os.ReadFile(*inputFile)
	if err != nil {
		panic("Unable to read the input file")
	}

	input := string(bytes)
	fmt.Printf("AoC2022, Day6, Part1 solution is: %d\n", run(&input, 4))
	fmt.Printf("AoC2022, Day6, Part2 solution is: %d\n", run(&input, 14))
}

func run(input *string, nrCharacters int) int {
	count := 0
	for i := nrCharacters; i < len(*input); i++ {
		tmpStr := (*input)[i-nrCharacters : i]
		if markerAppears(tmpStr) {
			count += i
			break
		}
	}

	return count
}

func markerAppears(tmpStr string) bool {
	seen := make(map[rune]bool)

	for _, c := range tmpStr {
		if _, exists := seen[c]; exists {
			return false
		}
		seen[c] = true
	}
	return true
}
