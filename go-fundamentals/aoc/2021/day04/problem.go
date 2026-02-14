package aoc2021day04

import (
	"flag"
	"fmt"
	"os"
	"strings"

	"github.com/povarna/generative-ai-with-go/fundamentals/utils"
)

var inputFile = flag.String("inputFile", "input.txt", "Relative path to the input file")

func Run() {
	flag.Parse()

	bytes, err := os.ReadFile(*inputFile)
	if err != nil {
		panic("Unable to read the input file")
	}

	input := string(bytes)
	fmt.Printf("AoC2021, Day4, Part1 solution is: %d\n", part1(&input))
	fmt.Printf("AoC2021, Day4, Part2 solution is: %d\n", part2(&input))
}

func part1(input *string) int {
	nums, boards := readInput(input)

	for _, n := range nums {
		for _, b := range boards {
			didWin := b.pickNum(n)
			if didWin {
				// multiply score of winning board by number that was just called
				return b.score() * n
			}
		}
	}

	return -1
}

func part2(input *string) int {
	nums, boards := readInput(input)

	lastWinningScore := -1
	alreadyWon := map[int]bool{}
	for _, n := range nums {
		for bi, b := range boards {
			if alreadyWon[bi] {
				continue
			}
			didWin := b.pickNum(n)
			if didWin {
				// WHICH BOARD WINS LAST
				lastWinningScore = b.score() * n

				// mark board as already won
				alreadyWon[bi] = true
			}
		}
	}

	return lastWinningScore

}

type BoardState struct {
	board  [][]int
	picked [][]bool
}

func (b *BoardState) pickNum(num int) bool {
	for r, rows := range b.board {
		for c, v := range rows {
			if v == num {
				b.picked[r][c] = true
			}
		}
	}

	for i := 0; i < len(b.board); i++ {
		isFullRow, isFullCol := true, true
		// board is square so this works fine, otherwise would need another pair of nested loops
		for j := 0; j < len(b.board); j++ {
			// check row at index i
			if !b.picked[i][j] {
				isFullRow = false
			}
			// check col at index j
			if !b.picked[j][i] {
				isFullCol = false
			}
		}
		if isFullRow || isFullCol {
			// returns true if is winning board
			return true
		}
	}

	// false for incomplete board
	return false
}

func newBoardState(board [][]int) BoardState {
	picked := make([][]bool, len(board))
	for i := range picked {
		picked[i] = make([]bool, len(board[0]))
	}
	return BoardState{
		board:  board,
		picked: picked,
	}
}

func (b *BoardState) score() int {
	var score int

	for r, rows := range b.board {
		for c, v := range rows {
			// adds up all the non-picked/marked cells
			if !b.picked[r][c] {
				score += v
			}
		}
	}

	return score
}

func readInput(input *string) ([]int, []BoardState) {
	lines := strings.Split(*input, "\n\n")

	var nums []int
	var boards []BoardState

	for v := range strings.SplitSeq(lines[0], ",") {
		nums = append(nums, utils.ToInt(strings.TrimSpace(v)))
	}

	for _, group := range lines[1:] {
		var board [][]int
		boardLines := strings.SplitSeq(strings.TrimSpace(group), "\n")

		for line := range boardLines {
			fields := strings.Fields(line)
			var row []int
			for _, f := range fields {
				row = append(row, utils.ToInt(f))
			}
			board = append(board, row)
		}

		boards = append(boards, newBoardState(board))
	}

	return nums, boards
}
