package containsduplicates

func ContainsDuplicates(numbers []int) bool {
	visited := make(map[int]int)

	for i, n := range numbers {
		_, exists := visited[n]
		if exists {
			return true
		}
		visited[n] = i
	}

	return false
}
