package twosum

func TwoSum(nums []int, target int) []int {
	visited := make(map[int]int)

	for i, num := range nums {
		expected_value := target - num
		if j, exists := visited[expected_value]; exists {
			return []int{j, i}
		}
		visited[num] = i
	}

	return []int{}
}
