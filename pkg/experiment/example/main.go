package main

import (
	"fmt"

	"github.com/intelsdi-x/swan/pkg/experiment"
)

func main() {
	fmt.Println("Nested Set() and Range()")

	var iteration int
	experiment.Set("one,two,three", func() {
		experiment.Range("1-3", func(a string, b int) {
			fmt.Printf("iteration: %d\tset: %v\trange: %v\n", iteration, a, b)
			iteration++
		})
	})

	fmt.Println("Should produce identical result as Permute()")

	iteration = 0
	experiment.Permutate("one,two,three", "1-3", func(a string, b int) {
		fmt.Printf("iteration: %d\tset: %v\trange: %v\n", iteration, a, b)
		iteration++
	})
}
