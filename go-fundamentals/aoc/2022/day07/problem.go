package aoc2022day07

import (
	"flag"
	"fmt"
	"math"
	"os"
	"strconv"
	"strings"
)

type Node struct {
	Name     string
	IsDir    bool
	Size     int
	Parent   *Node
	Children map[string]*Node
}

func NewDir(name string, parent *Node) *Node {
	return &Node{
		Name:     name,
		IsDir:    true,
		Size:     0,
		Parent:   parent,
		Children: make(map[string]*Node),
	}
}

func NewFile(name string, size int, parent *Node) *Node {
	return &Node{
		Name:     name,
		IsDir:    false,
		Size:     size,
		Parent:   parent,
		Children: nil,
	}
}

func (n *Node) GetTotalSize() int {
	if !n.IsDir {
		return n.Size
	}

	total := 0
	for _, child := range n.Children {
		total += child.GetTotalSize()
	}
	return total
}

var inputFile = flag.String("inputFile", "input.txt", "Relative path to the input file")

func Run() {
	flag.Parse()

	bytes, err := os.ReadFile(*inputFile)

	if err != nil {
		panic("Unable to read the input file")
	}

	input := string(bytes)

	fmt.Printf("AoC2022, Day07, Part1 is: %d \n", part1(&input))
	fmt.Printf("AoC2022, Day07, Part2 is: %d \n", part2(&input))
}

func part1(input *string) int {
	lines := strings.Split(*input, "\n")
	root := buildFileSystem(lines)
	return sumSmallDirectories(root, 100000)
}

func part2(input *string) int {
	lines := strings.Split(*input, "\n")
	root := buildFileSystem(lines)
	const totalDiskSpace = 70000000
	const requiredSpace = 30000000

	usedSpace := root.GetTotalSize()
	currentlyFree := totalDiskSpace - usedSpace
	needToFree := requiredSpace - currentlyFree

	return findTheSmallestToDelete(root, needToFree)
}

func buildFileSystem(lines []string) *Node {
	root := NewDir("/", nil)
	current := root

	for _, line := range lines {

		if strings.HasPrefix(line, "$ cd") {
			target := strings.TrimPrefix(line, "$ cd ")

			if target == "/" {
				current = root
			} else if target == ".." {
				if current.Parent != nil {
					current = current.Parent
				}
			} else {
				// cd into directory
				if child, exists := current.Children[target]; exists {
					current = child
				} else {
					// Creates another directory if it doesn't exist
					newDir := NewDir(target, current)
					current.Children[target] = newDir
					current = newDir
				}
			}
		} else if line == "$ ls" {
			// ls command -> next lines are output
		} else if strings.HasPrefix(line, "dir ") {
			dirName := strings.TrimPrefix(line, "dir ")
			if _, exists := current.Children[dirName]; !exists {
				current.Children[dirName] = NewDir(dirName, current)
			}
		} else if line != "" && !strings.HasPrefix(line, "$") {
			// File listing: <size> <name>
			parts := strings.Split(line, " ")
			if len(parts) == 2 {
				size, _ := strconv.Atoi(parts[0])
				fileName := parts[1]
				current.Children[fileName] = NewFile(fileName, size, current)
			}
		}

	}
	return root
}

func sumSmallDirectories(node *Node, limit int) int {
	if !node.IsDir {
		return 0
	}

	total := 0
	dirSize := node.GetTotalSize()

	if dirSize <= limit {
		total += dirSize
	}

	for _, child := range node.Children {
		if child.IsDir {
			total += sumSmallDirectories(child, limit)
		}
	}
	return total
}

func findTheSmallestToDelete(node *Node, minSize int) int {
	if !node.IsDir {
		return math.MaxInt
	}

	currentSize := node.GetTotalSize()
	smallest := math.MaxInt

	if currentSize >= minSize {
		smallest = currentSize
	}
	for _, child := range node.Children {
		if child.IsDir {
			childResult := findTheSmallestToDelete(child, minSize)
			if childResult < smallest {
				smallest = childResult
			}
		}
	}

	return smallest

}
