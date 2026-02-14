package productexceptself

func ProductExceptSelf(nums []int) []int {
	result := make([]int, len(nums))

	leftMult(nums, result)
	rightMult(nums, result)

	return result
}

func leftMult(nums []int, result []int) {
	mul := 1
	for i := 0; i < len(nums); i++ {
		result[i] = mul
		mul *= nums[i]
	}
}

func rightMult(nums []int, result []int) {
	mul := 1
	for i := len(nums) - 1; i >= 0; i-- {
		result[i] *= mul
		mul *= nums[i]
	}
}
