package utils

import "strconv"

func ToInt(s string) int {
	n, err := strconv.Atoi(s)
	if err != nil {
		panic("Unable to covert number to string")
	}

	return n
}
