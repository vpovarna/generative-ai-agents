package aoc2021day03

import (
	"flag"
	"fmt"
	"os"
	"strconv"
	"strings"
)

var inputFile = flag.String("inputFile", "input.txt", "Relative Path for the input file")

func Run() {
	flag.Parse()

	bytes, err := os.ReadFile(*inputFile)

	if err != nil {
		panic("Unable to read the input file")
	}

	input := string(bytes)
	fmt.Printf("AoC2021, Day1, Part1 solution is: %d\n", part1(&input))
	fmt.Printf("AoC2021, Day1, Part2 solution is: %d\n", part2(&input))
}

func part1(input *string) int {
	lines := strings.Split(*input, "\n")

	var gamma, epsilon string

	n := len(lines[0])
	for i := range n {
		countOnes, countZeros := 0, 0

		for _, line := range lines {
			a := string(line[i])
			if a == "0" {
				countZeros += 1
			} else {
				countOnes += 1
			}
		}

		if countOnes > countZeros {
			gamma += "0"
			epsilon += "1"
		} else {
			gamma += "1"
			epsilon += "0"
		}
	}

	e, err := strconv.ParseInt(epsilon, 2, 64)
	if err != nil {
		panic(err)
	}
	g, err := strconv.ParseInt(gamma, 2, 64)
	if err != nil {
		panic(err)
	}
	return int(e * g)
}

func part2(input *string) int {

	lines := strings.Split(*input, "\n")

	for i := 0; i < len(lines[0]) && len(lines) > 1; i++ {
		countOnes, countZeros := 0, 0
		for _, line := range lines {
			if line[i] == '1' {
				countOnes += 1
			} else {
				countZeros += 1
			}
		}
		if countOnes >= countZeros {
			lines = filter_lines(lines, i, '1')
		} else {
			lines = filter_lines(lines, i, '0')
		}
	}

	oxygenGeneratorRating := lines[0]
	oxygenGeneratorRatingInt, err := strconv.ParseInt(oxygenGeneratorRating, 2, 64)
	if err != nil {
		panic(err)
	}

	lines = strings.Split(*input, "\n")
	for i := 0; i < len(lines[0]) && len(lines) > 1; i++ {
		countOnes, countZeros := 0, 0
		for _, line := range lines {
			if line[i] == '1' {
				countOnes += 1
			} else {
				countZeros += 1
			}
		}
		if countOnes >= countZeros {
			lines = filter_lines(lines, i, '0')
		} else {
			lines = filter_lines(lines, i, '1')
		}
	}

	co2ScrubberRating := lines[0]
	co2ScrubberRatingInt, err := strconv.ParseInt(co2ScrubberRating, 2, 64)
	if err != nil {
		panic(err)
	}

	return int(oxygenGeneratorRatingInt * co2ScrubberRatingInt)
}

func filter_lines(lines []string, index int, value rune) []string {
	new_lines := []string{}
	for _, line := range lines {
		if rune(line[index]) == value {
			new_lines = append(new_lines, line)
		}
	}
	return new_lines
}
