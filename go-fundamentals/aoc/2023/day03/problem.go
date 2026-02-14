package aoc2023day03

import (
	"flag"
	"fmt"
	"os"
	"regexp"
	"slices"
	"strings"

	"github.com/povarna/generative-ai-with-go/fundamentals/utils"
)

var inputFile = flag.String("inputFile", "input.txt", "Relative path to the input file")
var digits = []string{"0", "1", "2", "3", "4", "5", "6", "7", "8", "9"}
var numReg = regexp.MustCompile("[0-9]")

func Run() {
	flag.Parse()

	bytes, err := os.ReadFile(*inputFile)
	if err != nil {
		panic("Unable to read the input file")
	}

	input := string(bytes)
	fmt.Printf("AoC2023, Day3, Part1 solution is: %d\n", part1(&input))
	fmt.Printf("AoC2023, Day3, Part2 solution is: %d\n", part2(&input))
}

func part1(input *string) int {
	grid := parseInput(input)

	m := len(grid)
	n := len(grid[0])
	total := 0

	for i := range m {
		var number string
		var hasSymbol bool
		for j := range n {
			currentDigit := string(grid[i][j])
			if !slices.Contains(digits, currentDigit) {
				if number != "" {
					if hasSymbol {
						total += utils.ToInt(number)
					}
					number = ""
					hasSymbol = false
				}
				continue
			}
			number += currentDigit
			neighbors := getNeighbors(i, j)
			for _, pair := range neighbors {
				row, col := pair[0], pair[1]
				if row < 0 || row >= m || col < 0 || col >= n || string(grid[row][col]) == "." || slices.Contains(digits, string(grid[row][col])) {
					continue
				}
				hasSymbol = true
			}

		}

		if number != "" && hasSymbol {
			total += utils.ToInt(number)
		}
	}

	return total
}

func part2(input *string) int {
	grid := parseInput(input)
	seen := map[[2]int]bool{}

	m := len(grid)
	n := len(grid[0])
	total := 0

	for i := range grid {
		for j, val := range grid[i] {
			if string(val) == "*" {
				nums := []int{}
				for _, pairs := range getNeighbors(i, j) {
					row, col := pairs[0], pairs[1]

					if row >= 0 && row < m && col >= 0 && col < n {
						foundNumber := getNumber(grid, row, col, seen)
						if foundNumber != -1 {
							nums = append(nums, foundNumber)
						}
					}
				}
				if len(nums) == 2 {
					total += nums[0] * nums[1]
				}
			}
		}
	}

	return total
}

func getNumber(matrix []string, row, col int, seen map[[2]int]bool) int {
	val := string(matrix[row][col])
	if seen[[2]int{row, col}] || !numReg.MatchString(val) {
		return -1
	}
	currentCol := col
	numStr := string(matrix[row][currentCol])
	currentCol -= 1
	for currentCol >= 0 {
		if numReg.MatchString(string(matrix[row][currentCol])) {
			numStr = fmt.Sprintf("%s%s", string(matrix[row][currentCol]), numStr)
			seen[[2]int{row, currentCol}] = true
		} else {
			break
		}
		currentCol -= 1
	}
	currentCol = col
	currentCol += 1
	for currentCol < len(matrix[row]) {
		if numReg.MatchString(string(matrix[row][currentCol])) {
			numStr = fmt.Sprintf("%s%s", numStr, string(matrix[row][currentCol]))
			seen[[2]int{row, currentCol}] = true
		} else {
			break
		}
		currentCol += 1
	}

	return utils.ToInt(numStr)
}

func getNeighbors(i, j int) [][2]int {
	return [][2]int{
		{i - 1, j - 1},
		{i - 1, j},
		{i - 1, j + 1},
		{i, j - 1},
		{i, j + 1},
		{i + 1, j - 1},
		{i + 1, j},
		{i + 1, j + 1},
	}
}

func parseInput(input *string) []string {
	lines := strings.Split(*input, "\n")

	grid := []string{}

	for _, line := range lines {
		if line != "" {
			grid = append(grid, line)
		}
	}

	return grid
}
