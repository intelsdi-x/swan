package main

import (
	"fmt"

	"github.com/intelsdi-x/swan/pkg/experiment"
)

func main() {

	var iteration int
	repetition := func(a string, b int) {
		fmt.Printf("iteration: %d\tset: %v\trange: %v\n", iteration, a, b)
		iteration++
	}

	fmt.Println("Nested Iterate() calls")
	experiment.Iterate("one,two,three", func() {
		experiment.Iterate("1-3", repetition)
	})

	fmt.Println("Should produce identical result as a Permutate() call")
	iteration = 0
	experiment.Permutate("one,two,three", "1-3", repetition)

	fmt.Println("Dry run should ouput nothing")
	experiment.Permutate("one,two,three", "1-3", "1-100", experiment.DryRun)
	fmt.Printf("Number of iterations that would have been executed: %d", experiment.GetNumberOfIterations())
}
