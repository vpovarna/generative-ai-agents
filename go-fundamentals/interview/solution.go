package interview

import (
	"flag"
	"fmt"
	"os"
	"strconv"
	"strings"
)

var inputFile = flag.String("inputFile", "input.txt", "Relative path to the input file")

type Node struct {
	name     string
	isDir    bool
	size     int
	children map[string]*Node
	parent   *Node
}

func NewNode(name string, isDir bool, size int) *Node {
	return &Node{
		name:     name,
		isDir:    isDir,
		size:     size,
		children: map[string]*Node{},
	}
}

func buildFileSystem(input string) *Node {
	lines := strings.Split(input, "\n")

	root := NewNode("/", true, 0)
	current := root

	for _, line := range lines {
		parts := strings.Fields(line)

		// All the commands contains only two strings
		switch parts[0] {
		case "cd":
			dirname := parts[1]
			if dirname == "/" {
				current = root
			} else if dirname == ".." {
				if current.parent != nil {
					current = current.parent
				}
			} else {
				if _, exists := current.children[dirname]; exists {
					// switch parent to the new dir
					current = current.children[dirname]
				} else {
					fmt.Printf("Directory: %s doesn't exist \n", dirname)
				}
			}
		case "dir":
			dirname := parts[1]
			if _, exists := current.children[dirname]; !exists {
				newDir := NewNode(dirname, true, 0)
				newDir.parent = current
				current.children[dirname] = newDir
			}
		default:
			// File <size> <filename>
			size, err := strconv.Atoi(parts[0])
			if err != nil {
				fmt.Println("Invalid file size")
			}

			filename := parts[1]

			file := NewNode(filename, false, size)
			file.parent = current
			current.children[filename] = file
		}

	}

	return root
}

func calculateSize(node *Node) int {
	if !node.isDir {
		return node.size
	}

	total := 0

	for _, child := range node.children {
		total += calculateSize(child)
	}
	return total
}

func findDirsUnderLimit(node *Node, limit int) []int {
	if !node.isDir {
		return []int{}
	}

	var result []int
	currentSize := 0

	for _, child := range node.children {
		if child.isDir {
			result = append(result, findDirsUnderLimit(child, limit)...)
			currentSize += calculateSize(child)
		} else {
			currentSize += child.size
		}
	}
	if currentSize <= limit {
		result = append(result, currentSize)
	}
	return result
}

// Return the directory names and the size for each of them.
type DirInfo struct {
	path string
	size int
}

func findDirsSize(node *Node, name string) []DirInfo {
	if !node.isDir {
		return []DirInfo{}
	}

	var result []DirInfo
	currentSize := 0

	// Calculate current directory size and collect subdirectory info
	for _, child := range node.children {
		if child.isDir {
			// Recursively collect info from subdirectories
			result = append(result, findDirsSize(child, child.name)...)
			// Add subdirectory size to current
			currentSize += calculateSize(child)
		} else {
			// Add file size to current directory
			currentSize += child.size
		}
	}

	// Add current directory info
	result = append(result, DirInfo{
		path: name,
		size: currentSize,
	})

	return result
}

func Run() {
	flag.Parse()

	bytes, err := os.ReadFile(*inputFile)

	if err != nil {
		panic("Unable to read the input file")
	}

	input := string(bytes)
	root := buildFileSystem(input)

	dirs := findDirsSize(root, "/")
	fmt.Println(dirs)

	sizes := findDirsUnderLimit(root, 100000)

	sum := 0
	for _, size := range sizes {
		sum += size
	}

	fmt.Println(sum)
}
