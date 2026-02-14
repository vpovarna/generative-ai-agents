package aoc2022day04

import (
	"flag"
	"fmt"
	"os"
	"strings"
)

var inputFile = flag.String("inputFile", "input.txt", "Relative path to the input file")

func Run() {
	flag.Parse()

	bytes, err := os.ReadFile(*inputFile)

	if err != nil {
		panic("Unable to read the input file")
	}

	input := string(bytes)

	fmt.Printf("AoC 2022, Day04, Part1 solution is: %d\n", part1(&input))
	fmt.Printf("AoC 2022, Day04, Part2 solution is: %d\n", part2(&input))
}

func part1(input *string) int {
	lines := strings.Split(*input, "\n")
	total := 0

	for _, line := range lines {
		var x1, y1, x2, y2 int
		fmt.Sscanf(line, "%d-%d,%d-%d", &x1, &y1, &x2, &y2)
		if (x1 <= x2 && y2 <= y1) || (x2 <= x1 && y1 <= y2) {
			total += 1
		}
	}

	return total
}

func part2(input *string) int {
	lines := strings.Split(*input, "\n")
	total := 0

	for _, line := range lines {
		var x1, y1, x2, y2 int
		fmt.Sscanf(line, "%d-%d,%d-%d", &x1, &y1, &x2, &y2)
		//  2, 3, 4
		//     3, 4, 5
		if (x1 <= x2 && x2 <= y1) || (x2 <= x1 && x1 <= y2) {
			total += 1
			//    2, 3, 4
			// 1, 2, 3,
		} else if (x1 <= y2 && y2 <= y1) || (x2 <= y1 && y1 <= y2) {
			total += 1
		}
	}

	return total
}
