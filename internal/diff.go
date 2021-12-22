package internal

import (
	"fmt"
	"unicode/utf8"
)

//https://blog.robertelder.org/diff-algorithm/
//https://blog.jcoglan.com/2017/02/12/the-myers-diff-algorithm-part-1/
//https://blog.jcoglan.com/2017/02/12/the-myers-diff-algorithm-part-2/
//visualization: https://www.nathaniel.ai/myers-diff/

//The Greedy Principle:
//furthest reaching D-paths are obtained by greedily extending furthest reaching (D − 1)-paths.

//This suggests computing the endpoints of D-paths in the relevant D+1 diagonals for successively increasing values of D
//until the furthest reaching path in diagonal N − M reaches (N,M).

type arr struct {
	min int
	x   []int
}

func initArr(n, m int) *arr {
	max := n + m
	min := -max
	return &arr{
		min: min,
	}
}

func (a *arr) get(k int) int {
	pos := k + (-a.min)
	return a.x[pos]
}

func (a *arr) put(k, i int) {
	pos := k + (-a.min)
	a.x[pos] = i
}

func diff(a, b []byte) ([]byte, error) {
	if !utf8.Valid(a) {
		return nil, fmt.Errorf("base string is invalid utf8")
	}
	if !utf8.Valid(b) {
		return nil, fmt.Errorf("Derived string is nvalid utf8")
	}
	n, m := len(a), len(b)
	// k can range from -n to -m
	max := n + m
	V := initArr(n, m)
	V.put(1, 0)
	for d := 0; d <= max; d++ {
		for k := -d; k <= d; k += 2 {
			x := 0
			if (k == -d) || (k != d && V.get(k-1) < V.get(k+1)) { //optimizing for x
				x = V.get(k + 1)
			} else {
				x = V.get(k-1) + 1
			}
			y := x - k
			for x < n && y < m && a[x+1] == b[y+1] {
				x, y = x+1, y+1
			}
			V.put(k, x)
			if x >= n && y >= m {
				break
			}
		}
	}

	//comebcak
	//todo
	return nil, nil
}
