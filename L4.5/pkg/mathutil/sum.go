package mathutil

import "fmt"

func Sum(a, b int) int {
	return a + b
}

func SumBad(a, b int) int {
	var s string
	s = fmt.Sprintf("%d", a+b)

	var x int
	fmt.Sscanf(s, "%d", &x)
	return x
}
