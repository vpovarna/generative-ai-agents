package main

import (
	"fmt"

	productexceptself "github.com/povarna/generative-ai-with-go/fundamentals/leetcode/0238_product_except_self"
)

func pointers_example() {
	i := 42

	p := &i
	fmt.Println(*p)
	*p = 21
	fmt.Println(i)

	j := 2071

	p = &j
	*p = *p + 100
	fmt.Println(j)

}

func arrays() {
	var s []int
	printSlice(s)
	s = append(s, 0)
	printSlice(s)

	s = append(s, 2, 3, 4)
	printSlice(s)

}

func printSlice(s []int) {
	fmt.Printf("len=%d cap=%d %v\n", len(s), cap(s), s)
}

func main() {
	nums := []int{1, 2, 3, 4}
	fmt.Println(productexceptself.ProductExceptSelf(nums))
}
