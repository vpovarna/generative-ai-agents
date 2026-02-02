package aoc2023day01

import (
	"flag"
	"fmt"
	"os"
	"strconv"
	"strings"
	"unicode"

	"github.com/povarna/generative-ai-with-go/fundamentals/utils"
)

var inputFile = flag.String("inputFile", "input.txt", "Relative path to the input file")

func Run() {
	flag.Parse()

	bytes, err := os.ReadFile(*inputFile)

	if err != nil {
		panic("Unable to load the input file")
	}
	input := string(bytes)

	fmt.Printf("AoC2023, Day1, Part1 solution is: %d\n", part1(&input))
	fmt.Printf("AoC2023, Day1, Part2 solution is: %d\n", part2(&input))
}

func part1(input *string) int {
	lines := strings.SplitSeq(*input, "\n")

	result := 0

	for line := range lines {
		acc := ""
		for i := 0; i < len(line); i++ {
			n := rune(line[i])
			if unicode.IsDigit(n) {
				acc = fmt.Sprintf("%s%s", acc, string(n))
			}
		}
		if len(acc) == 1 {
			result += utils.ToInt(fmt.Sprintf("%s%s", acc, acc))
		} else {
			result += utils.ToInt(fmt.Sprintf("%s%s", string(acc[0]), string(acc[len(acc)-1])))
		}

	}

	return result
}

func part2(input *string) any {
	lines := strings.SplitSeq(*input, "\n")
	prefixes := map[string]int{
		"one":   1,
		"two":   2,
		"three": 3,
		"four":  4,
		"five":  5,
		"six":   6,
		"seven": 7,
		"eight": 8,
		"nine":  9,
	}

	for i := range 10 {
		t := strconv.Itoa(i)
		prefixes[t] = i
	}

	result := 0

	for line := range lines {
		n := len(line)
		acc := ""
		for i := range n {
			for prefix, val := range prefixes {
				if doesStringHavePrefix(line[i:], prefix) {
					acc = fmt.Sprintf("%s%s", acc, fmt.Sprint(val))
				}
			}
		}
		if len(acc) == 1 {
			result += utils.ToInt(fmt.Sprintf("%s%s", acc, acc))
		} else {
			result += utils.ToInt(fmt.Sprintf("%s%s", string(acc[0]), string(acc[len(acc)-1])))
		}

	}

	return result
}

func doesStringHavePrefix(s, prefix string) bool {
	if len(s) < len(prefix) {
		return false
	}
	return s[:(len(prefix))] == prefix
}
