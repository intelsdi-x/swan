package main

import (
	"fmt"

	"github.com/intelsdi-x/swan/pkg/experiment"
)

func main() {
	fmt.Println("Nested Iterate() calls")

	var iteration int
	repetition := func(a string, b int) {
		fmt.Printf("iteration: %d\tset: %v\trange: %v\n", iteration, a, b)
		iteration++
	}

	experiment.Iterate("one,two,three", func() {
		experiment.Iterate("1-3", repetition)
	})

	fmt.Println("Should produce identical result as a Permutate() call")

	iteration = 0
	experiment.Permutate("one,two,three", "1-3", repetition)
}
