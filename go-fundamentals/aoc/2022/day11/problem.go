package aoc2022day11

import (
	"fmt"
	"slices"
)

type Monkey struct {
	id                   int
	items                []int
	operation            func(int) int
	testDivisibleBy      int
	trueMonkeyCondition  int
	falseMonkeyCondition int
}

func initExample() []Monkey {
	return []Monkey{
		{
			id:    0,
			items: []int{79, 98},
			operation: func(x int) int {
				return x * 19
			},
			testDivisibleBy:      23,
			trueMonkeyCondition:  2,
			falseMonkeyCondition: 3,
		},
		{
			id:    1,
			items: []int{54, 65, 75, 74},
			operation: func(x int) int {
				return x + 6
			},
			testDivisibleBy:      19,
			trueMonkeyCondition:  2,
			falseMonkeyCondition: 0,
		},
		{
			id:    2,
			items: []int{79, 60, 97},
			operation: func(x int) int {
				return x * x
			},
			testDivisibleBy:      13,
			trueMonkeyCondition:  1,
			falseMonkeyCondition: 3,
		},
		{
			id:    3,
			items: []int{74},
			operation: func(x int) int {
				return x + 3
			},
			testDivisibleBy:      17,
			trueMonkeyCondition:  0,
			falseMonkeyCondition: 1,
		},
	}
}

func inputExample() []Monkey {
	return []Monkey{
		{
			id:    0,
			items: []int{61},
			operation: func(x int) int {
				return x * 11
			},
			testDivisibleBy:      5,
			trueMonkeyCondition:  7,
			falseMonkeyCondition: 4,
		},
		{
			id:    1,
			items: []int{76, 92, 53, 93, 79, 86, 81},
			operation: func(x int) int {
				return x + 4
			},
			testDivisibleBy:      2,
			trueMonkeyCondition:  2,
			falseMonkeyCondition: 6,
		},
		{
			id:    2,
			items: []int{91, 99},
			operation: func(x int) int {
				return x * 19
			},
			testDivisibleBy:      13,
			trueMonkeyCondition:  5,
			falseMonkeyCondition: 0,
		},
		{
			id:    3,
			items: []int{58, 67, 66},
			operation: func(x int) int {
				return x * x
			},
			testDivisibleBy:      7,
			trueMonkeyCondition:  6,
			falseMonkeyCondition: 1,
		},
		{
			id:    4,
			items: []int{94, 54, 62, 73},
			operation: func(x int) int {
				return x + 1
			},
			testDivisibleBy:      19,
			trueMonkeyCondition:  3,
			falseMonkeyCondition: 7,
		},
		{
			id:    5,
			items: []int{59, 95, 51, 58, 58},
			operation: func(x int) int {
				return x + 3
			},
			testDivisibleBy:      11,
			trueMonkeyCondition:  0,
			falseMonkeyCondition: 4,
		},
		{
			id:    6,
			items: []int{87, 69, 92, 56, 91, 93, 88, 73},
			operation: func(x int) int {
				return x + 8
			},
			testDivisibleBy:      3,
			trueMonkeyCondition:  5,
			falseMonkeyCondition: 2,
		},
		{
			id:    7,
			items: []int{71, 57, 86, 67, 96, 95},
			operation: func(x int) int {
				return x + 7
			},
			testDivisibleBy:      17,
			trueMonkeyCondition:  3,
			falseMonkeyCondition: 1,
		},
	}
}

func Run() {
	fmt.Printf("AoC2022, Day11, Part1 solution is: %d\n", part1())
	fmt.Printf("AoC2022, Day11, Part2 solution is: %d\n", part2())
}

func part1() int {
	monkeys := inputExample()

	monkeysItemsCount := []int{0, 0, 0, 0, 0, 0, 0, 0}

	for range 20 {
		for id, monkey := range monkeys {
			monkeysItemsCount[id] += len(monkey.items)

			for _, id := range monkey.items {
				worryLevel := monkey.operation(id) / 3
				if worryLevel%monkey.testDivisibleBy == 0 {
					monkeys[monkey.trueMonkeyCondition].items = append(monkeys[monkey.trueMonkeyCondition].items, worryLevel)
				} else {
					monkeys[monkey.falseMonkeyCondition].items = append(monkeys[monkey.falseMonkeyCondition].items, worryLevel)
				}
			}
			monkeys[id].items = []int{}
		}
	}

	slices.SortFunc(monkeysItemsCount, func(x, y int) int {
		return y - x
	})

	return monkeysItemsCount[0] * monkeysItemsCount[1]
}

func part2() int {
	monkeys := inputExample()

	monkeysItemsCount := []int{0, 0, 0, 0, 0, 0, 0, 0}

	bigMod := 1
	for _, monkey := range monkeys {
		bigMod *= monkey.testDivisibleBy
	}

	for range 10000 {
		for id, monkey := range monkeys {
			monkeysItemsCount[id] += len(monkey.items)

			for _, id := range monkey.items {
				worryLevel := monkey.operation(id) % bigMod
				if worryLevel%monkey.testDivisibleBy == 0 {
					monkeys[monkey.trueMonkeyCondition].items = append(monkeys[monkey.trueMonkeyCondition].items, worryLevel)
				} else {
					monkeys[monkey.falseMonkeyCondition].items = append(monkeys[monkey.falseMonkeyCondition].items, worryLevel)
				}
			}
			monkeys[id].items = []int{}
		}
	}

	slices.SortFunc(monkeysItemsCount, func(x, y int) int {
		return y - x
	})

	return monkeysItemsCount[0] * monkeysItemsCount[1]
}
