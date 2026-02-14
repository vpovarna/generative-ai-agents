package aoc2021day05

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
	fmt.Printf("AoC2021, Day5, Part1 solution is: %d\n", run(&input, false))
	fmt.Printf("AoC2021, Day5, Part2 solution is: %d\n", run(&input, true))
}


func run(input *string, part2 bool) int {
	lines := strings.Split(*input, "\n")
	var coors [][4]int

	for _, line := range lines {
		var x1, x2, y1, y2 int
		_, err := fmt.Sscanf(line, "%d,%d -> %d,%d", &x1, &y1, &x2, &y2)

		if err != nil {
			panic("Unable to read line")
		}
		coors = append(coors, [4]int{x1, y1, x2, y2})
	}

	// Find max values
	var endCol, endRow int
	for _, coor := range coors {
		if coor[0] > endRow {
			endRow = coor[0]
		}
		if coor[1] > endCol {
			endCol = coor[1]
		}
		if coor[2] > endRow {
			endRow = coor[2]
		}
		if coor[3] > endCol {
			endCol = coor[3]
		}
	}

	// building grid
	grid := make([][]int, endRow+1)

	for i := range grid {
		grid[i] = make([]int, endCol+1)
	}

	for _, c := range coors {
		if c[0] == c[2] {
			row := c[0]
			start, end := c[1], c[3]
			if c[1] > c[3] {
				start, end = end, start
			}
			for col := start; col <= end; col++ {
				grid[row][col]++
			}
		} else if c[1] == c[3] {
			col := c[1]
			start, end := c[0], c[2]
			if c[0] > c[2] {
				start, end = end, start
			}
			for row := start; row <= end; row++ {
				grid[row][col]++
			}
		} else if part2 {
			if c[1] > c[3] {
				c = [4]int{
					c[2], c[3],
					c[0], c[1],
				}
			}

			if c[0] < c[2] {
				for row := c[0]; row <= c[2]; row++ {
					col := c[1] + row - c[0]
					grid[row][col]++
				}
			} else {
				for row := c[0]; row >= c[2]; row-- {
					col := c[1] + (c[0] - row)
					grid[row][col]++
				}
			}
		}

	}

	var ans int
	for _, rows := range grid {
		for _, v := range rows {
			if v >= 2 {
				ans++
			}
		}
	}
	return ans
}
