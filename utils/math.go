package utils

func Min(x, y int) int {
	if x > y {
		return y
	}
	return x
}

func CeilForce(x, y int) int {
	res := x / y
	f := float64(x) / float64(y)
	if f > float64(res) {
		return res + 1
	} else {
		return res
	}
}
